package main

import (
	"os"

	"github.com/spf13/cobra"
	tm "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/amolabs/amoabci/cmd/amod/cmd"
	cfg "github.com/amolabs/amoabci/config"
)

/* Commands (expected hierarchy)
 *
 * amod |- run
 *      |- tendermint
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

	cli.PrepareBaseCmd(runCmd, "AMO", cfg.DefaultAMODirPath)
	cli.PrepareBaseCmd(tmCmd, "AMO", cfg.DefaultAMODirPath)

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(tmCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
