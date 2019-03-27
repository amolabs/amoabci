package types

import "github.com/tendermint/tendermint/crypto"

type DelegateValue struct {
	To        Currency       `json:"to"`
	Delegator crypto.Address `json:"delegator"`
}
