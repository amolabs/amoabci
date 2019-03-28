package query

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
)

var DelegateCmd = &cobra.Command{
	Use:   "delegate --holder <address> --delegator <address>",
	Short: "Get informations regarding to delegation of stake",
	Args:  cobra.NoArgs,
	RunE:  delegateFunc,
}

func delegateFunc(cmd *cobra.Command, args []string) error {
	var (
		holder, delegator         string
		holderAddr, delegatorAddr []byte
		err                       error
	)

	flags := cmd.Flags()

	if holder, err = flags.GetString("holder"); err != nil {
		return err
	}

	if delegator, err = flags.GetString("delegator"); err != nil {
		return err
	}

	holderAddr, err = hex.DecodeString(holder)
	if err != nil {
		return err
	}

	delegatorAddr, err = hex.DecodeString(delegator)
	if err != nil {
		return err
	}

	amount, err := rpc.QueryDelegate(holderAddr, delegatorAddr)
	if err != nil {
		return err
	}

	fmt.Println(amount)

	return nil
}

func init() {
	cmd := DelegateCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("holder", "f", "", "holder ...")
	cmd.Flags().StringP("delegator", "t", "", "delegator ...")

	cmd.MarkFlagRequired("holder")
	cmd.MarkFlagRequired("delegator")
}
