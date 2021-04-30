package types

import (
	"encoding/json"
	"fmt"
)

const (
	// hard-coded configs
	DefaultMaxValidators   = uint64(100)
	DefaultWeightValidator = float64(2)
	DefaultWeightDelegator = float64(1)

	DefaultMinStakingUnit = "1000000000000000000000000"

	DefaultBlkReward = "0"
	DefaultTxReward  = "10000000000000000000"

	// TODO: not fixed Default ratios yet
	DefaultPenaltyRatioM = float64(0.3)
	DefaultPenaltyRatioL = float64(0.3)

	DefaultLazinessWindow     = int64(10000)
	DefaultLazinessThreshold  = int64(8000)
	DefaultHibernateThreshold = int64(100)
	DefaultHibernatePeriod    = int64(10000)
	DefaultBlockBindingWindow = int64(10000)
	DefaultLockupPeriod       = int64(1000000)

	DefaultDraftOpenCount  = int64(10000)
	DefaultDraftCloseCount = int64(10000)
	DefaultDraftApplyCount = int64(10000)
	DefaultDraftDeposit    = "1000000000000000000000000"
	DefaultDraftQuorumRate = float64(0.3)
	DefaultDraftPassRate   = float64(0.51)
	DefaultDraftRefundRate = float64(0.2)

	DefaultUpgradeProtocolHeight  = int64(0)
	DefaultUpgradeProtocolVersion = uint64(0)
)

type AMOAppConfig struct {
	MaxValidators          uint64   `json:"max_validators"`
	WeightValidator        float64  `json:"weight_validator"`
	WeightDelegator        float64  `json:"weight_delegator"`
	MinStakingUnit         Currency `json:"min_staking_unit"`
	BlkReward              Currency `json:"blk_reward"`
	TxReward               Currency `json:"tx_reward"`
	PenaltyRatioM          float64  `json:"penalty_ratio_m"` // malicious validator
	PenaltyRatioL          float64  `json:"penalty_ratio_l"` // lazy validators
	LazinessWindow         int64    `json:"laziness_window"`
	LazinessThreshold      int64    `json:"laziness_threshold"`
	HibernateThreshold     int64    `json:"hibernate_threshold"`
	HibernatePeriod        int64    `json:"hibernate_period"`
	BlockBindingWindow     int64    `json:"block_binding_window"`
	LockupPeriod           int64    `json:"lockup_period"`
	DraftOpenCount         int64    `json:"draft_open_count"`
	DraftCloseCount        int64    `json:"draft_close_count"`
	DraftApplyCount        int64    `json:"draft_apply_count"`
	DraftDeposit           Currency `json:"draft_deposit"`
	DraftQuorumRate        float64  `json:"draft_quorum_rate"`
	DraftPassRate          float64  `json:"draft_pass_rate"`
	DraftRefundRate        float64  `json:"draft_refund_rate"`
	UpgradeProtocolHeight  int64    `json:"upgrade_protocol_height"`
	UpgradeProtocolVersion uint64   `json:"upgrade_protocol_version"`
}

func NewDefaultAMOAppConfig() (AMOAppConfig, error) {
	cfg := AMOAppConfig{
		MaxValidators:          DefaultMaxValidators,
		WeightValidator:        DefaultWeightValidator,
		WeightDelegator:        DefaultWeightDelegator,
		PenaltyRatioM:          DefaultPenaltyRatioM,
		PenaltyRatioL:          DefaultPenaltyRatioL,
		LazinessWindow:         DefaultLazinessWindow,
		LazinessThreshold:      DefaultLazinessThreshold,
		HibernateThreshold:     DefaultHibernateThreshold,
		HibernatePeriod:        DefaultHibernatePeriod,
		BlockBindingWindow:     DefaultBlockBindingWindow,
		LockupPeriod:           DefaultLockupPeriod,
		DraftOpenCount:         DefaultDraftOpenCount,
		DraftCloseCount:        DefaultDraftCloseCount,
		DraftApplyCount:        DefaultDraftApplyCount,
		DraftQuorumRate:        DefaultDraftQuorumRate,
		DraftPassRate:          DefaultDraftPassRate,
		DraftRefundRate:        DefaultDraftRefundRate,
		UpgradeProtocolHeight:  DefaultUpgradeProtocolHeight,
		UpgradeProtocolVersion: DefaultUpgradeProtocolVersion,
	}

	tmp, err := new(Currency).SetString(DefaultMinStakingUnit, 10)
	if err != nil {
		return cfg, err
	}
	cfg.MinStakingUnit = *tmp

	tmp, err = new(Currency).SetString(DefaultBlkReward, 10)
	if err != nil {
		return cfg, err
	}
	cfg.BlkReward = *tmp

	tmp, err = new(Currency).SetString(DefaultTxReward, 10)
	if err != nil {
		return cfg, err
	}
	cfg.TxReward = *tmp

	tmp, err = new(Currency).SetString(DefaultDraftDeposit, 10)
	if err != nil {
		return cfg, err
	}
	cfg.DraftDeposit = *tmp

	return cfg, nil
}

