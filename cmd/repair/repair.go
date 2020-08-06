package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tendermint/iavl"
	tmstate "github.com/tendermint/tendermint/state"
	tmstore "github.com/tendermint/tendermint/store"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo"
	astore "github.com/amolabs/amoabci/amo/store"
	atypes "github.com/amolabs/amoabci/amo/types"
)

const (
	merkleTreeCacheSize = 10000
)

func repair(amoRoot string, doFix bool, rewindMerkle bool) {
	fmt.Println("Inspecting data root:", amoRoot)

	//// open

	fn := amoRoot + "/data/state.json"
	flag := os.O_RDONLY
	if doFix {
		flag = os.O_RDWR
	}
	amoStateFile, err := os.OpenFile(fn, flag, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer amoStateFile.Close()

	merkleDB, err := tmdb.NewGoLevelDB("merkle", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer merkleDB.Close()

	indexDB, err := tmdb.NewGoLevelDB("index", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer indexDB.Close()

	bsdb, err := tmdb.NewGoLevelDB("blockstore", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer bsdb.Close()

	sdb, err := tmdb.NewGoLevelDB("state", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdb.Close()

	//// load

	amoMt, err := iavl.NewMutableTree(merkleDB, merkleTreeCacheSize)
	if err != nil {
		fmt.Println(err)
		return
	}
	amoMt.Load()

	amoStore, err := astore.NewStore(nil, 100, merkleDB, indexDB)
	if err != nil {
		fmt.Println(err)
		return
	}

	amoCfg := atypes.AMOAppConfig{}
	b := amoStore.GetAppConfig()
	if len(b) <= 0 {
		fmt.Println("couldn't find proper app config")
		return
	}
	err = json.Unmarshal(b, &amoCfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	amoState := amo.State{}
	amoState.LoadFrom(amoStore, amoCfg)

	tmBlockStore := tmstore.NewBlockStore(bsdb)
	tmBlockStoreState := tmstore.LoadBlockStoreStateJSON(bsdb)
	tmState := tmstate.LoadState(sdb)

	//// display

	display(amoMt, amoState, tmBlockStoreState, tmState)

	//// repair

	fmt.Println("Repair TM state from block store...")

	orgHeight := tmState.LastBlockHeight
	nblk := tmBlockStore.LoadBlock(tmState.LastBlockHeight + 1)
	for nblk != nil && matchStateAndBlock(tmState, nblk) == false {
		// rewind to find the matching state
		nblk = tmBlockStore.LoadBlock(tmState.LastBlockHeight)
		if nblk == nil {
			fmt.Println("Unable to get rewind target block")
			return
		}
		// tmState.height will be nblk.height - 1 after this.
		curBlk := tmBlockStore.LoadBlock(nblk.Height - 1)
		tmState, err = stateFromBlock(tmState, nblk, curBlk)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("Rewinded TM state by %d blocks\n",
		orgHeight-tmState.LastBlockHeight)

	if tmBlockStoreState.Height > tmState.LastBlockHeight+1 {
		tmBlockStoreState.Height = tmState.LastBlockHeight + 1
	}

	fmt.Println("Repair AMO merkle tree...")

	ver, _ := amoMt.Load()
	orgVersion := ver
	appHash := amoMt.Hash()
	prevHash := appHash // should be null?
	// amoMt.Version equals to the block height where the app hash is written.
	// Hence, greater by one than the last block height.
	for ver > tmState.LastBlockHeight+1 {
		prevHash = appHash
		ver, err = amoMt.LoadVersion(ver - 1)
		if err != nil {
			fmt.Println(err)
			return
		}
		appHash = amoMt.Hash()
		if rewindMerkle || bytes.Equal(appHash, prevHash) {
			// Ok. No change in appHash, so no change in index db. We don't
			// have to touch index db. Rewind was safe in this case.
		} else {
			fmt.Println("Unable to rewind merkle db")
			return
		}
	}
	fmt.Printf("Rewinded AMO merkle db by %d\n", orgVersion-ver)

	fmt.Println("Check TM state again with AMO merkle db...")
	if !bytes.Equal(tmState.AppHash, appHash) {
		tmState.AppHash = appHash
	}

	fmt.Println("Repair AMO state...")

	amoState.LastHeight = tmState.LastBlockHeight
	amoState.LastAppHash = tmState.AppHash
	// TODO: ProtocolVersion, NextDraftID

	display(amoMt, amoState, tmBlockStoreState, tmState)

	//// save

	if !doFix {
		fmt.Println("repair result not saved. provide -f flag to save.")
		return
	}

	fmt.Println("saving repair result...")
	tmBlockStoreState.Save(bsdb)
	tmstate.SaveState(sdb, tmState)

	// cleaning up tx index
	// If validator set has not changed, then these dbs have no change either.
	// indexDelegator = tmdb.NewPrefixDB(indexDB, []byte("delegator")
	// indexValidator = tmdb.NewPrefixDB(indexDB, []byte("validator")
	// indexEffStake =  tmdb.NewPrefixDB(indexDB, []byte("effstake")
	indexBlockTx := tmdb.NewPrefixDB(indexDB, []byte("blocktx"))
	indexTxBlock := tmdb.NewPrefixDB(indexDB, []byte("txblock"))
	defer indexBlockTx.Close()
	defer indexTxBlock.Close()

	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(amoState.LastHeight+1))
	iter, err := indexBlockTx.Iterator(b, nil)
	for ; iter.Valid(); iter.Next() {
		fmt.Println("cleaning height", binary.BigEndian.Uint64(iter.Key()))
		indexBlockTx.DeleteSync(iter.Key())
	}

	iter, err = indexTxBlock.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		txHash := hex.EncodeToString(iter.Key())
		h := int64(binary.BigEndian.Uint64(iter.Value()))
		if h > amoState.LastHeight {
			fmt.Println("cleaning tx", strings.ToUpper(txHash))
			indexTxBlock.DeleteSync(iter.Key())
		}
	}
}

func matchStateAndBlock(tmState tmstate.State, blk *tmtypes.Block) bool {
	if !tmState.LastBlockID.Equals(blk.LastBlockID) {
		fmt.Println("LastBlockID mismatch!!!!")
		return false
	}
	if !bytes.Equal(tmState.NextValidators.Hash(), blk.NextValidatorsHash) {
		fmt.Println("NextValidatorsHash mismatch!!!!")
		return false
	}
	if !bytes.Equal(tmState.Validators.Hash(), blk.ValidatorsHash) {
		fmt.Println("ValidatorsHash mismatch!!!!")
		return false
	}
	if !bytes.Equal(tmState.ConsensusParams.Hash(), blk.ConsensusHash) {
		fmt.Println("ConsensusParamsHash mismatch!!!!")
		return false
	}
	if !bytes.Equal(tmState.LastResultsHash, blk.LastResultsHash) {
		fmt.Println("LastResultsHash mismatch!!!!")
		return false
	}
	if !bytes.Equal(tmState.AppHash, blk.AppHash) {
		fmt.Println("AppHash mismatch!!!!")
		return false
	}
	return true
}

func stateFromBlock(tmState tmstate.State, nextBlk *tmtypes.Block, curBlk *tmtypes.Block) (tmstate.State, error) {
	if tmState.LastHeightValidatorsChanged >= nextBlk.Height ||
		tmState.LastHeightConsensusParamsChanged >= nextBlk.Height {
		return tmState, errors.New("unable to rewind")
	}
	tmState.LastBlockHeight = nextBlk.Height - 1
	tmState.LastBlockID = nextBlk.LastBlockID
	if curBlk != nil {
		tmState.LastBlockTime = curBlk.Time
	}
	// tmState.NextValidators
	// tmState.Validators
	// tmState.ConsensusParams
	tmState.LastResultsHash = nextBlk.LastResultsHash
	tmState.AppHash = nextBlk.AppHash
	return tmState, nil
}

func display(mt *iavl.MutableTree, amoState amo.State, tmBsj tmstore.BlockStoreStateJSON, tmState tmstate.State) {
	mtVersion := mt.Version()

	fmt.Printf("AMO merkle tree ----------------------------\n")
	fmt.Printf("  .Version         = %d\n", mtVersion)
	fmt.Printf("  .Hash            = %X\n", mt.Hash())
	mt.LazyLoadVersion(mtVersion - 1)
	fmt.Printf("  .Hash    (-1)    = %X\n", mt.Hash())
	mt.LazyLoadVersion(mtVersion - 2)
	fmt.Printf("  .Hash    (-2)    = %X\n", mt.Hash())
	mt.LazyLoadVersion(mtVersion - 3)
	fmt.Printf("  .Hash    (-3)    = %X\n", mt.Hash())
	mt.Load()
	fmt.Printf("AMO state ----------------------------------\n")
	fmt.Printf("  .LastHeight      = %d\n", amoState.LastHeight)
	fmt.Printf("  .LastAppHash     = %X\n", amoState.LastAppHash)
	fmt.Printf("TM block store state -----------------------\n")
	fmt.Printf("  .height          = %d\n", tmBsj.Height)
	fmt.Printf("TM state -----------------------------------\n")
	fmt.Printf("  .LastBlockHeight = %d\n", tmState.LastBlockHeight)
	fmt.Printf("  .LastBlockID     = %s\n", tmState.LastBlockID)
	fmt.Printf("  .LastBlockTime   = %s\n", tmState.LastBlockTime)
	fmt.Printf("  .NextValidators  = %X\n", tmState.NextValidators.Hash())
	fmt.Printf("  .Validators      = %X\n", tmState.Validators.Hash())
	fmt.Printf("  .LastValidators  = %X\n", tmState.LastValidators.Hash())
	fmt.Printf("   (vals changed)  = %d\n", tmState.LastHeightValidatorsChanged)
	fmt.Printf("  .ConsensusParams = %X\n", tmState.ConsensusParams.Hash())
	fmt.Printf("   (cons changed)  = %d\n", tmState.LastHeightConsensusParamsChanged)
	fmt.Printf("  .LastResultsHash = %X\n", tmState.LastResultsHash)
	fmt.Printf("  .AppHash         = %X\n", tmState.AppHash)

}
