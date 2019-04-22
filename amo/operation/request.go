package operation

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

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

func (o Request) Check(store *store.Store, sender crypto.Address) uint32 {
	parcel := store.GetParcel(o.Target)
	if parcel == nil {
		return code.TxCodeTargetNotExists
	}
	if bytes.Equal(parcel.Owner, sender) {
		return code.TxCodeSelfTransaction
	}
	if store.GetUsage(sender, o.Target) != nil {
		return code.TxCodeTargetAlreadyBought
	}
	if store.GetBalance(sender).LessThan(&o.Payment) {
		return code.TxCodeNotEnoughBalance
	}
	return code.TxCodeOK
}

func (o Request) Execute(store *store.Store, sender crypto.Address) uint32 {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode
	}
	balance := store.GetBalance(sender)
	balance.Sub(&o.Payment)
	store.SetBalance(sender, balance)
	request := types.RequestValue{
		Payment: o.Payment,
	}
	store.SetRequest(sender, o.Target, &request)
	return code.TxCodeOK
}
