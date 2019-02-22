package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

/* Commands (expected hierarchy)
 *
 * amoconsole |- key |- list
 *		  			 |- generate <nickname>
 *				 	 |- remove <nickname>
 */

var keyCmd = &cobra.Command{
	Use:     "key",
	Aliases: []string{"k"},
	Short:   "Manages the key(wallet)-related features",
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
	Short: "Shows all of keys stored on the local storage",
	Args:  cobra.NoArgs,
	RunE:  keyListFunc,
}

func keyListFunc(cmd *cobra.Command, args []string) error {
	return nil
}

var keyGenCmd = &cobra.Command{
	Use:   "generate [nickname]",
	Short: "Generates a key with a specified nickname",
	Args:  cobra.MinimumNArgs(1),
	RunE:  keyGenFunc,
}

func keyGenFunc(cmd *cobra.Command, args []string) error {
	var nickname string

	nickname = args[0]

	fmt.Println(nickname)

	return nil
}

var keyRemoveCmd = &cobra.Command{
	Use:   "remove [nickname]",
	Short: "Removes the specified key",
	Args:  cobra.MinimumNArgs(1),
	RunE:  keyRemoveFunc,
}

func keyRemoveFunc(cmd *cobra.Command, args []string) error {
	var nickname string

	nickname = args[0]

	fmt.Println(nickname)

	return nil
}
