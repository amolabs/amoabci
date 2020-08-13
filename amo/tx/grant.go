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

type GrantParam struct {
	Recipient crypto.Address   `json:"recipient"`
	Target    tmbytes.HexBytes `json:"target"`
	Custody   tmbytes.HexBytes `json:"custody"`
	Extra     json.RawMessage  `json:"extra,omitempty"`
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

	if len(txParam.Recipient) != crypto.AddressSize {
		return code.TxCodeBadParam, "improper recipient address"
	}

	return code.TxCodeOK, "ok"
}

func (t *TxGrant) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseGrantParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if len(txParam.Recipient) != crypto.AddressSize {
		return code.TxCodeBadParam, "improper recipient address", nil
	}

	grantor := t.GetSender()
	parcel := store.GetParcel(txParam.Target, false)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}

	if !bytes.Equal(parcel.Owner, grantor) &&
		!bytes.Equal(parcel.ProxyAccount, grantor) {
		return code.TxCodePermissionDenied, "permission denied", nil
	}

	usage := store.GetUsage(txParam.Recipient, txParam.Target, false)
	if usage != nil {
		return code.TxCodeAlreadyGranted, "parcel already granted", nil
	}

	request := store.GetRequest(txParam.Recipient, txParam.Target, false)
	if request == nil {
		return code.TxCodeRequestNotFound, "parcel not requested", nil
	}

	storageID := binary.BigEndian.Uint32(txParam.Target[:types.StorageIDLen])
	storage := store.GetStorage(storageID, false)
	if storage == nil || storage.Active == false {
		return code.TxCodeNoStorage, "no active storage for this parcel", nil
	}

	balance := store.GetBalance(parcel.Owner, false)
	if balance.Add(&request.Payment).LessThan(&storage.HostingFee) {
		return code.TxCodeNotEnoughBalance,
			"not enough balance for hosting fee", nil
	}

	store.DeleteRequest(txParam.Recipient, txParam.Target)

	store.SetUsage(txParam.Recipient, txParam.Target, &types.Usage{
		Custody: txParam.Custody,
		Extra: types.Extra{
			Register: request.Extra.Register,
			Request:  request.Extra.Request,
			Grant:    txParam.Extra,
		},
	})

	balance = store.GetBalance(parcel.Owner, false)
	balance.Add(&request.Payment).Sub(&storage.HostingFee)
	store.SetBalance(parcel.Owner, balance)
	balance = store.GetBalance(storage.Owner, false)
	balance.Add(&storage.HostingFee)
	store.SetBalance(storage.Owner, balance)
	balance = store.GetBalance(request.Dealer, false)
	balance.Add(&request.DealerFee)
	store.SetBalance(request.Dealer, balance)

	return code.TxCodeOK, "ok", []abci.Event{}
}
