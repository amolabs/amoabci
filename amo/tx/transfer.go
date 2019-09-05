package tx

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var _ Operation = TransferParam{}

type TransferParam struct {
	To     crypto.Address `json:"to"`
	Amount types.Currency `json:"amount"`
}

func (o TransferParam) Check(store *store.Store, sender crypto.Address) uint32 {
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

func (o TransferParam) Execute(store *store.Store, sender crypto.Address) (uint32, []tm.KVPair) {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode, nil
	}
	fromBalance := store.GetBalance(sender)
	toBalance := store.GetBalance(o.To)
	fromBalance.Sub(&o.Amount)
	toBalance.Add(&o.Amount)
	store.SetBalance(sender, fromBalance)
	store.SetBalance(o.To, toBalance)
	return code.TxCodeOK, nil
}
