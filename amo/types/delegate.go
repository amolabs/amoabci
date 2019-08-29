package types

import "github.com/tendermint/tendermint/crypto"

type Delegate struct {
	Delegatee crypto.Address `json:"delegatee"`
	Amount    Currency       `json:"amount"`
}

type DelegateEx struct {
	Delegator crypto.Address `json:"delegator"` // just for convenience
	*Delegate
}
