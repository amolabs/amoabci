package types

import (
	"github.com/tendermint/tendermint/crypto"
)

type Draft struct {
	Proposer crypto.Address `json:"proposer"`
	Config   AMOAppConfig   `json:"config"`
	Desc     string         `json:"desc"`

	OpenCount  int64    `json:"open_count"`
	CloseCount int64    `json:"close_count"`
	ApplyCount int64    `json:"apply_count"`
	Deposit    Currency `json:"deposit"`

	TallyQuorum  Currency `json:"tally_quorum"`
	TallyApprove Currency `json:"tally_approve"`
	TallyReject  Currency `json:"tally_reject"`
}

type DraftEx struct {
	*Draft
	Votes []*VoteInfo `json:"votes"`
}
