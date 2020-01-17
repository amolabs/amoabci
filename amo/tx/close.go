package tx

import (
	"bytes"
	"encoding/json"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

type CloseParam struct {
	Storage tm.HexBytes `json:"storage"`
}

func parseCloseParam(raw []byte) (CloseParam, error) {
	var param CloseParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxClose struct {
	TxBase
	Param CloseParam `json:"-"`
}

var _ Tx = &TxClose{}

func (t *TxClose) Check() (uint32, string) {
	// TOOD: check url format
	_, err := parseCloseParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxClose) Execute(s *store.Store) (uint32, string, []tm.KVPair) {
	param := t.Param
	sender := t.GetSender()

	sto := s.GetStorage(param.Storage, false)
	if sto == nil {
		return code.TxCodeNotFound, "not found", nil
	} else {
		if bytes.Equal(sender, sto.Owner) == false {
			return code.TxCodePermissionDenied, "permission denied", nil
		}
		// update fields
		sto.Active = false
	}
	// store
	err := s.SetStorage(param.Storage, sto)
	if err != nil {
		return code.TxCodeUnknown, err.Error(), nil
	}
	return code.TxCodeOK, "ok", nil
}
