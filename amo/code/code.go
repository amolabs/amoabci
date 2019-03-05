package code

const (
	TxCodeOK uint32 = iota
	TxCodeBadParam
	TxCodeNotEnoughBalance
	TxCodeSelfTransaction
	TxCodePermissionDenied
	TxCodeTargetAlreadyBought
	TxCodeTargetAlreadyExists
	TxCodeTargetNotExists
	TxCodeBadSignature
	TxCodeRequestNotExists
)
