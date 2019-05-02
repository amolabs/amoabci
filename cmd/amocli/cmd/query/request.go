package query

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/rpc"
)

var RequestCmd = &cobra.Command{
	Use:   "request <buyer_address> <parcel_id>",
	Short: "Requested parcel usage",
	Args:  cobra.MinimumNArgs(2),
	RunE:  requestFunc,
}

func requestFunc(cmd *cobra.Command, args []string) error {
	asJson, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	buyerAddr, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	targetHex, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}

	res, err := rpc.QueryRequest(buyerAddr, targetHex)
	if err != nil {
		return err
	}

	if asJson {
		fmt.Println(string(res))
		return nil
	}

	var requestValue types.RequestValue
	err = json.Unmarshal(res, &requestValue)
	if err != nil {
		return err
	}
	// fmt.Printf()

	return nil
}

func init() {
	cmd := RequestCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("buyer", "b", "", "buyer ...")
	cmd.Flags().StringP("target", "t", "", "target ...")

	cmd.MarkFlagRequired("buyer")
	cmd.MarkFlagRequired("target")
}