type AMOAppConfigGenesis struct {
	MaxValidators          uint64   `json:"max_validators"`
	WeightValidator        float64  `json:"weight_validator"`
	WeightDelegator        float64  `json:"weight_delegator"`
	MinStakingUnit         Currency `json:"min_staking_unit"`
	BlkReward              Currency `json:"blk_reward"`
	TxReward               Currency `json:"tx_reward"`
	PenaltyRatioM          float64  `json:"penalty_ratio_m"`
	PenaltyRatioL          float64  `json:"penalty_ratio_l"`
	LazinessCounterWindow  int64    `json:"laziness_counter_window"`
	LazinessThreshold      float64  `json:"laziness_threshold"`
	BlockBindingWindow     int64    `json:"block_binding_window"`
	LockupPeriod           int64    `json:"lockup_period"`
	DraftOpenCount         int64    `json:"draft_open_count"`
	DraftCloseCount        int64    `json:"draft_close_count"`
	DraftApplyCount        int64    `json:"draft_apply_count"`
	DraftDeposit           Currency `json:"draft_deposit"`
	DraftQuorumRate        float64  `json:"draft_quorum_rate"`
	DraftPassRate          float64  `json:"draft_pass_rate"`
	DraftRefundRate        float64  `json:"draft_refund_rate"`
	UpgradeProtocolHeight  int64    `json:"upgrade_protocol_height"`
	UpgradeProtocolVersion uint64   `json:"upgrade_protocol_version"`
}

func (bCfg *AMOAppConfigGenesis) Migrate() AMOAppConfig {
	var cfg AMOAppConfig
	// explict config change process
	cfg.LazinessWindow = bCfg.LazinessCounterWindow
	cfg.LazinessThreshold = int64(float64(bCfg.LazinessCounterWindow) * bCfg.LazinessThreshold)
	cfg.HibernateThreshold = DefaultHibernateThreshold
	cfg.HibernatePeriod = DefaultHibernatePeriod

	cfg.MaxValidators = bCfg.MaxValidators
	cfg.WeightValidator = bCfg.WeightValidator
	cfg.WeightDelegator = bCfg.WeightDelegator
	cfg.MinStakingUnit = bCfg.MinStakingUnit
	cfg.BlkReward = bCfg.BlkReward
	cfg.TxReward = bCfg.TxReward
	cfg.PenaltyRatioM = bCfg.PenaltyRatioM
	cfg.PenaltyRatioL = bCfg.PenaltyRatioL
	cfg.BlockBindingWindow = bCfg.BlockBindingWindow
	cfg.LockupPeriod = bCfg.LockupPeriod
	cfg.DraftOpenCount = bCfg.DraftOpenCount
	cfg.DraftCloseCount = bCfg.DraftCloseCount
	cfg.DraftApplyCount = bCfg.DraftApplyCount
	cfg.DraftDeposit = bCfg.DraftDeposit
	cfg.DraftQuorumRate = bCfg.DraftQuorumRate
	cfg.DraftPassRate = bCfg.DraftPassRate
	cfg.DraftRefundRate = bCfg.DraftRefundRate
	cfg.UpgradeProtocolHeight = bCfg.UpgradeProtocolHeight
	cfg.UpgradeProtocolVersion = bCfg.UpgradeProtocolVersion
	return cfg
}

