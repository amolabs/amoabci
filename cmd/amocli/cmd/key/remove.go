package key

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/keys"
	"github.com/amolabs/amoabci/client/util"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove <username>",
	Short: "Remove the specified key",
	Args:  cobra.MinimumNArgs(1),
	RunE:  removeFunc,
}

func removeFunc(cmd *cobra.Command, args []string) error {
	username := args[0]
	keyFile := util.DefaultKeyFilePath()

	kr, err := keys.GetKeyRing(keyFile)
	if err != nil {
		return err
	}
	err = kr.RemoveKey(username)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully removed the key with username: %s\n", username)
	return nil
}
