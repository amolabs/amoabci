package amo

const (
	TxCodeOK uint32 = iota
	TxCodeBadParam
	TxCodeNotEnoughBalance
	TxCodeAlreadyBought
	TxCodeSelfTransaction
)