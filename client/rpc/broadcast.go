package rpc

import (
	"github.com/amolabs/tendermint-amo/crypto"
	ctypes "github.com/amolabs/tendermint-amo/rpc/core/types"

	"github.com/amolabs/amoabci/amo/operation"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/keys"
)

// Transfer handles transfer transaction
func Transfer(to crypto.Address, amount *atypes.Currency, signer keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	return RPCBroadcastTxCommit(MakeMessage(operation.TxTransfer, signer, 0, operation.Transfer{
		To:     to,
		Amount: *amount,
	}, true))
}
