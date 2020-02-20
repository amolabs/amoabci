package cmd

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"

	"github.com/amolabs/amoabci/amo"
	"github.com/amolabs/amoabci/amo/store"
)

/* Commands (expected hierarchy)
 *
 * amod |- run
 */

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		amoDirPath, err := cmd.Flags().GetString("amo-home")
		if err != nil {
			return err
		}

		node, err := initApp(amoDirPath)
		if err != nil {
			return err
		}

		node.Start()
		defer func() {
			node.Stop()
			node.Wait()
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		os.Exit(0)

		return nil
	},
}

func initApp(amoDirPath string) (*nm.Node, error) {
	// amo dir
	if _, err := os.Stat(amoDirPath); os.IsNotExist(err) {
		os.Mkdir(amoDirPath, os.FileMode(0700))
	}

	// state file
	stateFilePath := filepath.Join(amoDirPath, defaultStateFile)
	stateFile, err := os.OpenFile(stateFilePath, os.O_CREATE, os.FileMode(0644))
	if err != nil {
		return nil, err
	}

	// TODO: do not use hard-coded value. use value from configuration.
	dataDirPath := filepath.Join(amoDirPath, defaultDataDir)

	merkleDB, err := store.NewDBProxy(defaultMerkleDB, dataDirPath)
	if err != nil {
		return nil, err
	}

	indexDB, err := store.NewDBProxy(defaultIndexDB, dataDirPath)
	if err != nil {
		return nil, err
	}

	incentiveDB, err := store.NewDBProxy(defaultIncentiveDB, dataDirPath)
	if err != nil {
		return nil, err
	}

	groupCounterDB, err := store.NewDBProxy(defaultGroupCounterDB, dataDirPath)
	if err != nil {
		return nil, err
	}

	// logger
	appLogger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	// create app
	app := amo.NewAMOApp(
		stateFile,
		merkleDB, indexDB, incentiveDB, groupCounterDB,
		appLogger.With("module", "abci-app"),
	)

	node, err := newTM(app)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func newTM(app abci.Application) (*nm.Node, error) {
	// parse config
	config := cfg.DefaultConfig()
	err := viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	config.SetRoot(config.RootDir)
	cfg.EnsureRoot(config.RootDir)
	err = config.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// logger
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
	if err != nil {
		return nil, err
	}

	// read private validator
	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)

	// read node key
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	// create node
	return nm.NewNode(
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger,
	)
}

func init() {
	// init here
	RunCmd.Flags().String("amo-home", defaultAMODirPath, "AMO home directory")
}
