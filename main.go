package main

import (
	"github.com/amolabs/amoabci/amo/util"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func main() {
	err := initApp()
	if err != nil {
		panic(err)
	}
}

func initApp() error {
	db := dbm.NewMemDB()
	/*
		db, err := dbm.NewGoLevelDB("state", path.Join(util.RootName, "state"))
		if err != nil {
			panic(err)
		}
	*/
	_, err := util.StartInProcess(db)
	return err
}
