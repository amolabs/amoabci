package operation

import (
	"bytes"
	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	dtypes "github.com/amolabs/amoabci/amo/store/types"
	"github.com/amolabs/tendermint-amo/crypto"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"strconv"
)

var _ Operation = Grant{}

type Grant struct {
	Target  cmn.HexBytes
	Grantee crypto.Address
	Custody cmn.HexBytes
}

func (o Grant) Check(store *store.Store, signer crypto.Address) uint32 {
	parcel := store.GetParcel(o.Target)
	if !bytes.Equal(parcel.Owner, signer) {
		return code.TxCodePermissionDenied
	}
	if store.GetRequest(o.Grantee, o.Target) == nil {
		return code.TxCodeRequestNotExists
	}
	usage := store.GetUsage(o.Grantee, o.Target)
	if usage != nil {
		return code.TxCodeTargetAlreadyExists
	}
	return code.TxCodeOK
}

func (o Grant) Execute(store *store.Store, signer crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, signer); resCode != code.TxCodeOK {
		return resCode, nil
	}
	request := store.GetRequest(o.Grantee, o.Target)
	store.DeleteRequest(o.Grantee, o.Target)
	balance := store.GetBalance(signer)
	balance += request.Payment
	store.SetBalance(signer, balance)
	usage := dtypes.UsageValue{
		Custody: o.Custody,
	}
	store.SetUsage(o.Grantee, o.Target, &usage)
	tags := []cmn.KVPair{
		{Key: []byte("target"), Value: []byte(o.Target.String())},
		{Key: []byte(signer.String()), Value: []byte(strconv.FormatUint(uint64(balance), 10))},
	}
	return code.TxCodeOK, tags
}
