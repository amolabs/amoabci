package rpc

import (
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/amolabs/amoabci/amo/tx"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/client/keys"
)

// Transfer handles transfer transaction
func Transfer(to crypto.Address, amount *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxTransfer, 0, tx.Transfer{
		To:     to,
		Amount: *amount,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Stake(amount *atypes.Currency, vKey cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxStake, 0, tx.Stake{
		Amount:    *amount,
		Validator: vKey,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Withdraw(amount *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxWithdraw, 0, tx.Withdraw{
		Amount: *amount,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Delegate(to crypto.Address, amount *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxDelegate, 0, tx.Delegate{
		To:     to,
		Amount: *amount,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Retract(amount *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxRetract, 0, tx.Retract{
		Amount: *amount,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Register(target cmn.HexBytes, custody cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxRegister, 0, tx.Register{
		Target:  target,
		Custody: custody,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Request(target cmn.HexBytes, payment *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxRequest, 0, tx.Request{
		Target:  target,
		Payment: *payment,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Cancel(target cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxCancel, 0, tx.Cancel{
		Target: target,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Grant(target cmn.HexBytes, grantee crypto.Address, custody cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxGrant, 0, tx.Grant{
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
	msg, err := MakeMessage(tx.TxRevoke, 0, tx.Revoke{
		Target:  target,
		Grantee: grantee,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}

func Discard(target cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	msg, err := MakeMessage(tx.TxDiscard, 0, tx.Discard{
		Target: target,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(msg)
}
