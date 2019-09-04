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
	TxCodeDelegationNotExists
	TxCodeNoStake
	TxCodeBadValidator
	TxCodeLastValidator
	TxCodeDelegateExists
)

const (
	QueryCodeOK uint32 = iota
	QueryCodeBadPath
	QueryCodeNoKey
	QueryCodeBadKey
	QueryCodeNoMatch
)
