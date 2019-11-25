package code

const (
	TxCodeOK uint32 = iota
	TxCodeBadParam
	TxCodeOldTx
	TxCodeNotEnoughBalance
	TxCodeSelfTransaction
	TxCodePermissionDenied
	TxCodeAlreadyGranted
	TxCodeAlreadyRegistered
	TxCodeParcelNotFound
	TxCodeBadSignature
	TxCodeRequestNotFound
	TxCodeMultipleDelegates
	TxCodeDelegateNotFound
	TxCodeNoStake
	TxCodeHeightTaken
	TxCodeBadValidator
	TxCodeLastValidator
	TxCodeDelegateExists
	TxCodeStakeLocked
	TxCodeUnknown
)

const (
	QueryCodeOK uint32 = iota
	QueryCodeBadPath
	QueryCodeNoKey
	QueryCodeBadKey
	QueryCodeNoMatch
)
