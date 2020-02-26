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

	if !txParam.Amount.GreaterThan(zero) {
		return code.TxCodeInvalidAmount, "invalid amount", nil
	}

	// check minimum staking unit first
	tmp := new(types.Currency)
	tmp.Mod(&txParam.Amount.Int, &ConfigAMOApp.MinStakingUnit.Int)
	if !tmp.Equals(new(types.Currency).Set(0)) {
		return code.TxCodeImproperStakeAmount, "improper stake amount", nil
	}

	balance := store.GetBalance(t.GetSender(), false)
	if balance.LessThan(&txParam.Amount) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}

	balance.Sub(&txParam.Amount)

	// just to check if existing validator key matches to the one of sender
	stake := store.GetStake(t.GetSender(), false)
	if stake != nil && !bytes.Equal(stake.Validator[:], txParam.Validator[:]) {
		return code.TxCodePermissionDenied, "validator key mismatch", nil
	}

	var k ed25519.PubKeyEd25519
	copy(k[:], txParam.Validator)
	stake = &types.Stake{
		Amount:    txParam.Amount,
		Validator: k,
	}

	err = store.SetLockedStake(t.GetSender(), stake, ConfigAMOApp.LockupPeriod)
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
