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
	Recipient crypto.Address   `json:"recipient,omitempty"`
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

	rpkSize := len(txParam.Recipient)
	if rpkSize != 0 && rpkSize != crypto.AddressSize {
		return code.TxCodeBadParam, "improper recipient address"
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

	var (
		requestor crypto.Address   = t.GetSender()
		canceler  crypto.Address   = t.GetSender()
		recipient crypto.Address   = t.GetSender()
		target    tmbytes.HexBytes = txParam.Target
	)

	rpkSize := len(txParam.Recipient)
	if rpkSize != 0 {
		if rpkSize != crypto.AddressSize {
			return code.TxCodeBadParam, "improper recipient address", nil
		}
		recipient = txParam.Recipient
	}

	request := store.GetRequest(recipient, target, false)
	if request == nil {
		return code.TxCodeRequestNotFound, "request not found", nil
	}

	// permission check
	if rpkSize != 0 {
		requestor = request.Agency
	}
	if !bytes.Equal(requestor, canceler) {
		return code.TxCodePermissionDenied, "permission denied", nil
	}

	usage := store.GetUsage(recipient, target, false)
	if usage != nil {
		return code.TxCodeAlreadyGranted, "parcel already granted", nil
	}

	store.DeleteRequest(recipient, target)

	balance := store.GetBalance(canceler, false)
	balance.Add(&request.Payment)
	balance.Add(&request.DealerFee)
	store.SetBalance(canceler, balance)

	return code.TxCodeOK, "ok", []abci.Event{}
}
