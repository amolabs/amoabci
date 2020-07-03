package tx

import (
	//"bytes"
	//"encoding/binary"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

//// claim

type ClaimParam struct {
	Target   string          `json:"target"`
	Document json.RawMessage `json:"document"`
}

func parseClaimParam(raw []byte) (ClaimParam, error) {
	var param ClaimParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxClaim struct {
	TxBase
	Param ClaimParam `json:"-"`
}

var _ Tx = &TxClaim{}

func (t *TxClaim) Check() (uint32, string) {
	_, err := parseClaimParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxClaim) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseClaimParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	entry := &types.DIDEntry{
		Owner:    t.GetSender(),
		Document: txParam.Document,
	}
	store.SetDIDEntry(txParam.Target, entry)

	return code.TxCodeOK, "ok", []abci.Event{}
}

//// dismiss

type DismissParam struct {
	Target string `json:"target"`
}

func parseDismissParam(raw []byte) (DismissParam, error) {
	var param DismissParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxDismiss struct {
	TxBase
	Param DismissParam `json:"-"`
}

var _ Tx = &TxDismiss{}

func (t *TxDismiss) Check() (uint32, string) {
	_, err := parseDismissParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxDismiss) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseClaimParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	store.DeleteDIDEntry(txParam.Target)

	return code.TxCodeOK, "ok", []abci.Event{}
}
