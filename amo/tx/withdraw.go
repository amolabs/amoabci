package tx

import (
	"encoding/json"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type WithdrawParam struct {
	Amount types.Currency `json:"amount"`
}

func parseWithdrawParam(raw []byte) (WithdrawParam, error) {
	var param WithdrawParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

func CheckWithdraw(t Tx) (uint32, string) {
	// TODO: check format
	//txParam, err := parseWithdrawParam(t.Payload)
	_, err := parseWithdrawParam(t.Payload)
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func ExecuteWithdraw(t Tx, store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseWithdrawParam(t.Payload)
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	stake := store.GetStake(t.Sender)
	if stake == nil {
		return code.TxCodeNoStake, "no stake", nil
	}
	if stake.Amount.Sub(&txParam.Amount).Sign() == -1 {
		return code.TxCodeNotEnoughBalance, "not enough stake", nil
	}
	if err := store.SetStake(t.Sender, stake); err != nil {
		switch err {
		case code.TxErrBadParam:
			return code.TxCodeBadParam, err.Error(), nil
		case code.TxErrPermissionDenied:
			return code.TxCodePermissionDenied, err.Error(), nil
		case code.TxErrDelegateExists:
			return code.TxCodeDelegateExists, err.Error(), nil
		case code.TxErrLastValidator:
			return code.TxCodeLastValidator, err.Error(), nil
		default:
			return code.TxCodeUnknown, err.Error(), nil
		}
	}
	balance := store.GetBalance(t.Sender)
	balance.Add(&txParam.Amount)
	store.SetBalance(t.Sender, balance)
	return code.TxCodeOK, "ok", nil
}
