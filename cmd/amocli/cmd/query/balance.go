package query

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/rpc"
)

var BalanceCmd = &cobra.Command{
	Use:   "balance <address>",
	Short: "Coin balance of an account",
	Args:  cobra.MinimumNArgs(1),
	RunE:  balanceFunc,
}

func balanceFunc(cmd *cobra.Command, args []string) error {
	asJson, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	address, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	res, err := rpc.QueryBalance(crypto.Address(address))
	if err != nil {
		return err
	}

	if asJson {
		fmt.Println(string(res))
		return nil
	}

	var balance types.Currency
	err = json.Unmarshal(res, &balance)
	if err != nil {
		return err
	}
	fmt.Println(balance, "mote") // TODO: print AMO unit also

	return nil
}
