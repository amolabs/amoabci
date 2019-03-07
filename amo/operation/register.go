package operation

import (
	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/store/types"
	"github.com/amolabs/tendermint-amo/crypto"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
)

var _ Operation = Register{}

type Register struct {
	Target  cmn.HexBytes `json:"target"`
	Custody cmn.HexBytes `json:"custody"`
	// TODO: extra info
}

func (o Register) Check(store *store.Store, signer crypto.Address) uint32 {
	// TODO: permission check from PDSN
	if store.GetParcel(o.Target) != nil {
		return code.TxCodeTargetAlreadyExists
	}
	return code.TxCodeOK
}

func (o Register) Execute(store *store.Store, signer crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, signer); resCode != code.TxCodeOK {
		return resCode, nil
	}
	parcel := types.ParcelValue{
		Owner:   signer,
		Custody: o.Custody,
	}
	store.SetParcel(o.Target, &parcel)
	tags := []cmn.KVPair{
		{Key: []byte("owner"), Value: []byte(signer.String())},
		{Key: []byte("target"), Value: []byte(o.Target.String())},
	}
	return code.TxCodeOK, tags
}
