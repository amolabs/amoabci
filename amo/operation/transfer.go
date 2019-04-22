package operation

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	atypes "github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Transfer{}

type Transfer struct {
	To     crypto.Address  `json:"to"`
	Amount atypes.Currency `json:"amount"`
}

func (o Transfer) Check(store *store.Store, sender crypto.Address) uint32 {
	// TODO: make util for checking address size
	if len(o.To) != crypto.AddressSize {
		return code.TxCodeBadParam
	}
	fromBalance := store.GetBalance(sender)
	if fromBalance.LessThan(&o.Amount) {
		return code.TxCodeNotEnoughBalance
	}
	if bytes.Equal(sender, o.To) {
		return code.TxCodeSelfTransaction
	}
	return code.TxCodeOK
}

func (o Transfer) Execute(store *store.Store, sender crypto.Address) uint32 {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode
	}
	fromBalance := store.GetBalance(sender)
	toBalance := store.GetBalance(o.To)
	fromBalance.Sub(&o.Amount)
	toBalance.Add(&o.Amount)
	store.SetBalance(sender, fromBalance)
	store.SetBalance(o.To, toBalance)
	return code.TxCodeOK
}
