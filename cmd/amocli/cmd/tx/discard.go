package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
	"github.com/amolabs/amoabci/client/util"
)

var DiscardCmd = &cobra.Command{
	Use:   "discard <parcel_id>",
	Short: "Discard the registered data in store/parcel",
	Args:  cobra.MinimumNArgs(1),
	RunE:  discardFunc,
}

func discardFunc(cmd *cobra.Command, args []string) error {
	parcel, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Discard(parcel, key)
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
