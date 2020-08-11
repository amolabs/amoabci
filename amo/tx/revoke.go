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

type RevokeParam struct {
	Recipient crypto.Address   `json:"recipient"`
	Target    tmbytes.HexBytes `json:"target"`
}

func parseRevokeParam(raw []byte) (RevokeParam, error) {
	var param RevokeParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxRevoke struct {
	TxBase
	Param RevokeParam `json:"-"`
}

var _ Tx = &TxRevoke{}

func (t *TxRevoke) Check() (uint32, string) {
	txParam, err := parseRevokeParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	// TODO: check format

	if len(txParam.Recipient) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong recipient address"
	}

	return code.TxCodeOK, "ok"
}

// TODO: fix: use GetUsage
func (t *TxRevoke) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseRevokeParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if len(txParam.Recipient) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong recipient address", nil
	}

	revoker := t.GetSender()
	parcel := store.GetParcel(txParam.Target, false)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}
	if !bytes.Equal(parcel.Owner, revoker) &&
		!bytes.Equal(parcel.ProxyAccount, revoker) {
		return code.TxCodePermissionDenied, "permission denied", nil
	}

	usage := store.GetUsage(txParam.Recipient, txParam.Target, false)
	if usage == nil {
		return code.TxCodeUsageNotFound, "usage not found", nil
	}

	store.DeleteUsage(txParam.Recipient, txParam.Target)

	return code.TxCodeOK, "ok", []abci.Event{}
}
