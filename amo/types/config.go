package types

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
