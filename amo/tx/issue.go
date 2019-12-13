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
	Id        tm.HexBytes      `json:"id"`        // required
	Operators []crypto.Address `json:"operators"` // optional
	Desc      string           `json:"desc"`      // optional
	Total     types.Currency   `json:"total"`     // required
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

	udc := s.GetUDC(param.Id, false)
	if udc == nil {
		// TODO: check validator permission before creating new UDC
		udc = &types.UDC{
			Id:        param.Id,
			Issuer:    sender,
			Operators: param.Operators,
			Desc:      param.Desc,
			Total:     param.Total,
		}
	} else {
		if bytes.Equal(sender, udc.Issuer) == false {
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
		udc.Total.Add(&param.Total)
	}
	// TODO: update UDC balance
	// save
	err := s.SetUDC(param.Id, udc)
	if err != nil {
		return code.TxCodeUnknown, "failed to save UDC", nil
	}
	return code.TxCodeOK, "ok", nil
}
