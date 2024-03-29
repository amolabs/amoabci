package tx

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type BurnParam struct {
	UDC    uint32         `json:"udc"`
	Amount types.Currency `json:"amount"`
}

func parseBurnParam(raw []byte) (BurnParam, error) {
	var param BurnParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxBurn struct {
	TxBase
	Param BurnParam `json:"-"`
}

var _ Tx = &TxBurn{}

func (t *TxBurn) Check() (uint32, string) {
	_, err := parseTransferParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}
	return code.TxCodeOK, "ok"
}

func (t *TxBurn) Execute(s *store.Store) (uint32, string, []abci.Event) {
	param := t.Param

	if !param.Amount.GreaterThan(zero) {
		return code.TxCodeInvalidAmount, "invalid amount", nil
	}

	udc := s.GetUDC(param.UDC, false)
	if udc == nil {
		return code.TxCodeUDCNotFound, "UDC not found", nil
	}

	if !param.Amount.GreaterThan(zero) {
		return code.TxCodeInvalidAmount, "invalid amount", nil
	}

	udcLock := s.GetUDCLock(param.UDC, t.GetSender(), false)
	balance := s.GetUDCBalance(param.UDC, t.GetSender(), false)
	required := udcLock
	required.Add(&param.Amount)
	if balance.LessThan(required) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}
	balance.Sub(&param.Amount)
	s.SetUDCBalance(param.UDC, t.GetSender(), balance)
	udc.Total.Sub(&param.Amount)
	s.SetUDC(param.UDC, udc)

	return code.TxCodeOK, "ok", nil
}
