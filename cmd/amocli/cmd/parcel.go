package cmd

import (
	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/cmd/amocli/cmd/parcel"
)

/* Commands (expected hierarchy)
 *
 * amocli |- parcel |- upload <hex> --owner <address> --qualifier <json>
 *              |- retrieve <parcelID>
 *              |- query --start <timestamp> --end <timestamp> --owner <address> --qualifier <json>
 */

var parcelCmd = &cobra.Command{
	Use:   "parcel",
	Short: "Data parcel operations",
}

func init() {
	parcelCmd.AddCommand(
		parcel.UploadCmd,
		parcel.RetrieveCmd,
		parcel.QueryCmd,
	)
}
