package tx

import (
	"bytes"
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type DelegateParam struct {
	To     crypto.Address `json:"to"`
	Amount types.Currency `json:"amount"`
}

func parseDelegateParam(raw []byte) (DelegateParam, error) {
	var param DelegateParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

func CheckDelegate(t Tx) (uint32, string) {
	txParam, err := parseDelegateParam(t.Payload)
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	if len(txParam.To) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong recipient address size"
	}
	if bytes.Equal(txParam.To, t.Sender) {
		return code.TxCodeSelfTransaction, "tried to delegate to self"
	}
	return code.TxCodeOK, "ok"
}

func ExecuteDelegate(t Tx, store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseDelegateParam(t.Payload)
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	balance := store.GetBalance(t.Sender)
	if balance.LessThan(&txParam.Amount) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}
	balance.Sub(&txParam.Amount)

	stake := store.GetStake(txParam.To)
	if stake == nil || stake.Amount.Equals(zero) {
		return code.TxCodeNoStake, "no stake", nil
	}

	delegate := store.GetDelegate(t.Sender)
	if delegate == nil {
		delegate = &types.Delegate{
			Delegatee: txParam.To,
			Amount:    txParam.Amount,
		}
	} else if bytes.Equal(delegate.Delegatee, txParam.To) {
		delegate.Amount.Add(&txParam.Amount)
	} else {
		return code.TxCodeMultipleDelegates, "multiple delegate", nil
	}
	if err := store.SetDelegate(t.Sender, delegate); err != nil {
		switch err {
		case code.TxErrNoStake:
			return code.TxCodeNoStake, err.Error(), nil
		default:
			return code.TxCodeUnknown, err.Error(), nil
		}
	}
	store.SetBalance(t.Sender, balance)
	return code.TxCodeOK, "ok", nil
}
