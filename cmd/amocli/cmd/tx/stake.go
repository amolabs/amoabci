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
	Use:   "stake --amount <currency> --validator <ed25519>",
	Short: "Lock AMO coin as a stake of the coin holder",
	Args:  cobra.NoArgs,
	RunE:  stakeFunc,
}

func stakeFunc(cmd *cobra.Command, args []string) error {
	var argAmount, argValidator string
	var err error

	flags := cmd.Flags()

	if argAmount, err = flags.GetString("amount"); err != nil {
		return err
	}
	if argValidator, err = flags.GetString("validator"); err != nil {
		return err
	}

	amount, err := new(atypes.Currency).SetString(argAmount, 10)
	if err != nil {
		return err
	}

	validator, err := base64.StdEncoding.DecodeString(argValidator)
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Stake(amount, validator, key)
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
	cmd := StakeCmd
	cmd.Flags().SortFlags = false
	cmd.Flags().StringP("amount", "a", "", "decimal number")
	cmd.Flags().StringP("validator", "v", "", "base64 encoding of validagtor publie key")

	cmd.MarkFlagRequired("amount")
	cmd.MarkFlagRequired("validator")
}
