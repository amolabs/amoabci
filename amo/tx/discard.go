package tx

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

type DiscardParam struct {
	Target tmbytes.HexBytes `json:"target"`
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

func (t *TxDiscard) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseDiscardParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	parcel := store.GetParcel(txParam.Target, false)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}

	if !bytes.Equal(parcel.Owner, t.GetSender()) &&
		!bytes.Equal(parcel.ProxyAccount, t.GetSender()) {
		return code.TxCodePermissionDenied, "permission denied", nil
	}

	parcel.OnSale = false
	store.SetParcel(txParam.Target, parcel)

	events := []abci.Event{
		abci.Event{
			Type: "parcel",
			Attributes: []kv.Pair{
				{Key: []byte("id"), Value: []byte(txParam.Target.String())},
			},
		},
	}

	return code.TxCodeOK, "ok", events
}
