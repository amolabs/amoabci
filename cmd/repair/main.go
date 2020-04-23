package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "repair",
	Short: "AMO state repair tool",
	Args:  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		doFix, err := cmd.Flags().GetBool("fix")
		if err != nil {
			fmt.Println(err)
			return
		}
		amoRoot := args[0]
		repair(amoRoot, doFix)
	},
}

func main() {
	rootCmd.PersistentFlags().BoolP("fix", "f", false, "do fix")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
