package operation

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	atypes "github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Transfer{}

type Transfer struct {
	To     crypto.Address  `json:"to"`
	Amount atypes.Currency `json:"amount"`
}

func (o Transfer) Check(store *store.Store, signer crypto.Address) uint32 {
	// TODO: make util for checking address size
	if len(o.To) != crypto.AddressSize {
		return code.TxCodeBadParam
	}
	fromBalance := store.GetBalance(signer)
	if fromBalance.LessThan(&o.Amount) {
		return code.TxCodeNotEnoughBalance
	}
	if bytes.Equal(signer, o.To) {
		return code.TxCodeSelfTransaction
	}
	return code.TxCodeOK
}

func (o Transfer) Execute(store *store.Store, signer crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, signer); resCode != code.TxCodeOK {
		return resCode, nil
	}
	fromBalance := store.GetBalance(signer)
	toBalance := store.GetBalance(o.To)
	fromBalance.Sub(&o.Amount)
	toBalance.Add(&o.Amount)
	store.SetBalance(signer, fromBalance)
	store.SetBalance(o.To, toBalance)
	tags := []cmn.KVPair{
		{Key: []byte(signer.String()), Value: []byte(fromBalance.String())},
		{Key: []byte(o.To.String()), Value: []byte(toBalance.String())},
	}
	return code.TxCodeOK, tags
}
