package operation

import (
	"bytes"
	"strconv"

	"github.com/amolabs/tendermint-amo/crypto"
	cmn "github.com/amolabs/tendermint-amo/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Request{}

type Request struct {
	Target  cmn.HexBytes   `json:"target"`
	Payment types.Currency `json:"payment"`
	// TODO: Extra info
}

func (o Request) Check(store *store.Store, signer crypto.Address) uint32 {
	parcel := store.GetParcel(o.Target)
	if parcel == nil {
		return code.TxCodeTargetNotExists
	}
	if bytes.Equal(parcel.Owner, signer) {
		return code.TxCodeSelfTransaction
	}
	if store.GetUsage(signer, o.Target) != nil {
		return code.TxCodeTargetAlreadyBought
	}
	if store.GetBalance(signer) < o.Payment {
		return code.TxCodeNotEnoughBalance
	}
	return code.TxCodeOK
}

func (o Request) Execute(store *store.Store, signer crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, signer); resCode != code.TxCodeOK {
		return resCode, nil
	}
	balance := store.GetBalance(signer)
	balance -= o.Payment
	store.SetBalance(signer, balance)
	request := types.RequestValue{
		Payment: o.Payment,
	}
	store.SetRequest(signer, o.Target, &request)
	tags := []cmn.KVPair{
		{Key: []byte(signer.String()), Value: []byte(strconv.FormatUint(uint64(balance), 10))},
		{Key: []byte("target"), Value: []byte(o.Target.String())},
	}
	return code.TxCodeOK, tags
}