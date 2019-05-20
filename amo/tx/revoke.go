package tx

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

var _ Operation = Revoke{}

type Revoke struct {
	Grantee crypto.Address `json:"grantee"`
	Target  tm.HexBytes    `json:"target"`
}

// TODO: fix: use GetUsage
func (o Revoke) Check(store *store.Store, sender crypto.Address) uint32 {
	parcel := store.GetParcel(o.Target)
	if parcel == nil {
		return code.TxCodeParcelNotFound
	}
	if !bytes.Equal(parcel.Owner, sender) {
		return code.TxCodePermissionDenied
	}
	return code.TxCodeOK
}

func (o Revoke) Execute(store *store.Store, sender crypto.Address) (uint32, []tm.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	store.DeleteUsage(o.Grantee, o.Target)
	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(o.Target.String())},
	}
	return code.TxCodeOK, tags
}
