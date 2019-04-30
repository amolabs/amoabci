package cmd

import (
	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/cmd/amocli/cmd/key"
)

/* Commands (expected hierarchy)
 *
 * amocli |- key |- list
 *               |- generate <username>
 *               |- import <private key> --username <username>
 *               |- remove <username>
 */

var keyCmd = &cobra.Command{
	Use:     "key",
	Aliases: []string{"k"},
	Short:   "Manage local keyring",
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
