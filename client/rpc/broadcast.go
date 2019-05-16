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
	txMsg, err := MakeTx(tx.TxTransfer, 0, tx.Transfer{
		To:     to,
		Amount: *amount,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Stake(amount *atypes.Currency, vKey cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxStake, 0, tx.Stake{
		Amount:    *amount,
		Validator: vKey,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Withdraw(amount *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxWithdraw, 0, tx.Withdraw{
		Amount: *amount,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Delegate(to crypto.Address, amount *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxDelegate, 0, tx.Delegate{
		To:     to,
		Amount: *amount,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Retract(amount *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxRetract, 0, tx.Retract{
		Amount: *amount,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Register(target cmn.HexBytes, custody cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxRegister, 0, tx.Register{
		Target:  target,
		Custody: custody,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Request(target cmn.HexBytes, payment *atypes.Currency, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxRequest, 0, tx.Request{
		Target:  target,
		Payment: *payment,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Cancel(target cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxCancel, 0, tx.Cancel{
		Target: target,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Grant(target cmn.HexBytes, grantee crypto.Address, custody cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxGrant, 0, tx.Grant{
		Target:  target,
		Grantee: grantee,
		Custody: custody,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Revoke(target cmn.HexBytes, grantee crypto.Address, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxRevoke, 0, tx.Revoke{
		Target:  target,
		Grantee: grantee,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}

func Discard(target cmn.HexBytes, key keys.Key) (*ctypes.ResultBroadcastTxCommit, error) {
	txMsg, err := MakeTx(tx.TxDiscard, 0, tx.Discard{
		Target: target,
	}, key)

	if err != nil {
		return nil, err
	}

	return RPCBroadcastTxCommit(txMsg)
}
