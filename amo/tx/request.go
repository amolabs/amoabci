package tx

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type RequestParam struct {
	Recipient crypto.Address   `json:"recipient"`
	Target    tmbytes.HexBytes `json:"target"`
	Payment   types.Currency   `json:"payment"`
	Dealer    crypto.Address   `json:"dealer,omitempty"`
	DealerFee types.Currency   `json:"dealer_fee,omitempty"`
	Extra     json.RawMessage  `json:"extra,omitempty"`
}

func parseRequestParam(raw []byte) (RequestParam, error) {
	var param RequestParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxRequest struct {
	TxBase
	Param RequestParam `json:"-"`
}

var _ Tx = &TxRequest{}

func (t *TxRequest) Check() (uint32, string) {
	txParam, err := parseRequestParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	if len(txParam.Recipient) != crypto.AddressSize {
		return code.TxCodeBadParam, "improper recipient address"
	}

	return code.TxCodeOK, "ok"
}

func (t *TxRequest) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseRequestParam(t.getPayload())
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

	requestor := t.GetSender()
	if bytes.Equal(parcel.Owner, txParam.Recipient) {
		// add new code for this
		return code.TxCodeSelfTransaction, "requesting owned parcel", nil
	}

	usage := store.GetUsage(txParam.Recipient, txParam.Target, false)
	if usage != nil {
		return code.TxCodeAlreadyGranted, "parcel already granted", nil
	}

	request := store.GetRequest(txParam.Recipient, txParam.Target, false)
	if request != nil {
		return code.TxCodeAlreadyRequested, "parcel already requested", nil
	}

	storageID := binary.BigEndian.Uint32(txParam.Target[:types.StorageIDLen])
	storage := store.GetStorage(storageID, false)
	if storage == nil || storage.Active == false {
		return code.TxCodeNoStorage, "no active storage for this parcel", nil
	}

	if len(txParam.Dealer) == 0 {
		txParam.DealerFee.Set(0)
	} else if len(txParam.Dealer) != crypto.AddressSize {
		return code.TxCodeBadParam, "invalid dealer address", nil
	}

	balance := store.GetBalance(requestor, false)
	wanted, err := txParam.Payment.Clone()
	if err != nil {
		return code.TxCodeInvalidAmount, err.Error(), nil
	}
	wanted.Add(&txParam.DealerFee)
	if balance.LessThan(wanted) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}

	store.SetRequest(txParam.Recipient, txParam.Target, &types.Request{
		Agency:    requestor,
		Recipient: txParam.Recipient,
		Payment:   txParam.Payment,
		Dealer:    txParam.Dealer,
		DealerFee: txParam.DealerFee,
		Extra: types.Extra{
			Register: parcel.Extra.Register,
			Request:  txParam.Extra,
		},
	})

	balance.Sub(wanted)
	store.SetBalance(requestor, balance)

	return code.TxCodeOK, "ok", []abci.Event{}
}
