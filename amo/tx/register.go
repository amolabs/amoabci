package tx

import (
	"encoding/json"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type RegisterParam struct {
	Target  tm.HexBytes `json:"target"`
	Custody tm.HexBytes `json:"custody"`
	// TODO: extra info
}

func parseRegisterParam(raw []byte) (RegisterParam, error) {
	var param RegisterParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

func CheckRegister(t Tx) (uint32, string) {
	// TOOD: check format
	//txParam, err := parseRegisterParam(t.getPayload())
	_, err := parseRegisterParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func ExecuteRegister(t Tx, store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseRegisterParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if store.GetParcel(txParam.Target) != nil {
		return code.TxCodeAlreadyRegistered, "parcel already registered", nil
	}

	parcel := types.ParcelValue{
		Owner:   t.GetSender(),
		Custody: txParam.Custody,
	}
	store.SetParcel(txParam.Target, &parcel)
	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(txParam.Target.String())},
		{Key: []byte("parcel.owner"), Value: t.GetSender()},
	}
	return code.TxCodeOK, "ok", tags
}
