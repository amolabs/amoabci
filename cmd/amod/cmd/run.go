package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"syscall"
	"time"

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

		app, err := initApp(amoDirPath)
		if err != nil {
			return err
		}
		//defer app.Close()

		node, err := newTM(app)
		if err != nil {
			return err
		}

		cpuprof, _ := cmd.Flags().GetString("cpuprofile")
		if len(cpuprof) > 0 {
			f, err := os.Create(cpuprof)
			if err != nil {
				fmt.Println("unable to create cpu profile")
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				fmt.Println("unable to start cpu profile")
			}
			defer pprof.StopCPUProfile()
		}

		memprof, _ := cmd.Flags().GetString("memprofile")
		if len(memprof) > 0 {
			defer func() {
				mf, err := os.Create(memprof)
				if err != nil {
					fmt.Println("unable to create mem profile")
				}
				if err := pprof.WriteHeapProfile(mf); err != nil {
					fmt.Println("unable to write mem heap profile")
				}
				mf.Close()
			}()
		}

		node.Start()
		defer func() {
			node.Stop()
			node.ProxyApp().Stop()
			node.Wait()
			// XXX: I couldn't find the proper stopping sequence yet. So, just
			// wait until the TM closes all.
			time.Sleep(200000000) // 100ms
			//app.Close()
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		return nil
	},
}

func initApp(amoDirPath string) (*amo.AMOApp, error) {
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
	merkleDB := tmdb.NewDB(defaultMerkleDB,
		tmdb.BackendType(config.DBBackend), dataDirPath)
	indexDB := tmdb.NewDB(defaultIndexDB,
		tmdb.BackendType(config.DBBackend), dataDirPath)
	groupCounterDB := tmdb.NewDB(defaultGroupCounterDB,
		tmdb.BackendType(config.DBBackend), dataDirPath)

	// logger
	appLogger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	appLogger, err = tmflags.ParseLogLevel(config.LogLevel, appLogger,
		cfg.DefaultLogLevel())

	// create app
	app := amo.NewAMOApp(
		stateFile,
		merkleDB, indexDB, groupCounterDB,
		appLogger.With("module", "abci-app"),
	)

	return app, nil
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
	RunCmd.Flags().String("cpuprofile", "", "write cpu profile to `file`")
	RunCmd.Flags().String("memprofile", "", "write mem profile to `file`")
}
