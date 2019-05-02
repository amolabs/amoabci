package parcel

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	client "github.com/amolabs/amoabci/client/parcel"
)

var UploadCmd = &cobra.Command{
	Use:   "upload <hex> --owner <address> --qualifier <json>",
	Short: "Upload data parcel",
	//Args: cobra.NoArgs,
	Args: cobra.MinimumNArgs(1),
	RunE: uploadFunc,
}

func init() {
	cmd := UploadCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().StringP("owner", "o", "", "owner of the uploading data")
	cmd.Flags().StringP("qualifier", "q", "", "extra data info")

	cmd.MarkFlagRequired("owner")
}

func uploadFunc(cmd *cobra.Command, args []string) error {
	var (
		owner, qualifier string
		data             []byte
		err              error
	)

	flags := cmd.Flags()

	data, err = hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	owner, err = flags.GetString("owner")
	if err != nil {
		return err
	}

	qualifier, err = flags.GetString("qualifier")
	if err != nil {
		return err
	}

	result, err := client.Upload(owner, data, qualifier)
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
