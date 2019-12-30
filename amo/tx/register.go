package tx

import (
	"encoding/json"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type RegisterParam struct {
	Target       tm.HexBytes `json:"target"`
	Custody      tm.HexBytes `json:"custody"`
	ProxyAccount tm.HexBytes `json:"proxy_account"`

	Extra json.RawMessage `json:"extra"`
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
	// TOOD: check format
	//txParam, err := parseRegisterParam(t.getPayload())
	_, err := parseRegisterParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxRegister) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseRegisterParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	parcel := store.GetParcel(txParam.Target, false)
	if parcel != nil {
		return code.TxCodeAlreadyRegistered, "parcel already registered", nil
	}

	store.SetParcel(txParam.Target, &types.ParcelValue{
		Owner:        t.GetSender(),
		Custody:      txParam.Custody,
		ProxyAccount: txParam.ProxyAccount,

		Extra: types.Extra{
			Register: txParam.Extra,
		},
	})

	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(txParam.Target.String())},
		{Key: []byte("parcel.owner"), Value: t.GetSender()},
	}

	return code.TxCodeOK, "ok", tags
}
