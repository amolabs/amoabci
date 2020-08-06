package config

import (
	"os"
	"path/filepath"
	"syscall"

	tmCfg "github.com/tendermint/tendermint/config"
)

var (
	defaultAMODir     = ".amo"
	DefaultAMODirPath = filepath.Join(os.ExpandEnv("$HOME"), defaultAMODir)

	defaultDataDir   = "data"
	defaultConfigDir = "config"
	defaultMerkleDB  = "merkle"
	defaultIndexDB   = "index"
	defaultStateFile = "state.json"

	DefaultConfigFileName = "config.toml"
)

type AMOConfig struct {
	DataDir      string `mapstructure:"data_dir"`
	ConfigDir    string `mapstructure:"config_dir"`
	MerkleDB     string `mapstructure:"merkle_db"`
	IndexDB      string `mapstructure:"index_db"`
	StateFile    string `mapstructure:"state_file"`
	RLimitNoFile uint64 `mapstructure:"rlimit_nofile"`

	*tmCfg.Config
}

type Config struct {
	AMOConfig    `mapstructure:"amo"`
	tmCfg.Config `mapstructure:",squash"`
}

func DefaultConfig() *Config {
	var defaultrLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &defaultrLimit)
	if err != nil {
		panic(err)
	}

	return &Config{
		AMOConfig: AMOConfig{
			DataDir:      defaultDataDir,
			ConfigDir:    defaultConfigDir,
			MerkleDB:     defaultMerkleDB,
			IndexDB:      defaultIndexDB,
			StateFile:    defaultStateFile,
			RLimitNoFile: defaultrLimit.Cur, // soft limit
		},
		Config: *tmCfg.DefaultConfig(),
	}
}

func (cfg *Config) ExportToTmCfg() *tmCfg.Config {
	return &tmCfg.Config{
		BaseConfig:      cfg.BaseConfig,
		RPC:             cfg.RPC,
		P2P:             cfg.P2P,
		Mempool:         cfg.Mempool,
		FastSync:        cfg.FastSync,
		Consensus:       cfg.Consensus,
		TxIndex:         cfg.TxIndex,
		Instrumentation: cfg.Instrumentation,
	}
}
