package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/amolabs/amoabci/client/util"

	"github.com/spf13/cobra"

	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/keys"
	"github.com/amolabs/amoabci/client/rpc"
)

var TransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer the specified amount of money to <address>",
	Args:  cobra.NoArgs,
	RunE:  transferFunc,
}

func transferFunc(cmd *cobra.Command, args []string) error {
	var (
		key    keys.Key
		to     string
		tmp    uint64
		amount atypes.Currency
		err    error
	)

	flags := cmd.Flags()

	if to, err = flags.GetString("to"); err != nil {
		return err
	}
	if tmp, err = flags.GetUint64("amount"); err != nil {
		return err
	}
	amount = atypes.Currency(tmp)

	// get the key to sign this tx
	key, err = keys.GetKeyToSign(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	if key.Encrypted {
		err = keys.GetDecryptedKey(&key)
		if err != nil {
			return err
		}
	}

	toAddr, err := hex.DecodeString(to)
	if err != nil {
		return err
	}

	result, err := rpc.Transfer(toAddr, &amount, key)
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
	cmd.Flags().Uint64P("amount", "a", 0, "actual amount of coin to transfer")

	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("amount")
}
