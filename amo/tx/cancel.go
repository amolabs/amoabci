package tx

import (
	"encoding/json"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

type CancelParam struct {
	Target tm.HexBytes `json:"target"`
}

func parseCancelParam(raw []byte) (CancelParam, error) {
	var param CancelParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

func CheckCancel(t Tx) (uint32, string) {
	// TODO: check parcel id format in the future
	//txParam, err := parseCancelParam(t.Payload)
	_, err := parseCancelParam(t.Payload)
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func ExecuteCancel(t Tx, store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseCancelParam(t.Payload)
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	request := store.GetRequest(t.Sender, txParam.Target)
	if request == nil {
		return code.TxCodeRequestNotFound, "request not found", nil
	}
	store.DeleteRequest(t.Sender, txParam.Target)
	balance := store.GetBalance(t.Sender)
	balance.Add(&request.Payment)
	store.SetBalance(t.Sender, balance)
	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(txParam.Target.String())},
	}
	return code.TxCodeOK, "ok", tags
}
