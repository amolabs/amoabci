package parcel

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	client "github.com/amolabs/amoabci/client/parcel"
)

var QueryCmd = &cobra.Command{
	Use:   "query --start <timestamp> --end <timestamp> --owner <address> --qualifier <json>",
	Short: "query data generated between start, end timestamps",
	//Args: cobra.NoArgs,
	Args: cobra.NoArgs,
	RunE: queryFunc,
}

func init() {
	cmd := QueryCmd
	cmd.Flags().SortFlags = false

	cmd.Flags().Uint64P("start", "s", 0, "specify the start timestamp(epoch)")
	cmd.Flags().Uint64P("end", "e", 0, "specify the end timestamp(epoch)")

	cmd.Flags().StringP("owner", "o", "", "owner of the uploaded data")
	cmd.Flags().StringP("qualifier", "q", "", "extra data info")

	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")
}

func queryFunc(cmd *cobra.Command, args []string) error {
	var (
		start, end       uint64
		owner, qualifier string
		err              error
	)

	flags := cmd.Flags()

	start, err = flags.GetUint64("start")
	if err != nil {
		return err
	}

	end, err = flags.GetUint64("end")
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

	result, err := client.Query(start, end, owner, qualifier)
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
