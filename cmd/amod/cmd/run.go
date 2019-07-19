package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/abci/server"
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
	// TODO: do not use hard-coded value. use value from configuration.
	db, err := dbm.NewGoLevelDB("store", "data/state")
	if err != nil {
		return err
	}
	index, err := dbm.NewGoLevelDB("index", "data/index")
	if err != nil {
		return err
	}
	app := amo.NewAMOApp(db, index, appLogger.With("module", "abci-app"))
	srv, err := server.NewServer("tcp://0.0.0.0:26658", "socket", app)
	if err != nil {
		return err
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		return err
	}
	cmn.TrapSignal(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "abci-server"),
		func() {
			// Cleanup
			srv.Stop()
		})

	select {}
}

func init() {
	// init here if needed
}
