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

type RequestParam struct {
	Target    tm.HexBytes     `json:"target"`
	Payment   types.Currency  `json:"payment"`
	Dealer    crypto.Address  `json:"dealer,omitempty"`
	DealerFee types.Currency  `json:"dealer_fee,omitempty"`
	Extra     json.RawMessage `json:"extra"`
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
	// TOOD: check format
	//txParam, err := parseRequestParam(t.Payload)
	_, err := parseRequestParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxRequest) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseRequestParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	parcel := store.GetParcel(txParam.Target, false)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}

	if bytes.Equal(parcel.Owner, t.GetSender()) {
		// add new code for this
		return code.TxCodeSelfTransaction, "requesting owned parcel", nil
	}

	usage := store.GetUsage(t.GetSender(), txParam.Target, false)
	if usage != nil {
		return code.TxCodeAlreadyGranted, "parcel already granted", nil
	}

	request := store.GetRequest(t.GetSender(), txParam.Target, false)
	if request != nil {
		return code.TxCodeAlreadyRequested, "parcel already requested", nil
	}

	if len(txParam.Dealer) == 0 {
		txParam.DealerFee.Set(0)
	} else if len(txParam.Dealer) != crypto.AddressSize {
		return code.TxCodeBadParam, "invalid dealer address", nil
	}

	balance := store.GetBalance(t.GetSender(), false)
	wanted := txParam.Payment
	wanted.Add(&txParam.DealerFee)
	if balance.LessThan(&wanted) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}

	store.SetRequest(t.GetSender(), txParam.Target, &types.Request{
		Payment:   txParam.Payment,
		Dealer:    txParam.Dealer,
		DealerFee: txParam.DealerFee,
		Extra: types.Extra{
			Register: parcel.Extra.Register,
			Request:  txParam.Extra,
		},
	})

	balance.Sub(&wanted)
	store.SetBalance(t.GetSender(), balance)

	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(txParam.Target.String())},
	}

	return code.TxCodeOK, "ok", tags
}
