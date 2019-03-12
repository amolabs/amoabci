package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/client/rpc"
)

/* Commands (expected hierarchy)
 *
 * amocli |- query |- balance
 */

var queryCmd = &cobra.Command{
	Use:     "query",
	Aliases: []string{"q"},
	Short:   "Performs the query specified by users",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}

		return nil
	},
}

var queryBalanceCmd = &cobra.Command{
	Use:   "balance [address]",
	Short: "Show balance of an address",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		address, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		balance, err := rpc.QueryBalance(crypto.Address(address))
		if err != nil {
			return err
		}

		fmt.Println(balance)

		return nil
	},
}

func init() {
	// init here if needed
	addressCmd := queryBalanceCmd
	cmd := queryCmd
	cmd.AddCommand(addressCmd)
}
