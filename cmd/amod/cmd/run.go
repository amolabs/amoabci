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
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		amoDirPath, err := cmd.Flags().GetString("home")
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

	dataDirPath := filepath.Join(config.RootDir, defaultDataDir)

	// state file
	stateFilePath := filepath.Join(dataDirPath, defaultStateFile)
	stateFile, err := os.OpenFile(stateFilePath, os.O_CREATE, os.FileMode(0644))
	if err != nil {
		return nil, err
	}

	// TODO: do not use hard-coded value. use value from configuration.
	merkleDB, err := tmdb.NewGoLevelDB(defaultMerkleDB, dataDirPath)
	if err != nil {
		return nil, err
	}

	indexDB, err := tmdb.NewGoLevelDB(defaultIndexDB, dataDirPath)
	if err != nil {
		return nil, err
	}

	incentiveDB, err := tmdb.NewGoLevelDB(defaultIncentiveDB, dataDirPath)
	if err != nil {
		return nil, err
	}

	groupCounterDB, err := tmdb.NewGoLevelDB(defaultGroupCounterDB, dataDirPath)
	if err != nil {
		return nil, err
	}

	// logger
	appLogger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	// create app
	app, err := amo.NewAMOApp(
		stateFile,
		merkleDB, indexDB, incentiveDB, groupCounterDB,
		appLogger.With("module", "abci-app"),
	)
	if err != nil {
		return nil, err
	}

	node, err := newTM(app, config)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func newTM(app abci.Application, config *cfg.Config) (*nm.Node, error) {
	// logger
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger, err := tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
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
}
