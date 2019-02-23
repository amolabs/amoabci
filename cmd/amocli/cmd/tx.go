package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/cmd/amocli/tx"
)

/* Commands (expected hierarchy)
 *
 * amocli |- tx |- transfer --from <address> --to <address> --amount <number>
 *		    	|- purchase --from <address> --file <hash>
 */

var txCmd = &cobra.Command{
	Use:     "tx",
	Aliases: []string{"t"},
	Short:   "Perform a transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}

		return nil
	},
}

var txTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer the specified amount of money from <address> to <address>",
	Args:  cobra.NoArgs,
	RunE:  txTransferFunc,
}

var txPurchaseCmd = &cobra.Command{
	Use:   "purchase",
	Short: "Purchase the file specified with file's <hash>",
	Args:  cobra.NoArgs,
	RunE:  txPurchaseFunc,
}

func init() {
	transferCmd := txTransferCmd
	transferCmd.Flags().SortFlags = false

	transferCmd.Flags().String("from", "", "ex) a8cxVrk1ju91UaJf7U1Hscgn3sRqzfmjgg")
	transferCmd.Flags().String("to", "", "ex) aH2JdDUP5NoFmeEQEqDREZnkmCh8V7co7y")
	transferCmd.Flags().Uint64("amount", 0, "specify 'amount'")
	transferCmd.MarkFlagRequired("from")
	transferCmd.MarkFlagRequired("to")
	transferCmd.MarkFlagRequired("amount")

	purchaseCmd := txPurchaseCmd
	purchaseCmd.Flags().SortFlags = false

	purchaseCmd.Flags().String("from", "", "ex) a8cxVrk1ju91UaJf7U1Hscgn3sRqzfmjgg")
	purchaseCmd.Flags().String("file_hash", "", "ex) 0xb94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9")
	purchaseCmd.MarkFlagRequired("from")
	purchaseCmd.MarkFlagRequired("file_hash")

	cmd := txCmd
	cmd.AddCommand(transferCmd, purchaseCmd)
}

func txTransferFunc(cmd *cobra.Command, args []string) error {
	var from, to string
	var amount uint64
	var err error

	flags := cmd.Flags()

	if from, err = flags.GetString("from"); err != nil {
		return err
	}
	if to, err = flags.GetString("to"); err != nil {
		return err
	}
	if amount, err = flags.GetUint64("amount"); err != nil {
		return err
	}

	fromAddr := atypes.NewAddress([]byte(from))
	toAddr := atypes.NewAddress([]byte(to))

	result, err := tx.Transfer(*fromAddr, *toAddr, &amount)
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

func txPurchaseFunc(cmd *cobra.Command, args []string) error {
	var from, fileHashString string
	var err error

	flags := cmd.Flags()

	if from, err = flags.GetString("from"); err != nil {
		return err
	}
	if fileHashString, err = flags.GetString("file_hash"); err != nil {
		return err
	}

	fileHashString = strings.TrimLeft(fileHashString, "0x")

	fromAddr := atypes.NewAddress([]byte(from))
	fileHash := atypes.NewHashFromHexString(fileHashString)

	result, err := tx.Purchase(*fromAddr, *fileHash)
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
