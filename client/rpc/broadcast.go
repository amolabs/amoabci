package rpc

import (
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/amolabs/amoabci/amo/operation"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/keys"
)

// Transfer handles transfer transaction
func Transfer(to crypto.Address, amount *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxTransfer, 0, operation.Transfer{
		To:     to,
		Amount: *amount,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Register(target cmn.HexBytes, custody cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxRegister, 0, operation.Register{
		Target:  target,
		Custody: custody,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Request(target cmn.HexBytes, payment *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxRequest, 0, operation.Request{
		Target:  target,
		Payment: *payment,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Cancel(target cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxCancel, 0, operation.Cancel{
		Target: target,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Grant(target cmn.HexBytes, grantee crypto.Address, custody cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxGrant, 0, operation.Grant{
		Target:  target,
		Grantee: grantee,
		Custody: custody,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Revoke(target cmn.HexBytes, grantee crypto.Address, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxRevoke, 0, operation.Revoke{
		Target:  target,
		Grantee: grantee,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Discard(target cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxDiscard, 0, operation.Discard{
		Target: target,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}
