package cmd

import (
	dbm "github.com/amolabs/tendermint-amo/libs/db"
	"os"

	"github.com/amolabs/amoabci/amo"
	"github.com/amolabs/tendermint-amo/abci/server"
	"github.com/amolabs/tendermint-amo/abci/types"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"github.com/amolabs/tendermint-amo/libs/log"

	"github.com/spf13/cobra"
)

/* Commands (expected hierarchy)
 *
 * amod |- run
 */

var runCmd = &cobra.Command{
	Use:	"run",
	Short:	"Execute the daemon",
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
	var app types.Application
	db := dbm.NewMemDB()
	app = amo.NewAMOApplication(db)
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
