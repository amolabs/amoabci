package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigCheckValue(t *testing.T) {
	cfg := AMOAppConfig{
		MaxValidators:           uint64(100),
		WeightValidator:         uint64(2),
		WeightDelegator:         uint64(1),
		MinStakingUnit:          *new(Currency).Set(100),
		BlkReward:               *new(Currency).Set(1000),
		TxReward:                *new(Currency).Set(1000),
		PenaltyRatioM:           float64(0.1),
		PenaltyRatioL:           float64(0.1),
		LazinessCounterWindow:   int64(10000),
		LazinessThreshold:       float64(0.9),
		BlockBoundTxGracePeriod: uint64(10000),
		LockupPeriod:            uint64(10000),
		DraftOpenCount:          uint64(10000),
		DraftCloseCount:         uint64(10000),
		DraftApplyCount:         uint64(10000),
		DraftDeposit:            *new(Currency).Set(1000),
		DraftQuorumRate:         float64(0.1),
		DraftPassRate:           float64(0.7),
		DraftRefundRate:         float64(0.2),
	}

	payload := []byte(`{"non_existing_config": "0"}`)
	ok, _ := cfg.Check(payload)
	assert.False(t, ok)

	payload = []byte(`{"lockup_period": 100}`)
	ok, _ = cfg.Check(payload)
	assert.False(t, ok)

	payload = []byte(`{"blk_reward": "-1"}`)
	ok, _ = cfg.Check(payload)
	assert.False(t, ok)

	payload = []byte(`{"lockup_period": 100000}`)
	ok, changedCfg := cfg.Check(payload)
	assert.True(t, ok)
	assert.NotEqual(t, changedCfg.LockupPeriod, cfg.LockupPeriod)

	payload = []byte(`{"blk_reward": "0"}`)
	ok, changedCfg = cfg.Check(payload)
	assert.True(t, ok)
	assert.NotEqual(t, changedCfg.BlkReward, cfg.BlkReward)

	payload = []byte(`{"blk_reward": "100", "lockup_period": 1000000}`)
	ok, changedCfg = cfg.Check(payload)
	assert.True(t, ok)
	assert.NotEqual(t, changedCfg.BlkReward, cfg.BlkReward)
	assert.NotEqual(t, changedCfg.LockupPeriod, cfg.LockupPeriod)
}
