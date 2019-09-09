package tx

import (
	"bytes"
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

type RevokeParam struct {
	Grantee crypto.Address `json:"grantee"`
	Target  tm.HexBytes    `json:"target"`
}

func parseRevokeParam(raw []byte) (RevokeParam, error) {
	var param RevokeParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

// TODO: fix: use GetUsage
func CheckRevoke(t Tx) (uint32, string) {
	txParam, err := parseRevokeParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	// TODO: check format

	if len(txParam.Grantee) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong grantee address size"
	}

	return code.TxCodeOK, "ok"
}

func ExecuteRevoke(t Tx, store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseRevokeParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	parcel := store.GetParcel(txParam.Target)
	if parcel == nil {
		return code.TxCodeParcelNotFound, "parcel not found", nil
	}
	if !bytes.Equal(parcel.Owner, t.GetSender()) {
		return code.TxCodePermissionDenied, "parcel not owned", nil
	}

	store.DeleteUsage(txParam.Grantee, txParam.Target)
	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(txParam.Target.String())},
	}
	return code.TxCodeOK, "ok", tags
}
