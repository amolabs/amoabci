package tx

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Delegate{}

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
	if delegate != nil && !bytes.Equal(delegate.Delegatee, o.To) {
		return code.TxCodeMultipleDelegates
	}
	stake := store.GetStake(o.To)
	if stake == nil || stake.Amount.Equals(zero) {
		return code.TxCodeNoStake
	}
	return code.TxCodeOK
}

func (o Delegate) Execute(store *store.Store, sender crypto.Address) (uint32, []tm.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	balance := store.GetBalance(sender)
	balance.Sub(&o.Amount)
	delegate := store.GetDelegate(sender)
	if delegate == nil {
		delegate = &types.Delegate{
			Delegatee: o.To,
			Amount:    o.Amount,
		}
	} else {
		delegate.Amount.Add(&o.Amount)
	}
	store.SetBalance(sender, balance)
	store.SetDelegate(sender, delegate)
	// TODO Update delegation state
	return code.TxCodeOK, nil
}
