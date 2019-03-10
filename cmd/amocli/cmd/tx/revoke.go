package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
)

var RevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke ...",
	Args:  cobra.NoArgs,
	RunE:  revokeFunc,
}

func revokeFunc(cmd *cobra.Command, args []string) error {
	var (
		target, grantee string
		targetHex       []byte
		err             error
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

	result, err := rpc.Revoke(targetHex, granteeAddr, true)
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
	cmd := RevokeCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("target", "t", "", "target ...")
	cmd.Flags().StringP("grantee", "g", "", "grantee ...")

	cmd.MarkFlagRequired("target")
	cmd.MarkFlagRequired("grantee")
}
