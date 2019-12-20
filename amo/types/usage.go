package types

import (
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type UsageValue struct {
	Custody cmn.HexBytes `json:"custody"`
	Extra   Extra        `json:"extra,omitempty"`
}

type UsageValueEx struct {
	*UsageValue
	Buyer crypto.Address `json:"buyer"`
}
