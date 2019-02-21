package main

import (
	dbm "github.com/amolabs/tendermint-amo/libs/db"
	"os"

	"github.com/amolabs/amoabci/amo"
	"github.com/amolabs/tendermint-amo/abci/server"
	"github.com/amolabs/tendermint-amo/abci/types"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"github.com/amolabs/tendermint-amo/libs/log"
)

func main() {
	err := initApp()
	if err != nil {
		panic(err)
	}
}

func initApp() error {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	var app types.Application
	db := dbm.NewMemDB()
	// TODO: load amo home
	storeRoot := "blockchain"
	app = amo.NewAMOApplication(db, storeRoot)
	srv, err := server.NewServer("tcp://0.0.0.0:46658", "socket", app)
	if err != nil {
		return err
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		return err
	}
	cmn.TrapSignal(func() {
		// Cleanup
		err := srv.Stop()
		if err != nil {
			panic(err)
		}
	})
	return nil
}
