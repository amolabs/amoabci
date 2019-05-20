package tx

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Grant{}

type Grant struct {
	Target  tm.HexBytes    `json:"target"`
	Grantee crypto.Address `json:"grantee"`
	Custody tm.HexBytes    `json:"custody"`
}

func (o Grant) Check(store *store.Store, sender crypto.Address) uint32 {
	parcel := store.GetParcel(o.Target)
	if !bytes.Equal(parcel.Owner, sender) {
		return code.TxCodePermissionDenied
	}
	if store.GetRequest(o.Grantee, o.Target) == nil {
		return code.TxCodeRequestNotFound
	}
	usage := store.GetUsage(o.Grantee, o.Target)
	if usage != nil {
		return code.TxCodeAlreadyGranted
	}
	return code.TxCodeOK
}

func (o Grant) Execute(store *store.Store, sender crypto.Address) (uint32, []tm.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	request := store.GetRequest(o.Grantee, o.Target)
	store.DeleteRequest(o.Grantee, o.Target)
	balance := store.GetBalance(sender)
	balance.Add(&request.Payment)
	store.SetBalance(sender, balance)
	usage := types.UsageValue{
		Custody: o.Custody,
	}
	store.SetUsage(o.Grantee, o.Target, &usage)
	tags := []tm.KVPair{
		{Key: []byte("parcel.id"), Value: []byte(o.Target.String())},
	}
	return code.TxCodeOK, tags
}
