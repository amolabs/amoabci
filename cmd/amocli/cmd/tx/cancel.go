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
	Use:   "cancel <parcel_id>",
	Short: "Cancel the request of parcel in store/request",
	Args:  cobra.MinimumNArgs(1),
	RunE:  cancelFunc,
}

func cancelFunc(cmd *cobra.Command, args []string) error {
	parcel, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Cancel(parcel, key)
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
