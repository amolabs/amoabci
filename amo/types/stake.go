package types

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type Stake struct {
	Validator ed25519.PubKeyEd25519 `json:"validator"`
	Amount    Currency              `json:"amount"`
}

type StakeEx struct {
	Validator string        `json:"validator"`
	Amount    Currency      `json:"amount"`
	Delegates []*DelegateEx `json:"delegates,omitempty"`
}
