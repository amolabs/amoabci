package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/amolabs/amoabci/amo"
)

/* Commands (expected hierarchy)
 *
 * amod |- run
 */

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initApp()
		if err != nil {
			return err
		}
		return nil
	},
}

func initApp() error {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	appLogger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	var app types.Application
	// TODO: do not use hard-coded value. use value from configuration.
	db, err := dbm.NewGoLevelDB("store", "blockchain/store")
	if err != nil {
		return err
	}
	index, err := dbm.NewGoLevelDB("index", "blockchain/index")
	if err != nil {
		return err
	}
	app = amo.NewAMOApplication(db, index, appLogger.With("module", "abci-app"))
	srv, err := server.NewServer("tcp://0.0.0.0:26658", "socket", app)
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

func init() {
	// init here if needed
}
