package tx

import (
	"bytes"
	"encoding/json"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

type DiscardParam struct {
	Target tm.HexBytes `json:"target"`
}

func parseDiscardParam(raw []byte) (DiscardParam, error) {
	var param DiscardParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxDiscard struct {
	TxBase
	Param DiscardParam `json:"-"`
}

var _ Tx = &TxDiscard{}

func (t *TxDiscard) Check() (uint32, string) {
	// TOOD: check format
	//txParam, err := parseDiscardParam(t.getPayload())
	_, err := parseDiscardParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxDiscard) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseDiscardParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	parcel := store.GetParcel(txParam.Target, false)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}
	if !bytes.Equal(parcel.Owner, t.GetSender()) {
		return code.TxCodePermissionDenied, "parcel not owned", nil
	}

	store.DeleteParcel(txParam.Target)
	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(txParam.Target.String())},
	}
	return code.TxCodeOK, "ok", tags
}
