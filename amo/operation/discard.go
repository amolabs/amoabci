package operation

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

var _ Operation = Discard{}

type Discard struct {
	Target cmn.HexBytes `json:"target"`
}

func (o Discard) Check(store *store.Store, sender crypto.Address) uint32 {
	parcel := store.GetParcel(o.Target)
	if parcel == nil {
		return code.TxCodeTargetNotExists
	}
	if !bytes.Equal(parcel.Owner, sender) {
		return code.TxCodePermissionDenied
	}
	return code.TxCodeOK
}

func (o Discard) Execute(store *store.Store, sender crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	store.DeleteParcel(o.Target)
	tags := []cmn.KVPair{
		{Key: []byte("target"), Value: []byte(o.Target.String())},
	}
	return code.TxCodeOK, tags
}
