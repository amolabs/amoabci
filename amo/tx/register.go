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

type RegisterParam struct {
	Target       tmbytes.HexBytes `json:"target"`
	Custody      tmbytes.HexBytes `json:"custody"`
	ProxyAccount tmbytes.HexBytes `json:"proxy_account,omitempty"`
	Extra        json.RawMessage  `json:"extra,omitempty"`
}

func parseRegisterParam(raw []byte) (RegisterParam, error) {
	var param RegisterParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxRegister struct {
	TxBase
	Param RegisterParam `json:"-"`
}

var _ Tx = &TxRegister{}

func (t *TxRegister) Check() (uint32, string) {
	txParam, err := parseRegisterParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}
	// XXX: If len(txParam.Target) == types.StorageIDLen, then there is no room
	// for in-storage ID for a parcel. Invalid parcel ID in that case.
	if len(txParam.Target) <= types.StorageIDLen {
		return code.TxCodeBadParam, "parcel id too short"
	}
	if len(txParam.ProxyAccount) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong proxy account address size"
	}
	return code.TxCodeOK, "ok"
}

func (t *TxRegister) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseRegisterParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	storageID := binary.BigEndian.Uint32(txParam.Target[:types.StorageIDLen])
	storage := store.GetStorage(storageID, false)
	if storage == nil || storage.Active == false {
		return code.TxCodeNoStorage, "no active storage for this parcel", nil
	}

	sender := t.GetSender()
	parcel := store.GetParcel(txParam.Target, false)

	if parcel == nil {
		if store.GetBalance(sender, false).LessThan(&storage.RegistrationFee) {
			return code.TxCodeNotEnoughBalance, "not enough balance for registration fee", nil
		}

		balance := store.GetBalance(sender, false)
		balance.Sub(&storage.RegistrationFee)
		store.SetBalance(sender, balance)
		balance = store.GetBalance(storage.Owner, false)
		balance.Add(&storage.RegistrationFee)
		store.SetBalance(storage.Owner, balance)
	} else {
		if !bytes.Equal(sender, parcel.Owner) &&
			!bytes.Equal(sender, parcel.ProxyAccount) {
			return code.TxCodePermissionDenied, "permission denied", nil
		}
	}

	store.SetParcel(txParam.Target, &types.Parcel{
		Owner:        sender,
		Custody:      txParam.Custody,
		ProxyAccount: txParam.ProxyAccount,
		Extra: types.Extra{
			Register: txParam.Extra,
		},
		OnSale: true,
	})

	return code.TxCodeOK, "ok", []abci.Event{}
}
