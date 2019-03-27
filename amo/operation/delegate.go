package operation

import (
	"bytes"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type Delegate struct {
	To     crypto.Address `json:"to"`
	Amount types.Currency `json:"amount"`
}

func (o Delegate) Check(store *store.Store, sender crypto.Address) uint32 {
	if bytes.Equal(o.To, sender) {
		return code.TxCodeSelfTransaction
	}
	balance := store.GetBalance(sender)
	if balance.LessThan(&o.Amount) {
		return code.TxCodeNotEnoughBalance
	}
	delegate := store.GetDelegate(sender)
	if delegate != nil && !bytes.Equal(delegate.Delegator, o.To) {
		return code.TxCodeMultipleDelegates
	}
	return code.TxCodeOK
}

func (o Delegate) Execute(store *store.Store, sender crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	balance := store.GetBalance(sender)
	balance.Sub(&o.Amount)
	delegate := store.GetDelegate(sender)
	if delegate == nil {
		delegate = &types.DelegateValue{
			Amount:    o.Amount,
			Delegator: o.To,
		}
	} else {
		delegate.Amount.Add(&o.Amount)
	}
	store.SetBalance(sender, balance)
	store.SetDelegate(sender, delegate)
	// TODO Update delegation state
	tags := []cmn.KVPair{
		{Key: []byte(sender.String()), Value: []byte(balance.String())},
	}
	return code.TxCodeOK, tags
}
