package types

import (
	"github.com/tendermint/tendermint/crypto"
)

type RequestValue struct {
	Payment Currency `json:"payment"`
	Extra   Extra    `json:"extra,omitempty"`
}

type RequestValueEx struct {
	*RequestValue
	Buyer crypto.Address `json:"buyer"`
}
