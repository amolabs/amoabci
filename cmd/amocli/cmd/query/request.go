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

	buyer, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	parcel, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}

	res, err := rpc.QueryRequest(buyer, parcel)
	if err != nil {
		return err
	}

	if asJson {
		fmt.Println(string(res))
		return nil
	}

	if res == nil || len(res) == 0 || string(res) == "null" {
		fmt.Println("no request")
	} else {
		var request types.RequestValue
		err = json.Unmarshal(res, &request)
		if err != nil {
			return err
		}
		fmt.Printf("payment: %s\nexpire: %s\n", request.Payment, request.Exp)
	}

	return nil
}
