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
		rewindMerkle, err := cmd.Flags().GetBool("merkle")
		if err != nil {
			fmt.Println(err)
			return
		}
		amoRoot := args[0]
		repair(amoRoot, doFix, rewindMerkle)
	},
}

func main() {
	rootCmd.PersistentFlags().BoolP("fix", "f", false, "do fix")
	rootCmd.PersistentFlags().BoolP("merkle", "m", false, "force-rewind merkle db")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
