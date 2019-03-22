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

var GenCmd = &cobra.Command{
	Use:   "generate <nickname>",
	Short: "Generate a key with a specified nickname",
	Args:  cobra.MinimumNArgs(1),
	RunE:  genFunc,
}

func init() {
	cmd := GenCmd
	cmd.Flags().SortFlags = false
	cmd.Flags().BoolP("encrypt", "e", true, "encrypt the private key with passphrase")
}

func genFunc(cmd *cobra.Command, args []string) error {
	nickname := args[0]
	keyFile := util.DefaultKeyFilePath()
	flags := cmd.Flags()

	keyStatus := keys.Check(nickname, keyFile)
	if keyStatus > keys.NoExists {
		return errors.New("The key already exists")
	}

	encrypt, err := flags.GetBool("encrypt")
	if err != nil {
		return err
	}

	var passphrase []byte

	if encrypt {
		fmt.Printf("Type passphrase: ")
		passphrase, err = terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return err
		}
	}

	err = keys.Generate(nickname, passphrase, encrypt, keyFile)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully generated the key with nickname: %s\n", nickname)

	return nil
}
