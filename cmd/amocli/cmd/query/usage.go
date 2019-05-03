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

	if res == nil || len(res) == 0 || string(res) == "null" {
		fmt.Printf("no usage")
	} else {
		var usage types.UsageValue
		err = json.Unmarshal(res, &usage)
		if err != nil {
			return err
		}
		fmt.Printf("custody: %s\nexpire: %s", usage.Custody, usage.Exp)
	}

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
