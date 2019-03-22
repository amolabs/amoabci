package cmd

import (
	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/cmd/amocli/cmd/key"
)

/* Commands (expected hierarchy)
 *
 * amocli |- key |- list
 *               |- generate <nickname>
 *               |- import <private key> --nickname <nickname>
 *               |- remove <nickname>
 */

var keyCmd = &cobra.Command{
	Use:     "key",
	Aliases: []string{"k"},
	Short:   "Manage the key(wallet)-related features",
}

func init() {
	cmd := keyCmd

	cmd.AddCommand(
		key.ListCmd,
		key.ImportCmd,
		key.GenCmd,
		key.RemoveCmd,
	)
}
