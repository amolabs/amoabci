package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/rpc"
	"github.com/amolabs/amoabci/client/util"
)

var TransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer the specified amount of money to <address>",
	Args:  cobra.NoArgs,
	RunE:  transferFunc,
}

func transferFunc(cmd *cobra.Command, args []string) error {
	var (
		to      string
		balance string
		amount  *atypes.Currency
		err     error
	)

	flags := cmd.Flags()

	if to, err = flags.GetString("to"); err != nil {
		return err
	}
	if balance, err = flags.GetString("amount"); err != nil {
		return err
	}

	amount, err = new(atypes.Currency).SetString(balance, 10)
	if err != nil {
		return err
	}

	toAddr, err := hex.DecodeString(to)
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Transfer(toAddr, amount, key)
	if err != nil {
		return err
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	fmt.Println(string(resultJSON))

	return nil
}

func init() {
	cmd := TransferCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("to", "t", "", "ex) 63A972C247D1DEBCEF2DDCF5D4E0848A42AFA529")
	cmd.Flags().StringP("amount", "a", "", "actual amount of coin to transfer; base 10")

	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("amount")
}
