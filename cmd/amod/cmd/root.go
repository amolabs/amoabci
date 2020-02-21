package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	defaultAMODir = ".amo"

	defaultDataDir   = "data"
	defaultConfigDir = "config"

	defaultMerkleDB       = "merkle"
	defaultIndexDB        = "index"
	defaultIncentiveDB    = "incentive"
	defaultGroupCounterDB = "group_counter"

	defaultStateFile = "state.json"

	DefaultAMODirPath = filepath.Join(os.ExpandEnv("$HOME"), defaultAMODir)
)

var RootCmd = &cobra.Command{
	Use:   "amod",
	Short: "AMO daemon management",
}

func init() {
	// init here
}
