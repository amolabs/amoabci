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

type TxWithdraw struct {
	TxBase
	Param WithdrawParam `json:"-"`
}

func (t *TxWithdraw) Check() (uint32, string) {
	// TODO: check format
	//txParam, err := parseWithdrawParam(t.Payload)
	_, err := parseWithdrawParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxWithdraw) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseWithdrawParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if txParam.Amount.Equals(zero) || txParam.Amount.LessThan(zero) {
		return code.TxCodeUnavailableAmount, "unavailable amount", nil
	}

	stake := store.GetStake(t.GetSender(), false)
	if stake == nil {
		return code.TxCodeNoStake, "no stake", nil
	}
	// this is just for rich error reporting
	if stake.Amount.Sub(&txParam.Amount).Sign() == -1 {
		return code.TxCodeNotEnoughBalance, "not enough stake", nil
	}
	// total stake for this account is enough for withdrawal, but not unlocked
	// stake.
	unlocked := store.GetUnlockedStake(t.GetSender(), false)
	if unlocked == nil || unlocked.Amount.Sub(&txParam.Amount).Sign() == -1 {
		return code.TxCodeStakeLocked, "stake locked", nil
	}

	if err := store.SetUnlockedStake(t.GetSender(), unlocked); err != nil {
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
	balance := store.GetBalance(t.GetSender(), false)
	balance.Add(&txParam.Amount)
	store.SetBalance(t.GetSender(), balance)
	return code.TxCodeOK, "ok", nil
}
