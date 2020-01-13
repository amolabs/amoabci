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

func (cfg *AMOAppConfig) Check(txRaw json.RawMessage) (bool, AMOAppConfig) {
	var txMap map[string]interface{}

	cfgMap, err := cfg.getMap()
	if err != nil {
		return false, AMOAppConfig{}
	}

	err = json.Unmarshal(txRaw, &txMap)
	if err != nil {
		return false, AMOAppConfig{}
	}

	for key, _ := range txMap["config"].(map[string]interface{}) {
		_, exist := cfgMap[key]
		if !exist {
			return false, AMOAppConfig{}
		}
	}

	tmp := struct {
		Config AMOAppConfig `json:"config"`
	}{Config: *cfg}

	err = json.Unmarshal(txRaw, &tmp)
	if err != nil {
		return false, AMOAppConfig{}
	}

	if cmp(tmp.Config.MaxValidators, ">", uint64(0)) &&
		cmp(tmp.Config.WeightValidator, ">", uint64(0)) &&
		cmp(tmp.Config.WeightDelegator, ">", uint64(0)) &&

		cmp(tmp.Config.MinStakingUnit, ">", "0") &&

		cmp(tmp.Config.BlkReward, ">=", "0") &&
		cmp(tmp.Config.TxReward, ">=", "0") &&

		cmp(tmp.Config.PenaltyRatioM, ">", float64(0)) &&
		cmp(tmp.Config.PenaltyRatioL, ">", float64(0)) &&

		cmp(tmp.Config.LazinessCounterWindow, ">=", int64(10000)) &&
		cmp(tmp.Config.LazinessThreshold, ">", float64(0)) &&

		cmp(tmp.Config.BlockBoundTxGracePeriod, ">=", uint64(10000)) &&
		cmp(tmp.Config.LockupPeriod, ">=", uint64(10000)) &&

		cmp(tmp.Config.DraftOpenCount, ">=", uint64(10000)) &&
		cmp(tmp.Config.DraftCloseCount, ">=", uint64(10000)) &&
		cmp(tmp.Config.DraftApplyCount, ">=", uint64(10000)) &&
		cmp(tmp.Config.DraftDeposit, ">=", "0") &&
		cmp(tmp.Config.DraftQuorumRate, ">", float64(0)) &&
		cmp(tmp.Config.DraftPassRate, ">", float64(0)) &&
		cmp(tmp.Config.DraftRefundRate, ">", float64(0)) {

		return true, tmp.Config
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
