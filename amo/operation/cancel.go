package operation

import (
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

var _ Operation = Cancel{}

type Cancel struct {
	Target cmn.HexBytes `json:"target"`
}

func (o Cancel) Check(store *store.Store, sender crypto.Address) uint32 {
	request := store.GetRequest(sender, o.Target)
	if request == nil {
		return code.TxCodeTargetNotExists
	}
	return code.TxCodeOK
}

func (o Cancel) Execute(store *store.Store, sender crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	request := store.GetRequest(sender, o.Target)
	store.DeleteRequest(sender, o.Target)
	balance := store.GetBalance(sender)
	balance.Add(&request.Payment)
	store.SetBalance(sender, balance)
	tags := []cmn.KVPair{
		{Key: []byte("target"), Value: []byte(o.Target.String())},
		{Key: []byte(sender.String()), Value: []byte(balance.String())},
	}
	return code.TxCodeOK, tags
}
