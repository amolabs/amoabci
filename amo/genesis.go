package amo

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type GenAmoAppState struct {
	State    State              `json:"state"`
	Config   types.AMOAppConfig `json:"-"`
	Balances []GenAccBalance    `json:"balances"`
	Stakes   []GenAccStake      `json:"stakes"`
}

type GenAccBalance struct {
	Owner  crypto.Address `json:"owner"`
	Amount types.Currency `json:"amount"`
}

type GenAccStake struct {
	Holder    crypto.Address `json:"holder"`
	Amount    types.Currency `json:"amount"`
	Validator []byte         `json:"validator"`
}

func ParseGenesisStateBytes(data []byte) (*GenAmoAppState, error) {
	genState := GenAmoAppState{}
	if len(data) > 0 {
		err := json.Unmarshal(data, &genState)
		if err != nil {
			return &genState, err
		}
	}
	if genState.State.ProtocolVersion == 0 {
		genState.State.ProtocolVersion = uint64(AMOGenesisProtocolVersion)
	}

	err := checkProtocolVersion(genState.State.ProtocolVersion, AMOProtocolVersion)
	if err != nil {
		panic(err)
	}

	// To avoid conflicts while unmarshaling config
	var genConfig struct {
		Config types.AMOAppConfig `json:"config"`
	}
	if len(data) > 0 {
		err := json.Unmarshal(data, &genConfig)
		if err != nil {
			return &genState, err
		}
	}
	genState.Config = genConfig.Config
	if genState.Config.MaxValidators == 0 {
		genState.Config.MaxValidators = types.DefaultMaxValidators
	}
	if genState.Config.WeightValidator == 0 {
		genState.Config.WeightValidator = types.DefaultWeightValidator
	}
	if genState.Config.WeightDelegator == 0 {
		genState.Config.WeightDelegator = types.DefaultWeightDelegator
	}
	if genState.Config.MinStakingUnit.Equals(types.Zero) {
		msu, err := new(types.Currency).SetString(types.DefaultMinStakingUnit, 10)
		if err != nil {
			return nil, err
		}
		genState.Config.MinStakingUnit = *msu
	}
	if genState.Config.BlkReward.Equals(types.Zero) {
		br, err := new(types.Currency).SetString(types.DefaultBlkReward, 10)
		if err != nil {
			return nil, err
		}
		genState.Config.BlkReward = *br
	}
	if genState.Config.TxReward.Equals(types.Zero) {
		tr, err := new(types.Currency).SetString(types.DefaultTxReward, 10)
		if err != nil {
			return nil, err
		}
		genState.Config.TxReward = *tr
	}
	if genState.Config.PenaltyRatioM == 0 {
		genState.Config.PenaltyRatioM = types.DefaultPenaltyRatioM
	}
	if genState.Config.PenaltyRatioL == 0 {
		genState.Config.PenaltyRatioL = types.DefaultPenaltyRatioL
	}
	if genState.Config.LazinessWindow == 0 {
		genState.Config.LazinessWindow = types.DefaultLazinessWindow
	}
	if genState.Config.LazinessThreshold == 0 {
		genState.Config.LazinessThreshold = types.DefaultLazinessThreshold
	}
	if genState.Config.HibernateThreshold == 0 {
		genState.Config.HibernateThreshold = types.DefaultHibernateThreshold
	}
	if genState.Config.HibernatePeriod == 0 {
		genState.Config.HibernatePeriod = types.DefaultHibernatePeriod
	}
	if genState.Config.BlockBindingWindow == 0 {
		genState.Config.BlockBindingWindow = types.DefaultBlockBindingWindow
	}
	if genState.Config.LockupPeriod == 0 {
		genState.Config.LockupPeriod = types.DefaultLockupPeriod
	}
	if genState.Config.DraftOpenCount == 0 {
		genState.Config.DraftOpenCount = types.DefaultDraftOpenCount
	}
	if genState.Config.DraftCloseCount == 0 {
		genState.Config.DraftCloseCount = types.DefaultDraftCloseCount
	}
	if genState.Config.DraftApplyCount == 0 {
		genState.Config.DraftApplyCount = types.DefaultDraftApplyCount
	}
	if genState.Config.DraftDeposit.Equals(types.Zero) {
		dd, err := new(types.Currency).SetString(types.DefaultDraftDeposit, 10)
		if err != nil {
			return nil, err
		}
		genState.Config.DraftDeposit = *dd
	}
	if genState.Config.DraftQuorumRate == 0 {
		genState.Config.DraftQuorumRate = types.DefaultDraftQuorumRate
	}
	if genState.Config.DraftPassRate == 0 {
		genState.Config.DraftPassRate = types.DefaultDraftPassRate
	}
	if genState.Config.DraftRefundRate == 0 {
		genState.Config.DraftRefundRate = types.DefaultDraftRefundRate
	}
	if genState.Config.UpgradeProtocolHeight == 0 {
		genState.Config.UpgradeProtocolHeight = types.DefaultUpgradeProtocolHeight
	}
	if genState.Config.UpgradeProtocolVersion == 0 {
		genState.Config.UpgradeProtocolVersion = types.DefaultUpgradeProtocolVersion
	}

	return &genState, nil
}

func FillGenesisState(st *State, s *store.Store, genState *GenAmoAppState) error {
	err := s.Purge()
	if err != nil {
		return err
	}

	// state
	st.ProtocolVersion = genState.State.ProtocolVersion

	// app config
	// TODO: use reflect package
	b, err := json.Marshal(genState.Config)
	if err != nil {
		return err
	}
	err = s.SetAppConfig(b)
	if err != nil {
		return err
	}

	// balances
	for _, accBal := range genState.Balances {
		s.SetBalance(accBal.Owner, &accBal.Amount)
	}

	// stakes
	for _, accStake := range genState.Stakes {
		var val25519 ed25519.PubKeyEd25519
		copy(val25519[:], accStake.Validator)
		s.SetUnlockedStake(accStake.Holder, &types.Stake{
			Amount:    accStake.Amount,
			Validator: val25519,
		})
	}

	return nil
}
