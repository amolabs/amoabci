package code

const (
	TxCodeOK uint32 = iota
	TxCodeBadParam
	TxCodeTooOldTx
	TxCodeAlreadyProcessedTx
	TxCodeInvalidAmount
	TxCodeNotEnoughBalance
	TxCodeSelfTransaction
	TxCodePermissionDenied
	TxCodeAlreadyRegistered
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
	TxCodeVoteNotOpened
	TxCodeAlreadyVoted

	TxCodeUnknown
)

const (
	QueryCodeOK uint32 = iota
	QueryCodeBadPath
	QueryCodeNoKey
	QueryCodeBadKey
	QueryCodeNoMatch
)
