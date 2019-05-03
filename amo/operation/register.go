package operation

import (
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Register{}

type Register struct {
	Target  cmn.HexBytes `json:"target"`
	Custody cmn.HexBytes `json:"custody"`
	// TODO: extra info
}

func (o Register) Check(store *store.Store, sender crypto.Address) uint32 {
	// TODO: permission check from PDSN
	if store.GetParcel(o.Target) != nil {
		return code.TxCodeAlreadyRegistered
	}
	return code.TxCodeOK
}

func (o Register) Execute(store *store.Store, sender crypto.Address) uint32 {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode
	}
	parcel := types.ParcelValue{
		Owner:   sender,
		Custody: o.Custody,
	}
	store.SetParcel(o.Target, &parcel)
	return code.TxCodeOK
}
