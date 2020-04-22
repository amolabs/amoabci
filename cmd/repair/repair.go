package main

import (
	"bytes"
	//"encoding/hex"
	//"encoding/json"
	"fmt"
	"os"
	//"strings"
	"errors"

	"github.com/tendermint/iavl"
	tmstate "github.com/tendermint/tendermint/state"
	tmstore "github.com/tendermint/tendermint/store"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo"
	//amostore "github.com/amolabs/amoabci/amo/store"
)

const (
	merkleTreeCacheSize = 10000
)

func repair(amoRoot string, doFix bool) {
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

	amoState := amo.State{}
	err = amoState.LoadFrom(amoStateFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	amoMt, err := iavl.NewMutableTree(merkleDB, merkleTreeCacheSize)
	if err != nil {
		fmt.Println(err)
		return
	}
	amoMt.Load()

	tmBlockStore := tmstore.NewBlockStore(bsdb)
	tmBlockStoreState := tmstore.LoadBlockStoreStateJSON(bsdb)
	tmState := tmstate.LoadState(sdb)

	//// display

	display(amoMt, amoState, tmBlockStoreState, tmState)

	/*
		responses, err := tmstate.LoadABCIResponses(sdb, amoState.LastHeight)
		if err != nil {
			fmt.Println(err)
			return
		}
		b, err := json.MarshalIndent(responses, "", "  ")
		fmt.Println("abci responses for block", amoState.LastHeight, string(b))
	*/

	//// repair

	fmt.Println("Repair TM state from block store...")

	orgHeight := tmState.LastBlockHeight
	nblk := tmBlockStore.LoadBlock(tmState.LastBlockHeight + 1)
	if nblk == nil {
		fmt.Println("Seems to be at the tip of the chain")
		return
	}
	for matchStateAndBlock(tmState, nblk) == false {
		// rewind to find the matching state
		nblk = tmBlockStore.LoadBlock(tmState.LastBlockHeight + 1)
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
	fmt.Printf("Rewinded %d blocks\n", orgHeight-tmState.LastBlockHeight)

	if tmBlockStoreState.Height > tmState.LastBlockHeight {
		tmBlockStoreState.Height = tmState.LastBlockHeight
	}

	fmt.Println("Repair AMO merkle tree...")

	ver, _ := amoMt.Load()
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
		if !bytes.Equal(appHash, prevHash) {
			fmt.Println("Unable to rewind merkle db")
			return
		}
		// Ok. No change in appHash, so no change in index db. We don't have to
		// touch index db. Rewind was safe in this case.
	}

	fmt.Println("Repair AMO state...")

	amoState.MerkleVersion = amoMt.Version()
	amoState.LastHeight = tmState.LastBlockHeight
	amoState.Height = tmState.LastBlockHeight
	amoState.LastAppHash = tmState.AppHash
	amoState.AppHash = tmState.AppHash
	// TODO: ProtocolVersion, CounterDue, NextDraftID

	display(amoMt, amoState, tmBlockStoreState, tmState)

	//// save

	if !doFix {
		fmt.Println("repair result not saved. provide -f flag to save.")
		return
	}

	fmt.Println("saving repair result...")
	amoState.SaveTo(amoStateFile)
	amoMt.LoadVersionForOverwriting(amoState.MerkleVersion)
	tmBlockStoreState.Save(bsdb)
	tmstate.SaveState(sdb, tmState)
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
	fmt.Printf("  .MerkleVersion   = %d\n", amoState.MerkleVersion)
	fmt.Printf("  .LastHeight      = %d\n", amoState.LastHeight)
	fmt.Printf("  .Height          = %d\n", amoState.Height)
	fmt.Printf("  .LastAppHash     = %X\n", amoState.LastAppHash)
	fmt.Printf("  .AppHash         = %X\n", amoState.AppHash)
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

/*
func repair(amoRoot string) {
	fmt.Println("Reparing data root:", amoRoot)

	fn := amoRoot + "/data/state.json"
	amoStateFile, err := os.OpenFile(fn, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer amoStateFile.Close()

	amoState := amo.State{}
	err = amoState.LoadFrom(amoStateFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	// tm dbs
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

	tmBlockStoreState := tmstore.LoadBlockStoreStateJSON(bsdb)
	tmBlockStore := tmstore.NewBlockStore(bsdb)
	tmState := tmstate.LoadState(sdb)

	// targets
	tmTargetHeight := tmState.LastBlockHeight

	fmt.Printf("Reset amoStore")
	err = os.RemoveAll(amoRoot + "/data/merkle.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.RemoveAll(amoRoot + "/data/index.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.RemoveAll(amoRoot + "/data/incentive.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.RemoveAll(amoRoot + "/data/group_counter.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	// amo dbs
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

	incentiveDB, err := tmdb.NewGoLevelDB("incentive", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer incentiveDB.Close()

	groupCounterDB, err := tmdb.NewGoLevelDB("group_counter", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer groupCounterDB.Close()

	fmt.Printf(" - Done !\n")

	fmt.Printf("Reset amoState")
	amoState = amo.State{}
	err = amoState.SaveTo(amoStateFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf(" - Done !\n")

	fmt.Printf("Clean up tmBlockStore:\n")
	for i := tmBlockStoreState.Height; i > tmTargetHeight; i-- {
		b := tmBlockStore.LoadBlock(i)

		fmt.Printf("  delete blockHeight = %d, blockHash = %x", b.Height, b.Hash())

		err = bsdb.DeleteSync([]byte(fmt.Sprintf("BH:%s",
			strings.ToLower(hex.EncodeToString(b.Hash())))),
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		targetKey := []byte(fmt.Sprintf("P:%v:", i))
		itr, err := bsdb.Iterator(targetKey, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer itr.Close()

		for ; itr.Valid(); itr.Next() {
			if !bytes.HasPrefix(itr.Key(), targetKey) {
				continue
			}

			err := bsdb.DeleteSync(itr.Key())
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		err = bsdb.DeleteSync([]byte(fmt.Sprintf("H:%v", i)))
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = bsdb.DeleteSync([]byte(fmt.Sprintf("C:%v", i)))
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = bsdb.DeleteSync([]byte(fmt.Sprintf("SC:%v", i)))
		if err != nil {
			fmt.Println(err)
			continue
		}

		btmp := tmBlockStore.LoadBlockByHash(b.Hash())
		if btmp != nil {
			fmt.Printf(" - not properly deleted %x\n", btmp.Hash())
		}

		fmt.Printf(" - Done !\n")
	}

	fmt.Printf("Fix tmBlockStoreState: Height = %d to Height = %d",
		tmBlockStoreState.Height,
		tmTargetHeight,
	)
	tmBlockStoreState.Height = tmTargetHeight
	tmBlockStoreState.Save(bsdb)
	fmt.Printf(" - Done !\n")
}
*/
