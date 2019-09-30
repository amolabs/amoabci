package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "amod",
	Short:            "AMO daemon management",
	PersistentPreRun: nil,
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

	rootCmd.AddCommand(runCmd)
	rootCmd.PersistentFlags().String("home", defaultAMODirPath, "AMO home directory")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
