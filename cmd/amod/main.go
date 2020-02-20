package main

import (
	"os"
	"path/filepath"

	"github.com/amolabs/amoabci/cmd/amod/cmd"
	"github.com/spf13/cobra"
	tm "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
)

/* Commands (expected hierarchy)
 *
 * amod |- run
 */

func main() {
	cobra.EnableCommandSorting = false

	rootCmd := cmd.RootCmd
	runCmd := cmd.RunCmd
	tmCmd := tm.RootCmd
	tmCmd.AddCommand(
		tm.GenValidatorCmd,
		tm.InitFilesCmd,
		tm.ProbeUpnpCmd,
		tm.LiteCmd,
		tm.ReplayCmd,
		tm.ReplayConsoleCmd,
		tm.ResetAllCmd,
		tm.ResetPrivValidatorCmd,
		tm.ShowValidatorCmd,
		tm.TestnetFilesCmd,
		tm.ShowNodeIDCmd,
		tm.GenNodeKeyCmd,
		tm.VersionCmd,
	)

	cli.PrepareBaseCmd(runCmd, "TM", os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultTendermintDir)))
	cli.PrepareBaseCmd(tmCmd, "TM", os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultTendermintDir)))

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(tmCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
