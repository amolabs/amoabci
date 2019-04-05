package types

import "github.com/tendermint/tendermint/crypto"

type Delegate struct {
	Holder    crypto.Address
	Amount    Currency       `json:"amount"`
	Delegator crypto.Address `json:"delegator"`
}
