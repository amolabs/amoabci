package cmd

import (
	"github.com/amolabs/amoabci/cmd/amocli/cmd/query"
	"github.com/spf13/cobra"
)

/* Commands (expected hierarchy)
 *
 * amocli |- query |- balance <address>
 *				   |
*				   |- parcel <parcelID>
*				   |- request --buyer <address> --target <parcelID>
*				   |- usage --buyer <address> --target <parcelID>
*/

var queryCmd = &cobra.Command{
	Use:     "query",
	Aliases: []string{"q"},
	Short:   "Perform queries specified by users",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	queryCmd.AddCommand(
		query.BalanceCmd,
		LineBreak,
		query.ParcelCmd,
		query.RequestCmd,
		query.UsageCmd,
	)
}
