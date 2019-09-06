package code

const (
	TxCodeOK uint32 = iota
	TxCodeBadParam
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
	TxCodeBadValidator
	TxCodeLastValidator
	TxCodeDelegateExists
	TxCodeUnknown
)

const (
	QueryCodeOK uint32 = iota
	QueryCodeBadPath
	QueryCodeNoKey
	QueryCodeBadKey
	QueryCodeNoMatch
)
