package query

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/rpc"
)

var UsageCmd = &cobra.Command{
	Use:   "usage <buyer_address> <parcel_id>",
	Short: "Granted parcel usage",
	Args:  cobra.MinimumNArgs(2),
	RunE:  usageFunc,
}

func usageFunc(cmd *cobra.Command, args []string) error {
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

	res, err := rpc.QueryUsage(buyerAddr, targetHex)
	if err != nil {
		return err
	}

	if asJson {
		fmt.Println(string(res))
		return nil
	}

	var usageValue types.UsageValue
	err = json.Unmarshal(res, &usageValue)
	if err != nil {
		return err
	}
	// fmt.Printf()

	return nil
}

func init() {
	cmd := UsageCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("buyer", "b", "", "buyer ...")
	cmd.Flags().StringP("target", "t", "", "target ...")

	cmd.MarkFlagRequired("buyer")
	cmd.MarkFlagRequired("target")
}
