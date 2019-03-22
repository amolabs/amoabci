package db

import (
	"github.com/spf13/cobra"
)

var QueryCmd = &cobra.Command{
	Use:   "query",
	Short: "query data",
	//Args: cobra.NoArgs,
	Args: cobra.MinimumNArgs(1),
	RunE: queryFunc,
}

func queryFunc(cmd *cobra.Command, args []string) error {

	return nil
}
