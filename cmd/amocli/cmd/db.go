package cmd

import (
	"github.com/amolabs/amoabci/cmd/amocli/cmd/db"
	"github.com/spf13/cobra"
)

/* Commands (expected hierarchy)
 *
 * amocli |- db |- upload <hex> --owner <address> --qualifier <json>
 *              |- retrieve <parcelID>
 *              |- query --start <timestamp> --end <timestamp> --owner <address> --qualifier <json>
 */

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Perform database-related operations",
}

func init() {
	dbCmd.AddCommand(
		db.UploadCmd,
		db.RetrieveCmd,
		db.QueryCmd,
	)
}
