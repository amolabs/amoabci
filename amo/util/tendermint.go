package util

import (
	"encoding/hex"
	"github.com/amolabs/amoabci/amo"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"os"
	"path/filepath"
	"strings"
)

const (
	RootName   = "blockchain"
	configFile = "config/config.toml"
)

type Context struct {
	Config *cfg.Config
	Logger log.Logger
}

func NewDefaultContext() *Context {
	return NewContext(
		cfg.DefaultConfig(),
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	)
}

func NewContext(config *cfg.Config, logger log.Logger) *Context {
	return &Context{config, logger}
}

func createConfigFile(ctx *Context, path string) {
	config := ctx.Config
	config.Consensus.CreateEmptyBlocks = false
	config.RPC.CORSAllowedOrigins = []string{"*"}
	cfg.EnsureRoot(config.RootDir)
	cfg.WriteConfigFile(path, config)
}

func loadConfigFile(config *cfg.Config, configFilePath string) error {
	v := viper.New()
	v.SetConfigFile(configFilePath)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	err = v.Unmarshal(config)
	if err != nil {
		return err
	}
	err = config.ValidateBasic()
	if err != nil {
		return err
	}
	return nil
}

func setOwners(app *amo.AMOApplication, owners []atypes.GenesisOwner) {
	for _, owner := range owners {
		encoded := strings.ToUpper(hex.EncodeToString(owner.Address))
		address := atypes.NewAddress([]byte(encoded))
		account := atypes.Account{
			Balance: owner.Amount,
			PurchasedFiles: make(atypes.HashSet),
		}
		app.SetAccount(*address, &account)
	}
}

func StartInProcess(db dbm.DB) (*node.Node, error) {
	ctx := NewDefaultContext()
	config := ctx.Config
	config.SetRoot(RootName)
	configFilePath := filepath.Join(ctx.Config.RootDir, configFile)
	if !cmn.FileExists(configFilePath) {
		createConfigFile(ctx, configFilePath)
	} else {
		// load config file
		if err := loadConfigFile(config, configFilePath); err != nil {
			return nil, err
		}
	}
	// Create config
	if err := InitFilesWithConfig(config, ctx.Logger); err != nil {
		panic(err)
	}
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	// Create AMO abci
	app := amo.NewAMOApplication(db)
	addRoutes()
	var genDoc atypes.AMOGenesisDoc
	err = genDoc.GenesisDocFromFile(config.GenesisFile())
	if err != nil {
		panic(err)
	}
	// Create tendermint and combine AMO abci
	tmNode, err := node.NewNode(
		config,
		privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(app),
		func() (*types.GenesisDoc, error) {
			return &genDoc.GenesisDoc, nil
		},
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(config.Instrumentation),
		ctx.Logger.With("module", "node"),
	)
	if err != nil {
		return nil, err
	}
	setOwners(app, genDoc.Owners)

	// TEST CODE
	buyer := app.GetBuyer(atypes.H1)
	(*buyer)[*atypes.SampleAddress] = true
	app.SetBuyer(atypes.H1, buyer)
	acc := app.GetAccount(*atypes.SampleAddress)
	acc.PurchasedFiles[atypes.H1] = true
	app.SetAccount(*atypes.SampleAddress, acc)
	// TEST CODE

	// Run
	err = tmNode.Start()
	if err != nil {
		return nil, err
	}

	cmn.TrapSignal(func() {
		if tmNode.IsRunning() {
			_ = tmNode.Stop()
		}
	})

	select {}
	return nil, nil
}

// From tendermint/cmd/tendermint/commands/init
func InitFilesWithConfig(config *cfg.Config, logger log.Logger) error {
	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	if cmn.FileExists(privValKeyFile) {
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
		logger.Info("Found private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	} else {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		logger.Info("Generated private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	}

	nodeKeyFile := config.NodeKeyFile()
	if cmn.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if cmn.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {
		genDoc := atypes.AMOGenesisDoc{}
		genDoc.ChainID = atypes.ChainID
		genDoc.GenesisTime = tmtime.Now()
		genDoc.ConsensusParams = types.DefaultConsensusParams()
		key := pv.GetPubKey()
		genDoc.Validators = []types.GenesisValidator{{
			Address: key.Address(),
			PubKey:  key,
			Power:   10,
		}}
		genDoc.Owners = []atypes.GenesisOwner{{
			Address: key.Address(),
			PubKey: key,
			Amount: 3000,
		}}
		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		logger.Info("Generated genesis file", "path", genFile)
	}

	return nil
}
