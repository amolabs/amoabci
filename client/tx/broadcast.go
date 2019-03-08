package tx

import (
	ctypes "github.com/amolabs/tendermint-amo/rpc/core/types"
	"github.com/amolabs/tendermint-amo/types"

	"github.com/amolabs/amoabci/amo/operation"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/util"
)

// Transfer handles transfer transaction
func Transfer(to types.Address, amount *atypes.Currency) (*ctypes.ResultBroadcastTxCommit, error) {
	return util.RPCBroadcastTxCommit(util.MakeMessage(operation.TxTransfer, operation.Transfer{
		To:     to,
		Amount: *amount,
	}))
}
