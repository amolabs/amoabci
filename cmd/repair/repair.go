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

	amoState := amo.State{}
	amoState.LoadFrom(amoStateFile)
	defer amoStateFile.Close()

	/*
		j, err := json.MarshalIndent(amoState, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(j))
	*/

	fmt.Println("AMO state height =", amoState.Height)
	fmt.Println("AppHash =", strings.ToUpper(
		hex.EncodeToString(amoState.AppHash)))

	targetHeight := amoState.LastHeight - 1
	targetVersion := targetHeight + 1

	merkleDB, err := tmdb.NewGoLevelDB("merkle", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer merkleDB.Close()

	mt, err := iavl.NewMutableTree(merkleDB, merkleTreeCacheSize)
	_, err = mt.LoadVersionForOverwriting(targetVersion)
	if err != nil {
		fmt.Println(err)
		return
	}

	amoState.MerkleVersion = targetVersion
	amoState.LastAppHash = mt.Hash()
	amoState.AppHash = mt.Hash()
	amoState.LastHeight = targetHeight
	amoState.Height = targetHeight

	fmt.Printf("fixed amoState.MerkleVersion = %d\n", amoState.MerkleVersion)
	fmt.Printf("fixed amoState.LastAppHash   = %x\n", amoState.LastAppHash)
	fmt.Printf("fixed amoState.AppHash       = %x\n", amoState.AppHash)
	fmt.Printf("fixed amoState.LastHeight    = %d\n", amoState.LastHeight)
	fmt.Printf("fixed amoState.Height        = %d\n", amoState.Height)

	err = amoState.SaveTo(amoStateFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("fixed amoState")

	bdb, err := tmdb.NewGoLevelDB("blockstore", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer bdb.Close()

	bsj := tmstore.LoadBlockStoreStateJSON(bdb)
	fmt.Printf("TM block store state height = %d\n", bsj.Height)

	bst := tmstore.NewBlockStore(bdb)
	lb := bst.LoadBlock(targetHeight + 1)
	for i := bsj.Height; i > targetHeight; i-- {
		b := bst.LoadBlock(i)

		fmt.Printf("BlockHeight = %d, BlockHash = %x\n", b.Height, b.Hash())

		err = bdb.DeleteSync([]byte(fmt.Sprintf("BH:%s", strings.ToLower(hex.EncodeToString(b.Hash())))))
		if err != nil {
			fmt.Println(err)
			continue
		}

		targetKey := []byte(fmt.Sprintf("P:%v:", i))
		itr, err := bdb.Iterator(targetKey, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer itr.Close()

		for ; itr.Valid(); itr.Next() {
			if !bytes.HasPrefix(itr.Key(), targetKey) {
				continue
			}

			err := bdb.DeleteSync(itr.Key())
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		err = bdb.DeleteSync([]byte(fmt.Sprintf("H:%v", i)))
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = bdb.DeleteSync([]byte(fmt.Sprintf("C:%v", i)))
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = bdb.DeleteSync([]byte(fmt.Sprintf("SC:%v", i)))
		if err != nil {
			fmt.Println(err)
			continue
		}

		btmp := bst.LoadBlockByHash(b.Hash())
		if btmp != nil {
			fmt.Printf("not properly deleted %x\n", btmp.Hash())
		}

		fmt.Printf("delete blockstore garbage block height = %d\n", b.Height)
	}

	fmt.Println("fixed BlockStore")

	bsj.Height = targetHeight
	bsj.Save(bdb)
	fmt.Printf("fixed blockStoreStateJSON.Height = %d\n", bsj.Height)

	fmt.Println("fixed BlockStoreStateJSON")

	db, err := tmdb.NewGoLevelDB("state", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	tmState := tmstate.LoadState(db)
	fmt.Printf("TM state block id   = %s\n", tmState.LastBlockID)
	fmt.Printf("TM state height     = %d\n", tmState.LastBlockHeight)
	fmt.Printf("TM state appHash    = %x\n", tmState.AppHash)
	fmt.Printf("TM state block time = %s\n", tmState.LastBlockTime)

	tmState.LastBlockID = lb.LastBlockID
	tmState.LastBlockHeight = targetHeight
	tmState.LastBlockTime = lb.Time

	//tmState.AppHash = amoState.AppHash
	tmstate.SaveState(db, tmState)

	fmt.Printf("fixed tmState.LastBlockID     = %s\n", tmState.LastBlockID)
	fmt.Printf("fixed tmState.LastBlockHeight = %d\n", tmState.LastBlockHeight)
	fmt.Printf("fixed tmState.AppHash         = %x\n", tmState.AppHash)
	fmt.Printf("fixed tmState.LastBlockTime   = %s\n", tmState.LastBlockTime)

	fmt.Println("fixed tmState")

	//return nil
}

func repair(amoRoot string) {
	fmt.Println("Reparing data root:", amoRoot)
}
