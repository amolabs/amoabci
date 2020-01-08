package blockchain

import (
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

func DistributeIncentive(
	store *store.Store,
	logger log.Logger,

	weightValidator, weightDelegator uint64,
	blkReward, txReward uint64,
	height, numDeliveredTxs int64,
	staker crypto.Address,
	feeAccumulated types.Currency,
) error {

	stake := store.GetStake(staker, true)
	if stake == nil {
		return errors.New("No stake, no reward.")
	}
	ds := store.GetDelegatesByDelegatee(staker, true)

	var tmp, tmp2 types.Currency
	var incentive, rTotal, rTx types.Currency

	// reward = BlkReward + TxReward * numDeliveredTxs
	// incentive = reward + fee

	// total reward
	rTotal.Set(blkReward)
	rTx.Set(txReward)
	tmp.SetInt64(numDeliveredTxs)
	tmp.Mul(&tmp.Int, &rTx.Int)
	rTotal.Add(&tmp)

	incentive.Set(0)
	incentive.Add(rTotal.Add(&feeAccumulated))

	// ignore 0 incentive
	if incentive.Equals(new(types.Currency).Set(0)) {
		return nil
	}

	// weighted sum
	var wsum, w big.Int
	w.SetUint64(weightValidator)
	wsum.Mul(&w, &stake.Amount.Int)
	w.SetUint64(weightDelegator)
	for _, d := range ds {
		tmp.Mul(&w, &d.Amount.Int)
		wsum.Add(&wsum, &tmp.Int)
	}
	// individual rewards
	tmp.Set(0) // subtotal for delegate holders
	for _, d := range ds {
		tmp2 = *partialAmount(weightDelegator, &d.Amount.Int, &wsum, &incentive)
		tmp.Add(&tmp2) // update subtotal
		b := store.GetBalance(d.Delegator, false).Add(&tmp2)
		store.SetBalance(d.Delegator, b)                     // update balance
		store.AddIncentiveRecord(height, d.Delegator, &tmp2) // update incentive record
		logger.Debug("Block reward",
			"delegator", hex.EncodeToString(d.Delegator), "reward", tmp2.Int64())
	}
	tmp2.Int.Sub(&incentive.Int, &tmp.Int) // calc validator reward
	b := store.GetBalance(staker, false).Add(&tmp2)
	store.SetBalance(staker, b)                     // update balance
	store.AddIncentiveRecord(height, staker, &tmp2) // update incentive record
	logger.Debug("Block reward",
		"proposer", hex.EncodeToString(staker), "reward", tmp2.Int64())

	return nil
}
