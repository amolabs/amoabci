package key

import (
	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/keys"
	"github.com/amolabs/amoabci/client/util"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all of keys stored on the local storage",
	Args:  cobra.NoArgs,
	RunE:  listFunc,
}

func listFunc(cmd *cobra.Command, args []string) error {
	keyFile := util.DefaultKeyFilePath()

	err := keys.List(keyFile)
	if err != nil {
		return err
	}

	return nil
}
