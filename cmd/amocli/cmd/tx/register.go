package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
	"github.com/amolabs/amoabci/client/util"
)

var RegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register parcel with extra information",
	Args:  cobra.NoArgs,
	RunE:  registerFunc,
}

func registerFunc(cmd *cobra.Command, args []string) error {
	var (
		target, custody       string
		targetHex, custodyHex []byte
		err                   error
	)

	flags := cmd.Flags()

	if target, err = flags.GetString("target"); err != nil {
		return err
	}

	if targetHex, err = hex.DecodeString(target); err != nil {
		return err
	}

	if custody, err = flags.GetString("custody"); err != nil {
		return err
	}

	if custodyHex, err = hex.DecodeString(custody); err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Register(targetHex, custodyHex, key)
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
	cmd := RegisterCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("target", "t", "", "")
	cmd.Flags().StringP("custody", "c", "", "")

	cmd.MarkFlagRequired("target")
	cmd.MarkFlagRequired("custody")
}
