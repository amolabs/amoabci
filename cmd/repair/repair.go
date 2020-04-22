package main

import (
	"bytes"
	"encoding/hex"
	//"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tendermint/iavl"
	tmstate "github.com/tendermint/tendermint/state"
	tmstore "github.com/tendermint/tendermint/store"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo"
	//amostore "github.com/amolabs/amoabci/amo/store"
)

const (
	merkleTreeCacheSize = 10000
)

func inspect(amoRoot string) {
	fmt.Println("Inspecting data root:", amoRoot)
	fn := amoRoot + "/data/state.json"
	amoStateFile, err := os.OpenFile(fn, os.O_RDONLY, 0)
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

	amoState := amo.State{}
	err = amoState.LoadFrom(amoStateFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	amoMerkleTree, err := iavl.NewMutableTree(merkleDB, merkleTreeCacheSize)
	if err != nil {
		fmt.Println(err)
		return
	}

	amoMerkleTree.Load()

	tmBlockStoreState := tmstore.LoadBlockStoreStateJSON(bsdb)
	tmState := tmstate.LoadState(sdb)

	fmt.Printf("amoMerkleTree\n")
	fmt.Printf("  .Version         = %d\n", amoMerkleTree.Version())
	fmt.Printf("  .Hash            = %x\n", amoMerkleTree.Hash())
	fmt.Printf("amoState\n")
	fmt.Printf("  .LastHeight      = %d\n", amoState.LastHeight)
	fmt.Printf("  .Height          = %d\n", amoState.Height)
	fmt.Printf("  .LastAppHash     = %x\n", amoState.LastAppHash)
	fmt.Printf("  .AppHash         = %x\n", amoState.AppHash)
	fmt.Printf("tmBlockStoreState\n")
	fmt.Printf("  .height          = %d\n", tmBlockStoreState.Height)
	fmt.Printf("tmState\n")
	fmt.Printf("  .LastBlockID     = %s\n", tmState.LastBlockID)
	fmt.Printf("  .LastBlockHeight = %d\n", tmState.LastBlockHeight)
	fmt.Printf("  .LastBlockTime   = %s\n", tmState.LastBlockTime)
	fmt.Printf("  .AppHash         = %x\n", tmState.AppHash)
}

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
