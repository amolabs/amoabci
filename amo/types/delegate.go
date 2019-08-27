package types

import "github.com/tendermint/tendermint/crypto"

type Delegate struct {
	Delegator crypto.Address `json:"-"`
	Delegatee crypto.Address `json:"delegatee"`
	Amount    Currency       `json:"amount"`
}
