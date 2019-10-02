package tx

import (
	"bytes"
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type TransferParam struct {
	To     crypto.Address `json:"to"`
	Amount types.Currency `json:"amount"`
}

func parseTransferParam(raw []byte) (TransferParam, error) {
	var param TransferParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxTransfer struct {
	TxBase
	Param TransferParam `json:"-"`
}

var _ Tx = &TxTransfer{}

func (t *TxTransfer) Check() (uint32, string) {
	txParam, err := parseTransferParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	if len(txParam.To) != crypto.AddressSize {
		return code.TxCodeBadParam, "wrong recipient address size"
	}
	if bytes.Equal(t.GetSender(), txParam.To) {
		return code.TxCodeSelfTransaction, "tried to transfer to self"
	}
	return code.TxCodeOK, "ok"
}

func (t *TxTransfer) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseTransferParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	fromBalance := store.GetBalance(t.GetSender(), fromStage)
	if fromBalance.LessThan(&txParam.Amount) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}
	toBalance := store.GetBalance(txParam.To, fromStage)
	fromBalance.Sub(&txParam.Amount)
	toBalance.Add(&txParam.Amount)
	store.SetBalance(t.GetSender(), fromBalance)
	store.SetBalance(txParam.To, toBalance)
	return code.TxCodeOK, "ok", nil
}
