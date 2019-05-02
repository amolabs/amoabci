package tx

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/rpc"
	"github.com/amolabs/amoabci/client/util"
)

var StakeCmd = &cobra.Command{
	Use:   "stake <validator_pubkey> <amount>",
	Short: "Lock AMO coin and acquire a stake with a validator key",
	Args:  cobra.MinimumNArgs(2),
	RunE:  stakeFunc,
}

func stakeFunc(cmd *cobra.Command, args []string) error {
	val, err := base64.StdEncoding.DecodeString(args[0])
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

	result, err := rpc.Stake(amount, val, key)
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
