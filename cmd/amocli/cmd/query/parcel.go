package query

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
)

var ParcelCmd = &cobra.Command{
	Use:   "parcel <parcelID>",
	Short: "Get parcel's extra informations",
	Args:  cobra.MinimumNArgs(1),
	RunE:  parcelFunc,
}

func parcelFunc(cmd *cobra.Command, args []string) error {
	tmp := strings.TrimLeft(args[0], "0x")
	parcelID, err := hex.DecodeString(tmp)
	if err != nil {
		return err
	}

	parcelValue, err := rpc.QueryParcel(parcelID)
	if err != nil {
		return err
	}

	fmt.Println(parcelValue)

	return nil
}
