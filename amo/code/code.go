package code

import (
	"errors"
	"fmt"
)

// tx codes
const (
	TxCodeOK uint32 = iota
	TxCodeBadParam
	TxCodeImproperTx
	TxCodeInvalidAmount
	TxCodeNotEnoughBalance
	TxCodeSelfTransaction
	TxCodePermissionDenied
	TxCodeAlreadyRequested
	TxCodeAlreadyGranted
	TxCodeParcelNotFound
	TxCodeRequestNotFound
	TxCodeUsageNotFound
	TxCodeBadSignature
	TxCodeMultipleDelegates
	TxCodeDelegateNotFound
	TxCodeNoStake
	TxCodeImproperStakeAmount
	TxCodeHeightTaken
	TxCodeBadValidator
	TxCodeLastValidator
	TxCodeDelegateExists
	TxCodeStakeLocked
	TxCodeImproperDraftID
	TxCodeImproperDraftConfig
	TxCodeProposedDraft
	TxCodeNonExistingDraft
	TxCodeAnotherDraftInProcess
	TxCodeVoteNotOpen
	TxCodeAlreadyVoted
	TxCodeNoStorage
	TxCodeUDCNotFound
	TxCodeNotFound
	TxCodeUnknown uint32 = 1000
)

// query codes
const (
	QueryCodeOK      uint32 = 0
	QueryCodeBadPath uint32 = iota + 1001 // offset
	QueryCodeNoKey
	QueryCodeBadKey
	QueryCodeNoMatch
)

var errMap map[uint32]error = map[uint32]error{
	0: nil, // TxCodeOK, QueryCodeOK

	TxCodeBadParam:              errors.New("BadParam"),
	TxCodeImproperTx:            errors.New("ImproperTx"),
	TxCodeInvalidAmount:         errors.New("InvalidAmount"),
	TxCodeNotEnoughBalance:      errors.New("NotEnoughBalance"),
	TxCodeSelfTransaction:       errors.New("SelfTransaction"),
	TxCodePermissionDenied:      errors.New("PermissionDenied"),
	TxCodeAlreadyRequested:      errors.New("AlreadyRequested"),
	TxCodeAlreadyGranted:        errors.New("AlreadyGranted"),
	TxCodeParcelNotFound:        errors.New("ParcelNotFound"),
	TxCodeRequestNotFound:       errors.New("RequestNotFound"),
	TxCodeUsageNotFound:         errors.New("UsageNotFound"),
	TxCodeBadSignature:          errors.New("BadSignature"),
	TxCodeMultipleDelegates:     errors.New("MultipleDelegates"),
	TxCodeDelegateNotFound:      errors.New("DelegateNotFound"),
	TxCodeNoStake:               errors.New("NoStake"),
	TxCodeImproperStakeAmount:   errors.New("ImproperStakeAmount"),
	TxCodeHeightTaken:           errors.New("HeightTaken"),
	TxCodeBadValidator:          errors.New("BadValidator"),
	TxCodeLastValidator:         errors.New("LastValidator"),
	TxCodeDelegateExists:        errors.New("DelegateExists"),
	TxCodeStakeLocked:           errors.New("StakeLocked"),
	TxCodeImproperDraftID:       errors.New("ImproperDraftID"),
	TxCodeImproperDraftConfig:   errors.New("ImproperDraftConfig"),
	TxCodeProposedDraft:         errors.New("ProposedDraft"),
	TxCodeNonExistingDraft:      errors.New("NonExistingDraft"),
	TxCodeAnotherDraftInProcess: errors.New("AnotherDraftInProcess"),
	TxCodeVoteNotOpen:           errors.New("VoteNotOpen"),
	TxCodeAlreadyVoted:          errors.New("AlreadyVoted"),
	TxCodeNoStorage:             errors.New("NoStorage"),
	TxCodeUDCNotFound:           errors.New("UDCNotFound"),
	TxCodeNotFound:              errors.New("NotFound"),
	TxCodeUnknown:               errors.New("Unknown"),

	QueryCodeBadPath: errors.New("BadPath"),
	QueryCodeNoKey:   errors.New("NoKey"),
	QueryCodeBadKey:  errors.New("BadKey"),
	QueryCodeNoMatch: errors.New("NoMatch"),
}

func GetError(code uint32) error {
	if _, exist := errMap[code]; !exist {
		panic(fmt.Errorf("Error corresponding to Code %d doesn't exist", code))
	}
	return errMap[code]
}
