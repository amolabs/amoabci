package cmd

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/amolabs/amoabci/client/keys"
	"github.com/amolabs/amoabci/client/util"
)

/* Commands (expected hierarchy)
 *
 * amocli |- key |- list
 *				 |- generate <nickname>
 *			 	 |- remove <nickname>
 */

var keyCmd = &cobra.Command{
	Use:     "key",
	Aliases: []string{"k"},
	Short:   "Manage the key(wallet)-related features",
}

func init() {
	listCmd := keyListCmd

	genCmd := keyGenCmd
	genCmd.Flags().SortFlags = false
	genCmd.Flags().BoolP("encrypt", "e", true,
		"encrypt the private key with passphrase. default: true")

	removeCmd := keyRemoveCmd

	cmd := keyCmd
	cmd.AddCommand(listCmd, genCmd, removeCmd)
}

var keyListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all of keys stored on the local storage",
	Args:  cobra.NoArgs,
	RunE:  keyListFunc,
}

func keyListFunc(cmd *cobra.Command, args []string) error {
	keyFile := util.DefaultKeyFilePath()

	err := keys.List(keyFile)
	if err != nil {
		return err
	}

	return nil
}

var keyGenCmd = &cobra.Command{
	Use:   "generate <nickname>",
	Short: "Generate a key with a specified nickname",
	Args:  cobra.MinimumNArgs(1),
	RunE:  keyGenFunc,
}

func keyGenFunc(cmd *cobra.Command, args []string) error {
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

var keyRemoveCmd = &cobra.Command{
	Use:   "remove <nickname>",
	Short: "Remove the specified key",
	Args:  cobra.MinimumNArgs(1),
	RunE:  keyRemoveFunc,
}

func keyRemoveFunc(cmd *cobra.Command, args []string) error {
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
