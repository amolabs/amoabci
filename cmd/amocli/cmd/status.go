package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amolabs/amoabci/client/rpc"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of AMO node",
	Long:  "Show status of AMO node including node info, pubkey, latest block hash, app hash, block height and time",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := rpc.RPCStatus()
		if err != nil {
			return err
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			return err
		}

		fmt.Println(string(resultJSON))

		return nil
	},
}

func init() {
	// init here if needed
}
