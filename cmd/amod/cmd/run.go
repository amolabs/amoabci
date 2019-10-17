package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/abci/server"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/amolabs/amoabci/amo"
	"github.com/amolabs/amoabci/amo/store"
)

/* Commands (expected hierarchy)
 *
 * amod |- run
 */

var (
	defaultAMODir = ".amo"

	defaultDataDir = "data"

	defaultMerkleDB    = "merkle"
	defaultIndexDB     = "index"
	defaultIncentiveDB = "incentive"

	defaultStateFile = "state.json"

	defaultAMODirPath = filepath.Join(os.ExpandEnv("$HOME"), defaultAMODir) // $HOME/.amo/
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		amoDirPath, err := cmd.Flags().GetString("home")
		if err != nil {
			return err
		}

		err = initApp(amoDirPath)
		if err != nil {
			return err
		}
		return nil
	},
}

func initApp(amoDirPath string) error {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	appLogger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	if _, err := os.Stat(amoDirPath); os.IsNotExist(err) {
		os.Mkdir(amoDirPath, os.FileMode(0700))
	}

	stateFilePath := filepath.Join(amoDirPath, defaultStateFile)

	stateFile, err := os.OpenFile(stateFilePath, os.O_CREATE, os.FileMode(0644))
	if err != nil {
		return err
	}

	// TODO: do not use hard-coded value. use value from configuration.
	dataDirPath := filepath.Join(amoDirPath, defaultDataDir)

	merkleDB, err := store.NewDBProxy(defaultMerkleDB, dataDirPath)
	if err != nil {
		return err
	}

	indexDB, err := store.NewDBProxy(defaultIndexDB, dataDirPath)
	if err != nil {
		return err
	}

	incentiveDB, err := store.NewDBProxy(defaultIncentiveDB, dataDirPath)
	if err != nil {
		return err
	}

	app := amo.NewAMOApp(
		stateFile,
		merkleDB, indexDB, incentiveDB,
		appLogger.With("module", "abci-app"),
	)

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
	// init here
}
