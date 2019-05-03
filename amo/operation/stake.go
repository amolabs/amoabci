package operation

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var _ Operation = Stake{}

type Stake struct {
	Amount    types.Currency `json:"amount"`
	Validator cmn.HexBytes   `json:"validator"`
}

func (o Stake) Check(store *store.Store, sender crypto.Address) uint32 {
	balance := store.GetBalance(sender)
	if balance.LessThan(&o.Amount) {
		return code.TxCodeNotEnoughBalance
	}
	if len(o.Validator) != ed25519.PubKeyEd25519Size {
		return code.TxCodeBadValidator
	}
	return code.TxCodeOK
}

func (o Stake) Execute(store *store.Store, sender crypto.Address) uint32 {
	if resCode := o.Check(store, sender); resCode != code.TxCodeOK {
		return resCode
	}
	balance := store.GetBalance(sender)
	balance.Sub(&o.Amount)
	stake := store.GetStake(sender)
	if stake == nil {
		var k ed25519.PubKeyEd25519
		copy(k[:], o.Validator)
		stake = &types.Stake{
			Amount:    o.Amount,
			Validator: k,
		}
	} else {
		stake.Amount.Add(&o.Amount)
		copy(stake.Validator[:], o.Validator)
	}
	if err := store.SetStake(sender, stake); err != nil {
		return code.TxCodeBadValidator
	}
	store.SetBalance(sender, balance)
	return code.TxCodeOK
}
