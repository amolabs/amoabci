package amo

import (
	"encoding/json"
)

const (
	// hard-coded configs
	defaultMaxValidators   = 100
	defaultWeightValidator = uint64(2)
	defaultWeightDelegator = uint64(1)

	defaultMinStakingUnit = "1000000000000000000000000"

	defaultBlkReward = "0"
	defaultTxReward  = "10000000000000000000"

	// TODO: not fixed default ratios yet
	defaultPenaltyRatioM = float64(0.3)
	defaultPenaltyRatioL = float64(0.3)

	defaultLazinessCounterWindow = int64(300)
	defaultLazinessThreshold     = float64(0.8)

	defaultBlockBoundTxGracePeriod = uint64(1000)
	defaultLockupPeriod            = uint64(1000000)

	defaultDraftOpenCount  = uint64(1000)
	defaultDraftCloseCount = uint64(100)
	defaultDraftApplyCount = uint64(1000)
	defaultDraftDeposit    = "1000000000000000000000000"
	defaultDraftQuorumRate = float64(0.3)
	defaultDraftPassRate   = float64(0.51)
	defaultDraftRefundRate = float64(0.2)
)

type AMOAppConfig struct {
	MaxValidators   uint64 `json:"max_validators"`
	WeightValidator uint64 `json:"weight_validator"`
	WeightDelegator uint64 `json:"weight_delegator"`

	MinStakingUnit string `json:"min_staking_unit"`

	BlkReward string `json:"blk_reward"`
	TxReward  string `json:"tx_reward"`

	PenaltyRatioM float64 `json:"penalty_ratio_m"` // malicious validator
	PenaltyRatioL float64 `json:"penalty_ratio_l"` // lazy validators

	LazinessCounterWindow int64   `json:"laziness_counter_window"`
	LazinessThreshold     float64 `json:"laziness_threshold"`

	BlockBoundTxGracePeriod uint64 `json:"block_bound_tx_grace_period"`
	LockupPeriod            uint64 `json:"lockup_period"`

	DraftOpenCount  uint64  `json:"draft_open_count"`
	DraftCloseCount uint64  `json:"draft_close_count"`
	DraftApplyCount uint64  `json:"draft_apply_count"`
	DraftDeposit    string  `json:"draft_deposit"`
	DraftQuorumRate float64 `json:"draft_quorum_rate"`
	DraftPassRate   float64 `json:"draft_pass_rate"`
	DraftRefundRate float64 `json:"draft_refund_rate"`
}

func (app *AMOApp) loadAppConfig() error {
	cfg := AMOAppConfig{
		defaultMaxValidators,
		defaultWeightValidator,
		defaultWeightDelegator,
		defaultMinStakingUnit,
		defaultBlkReward,
		defaultTxReward,
		defaultPenaltyRatioM,
		defaultPenaltyRatioL,
		defaultLazinessCounterWindow,
		defaultLazinessThreshold,
		defaultBlockBoundTxGracePeriod,
		defaultLockupPeriod,
		defaultDraftOpenCount,
		defaultDraftCloseCount,
		defaultDraftApplyCount,
		defaultDraftDeposit,
		defaultDraftQuorumRate,
		defaultDraftPassRate,
		defaultDraftRefundRate,
	}

	b := app.store.GetAppConfig()

	// if config exists
	if len(b) > 0 {
		err := json.Unmarshal(b, &cfg)
		if err != nil {
			return err
		}
	}

	app.config = cfg

	return nil
}
