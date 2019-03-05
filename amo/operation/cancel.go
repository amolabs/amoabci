package operation

import (
	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/db"
	"github.com/amolabs/tendermint-amo/crypto"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"strconv"
)

var _ Operation = Cancel{}

type Cancel struct {
	Target cmn.HexBytes `json:"target"`
}

func (o Cancel) Check(store *db.Store, signer crypto.Address) uint32 {
	request := store.GetRequest(signer, o.Target)
	if request == nil {
		return code.TxCodeTargetNotExists
	}
	return code.TxCodeOK
}

func (o Cancel) Execute(store *db.Store, signer crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, signer); resCode != code.TxCodeOK {
		return resCode, nil
	}
	request := store.GetRequest(signer, o.Target)
	store.DeleteRequest(signer, o.Target)
	balance := store.GetBalance(signer)
	balance += request.Payment
	store.SetBalance(signer, balance)
	tags := []cmn.KVPair{
		{Key: []byte("target"), Value: []byte(o.Target.String())},
		{Key: []byte(signer.String()), Value: []byte(strconv.FormatUint(uint64(balance), 10))},
	}
	return code.TxCodeOK, tags
}
