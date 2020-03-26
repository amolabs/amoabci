package tx

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type StakeParam struct {
	Validator tmbytes.HexBytes `json:"validator"`
	Amount    types.Currency   `json:"amount"`
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

func (t *TxStake) Execute(store *store.Store) (uint32, string, []abci.Event) {
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
		case code.GetError(code.TxCodeBadParam):
			return code.TxCodeBadParam, err.Error(), nil
		case code.GetError(code.TxCodePermissionDenied):
			return code.TxCodePermissionDenied, err.Error(), nil
		case code.GetError(code.TxCodeDelegateExists):
			return code.TxCodeDelegateExists, err.Error(), nil
		case code.GetError(code.TxCodeLastValidator):
			return code.TxCodeLastValidator, err.Error(), nil
		default:
			return code.TxCodeUnknown, err.Error(), nil
		}
	}

	store.SetBalance(t.GetSender(), balance)

	return code.TxCodeOK, "ok", nil
}
