package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "amod",
	Short: "AMO daemon management",
}

func init() {
	// init here
}
