package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
	"github.com/amolabs/amoabci/client/util"
)

var GrantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Grant the request of parcel in store/request by data owner",
	Args:  cobra.NoArgs,
	RunE:  grantFunc,
}

func grantFunc(cmd *cobra.Command, args []string) error {
	var (
		target, grantee, custody string
		targetHex, custodyHex    []byte
		err                      error
	)

	flags := cmd.Flags()

	if target, err = flags.GetString("target"); err != nil {
		return err
	}

	if targetHex, err = hex.DecodeString(target); err != nil {
		return err
	}

	if grantee, err = flags.GetString("grantee"); err != nil {
		return err
	}

	granteeAddr, err := hex.DecodeString(grantee)
	if err != nil {
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

	result, err := rpc.Grant(targetHex, granteeAddr, custodyHex, key)
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
	cmd := GrantCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("target", "t", "", "target ...")
	cmd.Flags().StringP("grantee", "g", "", "grantee ...")
	cmd.Flags().StringP("custody", "c", "", "custody ...")

	cmd.MarkFlagRequired("target")
	cmd.MarkFlagRequired("grantee")
	cmd.MarkFlagRequired("custody")
}
