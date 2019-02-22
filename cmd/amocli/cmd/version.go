package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VERSION represents the general version of this app
const VERSION = "0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(VERSION)
	},
}

func init() {
	// init here if needed
}
