package tx

import (
	"bytes"
	"encoding/json"

	"github.com/tendermint/tendermint/crypto/ed25519"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type StakeParam struct {
	Validator tm.HexBytes    `json:"validator"`
	Amount    types.Currency `json:"amount"`
}

func parseStakeParam(raw []byte) (StakeParam, error) {
	var param StakeParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxStake struct {
	TxBase
	Param StakeParam `json:"-"`
}

var _ Tx = &TxStake{}

func (t *TxStake) Check() (uint32, string) {
	txParam, err := parseStakeParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	// TODO: check format

	if len(txParam.Validator) != ed25519.PubKeyEd25519Size {
		return code.TxCodeBadValidator, "bad validator key"
	}
	return code.TxCodeOK, "ok"
}

func (t *TxStake) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseStakeParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	balance := store.GetBalance(t.GetSender())
	if balance.LessThan(&txParam.Amount) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}

	balance.Sub(&txParam.Amount)
	stake := store.GetStake(t.GetSender())
	if stake == nil {
		var k ed25519.PubKeyEd25519
		copy(k[:], txParam.Validator)
		stake = &types.Stake{
			Amount:    txParam.Amount,
			Validator: k,
		}
	} else if bytes.Equal(stake.Validator[:], txParam.Validator[:]) {
		stake.Amount.Add(&txParam.Amount)
	} else {
		return code.TxCodePermissionDenied, "validator key mismatch", nil
	}
	err = store.SetLockedStake(t.GetSender(), stake, int64(ConfigLockupPeriod))
	if err != nil {
		switch err {
		case code.TxErrBadParam:
			return code.TxCodeBadParam, err.Error(), nil
		case code.TxErrPermissionDenied:
			return code.TxCodePermissionDenied, err.Error(), nil
		case code.TxErrDelegateExists:
			return code.TxCodeDelegateExists, err.Error(), nil
		case code.TxErrLastValidator:
			return code.TxCodeLastValidator, err.Error(), nil
		default:
			return code.TxCodeUnknown, err.Error(), nil
		}
	}
	store.SetBalance(t.GetSender(), balance)
	return code.TxCodeOK, "ok", nil
}
