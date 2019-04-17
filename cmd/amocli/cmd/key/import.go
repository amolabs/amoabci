package key

import (
	"encoding/base64"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/amolabs/amoabci/client/keys"
	"github.com/amolabs/amoabci/client/util"
)

var ImportCmd = &cobra.Command{
	Use:   "import <private key> --username <username>",
	Short: "Import a raw private key from base64-formatted string",
	Args:  cobra.MinimumNArgs(1),
	RunE:  importFunc,
}

func init() {
	cmd := ImportCmd
	cmd.Flags().SortFlags = false
	cmd.Flags().BoolP("encrypt", "e", true, "encrypt the private key with passphrase")
	cmd.Flags().StringP("username", "n", "", "specify a username for the key")

	cmd.MarkFlagRequired("username")
}

func importFunc(cmd *cobra.Command, args []string) error {
	var (
		privKey    []byte
		username   string
		encrypt    bool
		passphrase []byte
		err        error
	)

	keyFile := util.DefaultKeyFilePath()
	flags := cmd.Flags()

	privKey, err = base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return err
	}

	username, err = flags.GetString("username")
	if err != nil {
		return err
	}

	encrypt, err = flags.GetBool("encrypt")
	if err != nil {
		return err
	}

	if encrypt {
		fmt.Printf("Type passphrase: ")
		passphrase, err = terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return err
		}
	}

	kr, err := keys.GetKeyRing(keyFile)
	if err != nil {
		return err
	}
	_, err = kr.ImportPrivKey(privKey, username, passphrase, encrypt)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully imported the key with username: %s\n", username)

	return nil
}
