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
	kr, err := keys.GetKeyRing(keyFile)
	if err != nil {
		return err
	}

	kr.PrintKeyList()

	return nil
}
