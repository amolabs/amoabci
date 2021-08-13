package tx

import (
	//"bytes"
	//"encoding/binary"
	//"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

//// claim

type TxClaimV6 struct {
	TxBase
	Param ClaimParam `json:"-"`
}

var _ Tx = &TxClaimV6{}

func (t *TxClaimV6) Check() (uint32, string) {
	_, err := parseClaimParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxClaimV6) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseClaimParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	entry := &types.DIDEntry{
		Document: txParam.Document,
	}
	store.SetDIDEntry(txParam.Target, entry)

	return code.TxCodeOK, "ok", []abci.Event{}
}

//// dismiss

type TxDismissV6 struct {
	TxBase
	Param DismissParam `json:"-"`
}

var _ Tx = &TxDismissV6{}

func (t *TxDismissV6) Check() (uint32, string) {
	_, err := parseDismissParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxDismissV6) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseClaimParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	store.DeleteDIDEntry(txParam.Target)

	return code.TxCodeOK, "ok", []abci.Event{}
}

