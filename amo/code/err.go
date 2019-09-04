package code

import (
	"errors"
)

var (
	TxErrBadParam            = errors.New("BadParam")
	TxErrNotEnoughBalance    = errors.New("NotEnoughBalance")
	TxErrSelfTransaction     = errors.New("SelfTransaction")
	TxErrPermissionDenied    = errors.New("PermissionDenied")
	TxErrAlreadyGranted      = errors.New("AlreadyGranted")
	TxErrAlreadyRegistered   = errors.New("AlreadyRegistered")
	TxErrParcelNotFound      = errors.New("ParcelNotFound")
	TxErrBadSignature        = errors.New("BadSignature")
	TxErrRequestNotFound     = errors.New("RequestNotFound")
	TxErrMultipleDelegates   = errors.New("MultipleDelegates")
	TxErrDelegationNotExists = errors.New("DelegationNotExists")
	TxErrNoStake             = errors.New("NoStake")
	TxErrBadValidator        = errors.New("BadValidator")
	TxErrLastValidator       = errors.New("LastValidator")
	TxErrDelegateExists      = errors.New("DelegateExists")
)
