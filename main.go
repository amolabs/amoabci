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
	_, err := util.StartInProcess(db)
	return err
}
