package cmd

import (
	"github.com/spf13/cobra"
)

/* Commands (expected hierarchy)
 *
 * amocli |- db |- upload
 *              |- retrieve
 *              |- query
 */

var pdbCmd = &cobra.Command{
	Use:   "db",
	Short: "Perform database-related operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	pdbCmd.AddCommand()
}
