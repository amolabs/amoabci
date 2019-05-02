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
	Use:   "transfer <address> <amount>",
	Short: "Transfer the specified amount of money to <address>",
	Args:  cobra.MinimumNArgs(2),
	RunE:  transferFunc,
}

func transferFunc(cmd *cobra.Command, args []string) error {
	recp, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	amount, err := new(atypes.Currency).SetString(args[1], 10)
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Transfer(recp, amount, key)
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
