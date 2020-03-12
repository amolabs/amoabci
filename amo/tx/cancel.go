package tx

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
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

type TxCancel struct {
	TxBase
	Param CancelParam `json:"-"`
}

var _ Tx = &TxCancel{}

func (t *TxCancel) Check() (uint32, string) {
	// TODO: check parcel id format in the future
	//txParam, err := parseCancelParam(t.getPayload())
	_, err := parseCancelParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxCancel) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseCancelParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	parcel := store.GetParcel(txParam.Target, false)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}

	request := store.GetRequest(t.GetSender(), txParam.Target, false)
	if request == nil {
		return code.TxCodeRequestNotFound, "request not found", nil
	}

	usage := store.GetUsage(t.GetSender(), txParam.Target, false)
	if usage != nil {
		return code.TxCodeAlreadyGranted, "parcel already granted", nil
	}

	store.DeleteRequest(t.GetSender(), txParam.Target)

	balance := store.GetBalance(t.GetSender(), false)
	balance.Add(&request.Payment)
	balance.Add(&request.DealerFee)
	store.SetBalance(t.GetSender(), balance)

	events := []abci.Event{
		abci.Event{
			Type: "parcel",
			Attributes: []tm.KVPair{
				{Key: []byte("id"), Value: []byte(txParam.Target.String())},
			},
		},
	}

	return code.TxCodeOK, "ok", events
}
