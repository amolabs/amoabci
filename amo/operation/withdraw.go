package operation

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Withdraw{}

type Withdraw struct {
	Amount types.Currency `json:"amount"`
}

func (o Withdraw) Check(store *store.Store, sender crypto.Address) uint32 {
	stake := store.GetStake(sender)
	if stake == nil || stake.Amount.LessThan(&o.Amount) {
		return code.TxCodeNotEnoughBalance
	}
	return code.TxCodeOK
}

func (o Withdraw) Execute(store *store.Store, sender crypto.Address) uint32 {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode
	}
	stake := store.GetStake(sender)
	balance := store.GetBalance(sender)
	stake.Amount.Sub(&o.Amount)
	balance.Add(&o.Amount)
	store.SetStake(sender, stake)
	store.SetBalance(sender, balance)
	return code.TxCodeOK
}
