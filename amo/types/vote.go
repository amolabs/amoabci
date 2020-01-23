package types

import (
	"github.com/tendermint/tendermint/crypto"
)

type Vote struct {
	Approve bool `json:"approve"`
}

type VoteInfo struct {
	Voter crypto.Address `json:"voter"`
	*Vote
}
