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

var RequestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request parcel to purchase with payment as offer amount and extra info",
	Args:  cobra.NoArgs,
	RunE:  requestFunc,
}

func requestFunc(cmd *cobra.Command, args []string) error {
	var (
		target    string
		payment   *atypes.Currency
		targetHex []byte
		balance   string
		err       error
	)

	flags := cmd.Flags()

	if target, err = flags.GetString("target"); err != nil {
		return err
	}

	if targetHex, err = hex.DecodeString(target); err != nil {
		return err
	}

	if balance, err = flags.GetString("payment"); err != nil {
		return err
	}

	payment, err = new(atypes.Currency).SetString(balance, 10)
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Request(targetHex, payment, key)
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
	cmd := RequestCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("target", "t", "", "")
	cmd.Flags().StringP("payment", "p", "", "")

	cmd.MarkFlagRequired("target")
	cmd.MarkFlagRequired("payment")
}
