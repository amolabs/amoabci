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

type TxDelegate struct {
	TxBase
	Param DelegateParam `json:"-"`
}

var _ Tx = &TxDelegate{}

func (t *TxDelegate) Check() (uint32, string) {
	txParam, err := parseDelegateParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	if len(txParam.To) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong recipient address size"
	}
	if bytes.Equal(txParam.To, t.GetSender()) {
		return code.TxCodeSelfTransaction, "tried to delegate to self"
	}
	return code.TxCodeOK, "ok"
}

func (t *TxDelegate) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseDelegateParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if txParam.Amount.Equals(zero) || txParam.Amount.LessThan(zero) {
		return code.TxCodeUnavailableAmount, "unavailable amount", nil
	}

	balance := store.GetBalance(t.GetSender(), false)
	if balance.LessThan(&txParam.Amount) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}
	balance.Sub(&txParam.Amount)

	stake := store.GetStake(txParam.To, false)
	if stake == nil || stake.Amount.Equals(types.Zero) {
		return code.TxCodeNoStake, "no stake", nil
	}

	delegate := store.GetDelegate(t.GetSender(), false)
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
	if err := store.SetDelegate(t.GetSender(), delegate); err != nil {
		switch err {
		case code.TxErrNoStake:
			return code.TxCodeNoStake, err.Error(), nil
		default:
			return code.TxCodeUnknown, err.Error(), nil
		}
	}
	store.SetBalance(t.GetSender(), balance)
	return code.TxCodeOK, "ok", nil
}
