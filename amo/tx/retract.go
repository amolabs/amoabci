package tx

import (
	"encoding/json"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type RetractParam struct {
	Amount types.Currency `json:"amount"`
}

func parseRetractParam(raw []byte) (RetractParam, error) {
	var param RetractParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxRetract struct {
	TxBase
	Param RetractParam `json:"-"`
}

var _ Tx = &TxRetract{}

func (t *TxRetract) Check() (uint32, string) {
	_, err := parseRetractParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxRetract) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseRetractParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if !txParam.Amount.GreaterThan(zero) {
		return code.TxCodeInvalidAmount, "invalid amount", nil
	}

	delegate := store.GetDelegate(t.GetSender(), false)
	if delegate == nil {
		return code.TxCodeDelegateNotFound, "delegate not found", nil
	}
	if delegate.Amount.LessThan(&txParam.Amount) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}

	delegate.Amount.Sub(&txParam.Amount)
	store.SetDelegate(t.GetSender(), delegate)
	balance := store.GetBalance(t.GetSender(), false)
	balance.Add(&txParam.Amount)
	store.SetBalance(t.GetSender(), balance)
	return code.TxCodeOK, "ok", nil
}
