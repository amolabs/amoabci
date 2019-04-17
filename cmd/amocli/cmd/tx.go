package cmd

import (
	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/cmd/amocli/cmd/tx"
)

/* Commands (expected hierarchy)
 *
 * amocli |- tx |- transfer --to <address> --amount <uint64>
 *				|
 *              |- stake --amount <currency> --validator <ed25519>
 *              |- withdraw <currency>
 *              |- delegate --to <address> --amount <currency>
 *              |- retract --from <address> --amount <currecncy>
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
	Short:   "Send signed transactions",
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
		tx.StakeCmd,
		tx.WithdrawCmd,
		tx.DelegateCmd,
		tx.RetractCmd,
		LineBreak,
		tx.RegisterCmd,
		tx.RequestCmd,
		tx.GrantCmd,
		LineBreak,
		tx.DiscardCmd,
		tx.CancelCmd,
		tx.RevokeCmd,
	)
	txCmd.PersistentPreRun = readUserPass
	txCmd.PersistentFlags().String("user", "", "username")
	txCmd.PersistentFlags().String("pass", "", "passphrase of an encrypted key")
}

func readUserPass(cmd *cobra.Command, args []string) {
	beep := false
	for c := cmd; c != nil; c = c.Parent() {
		if c == txCmd {
			beep = true
			continue
		}
		run := c.PersistentPreRun
		if beep && run != nil {
			run(cmd, args)
			break
		}
	}

	username, err := cmd.Flags().GetString("user")
	if err == nil {
		tx.Username = username
	}

	passphrase, err := cmd.Flags().GetString("pass")
	if err == nil {
		tx.Passphrase = passphrase
	}
}
