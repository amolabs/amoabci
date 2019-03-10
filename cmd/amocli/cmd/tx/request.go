package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/rpc"
)

var RequestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request ...",
	Args:  cobra.NoArgs,
	RunE:  requestFunc,
}

func requestFunc(cmd *cobra.Command, args []string) error {
	var (
		target    string
		payment   atypes.Currency
		targetHex []byte
		tmp       uint64
		err       error
	)

	flags := cmd.Flags()

	if target, err = flags.GetString("target"); err != nil {
		return err
	}

	if targetHex, err = hex.DecodeString(target); err != nil {
		return err
	}

	if tmp, err = flags.GetUint64("payment"); err != nil {
		return err
	}

	payment = atypes.Currency(tmp)

	result, err := rpc.Request(targetHex, &payment, true)
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
	cmd.Flags().Uint64P("payment", "p", 0, "")

	cmd.MarkFlagRequired("target")
	cmd.MarkFlagRequired("payment")
}
