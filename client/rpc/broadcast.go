package rpc

import (
	"github.com/amolabs/tendermint-amo/crypto"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	ctypes "github.com/amolabs/tendermint-amo/rpc/core/types"

	"github.com/amolabs/amoabci/amo/operation"
	atypes "github.com/amolabs/amoabci/amo/types"
)

// Transfer handles transfer transaction
func Transfer(to crypto.Address, amount *atypes.Currency, sign bool) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxTransfer, 0, operation.Transfer{
		To:     to,
		Amount: *amount,
	}, sign)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Register(target cmn.HexBytes, custody cmn.HexBytes, sign bool) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxRegister, 0, operation.Register{
		Target:  target,
		Custody: custody,
	}, sign)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Request(target cmn.HexBytes, payment *atypes.Currency, sign bool) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxRequest, 0, operation.Request{
		Target:  target,
		Payment: *payment,
	}, sign)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Cancel(target cmn.HexBytes, sign bool) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxCancel, 0, operation.Cancel{
		Target: target,
	}, sign)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Grant(target cmn.HexBytes, grantee crypto.Address, custody cmn.HexBytes, sign bool) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxGrant, 0, operation.Grant{
		Target:  target,
		Grantee: grantee,
		Custody: custody,
	}, sign)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Revoke(target cmn.HexBytes, grantee crypto.Address, sign bool) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxRevoke, 0, operation.Revoke{
		Target:  target,
		Grantee: grantee,
	}, sign)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Discard(target cmn.HexBytes, sign bool) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(operation.TxDiscard, 0, operation.Discard{
		Target: target,
	}, sign)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}
