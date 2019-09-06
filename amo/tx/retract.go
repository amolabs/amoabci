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

func CheckRetract(t Tx) (uint32, string) {
	_, err := parseRetractParam(t.Payload)
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func ExecuteRetract(t Tx, store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseRetractParam(t.Payload)
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	delegate := store.GetDelegate(t.Sender)
	if delegate == nil {
		return code.TxCodeDelegateNotFound, "delegate not found", nil
	}
	if delegate.Amount.LessThan(&txParam.Amount) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}

	delegate.Amount.Sub(&txParam.Amount)
	store.SetDelegate(t.Sender, delegate)
	balance := store.GetBalance(t.Sender)
	balance.Add(&txParam.Amount)
	store.SetBalance(t.Sender, balance)
	return code.TxCodeOK, "ok", nil
}
