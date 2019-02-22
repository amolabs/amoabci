package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "amoconsole",
	Short:             "Console app for a user to interact with 'amo' daemon",
	PersistentPreRunE: loadConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}

		return nil
	},
}

// Execute function is the main gate to this app
func Execute() {
	cobra.EnableCommandSorting = false

	rootCmd.AddCommand(
		versionCmd,
		statusCmd,
		keyCmd,
		txCmd,
		queryCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func loadConfig(cmd *cobra.Command, args []string) error {
	return nil
}
