package query

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
)

var RequestCmd = &cobra.Command{
	Use:   "request --buyer <address> --target <parcelID>",
	Short: "Get buyer's request information regarding to a parcel",
	Args:  cobra.NoArgs,
	RunE:  requestFunc,
}

func requestFunc(cmd *cobra.Command, args []string) error {
	var (
		buyer, target        string
		buyerAddr, targetHex []byte
		err                  error
	)

	flags := cmd.Flags()

	if buyer, err = flags.GetString("buyer"); err != nil {
		return err
	}

	if target, err = flags.GetString("target"); err != nil {
		return err
	}

	buyerAddr, err = hex.DecodeString(buyer)
	if err != nil {
		return err
	}

	targetHex, err = hex.DecodeString(target)
	if err != nil {
		return err
	}

	requestValue, err := rpc.QueryRequest(buyerAddr, targetHex)
	if err != nil {
		return err
	}

	fmt.Println(requestValue)

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
