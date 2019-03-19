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

func (o Cancel) Check(store *store.Store, signer crypto.Address) uint32 {
	request := store.GetRequest(signer, o.Target)
	if request == nil {
		return code.TxCodeTargetNotExists
	}
	return code.TxCodeOK
}

func (o Cancel) Execute(store *store.Store, signer crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, signer); resCode != code.TxCodeOK {
		return resCode, nil
	}
	request := store.GetRequest(signer, o.Target)
	store.DeleteRequest(signer, o.Target)
	balance := store.GetBalance(signer)
	balance.Add(&request.Payment)
	store.SetBalance(signer, balance)
	tags := []cmn.KVPair{
		{Key: []byte("target"), Value: []byte(o.Target.String())},
		{Key: []byte(signer.String()), Value: []byte(balance.String())},
	}
	return code.TxCodeOK, tags
}
