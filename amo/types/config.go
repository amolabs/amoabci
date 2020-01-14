package types

import (
	"encoding/json"
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

func (cfg *AMOAppConfig) Check(txCfgRaw json.RawMessage) (bool, AMOAppConfig) {
	var txCfgMap map[string]interface{}

	// handle exception for allowing empty config field on purpose
	if len(txCfgRaw) == 0 {
		return true, *cfg
	}

	cfgMap, err := cfg.getMap()
	if err != nil {
		return false, AMOAppConfig{}
	}

	err = json.Unmarshal(txCfgRaw, &txCfgMap)
	if err != nil {
		return false, AMOAppConfig{}
	}

	for key, _ := range txCfgMap {
		_, exist := cfgMap[key]
		if !exist {
			return false, AMOAppConfig{}
		}
	}

	tmpCfg := *cfg

	err = json.Unmarshal(txCfgRaw, &tmpCfg)
	if err != nil {
		return false, AMOAppConfig{}
	}

	if cmp(tmpCfg.MaxValidators, ">", uint64(0)) &&
		cmp(tmpCfg.WeightValidator, ">", uint64(0)) &&
		cmp(tmpCfg.WeightDelegator, ">", uint64(0)) &&

		cmp(tmpCfg.MinStakingUnit, ">", "0") &&

		cmp(tmpCfg.BlkReward, ">=", "0") &&
		cmp(tmpCfg.TxReward, ">=", "0") &&

		cmp(tmpCfg.PenaltyRatioM, ">", float64(0)) &&
		cmp(tmpCfg.PenaltyRatioL, ">", float64(0)) &&

		cmp(tmpCfg.LazinessCounterWindow, ">=", int64(10000)) &&
		cmp(tmpCfg.LazinessThreshold, ">", float64(0)) &&

		cmp(tmpCfg.BlockBoundTxGracePeriod, ">=", uint64(10000)) &&
		cmp(tmpCfg.LockupPeriod, ">=", uint64(10000)) &&

		cmp(tmpCfg.DraftOpenCount, ">=", uint64(10000)) &&
		cmp(tmpCfg.DraftCloseCount, ">=", uint64(10000)) &&
		cmp(tmpCfg.DraftApplyCount, ">=", uint64(10000)) &&
		cmp(tmpCfg.DraftDeposit, ">=", "0") &&
		cmp(tmpCfg.DraftQuorumRate, ">", float64(0)) &&
		cmp(tmpCfg.DraftPassRate, ">", float64(0)) &&
		cmp(tmpCfg.DraftRefundRate, ">", float64(0)) {

		return true, tmpCfg
	}

	return false, AMOAppConfig{}
}

func (cfg *AMOAppConfig) getMap() (map[string]interface{}, error) {
	var mapConfig map[string]interface{}

	jsonConfig, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonConfig, &mapConfig)
	if err != nil {
		return nil, err
	}

	return mapConfig, nil
}

func cmp(targetA interface{}, operator string, targetB interface{}) bool {
	var equal, greater, less bool

	switch targetA.(type) {
	case string:
		a, err := new(Currency).SetString(targetA.(string), 10)
		if err != nil {
			return false
		}

		b, err := new(Currency).SetString(targetB.(string), 10)
		if err != nil {
			return false
		}

		equal = a.Equals(b)
		greater = a.GreaterThan(b)
		less = a.LessThan(b)

	case uint64:
		a := targetA.(uint64)
		b := targetB.(uint64)

		equal = (a == b)
		greater = (a > b)
		less = (a < b)

	case int64:
		a := targetA.(int64)
		b := targetB.(int64)

		equal = (a == b)
		greater = (a > b)
		less = (a < b)

	case float64:
		a := targetA.(float64)
		b := targetB.(float64)

		equal = (a == b)
		greater = (a > b)
		less = (a < b)
	}

	switch operator {
	case ">":
		return greater
	case ">=":
		return greater || equal
	case "==":
		return equal
	case "<=":
		return less || equal
	case "<":
		return less
	default:
		return false
	}
}
