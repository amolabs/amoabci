package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "repair",
	Short: "AMO merkle rewind tool",
	Args:  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		doFix, err := cmd.Flags().GetBool("fix")
		if err != nil {
			fmt.Println(err)
			return
		}
		rewindVersion, err := cmd.Flags().GetInt64("rewind_version")
		if err != nil {
			fmt.Println(err)
			return
		}
		amoRoot := args[0]
		rewind(amoRoot, doFix, rewindVersion)
	},
}

func main() {
	rootCmd.PersistentFlags().BoolP("fix", "f", false, "do fix")
	rootCmd.PersistentFlags().Int64P("rewind_version", "r", 0, "version to rewind")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
