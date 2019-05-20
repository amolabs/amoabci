package tx

import (
	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Register{}

type Register struct {
	Target  tm.HexBytes `json:"target"`
	Custody tm.HexBytes `json:"custody"`
	// TODO: extra info
}

func (o Register) Check(store *store.Store, sender crypto.Address) uint32 {
	// TODO: permission check from PDSN
	if store.GetParcel(o.Target) != nil {
		return code.TxCodeAlreadyRegistered
	}
	return code.TxCodeOK
}

func (o Register) Execute(store *store.Store, sender crypto.Address) (uint32, []tm.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	parcel := types.ParcelValue{
		Owner:   sender,
		Custody: o.Custody,
	}
	store.SetParcel(o.Target, &parcel)
	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(o.Target.String())},
		{Key: []byte("parcel.owner"), Value: sender},
	}
	return code.TxCodeOK, tags
}
