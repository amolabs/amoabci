package types

import (
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type Usage struct {
	Custody cmn.HexBytes `json:"custody"`
	Extra   Extra        `json:"extra,omitempty"`
}

type UsageEx struct {
	*Usage
	Buyer crypto.Address `json:"buyer"`
}
