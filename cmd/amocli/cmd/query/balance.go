package query

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/client/rpc"
)

var BalanceCmd = &cobra.Command{
	Use:   "balance [address]",
	Short: "Show balance of an address",
	Args:  cobra.MinimumNArgs(1),
	RunE:  balanceFunc,
}

func balanceFunc(cmd *cobra.Command, args []string) error {
	address, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	balance, err := rpc.QueryBalance(crypto.Address(address))
	if err != nil {
		return err
	}

	fmt.Println(balance)

	return nil
}
