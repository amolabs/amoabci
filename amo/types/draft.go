package types

import (
	"github.com/tendermint/tendermint/crypto"
)

type Draft struct {
	Proposer crypto.Address `json:"proposer"`
	Config   AMOAppConfig   `json:"config"`
	Desc     string         `json:"desc"`

	OpenCount  uint64   `json:"open_count"`
	CloseCount uint64   `json:"close_count"`
	ApplyCount uint64   `json:"apply_count"`
	Deposit    Currency `json:"deposit"`

	TallyQuorum  Currency `json:"tally_quorum"`
	TallyApprove Currency `json:"tally_approve"`
	TallyReject  Currency `json:"tally_reject"`
}

type DraftEx struct {
	Draft *Draft      `json:"draft"`
	Votes []*VoteInfo `json:"votes"`
}
