package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
	"github.com/amolabs/amoabci/client/util"
)

var RevokeCmd = &cobra.Command{
	Use:   "revoke <parcel_id> <address>",
	Short: "Delete the usage of parcel in store/usage",
	Args:  cobra.MinimumNArgs(2),
	RunE:  revokeFunc,
}

func revokeFunc(cmd *cobra.Command, args []string) error {
	parcel, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	grantee, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}

	key, err := GetRawKey(util.DefaultKeyFilePath())
	if err != nil {
		return err
	}

	result, err := rpc.Revoke(parcel, grantee, key)
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
