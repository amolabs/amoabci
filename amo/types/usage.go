package types

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bytes"
)

type Usage struct {
	Custody bytes.HexBytes `json:"custody"`
	Extra   Extra          `json:"extra,omitempty"`
}

type UsageEx struct {
	*Usage
	Buyer crypto.Address `json:"buyer"`
}
