package tx

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type LockParam struct {
	UDC    uint32         `json:"udc"`
	Holder crypto.Address `json:"holder"`
	Amount types.Currency `json:"amount"`
}

func parseLockParam(raw []byte) (LockParam, error) {
	var param LockParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxLock struct {
	TxBase
	Param LockParam `json:"-"`
}

var _ Tx = &TxLock{}

func (t *TxLock) Check() (uint32, string) {
	param := t.Param
	if len(param.Holder) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong size of operator address"
	}
	return code.TxCodeOK, "ok"
}

func (t *TxLock) Execute(s *store.Store) (uint32, string, []abci.Event) {
	param := t.Param
	sender := t.GetSender()

	if !param.Amount.GreaterThan(zero) {
		return code.TxCodeInvalidAmount, "invalid amount", nil
	}

	udc := s.GetUDC(param.UDC, false)
	if udc == nil {
		return code.TxCodeUDCNotFound, "UDC not found", nil
	}

	if bytes.Equal(sender, udc.Owner) == false {
		match := false
		for _, op := range udc.Operators {
			if bytes.Equal(sender, op) {
				match = true
				break
			}
		}
		if match == false {
			return code.TxCodePermissionDenied, "permission denied", nil
		}
	}

	err := s.SetUDCLock(param.UDC, param.Holder, &param.Amount)
	if err != nil {
		return code.TxCodeUnknown, "error setting internal db", nil
	}

	return code.TxCodeOK, "ok", nil
}
