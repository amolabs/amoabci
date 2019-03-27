package operation

import (
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"

)

type Stake struct {
	Amount types.Currency `json:"amount"`
}

func (o Stake) Check(store *store.Store, sender crypto.Address) uint32 {
	balance := store.GetBalance(sender)
	if balance.LessThan(&o.Amount) {
		return code.TxCodeNotEnoughBalance
	}
	return code.TxCodeOK
}

func (o Stake) Execute(store *store.Store, sender crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	balance := store.GetBalance(sender)
	balance.Sub(&o.Amount)
	stake := store.GetStake(sender)
	stake.Add(&o.Amount)
	store.SetBalance(sender, balance)
	store.SetStake(sender, stake)
	tags := []cmn.KVPair{
		{Key: []byte(sender.String()), Value: []byte(balance.String())},
	}
	return code.TxCodeOK, tags
}