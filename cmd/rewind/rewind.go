package main

import (
	"fmt"

	"github.com/tendermint/iavl"
	tmdb "github.com/tendermint/tm-db"
)

const (
	merkleTreeCacheSize = 10000
)

func rewind(amoRoot string, doFix bool, rewindVersion int64) {
	merkleDB := tmdb.NewDB("merkle", tmdb.RocksDBBackend, amoRoot+"/data")
	defer merkleDB.Close()

	amoMt, err := iavl.NewMutableTree(merkleDB, merkleTreeCacheSize)
	if err != nil {
		fmt.Println(err)
		return
	}
	ver, err := amoMt.Load()

	fmt.Println("current merkle version", ver)

	if rewindVersion > 0 && ver > rewindVersion && doFix {
		ver, err = amoMt.LoadVersionForOverwriting(rewindVersion)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("fixed merkle version", ver)
	}
}
