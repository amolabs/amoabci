package tx

import (
	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Retract{}

type Retract struct {
	Amount types.Currency `json:"amount"`
}

func (o Retract) Check(store *store.Store, sender crypto.Address) uint32 {
	delegate := store.GetDelegate(sender)
	if delegate == nil {
		return code.TxCodeDelegationNotExists
	}
	if delegate.Amount.LessThan(&o.Amount) {
		return code.TxCodeNotEnoughBalance
	}
	return code.TxCodeOK
}

func (o Retract) Execute(store *store.Store, sender crypto.Address) (uint32, []tm.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	delegate := store.GetDelegate(sender)
	delegate.Amount.Sub(&o.Amount)
	store.SetDelegate(sender, delegate)
	balance := store.GetBalance(sender)
	balance.Add(&o.Amount)
	store.SetBalance(sender, balance)
	return code.TxCodeOK, nil
}
