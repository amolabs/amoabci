package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
)

var LineBreak = &cobra.Command{Run: func(*cobra.Command, []string) {}}

var rootCmd = &cobra.Command{
	Use:              "amocli",
	Short:            "Console app for a user to interact with AMO daemon",
	PersistentPreRun: loadConfig,
}

// Execute function is the main gate to this app
func Execute() {
	cobra.EnableCommandSorting = false

	rootCmd.AddCommand(
		versionCmd,
		keyCmd,
		LineBreak,
		statusCmd,
		LineBreak,
		queryCmd,
		txCmd,
		dbCmd,
		LineBreak,
	)
	rootCmd.PersistentFlags().String("rpc", "0.0.0.0:26657", "node_ip:port")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func loadConfig(cmd *cobra.Command, args []string) {
	rpcArg, err := cmd.Flags().GetString("rpc")
	if err == nil {
		rpc.RpcRemote = "tcp://" + rpcArg
	}
}
