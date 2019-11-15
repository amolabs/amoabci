package blockchain

import (
	"encoding/hex"
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

// Convicts consist of
// - Malicious Validator: M
// - Lazy Validator: L

func PenalizeConvicts(
	store *store.Store,
	logger log.Logger,

	evidences []abci.Evidence,
	lazyValidators []crypto.Address,

	weightValidator, weightDelegator int64,
	penaltyRatioM, penaltyRatioL float64,
) error {

	// handle evidences
	for _, evidence := range evidences {
		validator := evidence.GetValidator().Address
		penalize(
			store, logger,
			weightValidator, weightDelegator,
			validator, penaltyRatioM,
		)
	}

	// handle lazyValidators
	for _, lazyValidator := range lazyValidators {
		penalize(
			store, logger,
			weightValidator, weightDelegator,
			lazyValidator, penaltyRatioL,
		)
	}

	return nil
}

func penalize(
	store *store.Store,
	logger log.Logger,

	weightValidator, weightDelegator int64,
	validator crypto.Address,
	ratio float64,
) {

	zeroAmount := new(types.Currency).Set(0)

	holder := store.GetHolderByValidator(validator, true)

	vs := store.GetStake(holder, true) // validator's stake
	if vs == nil {
		return
	}

	ds := store.GetDelegatesByDelegatee(holder, true) // delegators' stake

	es := store.GetEffStake(holder, true)
	esf := new(big.Float).SetInt(&es.Amount.Int)
	prf := new(big.Float).SetFloat64(ratio)

	penalty := types.Currency{}
	pf := esf.Mul(esf, prf) // penalty = effStake * penaltyRatio
	pf.Int(&penalty.Int)

	// weighted sum
	var (
		wsum, w   big.Int
		tmp, tmp2 types.Currency
	)
	w.SetInt64(weightValidator)
	wsum.Mul(&w, &vs.Amount.Int)
	w.SetInt64(weightDelegator)
	for _, d := range ds {
		tmp.Mul(&w, &d.Amount.Int)
		wsum.Add(&wsum, &tmp.Int)
	}

	// individual penalties for delegators
	tmp.Set(0) // subtotal for delegate holders
	for _, d := range ds {
		tmp2 = *partialAmount(weightDelegator, &d.Amount.Int, &wsum, &penalty)
		tmp.Add(&tmp2) // update subtotal
		d.Delegate.Amount.Sub(&tmp2)

		if d.Delegate.Amount.LessThan(zeroAmount) {
			d.Delegate.Amount.Set(0)
		}

		store.SetDelegate(d.Delegator, d.Delegate)
		logger.Debug("Evidence penalty",
			"delegator", hex.EncodeToString(d.Delegator), "penalty", tmp2.Int64())
	}
	tmp2.Int.Sub(&penalty.Int, &tmp.Int) // calc voter(validator) penalty
	store.SlashStakes(holder, tmp2, true)

	logger.Debug("Evidence penalty",
		"validator", hex.EncodeToString(holder), "penalty", tmp2.Int64())
}
