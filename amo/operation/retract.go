package operation

import (
	"bytes"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type Retract struct {
	From   crypto.Address `json:"from"`
	Amount types.Currency `json:"amount"`
}

var zero = new(types.Currency).Set(0)

func (o Retract) Check(store *store.Store, sender crypto.Address) uint32 {
	delegate := store.GetDelegate(sender)
	if delegate == nil {
		return code.TxCodeDelegationNotExists
	}
	if !bytes.Equal(delegate.Delegator, o.From) {
		return code.TxCodeBadParam
	}
	if delegate.Amount.LessThan(&o.Amount) {
		return code.TxCodeNotEnoughBalance
	}
	return code.TxCodeOK
}

func (o Retract) Execute(store *store.Store, sender crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	delegate := store.GetDelegate(sender)
	delegate.Amount.Sub(&o.Amount)
	if delegate.Amount.Equals(zero) {
		store.SetDelegate(sender, nil)
	} else {
		store.SetDelegate(sender, delegate)
	}
	balance := store.GetBalance(sender)
	balance.Add(&o.Amount)
	store.SetBalance(sender, balance)
	// TODO Update delegation state
	tags := []cmn.KVPair{
		{Key: []byte(sender.String()), Value: []byte(balance.String())},
	}
	return code.TxCodeOK, tags
}
