package types

import (
	"github.com/tendermint/tendermint/crypto"
)

type Request struct {
	Payment Currency `json:"payment"`
	Extra   Extra    `json:"extra,omitempty"`
}

type RequestEx struct {
	*Request
	Buyer crypto.Address `json:"buyer"`
}
