package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/amolabs/amoabci/client/util"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
)

var CancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel ...",
	Args:  cobra.NoArgs,
	RunE:  cancelFunc,
}

func cancelFunc(cmd *cobra.Command, args []string) error {
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

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Cancel(targetHex, key)
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
	cmd := CancelCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("target", "t", "", "target ...")
	cmd.MarkFlagRequired("target")
}
