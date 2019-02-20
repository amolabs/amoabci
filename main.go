package main

import (
	"github.com/amolabs/amoabci/amo"
	"github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	"os"
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
	db, err := amo.LoadDB()
	if err != nil {
		return err
	}
	app = amo.NewAMOApplication(db)
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
		if err!= nil {
			panic(err)
		}
	})
	return nil
}
