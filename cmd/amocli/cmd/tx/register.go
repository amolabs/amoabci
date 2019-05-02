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
	Use:   "register <parcel_id> <key_custody>",
	Short: "Register parcel with extra information",
	Args:  cobra.MinimumNArgs(2),
	RunE:  registerFunc,
}

func registerFunc(cmd *cobra.Command, args []string) error {
	parcel, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	custody, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Register(parcel, custody, key)
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
