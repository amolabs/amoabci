package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
)

var DiscardCmd = &cobra.Command{
	Use:   "discard",
	Short: "Discard ...",
	Args:  cobra.NoArgs,
	RunE:  discardFunc,
}

func discardFunc(cmd *cobra.Command, args []string) error {
	var (
		target    string
		targetHex []byte
		err       error
	)

	flags := cmd.Flags()

	if target, err = flags.GetString("target"); err != nil {
		return err
	}

	if targetHex, err = hex.DecodeString(target); err != nil {
		return err
	}

	result, err := rpc.Discard(targetHex, true)
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
	cmd := DiscardCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("target", "t", "", "target ...")
	cmd.MarkFlagRequired("target")
}
