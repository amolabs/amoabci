package key

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/amolabs/amoabci/client/keys"
	"github.com/amolabs/amoabci/client/util"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove <nickname>",
	Short: "Remove the specified key",
	Args:  cobra.MinimumNArgs(1),
	RunE:  removeFunc,
}

func removeFunc(cmd *cobra.Command, args []string) error {
	var (
		passphrase []byte
		err        error
	)

	nickname := args[0]
	keyFile := util.DefaultKeyFilePath()

	keyStatus := keys.Check(nickname, keyFile)
	if keyStatus < keys.Exists {
		return errors.New("The key doesn't exist")
	} else if keyStatus == keys.Encrypted {
		fmt.Printf("Type passphrase: ")
		passphrase, err = terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return err
		}
	}

	err = keys.Remove(nickname, passphrase, keyFile)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully removed the key with nickname: %s\n", nickname)
	return nil
}
