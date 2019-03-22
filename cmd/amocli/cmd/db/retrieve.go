package db

import (
	"github.com/spf13/cobra"
)

var RetrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve data",
	//Args: cobra.NoArgs,
	Args: cobra.MinimumNArgs(1),
	RunE: retrieveFunc,
}

func retrieveFunc(cmd *cobra.Command, args []string) error {

	return nil
}
