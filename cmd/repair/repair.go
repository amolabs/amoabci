package main

import (
	"encoding/hex"
	//"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tendermint/tendermint/state"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo"
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

	db, err := tmdb.NewGoLevelDB("state", amoRoot+"/data")
	if err != nil {
		fmt.Println(err)
		return
	}
	tmState := state.LoadState(db)
	fmt.Println("TM state height  =", tmState.LastBlockHeight)
	fmt.Println("AppHash =", strings.ToUpper(
		hex.EncodeToString(tmState.AppHash)))

	//return nil
}

func repair(amoRoot string) {
	fmt.Println("Reparing data root:", amoRoot)
}
