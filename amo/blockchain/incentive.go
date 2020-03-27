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

	weightValidator, weightDelegator float64,
	blkReward, txReward types.Currency,
	height, numDeliveredTxs int64,
	staker crypto.Address,
	feeAccumulated types.Currency,
) error {

	stake := store.GetStake(staker, true)
	if stake == nil {
		return errors.New("No stake, no reward.")
	}
	ds := store.GetDelegatesByDelegatee(staker, true)

	// itof
	sf := new(big.Float).SetInt(&stake.Amount.Int)

	var tmpc, tmpc2 types.Currency
	var incentive, rTotal, rTx types.Currency

	// reward = BlkReward + TxReward * numDeliveredTxs
	// incentive = reward + fee

	// total reward
	rTotal = blkReward
	rTx = txReward
	tmpc.SetInt64(numDeliveredTxs)
	tmpc.Mul(&tmpc.Int, &rTx.Int)
	rTotal.Add(&tmpc)

	incentive.Set(0)
	incentive.Add(rTotal.Add(&feeAccumulated))

	// ignore 0 incentive
	if incentive.Equals(new(types.Currency).Set(0)) {
		return nil
	}

	var (
		wsumf, wf big.Float // weighted sum
		tmpf      big.Float // tmp
	)

	wf.SetFloat64(weightValidator)
	wsumf.Mul(&wf, sf)
	wf.SetFloat64(weightDelegator)
	for _, d := range ds {
		df := new(big.Float).SetInt(&d.Amount.Int)
		tmpf.Mul(&wf, df)
		wsumf.Add(&wsumf, &tmpf)
	}

	// individual rewards
	tmpc.Set(0) // subtotal for delegate holders
	for _, d := range ds {
		df := new(big.Float).SetInt(&d.Amount.Int)
		tmpc2 = *partialAmount(weightDelegator, df, &wsumf, &incentive)
		tmpc.Add(&tmpc2) // update subtotal

		b := store.GetBalance(d.Delegator, false).Add(&tmpc2)
		store.SetBalance(d.Delegator, b)                      // update balance
		store.AddIncentiveRecord(height, d.Delegator, &tmpc2) // update incentive record
		logger.Debug("Block reward",
			"delegator", hex.EncodeToString(d.Delegator), "reward", tmpc2.String())
	}
	tmpc2.Int.Sub(&incentive.Int, &tmpc.Int) // calc validator reward
	b := store.GetBalance(staker, false).Add(&tmpc2)
	store.SetBalance(staker, b)                      // update balance
	store.AddIncentiveRecord(height, staker, &tmpc2) // update incentive record
	logger.Debug("Block reward",
		"proposer", hex.EncodeToString(staker), "reward", tmpc2.String())

	return nil
}
