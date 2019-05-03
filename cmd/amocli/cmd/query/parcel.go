package query

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/rpc"
)

var ParcelCmd = &cobra.Command{
	Use:   "parcel <parcelID>",
	Short: "Data parcel detail",
	Args:  cobra.MinimumNArgs(1),
	RunE:  parcelFunc,
}

func parcelFunc(cmd *cobra.Command, args []string) error {
	asJson, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	parcelID, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	res, err := rpc.QueryParcel(parcelID)
	if err != nil {
		return err
	}

	if asJson {
		fmt.Println(string(res))
		return nil
	}

	if res == nil || len(res) == 0 || string(res) == "null" {
		fmt.Println("no parcel")
	} else {
		var parcel types.ParcelValue
		err = json.Unmarshal(res, &parcel)
		if err != nil {
			return err
		}
		fmt.Printf("owner: %s\ncustody: %s\n", parcel.Owner, parcel.Custody)
	}

	return nil
}
