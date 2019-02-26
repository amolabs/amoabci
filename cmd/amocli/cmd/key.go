package cmd

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/amolabs/amoabci/cmd/amocli/keys"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	listCmd := keyListCmd
	genCmd := keyGenCmd
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
	err := keys.List()
	if err != nil {
		return err
	}

	return nil
}

var keyGenCmd = &cobra.Command{
	Use:   "generate [nickname]",
	Short: "Generate a key with a specified nickname",
	Args:  cobra.MinimumNArgs(1),
	RunE:  keyGenFunc,
}

func keyGenFunc(cmd *cobra.Command, args []string) error {
	nickname := args[0]

	exists, err := keys.CheckKey(nickname)
	if exists {
		return err
	}

	fmt.Printf("Type passphrase: ")
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	err = keys.GenerateKey(nickname, passphrase)
	if err != nil {
		return err
	}

	fmt.Printf("\nSuccessfully generated the key with nickname: %s\n", nickname)
	return nil
}

var keyRemoveCmd = &cobra.Command{
	Use:   "remove [nickname]",
	Short: "Remove the specified key",
	Args:  cobra.MinimumNArgs(1),
	RunE:  keyRemoveFunc,
}

func keyRemoveFunc(cmd *cobra.Command, args []string) error {
	nickname := args[0]

	exists, err := keys.CheckKey(nickname)
	if !exists {
		return err
	}

	fmt.Printf("Type passphrase: ")
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	err = keys.RemoveKey(nickname, passphrase)
	if err != nil {
		return err
	}

	fmt.Printf("\nSuccessfully removed the key with nickname: %s\n", nickname)
	return nil
}
