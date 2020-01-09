package amo

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type GenAmoAppState struct {
	Config   types.AMOAppConfig `json:"config"`
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
	err := json.Unmarshal(data, &genState)
	if err != nil {
		return nil, err
	}
	if genState.Config.MaxValidators == 0 {
		genState.Config.MaxValidators = defaultMaxValidators
	}
	if genState.Config.WeightValidator == 0 {
		genState.Config.WeightValidator = defaultWeightValidator
	}
	if genState.Config.WeightDelegator == 0 {
		genState.Config.WeightDelegator = defaultWeightDelegator
	}
	if genState.Config.MinStakingUnit == "" {
		genState.Config.MinStakingUnit = defaultMinStakingUnit
	}
	if genState.Config.BlkReward == "" {
		genState.Config.BlkReward = defaultBlkReward
	}
	if genState.Config.TxReward == "" {
		genState.Config.TxReward = defaultTxReward
	}
	if genState.Config.PenaltyRatioM == 0 {
		genState.Config.PenaltyRatioM = defaultPenaltyRatioM
	}
	if genState.Config.PenaltyRatioL == 0 {
		genState.Config.PenaltyRatioL = defaultPenaltyRatioL
	}
	if genState.Config.LazinessCounterWindow == 0 {
		genState.Config.LazinessCounterWindow = defaultLazinessCounterWindow
	}
	if genState.Config.LazinessThreshold == 0 {
		genState.Config.LazinessThreshold = defaultLazinessThreshold
	}
	if genState.Config.BlockBoundTxGracePeriod == 0 {
		genState.Config.BlockBoundTxGracePeriod = defaultBlockBoundTxGracePeriod
	}
	if genState.Config.LockupPeriod == 0 {
		genState.Config.LockupPeriod = defaultLockupPeriod
	}
	if genState.Config.DraftOpenCount == 0 {
		genState.Config.DraftOpenCount = defaultDraftOpenCount
	}
	if genState.Config.DraftCloseCount == 0 {
		genState.Config.DraftCloseCount = defaultDraftCloseCount
	}
	if genState.Config.DraftApplyCount == 0 {
		genState.Config.DraftApplyCount = defaultDraftApplyCount
	}
	if genState.Config.DraftDeposit == "" {
		genState.Config.DraftDeposit = defaultDraftDeposit
	}
	if genState.Config.DraftQuorumRate == 0 {
		genState.Config.DraftQuorumRate = defaultDraftQuorumRate
	}
	if genState.Config.DraftPassRate == 0 {
		genState.Config.DraftPassRate = defaultDraftPassRate
	}
	if genState.Config.DraftRefundRate == 0 {
		genState.Config.DraftRefundRate = defaultDraftRefundRate
	}

	return &genState, nil
}

func FillGenesisState(s *store.Store, genState *GenAmoAppState) error {
	err := s.Purge()
	if err != nil {
		return err
	}

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
