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

type IssueParam struct {
	UDC       uint32           `json:"udc"`       // required
	Desc      string           `json:"desc"`      // optional
	Operators []crypto.Address `json:"operators"` // optional
	Amount    types.Currency   `json:"amount"`    // required
}

func parseIssueParam(raw []byte) (IssueParam, error) {
	var param IssueParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxIssue struct {
	TxBase
	Param IssueParam `json:"-"`
}

func (t *TxIssue) Check() (uint32, string) {
	param := t.Param
	for _, op := range param.Operators {
		if len(op) != crypto.AddressSize {
			return code.TxCodeBadParam, "wrong size of operator address"
		}
		if bytes.Equal(t.GetSender(), op) {
			return code.TxCodeSelfTransaction,
				"operator is same as the issuer"
		}
	}
	return code.TxCodeOK, "ok"
}

func (t *TxIssue) Execute(s *store.Store) (uint32, string, []tm.KVPair) {
	param := t.Param
	sender := t.GetSender()

	udc := s.GetUDC(param.UDC, false)
	if udc == nil {
		stakes := s.GetTopStakes(ConfigAMOApp.MaxValidators, sender, false)
		if len(stakes) == 0 {
			return code.TxCodePermissionDenied, "permission denied", nil
		}
		udc = &types.UDC{
			Owner:     sender,
			Operators: param.Operators,
			Desc:      param.Desc,
			Total:     param.Amount,
		}
	} else {
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
		// update fields
		udc.Operators = param.Operators
		udc.Desc = param.Desc
		udc.Total.Add(&param.Amount)
	}
	// update UDC balance
	bal := s.GetUDCBalance(param.UDC, sender, false)
	if bal == nil {
		bal = new(types.Currency)
	}
	after := bal.Add(&param.Amount)
	err := s.SetUDCBalance(param.UDC, sender, after)
	if err != nil {
		return code.TxCodeUnknown, err.Error(), nil
	}
	// store UDC registry
	err = s.SetUDC(param.UDC, udc)
	if err != nil {
		return code.TxCodeUnknown, err.Error(), nil
	}
	return code.TxCodeOK, "ok", nil
}
