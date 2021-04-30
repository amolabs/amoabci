package tx

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type TransferParamV5 struct {
	To     crypto.Address   `json:"to"`
	UDC    uint32           `json:"udc,omitempty"`
	Amount types.Currency   `json:"amount,omitempty"`
	Parcel tmbytes.HexBytes `json:"parcel,omitempty"`
}

func parseTransferParamV5(raw []byte) (TransferParamV5, error) {
	var param TransferParamV5
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxTransferV5 struct {
	TxBase
	Param TransferParamV5 `json:"-"`
}

var _ Tx = &TxTransferV5{}

func (t *TxTransferV5) Check() (uint32, string) {
	txParam, err := parseTransferParamV5(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	if len(txParam.To) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong recipient address size"
	}
	if bytes.Equal(t.GetSender(), txParam.To) {
		return code.TxCodeSelfTransaction, "tried to transfer to self"
	}
	if len(txParam.Parcel) == 0 && txParam.Amount.Equals(zero) {
		return code.TxCodeBadParam, "either parcel or coin should be specifed"
	}

	return code.TxCodeOK, "ok"
}

func (t *TxTransferV5) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseTransferParamV5(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if len(txParam.Parcel) > 0 {
		return t.TransferParcel(store, txParam)
	} else if txParam.Amount.GreaterThan(zero) {
		return t.TransferCoin(store, txParam)
	} else {
		return code.TxCodeBadParam,
			"either parcel or coin should be specifed",
			nil
	}
}

func (t *TxTransferV5) TransferCoin(store *store.Store, txParam TransferParamV5) (uint32, string, []abci.Event) {
	udc := txParam.UDC
	udcLock := store.GetUDCLock(udc, t.GetSender(), false)
	fromBalance := store.GetUDCBalance(udc, t.GetSender(), false)
	required := udcLock
	required.Add(&txParam.Amount)
	if fromBalance.LessThan(required) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}
	toBalance := store.GetUDCBalance(txParam.UDC, txParam.To, false)
	fromBalance.Sub(&txParam.Amount)
	toBalance.Add(&txParam.Amount)
	store.SetUDCBalance(txParam.UDC, t.GetSender(), fromBalance)
	store.SetUDCBalance(txParam.UDC, txParam.To, toBalance)
	return code.TxCodeOK, "ok", nil
}

func (t *TxTransferV5) TransferParcel(store *store.Store, txParam TransferParamV5) (uint32, string, []abci.Event) {
	parcel := store.GetParcel(txParam.Parcel, false)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}
	sender := t.GetSender()
	if !bytes.Equal(sender, parcel.Owner) {
		return code.TxCodePermissionDenied, "permission denied", nil
	}
	parcel.Owner = txParam.To
	store.SetParcel(txParam.Parcel, parcel)
	return code.TxCodeOK, "ok", nil
}
