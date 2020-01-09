package types

import (
	"github.com/tendermint/tendermint/crypto"
)

type Vote struct {
	Approve bool     `json:"approve"`
	Power   Currency `json:"power"`
}

type VoteInfo struct {
	Voter  crypto.Address `json:"voter"`
	Record Vote           `json:"record"`
}
