package operation

import (
	"bytes"
	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/db"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/tendermint-amo/crypto"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"strconv"
)

var _ Operation = Transfer{}

type Transfer struct {
	To     crypto.Address  `json:"to"`
	Amount atypes.Currency `json:"amount"`
}

func (o Transfer) Check(store *db.Store, signer crypto.Address) uint32 {
	// TODO: make util for checking address size
	if len(o.To) != crypto.AddressSize {
		return code.TxCodeBadParam
	}
	fromBalance := store.GetBalance(signer)
	if fromBalance < o.Amount {
		return code.TxCodeNotEnoughBalance
	}
	if bytes.Equal(signer, o.To) {
		return code.TxCodeSelfTransaction
	}
	return code.TxCodeOK
}

func (o Transfer) Execute(store *db.Store, signer crypto.Address) (uint32, []cmn.KVPair) {
	if resCode := o.Check(store, signer); resCode != code.TxCodeOK {
		return resCode, nil
	}
	fromBalance := store.GetBalance(signer)
	toBalance := store.GetBalance(o.To)
	fromBalance -= o.Amount
	toBalance += o.Amount
	store.SetBalance(signer, fromBalance)
	store.SetBalance(o.To, toBalance)
	tags := []cmn.KVPair{
		{Key: []byte(signer.String()), Value: []byte(strconv.FormatUint(uint64(fromBalance), 10))},
		{Key: []byte(o.To.String()), Value: []byte(strconv.FormatUint(uint64(toBalance), 10))},
	}
	return code.TxCodeOK, tags
}
