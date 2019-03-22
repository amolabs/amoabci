package cmd

import (
	"github.com/amolabs/amoabci/cmd/amocli/cmd/tx"
	"github.com/spf13/cobra"
)

/* Commands (expected hierarchy)
 *
 * amocli |- tx |- transfer --to <address> --amount <uint64>
 *				|
 *		    	|- register --target <file> --custody <key>
 *				|- request --target <file> --payment <uint64>
 *				|- cancel --target <file>
 *				|
 *				|- grant --target <file> --grantee <address> --custody <key>
 *				|- revoke --target <file> --grantee <address>
 *				|- discard --target <file>
 */

var txCmd = &cobra.Command{
	Use:     "tx",
	Aliases: []string{"t"},
	Short:   "Perform transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	txCmd.AddCommand(
		tx.TransferCmd,
		LineBreak,
		tx.RegisterCmd,
		tx.RequestCmd,
		tx.GrantCmd,
		LineBreak,
		tx.DiscardCmd,
		tx.CancelCmd,
		tx.RevokeCmd,
	)
}
