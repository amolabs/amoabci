package parcel

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	client "github.com/amolabs/amoabci/client/parcel"
)

var RetrieveCmd = &cobra.Command{
	Use:   "retrieve <parcelID>",
	Short: "Retrieve data parcel with parcelID",
	//Args: cobra.NoArgs,
	Args: cobra.MinimumNArgs(1),
	RunE: retrieveFunc,
}

func retrieveFunc(cmd *cobra.Command, args []string) error {
	var (
		parcelID []byte
		err      error
	)

	parcelID, err = hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	result, err := client.Retrieve(parcelID)
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
