package operation

import (
	"bytes"
	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/db"
	"github.com/amolabs/tendermint-amo/crypto"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
)

var _ Operation = Revoke{}

type Revoke struct {
	Grantee crypto.Address `json:"grantee"`
	Target  cmn.HexBytes   `json:"target"`
}

func (o Revoke) Check(store *db.Store, signer crypto.Address) uint32 {
	parcel := store.GetParcel(o.Target)
	if !bytes.Equal(parcel.Owner, signer) {
		return code.TxCodePermissionDenied
	}
	return code.TxCodeOK
}

func (o Revoke) Execute(store *db.Store, signer crypto.Address) (uint32, []cmn.KVPair) {
	store.DeleteUsage(o.Grantee, o.Target)
	tags := []cmn.KVPair{
		{Key: []byte("grantee"), Value: []byte(o.Grantee.String())},
		{Key: []byte("target"), Value: []byte(o.Target.String())},
	}
	return code.TxCodeOK, tags
}
