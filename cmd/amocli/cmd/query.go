package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/cmd/amocli/tx"
)

/* Commands (expected hierarchy)
 *
 * amocli |- query |- address
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

var queryAddressCmd = &cobra.Command{
	Use:   "address [address]",
	Short: "Show general information of specified address",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := atypes.NewAddress([]byte(args[0]))
		targetInfo, err := tx.QueryAddressInfo(*target)
		if err != nil {
			return err
		}

		fmt.Println(string(targetInfo))

		return nil
	},
}

func init() {
	// init here if needed
	addressCmd := queryAddressCmd
	cmd := queryCmd
	cmd.AddCommand(addressCmd)
}
