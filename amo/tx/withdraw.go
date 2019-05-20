package tx

import (
	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

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

func (o Withdraw) Execute(store *store.Store, sender crypto.Address) (uint32, []tm.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	stake := store.GetStake(sender)
	if stake == nil {
		return code.TxCodeNoStake, nil
	}
	if stake.Amount.Sub(&o.Amount).Sign() == -1 {
		return code.TxCodeNotEnoughBalance, nil
	}
	if err := store.SetStake(sender, stake); err != nil {
		return code.TxCodeBadValidator, nil
	}
	balance := store.GetBalance(sender)
	balance.Add(&o.Amount)
	store.SetBalance(sender, balance)
	return code.TxCodeOK, nil
}
