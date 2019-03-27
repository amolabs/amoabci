package types

import "github.com/tendermint/tendermint/crypto"

type DelegateValue struct {
	Amount    Currency       `json:"amount"`
	Delegator crypto.Address `json:"delegator"`
}
