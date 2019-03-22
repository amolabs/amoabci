package db

import (
	"github.com/spf13/cobra"
)

var UploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload data into db",
	//Args: cobra.NoArgs,
	Args: cobra.MinimumNArgs(1),
	RunE: uploadFunc,
}

func uploadFunc(cmd *cobra.Command, args []string) error {

	return nil
}