func (cfg *AMOAppConfig) Check(
	blockHeight int64,
	protocolVersion uint64,
	txCfgRaw json.RawMessage,
) (AMOAppConfig, error) {
	var txCfgMap map[string]interface{}

	// handle exception for allowing empty config field on purpose
	if len(txCfgRaw) == 0 {
		return *cfg, nil
	}

	cfgMap, err := cfg.getMap()
	if err != nil {
		return AMOAppConfig{}, err
	}

	err = json.Unmarshal(txCfgRaw, &txCfgMap)
	if err != nil {
		return AMOAppConfig{}, err
	}

	// check UpgradeProtocol* first
	if len(txCfgMap) == 2 && existUpgradeProtocolCfg(txCfgMap, true) {
		tmpCfg := *cfg
		err = json.Unmarshal(txCfgRaw, &tmpCfg)
		if err != nil {
			return AMOAppConfig{}, err
		}

		blockHeight += cfg.DraftOpenCount + cfg.DraftCloseCount + cfg.DraftApplyCount
		protocolVersion += uint64(1)

		if !(tmpCfg.UpgradeProtocolHeight > blockHeight) {
			return AMOAppConfig{}, fmt.Errorf("%d: improper upgrade protocol height",
				tmpCfg.UpgradeProtocolHeight,
			)
		}

		if !(tmpCfg.UpgradeProtocolVersion == protocolVersion) {
			return AMOAppConfig{}, fmt.Errorf("%d: improper upgrade protocol version",
				tmpCfg.UpgradeProtocolVersion,
			)
		}

		return tmpCfg, nil
	}

	// check other configs
	if existUpgradeProtocolCfg(txCfgMap, false) {
		return AMOAppConfig{}, fmt.Errorf("upgrade protocol config is included")
	}
	for key := range txCfgMap {
		if _, exist := cfgMap[key]; !exist {
			return AMOAppConfig{}, fmt.Errorf("%s doesn't exist in config map", key)
		}
	}

	tmpCfg := *cfg
	err = json.Unmarshal(txCfgRaw, &tmpCfg)
	if err != nil {
		return AMOAppConfig{}, err
	}

	if cmp(tmpCfg.MaxValidators, ">", uint64(0)) &&
		cmp(tmpCfg.WeightValidator, ">", float64(0)) &&
		cmp(tmpCfg.WeightDelegator, ">", float64(0)) &&
		cmp(tmpCfg.MinStakingUnit, ">", *Zero) &&
		cmp(tmpCfg.BlkReward, ">=", *Zero) &&
		cmp(tmpCfg.TxReward, ">=", *Zero) &&
		cmp(tmpCfg.PenaltyRatioM, ">", float64(0)) &&
		cmp(tmpCfg.PenaltyRatioL, ">", float64(0)) &&
		cmp(tmpCfg.LazinessWindow, ">=", int64(10000)) &&
		cmp(tmpCfg.LazinessThreshold, ">", int64(0)) &&
		cmp(tmpCfg.HibernateThreshold, ">", int64(0)) &&
		cmp(tmpCfg.HibernatePeriod, ">", int64(0)) &&
		cmp(tmpCfg.BlockBindingWindow, ">=", int64(10000)) &&
		cmp(tmpCfg.LockupPeriod, ">=", int64(10000)) &&
		cmp(tmpCfg.DraftOpenCount, ">=", int64(10000)) &&
		cmp(tmpCfg.DraftCloseCount, ">=", int64(10000)) &&
		cmp(tmpCfg.DraftApplyCount, ">=", int64(10000)) &&
		cmp(tmpCfg.DraftDeposit, ">=", *Zero) &&
		cmp(tmpCfg.DraftQuorumRate, ">", float64(0)) &&
		cmp(tmpCfg.DraftPassRate, ">", float64(0)) &&
		cmp(tmpCfg.DraftRefundRate, ">", float64(0)) {
		return tmpCfg, nil
	}

	return AMOAppConfig{}, fmt.Errorf("couldn't finish checking config values successfully")
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
	case Currency:
		a := targetA.(Currency)
		b := targetB.(Currency)

		equal = a.Equals(&b)
		greater = a.GreaterThan(&b)
		less = a.LessThan(&b)

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

	default:
		return false
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

func existUpgradeProtocolCfg(cfgMap map[string]interface{}, andOpt bool) bool {
	_, existHeight := cfgMap["upgrade_protocol_height"]
	_, existVersion := cfgMap["upgrade_protocol_version"]
	if andOpt {
		return existHeight && existVersion
	}
	return existHeight || existVersion
}
