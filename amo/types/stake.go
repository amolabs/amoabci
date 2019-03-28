package types

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type Stake struct {
	Amount    Currency              `json:"amount"`
	Validator ed25519.PubKeyEd25519 `json:"validator"`
}
