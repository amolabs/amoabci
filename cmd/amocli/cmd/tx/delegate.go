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

var DelegateCmd = &cobra.Command{
	Use:   "delegate <address> <amount>",
	Short: "Lock sender's AMO coin as a delegated stake of the delegator",
	Args:  cobra.MinimumNArgs(2),
	RunE:  delegateFunc,
}

func delegateFunc(cmd *cobra.Command, args []string) error {
	amount, err := new(atypes.Currency).SetString(args[1], 10)
	if err != nil {
		return err
	}

	delegator, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Delegate(delegator, amount, key)
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
