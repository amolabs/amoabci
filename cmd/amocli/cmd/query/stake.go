package query

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
)

var StakeCmd = &cobra.Command{
	Use:   "stake <address>",
	Short: "Get stake of an address",
	Args:  cobra.MinimumNArgs(1),
	RunE:  stakeFunc,
}

func stakeFunc(cmd *cobra.Command, args []string) error {
	address, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	stake, err := rpc.QueryStake(address)
	if err != nil {
		return err
	}

	fmt.Printf("amount: %s, validator: %s\n", stake.Amount, stake.Validator)

	return nil
}
