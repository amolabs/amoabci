package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/amolabs/amoabci/amo"
)

func TestConfigCheckValue(t *testing.T) {
	cfg := AMOAppConfig{
		MaxValidators:         uint64(100),
		WeightValidator:       uint64(2),
		WeightDelegator:       uint64(1),
		MinStakingUnit:        *new(Currency).Set(100),
		BlkReward:             *new(Currency).Set(1000),
		TxReward:              *new(Currency).Set(1000),
		PenaltyRatioM:         float64(0.1),
		PenaltyRatioL:         float64(0.1),
		LazinessCounterWindow: int64(10000),
		LazinessThreshold:     float64(0.9),
		BlockBindingWindow:    int64(10000),
		LockupPeriod:          int64(10000),
		DraftOpenCount:        int64(10000),
		DraftCloseCount:       int64(10000),
		DraftApplyCount:       int64(10000),
		DraftDeposit:          *new(Currency).Set(1000),
		DraftQuorumRate:       float64(0.1),
		DraftPassRate:         float64(0.7),
		DraftRefundRate:       float64(0.2),
	}

	state := amo.State{Height: 1, ProtocolVersion: 2}

	payload := []byte(`{"non_existing_config": "0"}`)
	_, err := cfg.Check(state, payload)
	assert.Error(t, err)

	payload = []byte(`{"lockup_period": 100}`)
	_, err = cfg.Check(state, payload)
	assert.Error(t, err)

	payload = []byte(`{"blk_reward": "-1"}`)
	_, err = cfg.Check(state, payload)
	assert.Error(t, err)

	payload = []byte(`{"lockup_period": 100000}`)
	changedCfg, err := cfg.Check(state, payload)
	assert.NoError(t, err)
	assert.NotEqual(t, changedCfg.LockupPeriod, cfg.LockupPeriod)

	payload = []byte(`{"blk_reward": "0"}`)
	changedCfg, err = cfg.Check(state, payload)
	assert.NoError(t, err)
	assert.NotEqual(t, changedCfg.BlkReward, cfg.BlkReward)

	payload = []byte(`{"blk_reward": "100", "lockup_period": 1000000}`)
	changedCfg, err = cfg.Check(state, payload)
	assert.NoError(t, err)
	assert.NotEqual(t, changedCfg.BlkReward, cfg.BlkReward)
	assert.NotEqual(t, changedCfg.LockupPeriod, cfg.LockupPeriod)

	payload = []byte(`{"upgrade_protocol_height": 10}`)
	_, err = cfg.Check(state, payload)
	assert.Error(t, err)

	payload = []byte(`{"upgrade_protocol_version": 1}`)
	_, err = cfg.Check(state, payload)
	assert.Error(t, err)

	payload = []byte(`{"upgrade_protocol_version": 2}`)
	_, err = cfg.Check(state, payload)
	assert.Error(t, err)

	payload = []byte(`{"upgrade_protocol_version": 4}`)
	_, err = cfg.Check(state, payload)
	assert.Error(t, err)

	payload = []byte(`{"upgrade_protocol_height": 30005, "upgrade_protocol_version": 3}`)
	changedCfg, err = cfg.Check(state, payload)
	assert.NoError(t, err)
	assert.NotEqual(t, changedCfg.UpgradeProtocolHeight, cfg.UpgradeProtocolHeight)
	assert.NotEqual(t, changedCfg.UpgradeProtocolVersion, cfg.UpgradeProtocolVersion)
}
