package tx

import (
	ctypes "github.com/amolabs/tendermint-amo/rpc/core/types"
	"github.com/amolabs/tendermint-amo/types"

	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/cmd/amocli/util"
)

// Transfer handles transfer transaction
func Transfer(from, to types.Address, amount *atypes.Currency) (*ctypes.ResultBroadcastTxCommit, error) {
	return util.RPCBroadcastTxCommit(util.MakeMessage(atypes.TxTransfer, atypes.Transfer{
		From:   from,
		To:     to,
		Amount: *amount,
	}))
}

// Purchase handles purchase transaction
func Purchase(from types.Address, fileHash atypes.Hash) (*ctypes.ResultBroadcastTxCommit, error) {
	return util.RPCBroadcastTxCommit(util.MakeMessage(atypes.TxPurchase, atypes.Purchase{
		From:     from,
		FileHash: fileHash,
	}))
}
