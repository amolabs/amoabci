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
	tmCfg "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo"
	cfg "github.com/amolabs/amoabci/config"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		// get default config
		config := cfg.DefaultConfig()

		// parse flags
		amoDirPath, err := cmd.Flags().GetString("home")
		if err != nil {
			return err
		}

		// set root dir
		config.SetRoot(amoDirPath)
		tmCfg.EnsureRoot(config.RootDir)

		// parse and validate config
		configFile := filepath.Join(
			config.RootDir,
			config.ConfigDir,
			cfg.DefaultConfigFileName,
		)
		vp := viper.New()
		vp.SetConfigFile(configFile)
		err = vp.ReadInConfig()
		if err != nil {
			return err
		}
		err = vp.UnmarshalExact(config)
		if err != nil {
			return err
		}
		err = config.ValidateBasic()
		if err != nil {
			return err
		}

		// set RLIMIT_NOFILE
		var rLimit syscall.Rlimit
		err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return err
		}
		if config.RLimitNoFile <= rLimit.Max {
			rLimit.Cur = config.RLimitNoFile
		}
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return err
		}

		// logger
		logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
		logger, err = tmflags.ParseLogLevel(
			config.LogLevel,
			logger,
			tmCfg.DefaultLogLevel(),
		)

		app, err := initApp(config, logger)
		if err != nil {
			return err
		}
		//defer app.Close()

		node, err := newTM(app, config, logger)
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

func initApp(config *cfg.Config, logger log.Logger) (*amo.AMOApp, error) {
	dataDirPath := filepath.Join(config.RootDir, config.DataDir)

	stateFilePath := filepath.Join(dataDirPath, config.StateFile)
	stateFile, err := os.OpenFile(stateFilePath, os.O_CREATE, os.FileMode(0644))
	if err != nil {
		return nil, err
	}

	merkleDB := tmdb.NewDB(
		config.MerkleDB,
		tmdb.BackendType(config.DBBackend),
		dataDirPath,
	)
	indexDB := tmdb.NewDB(
		config.IndexDB,
		tmdb.BackendType(config.DBBackend),
		dataDirPath,
	)

	// create app
	// TODO: read checkpoint_interval from config
	app := amo.NewAMOApp(
		stateFile, 100,
		merkleDB, indexDB,
		logger.With("module", "abci-app"),
	)

	return app, nil
}

func newTM(app abci.Application, config *cfg.Config, logger log.Logger) (
	*nm.Node, error) {
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

	tmConfig := config.ExportToTmCfg()

	// create node
	return nm.NewNode(
		tmConfig,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(tmConfig),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger,
	)
}

func init() {
	RunCmd.Flags().String("cpuprofile", "", "write cpu profile to `file`")
	RunCmd.Flags().String("memprofile", "", "write mem profile to `file`")
}
