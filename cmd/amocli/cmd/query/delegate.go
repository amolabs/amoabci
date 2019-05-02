package query

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/rpc"
)

var DelegateCmd = &cobra.Command{
	Use:   "delegate <address>",
	Short: "Delegated stake of an account",
	Args:  cobra.MinimumNArgs(1),
	RunE:  delegateFunc,
}

func delegateFunc(cmd *cobra.Command, args []string) error {
	asJson, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	holderAddr, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	res, err := rpc.QueryDelegate(holderAddr)
	if err != nil {
		return err
	}

	if asJson {
		fmt.Println(string(res))
		return nil
	}

	if res == nil || len(res) == 0 || string(res) == "null" {
		fmt.Printf("no delegate")
	} else {
		var delegate types.Delegate
		err = json.Unmarshal(res, &delegate)
		if err != nil {
			return err
		}
		fmt.Printf("amount: %s,\ndelegator address: %s\n",
			delegate.Amount, delegate.Delegator)
	}

	return nil
}
