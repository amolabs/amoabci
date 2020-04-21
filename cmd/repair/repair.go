package main

import (
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
	amoStateFile.Close()

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

	merkleDB, err := tmdb.NewGoLevelDB("merkle", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	mt, err := iavl.NewMutableTree(merkleDB, merkleTreeCacheSize)
	merkleVersion, err := mt.LoadVersionForOverwriting(amoState.Height + 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	//merkleVersions := mt.Version()
	fmt.Println("fixed merkle version =", merkleVersion)

	db, err := tmdb.NewGoLevelDB("state", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	tmState := tmstate.LoadState(db)
	fmt.Println("TM state height  =", tmState.LastBlockHeight)
	fmt.Println("AppHash =", strings.ToUpper(
		hex.EncodeToString(tmState.AppHash)))

	bdb, err := tmdb.NewGoLevelDB("blockstore", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	bsj := tmstore.LoadBlockStoreStateJSON(bdb)
	fmt.Println("Block store height  =", bsj.Height)
	bsj.Height = amoState.Height
	bsj.Save(bdb)
	fmt.Println("fixed Block store height  =", bsj.Height)

	//return nil
}

func repair(amoRoot string) {
	fmt.Println("Reparing data root:", amoRoot)
}
