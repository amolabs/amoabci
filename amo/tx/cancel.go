package tx

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

type CancelParam struct {
	Recipient crypto.Address   `json:"recipient"`
	Target    tmbytes.HexBytes `json:"target"`
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
	txParam, err := parseCancelParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	if len(txParam.Recipient) != crypto.AddressSize {
		return code.TxCodeBadParam, "improper recipient address"
	}

	return code.TxCodeOK, "ok"
}

func (t *TxCancel) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseCancelParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if len(txParam.Recipient) != crypto.AddressSize {
		return code.TxCodeBadParam, "improper recipient address", nil
	}

	parcel := store.GetParcel(txParam.Target, false)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}

	canceler := t.GetSender()
	request := store.GetRequest(txParam.Recipient, txParam.Target, false)
	if request == nil {
		return code.TxCodeRequestNotFound, "request not found", nil
	}

	if !bytes.Equal(request.Agency, canceler) {
		return code.TxCodePermissionDenied, "permission denied", nil
	}

	usage := store.GetUsage(txParam.Recipient, txParam.Target, false)
	if usage != nil {
		return code.TxCodeAlreadyGranted, "parcel already granted", nil
	}

	store.DeleteRequest(txParam.Recipient, txParam.Target)

	balance := store.GetBalance(canceler, false)
	balance.Add(&request.Payment)
	balance.Add(&request.DealerFee)
	store.SetBalance(canceler, balance)

	return code.TxCodeOK, "ok", []abci.Event{}
}
