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

type GrantParam struct {
	Target  tm.HexBytes    `json:"target"`
	Grantee crypto.Address `json:"grantee"`
	Custody tm.HexBytes    `json:"custody"`
}

func parseGrantParam(raw []byte) (GrantParam, error) {
	var param GrantParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxGrant struct {
	TxBase
	Param GrantParam `json:"-"`
}

var _ Tx = &TxGrant{}

func (t *TxGrant) Check() (uint32, string) {
	txParam, err := parseGrantParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	// TODO: check format

	if len(txParam.Grantee) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong grantee address size"
	}

	return code.TxCodeOK, "ok"
}

func (t *TxGrant) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseGrantParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	parcel := store.GetParcel(txParam.Target, fromStage)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}
	if !bytes.Equal(parcel.Owner, t.GetSender()) {
		return code.TxCodePermissionDenied, "parcel not owned", nil
	}
	if store.GetUsage(txParam.Grantee, txParam.Target, fromStage) != nil {
		return code.TxCodeAlreadyGranted, "parcel already granted", nil
	}
	request := store.GetRequest(txParam.Grantee, txParam.Target, fromStage)
	if request == nil {
		return code.TxCodeRequestNotFound, "request not found", nil
	}

	store.DeleteRequest(txParam.Grantee, txParam.Target)
	balance := store.GetBalance(t.GetSender(), fromStage)
	balance.Add(&request.Payment)
	store.SetBalance(t.GetSender(), balance)
	usage := types.UsageValue{
		Custody: txParam.Custody,
	}
	store.SetUsage(txParam.Grantee, txParam.Target, &usage)
	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(txParam.Target.String())},
	}
	return code.TxCodeOK, "ok", tags
}
