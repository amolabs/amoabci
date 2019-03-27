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

var RetractCmd = &cobra.Command{
	Use:   "retract --from <address> --amount <currecncy>",
	Short: "Retract all or part of the AMO coin locked as a delegated stake",
	Args:  cobra.NoArgs,
	RunE:  retractFunc,
}

func init() {
	cmd := RetractCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("from", "f", "", "ex) 63A972C247D1DEBCEF2DDCF5D4E0848A42AFA529")
	cmd.Flags().StringP("amount", "a", "", "actual amount of coin to retract; base 10")

	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("amount")
}

func retractFunc(cmd *cobra.Command, args []string) error {
	var (
		from    string
		balance string
		amount  *atypes.Currency
		err     error
	)

	flags := cmd.Flags()

	from, err = flags.GetString("from")
	if err != nil {
		return err
	}

	balance, err = flags.GetString("amount")
	if err != nil {
		return err
	}

	amount, err = new(atypes.Currency).SetString(balance, 10)
	if err != nil {
		return err
	}

	fromAddr, err := hex.DecodeString(from)
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Retract(fromAddr, amount, key)
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
