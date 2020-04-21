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
		amoRoot := args[0]
		inspect(amoRoot)
		repair(amoRoot)
	},
}

func main() {
	//rootCmd.PersistentFlags().String("root", "",
	//	"data root (contains config and data)")
	//rootCmd.MarkFlagRequired("root")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
