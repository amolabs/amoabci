package cmd

import (
	"github.com/amolabs/amoabci/cmd/amocli/cmd/tx"
	"github.com/spf13/cobra"
)

/* Commands (expected hierarchy)
 *
 * amocli |- tx |- transfer --from <address> --to <address> --amount <number>
 *		    	|- purchase --from <address> --file <hash>
 */

var txCmd = &cobra.Command{
	Use:     "tx",
	Aliases: []string{"t"},
	Short:   "Perform a transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	txCmd.AddCommand(tx.TransferCmd)
}
