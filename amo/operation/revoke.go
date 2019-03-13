package operation

import (
	"bytes"
	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var _ Operation = Revoke{}

type Revoke struct {
	Grantee crypto.Address `json:"grantee"`
	Target  cmn.HexBytes   `json:"target"`
}

func (o Revoke) Check(store *store.Store, signer crypto.Address) uint32 {
	parcel := store.GetParcel(o.Target)
	if parcel == nil {
		return code.TxCodeTargetNotExists
	}
	if !bytes.Equal(parcel.Owner, signer) {
		return code.TxCodePermissionDenied
	}
	return code.TxCodeOK
}

func (o Revoke) Execute(store *store.Store, signer crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, signer); resCode != code.TxCodeOK {
		return resCode, nil
	}
	store.DeleteUsage(o.Grantee, o.Target)
	tags := []cmn.KVPair{
		{Key: []byte("grantee"), Value: []byte(o.Grantee.String())},
		{Key: []byte("target"), Value: []byte(o.Target.String())},
	}
	return code.TxCodeOK, tags
}
