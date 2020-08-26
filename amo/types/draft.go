package types

import (
	"encoding/json"

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

// DraftForQuery structure is an alternative one to contain AMOAppConfig
// properly without considering its internal structure while preventing errors
// driven from json.Unmarshal() function. That is, DraftForQuery structure can
// contain any type of AMOAppConfig regardless of its internal structure if it
// is well written in json format.
//
// NOTE: All of alive drafts(unclosed or unapplied ones) should be treated
// with Draft structure as it should follow AMOAppConfig speciffically set in
// its protocol. On the other hand, all of dead drafts(closed or applied ones)
// should be treated with DraftForQuery structure.

type DraftForQuery struct {
	Proposer crypto.Address  `json:"proposer"`
	Config   json.RawMessage `json:"config"`
	Desc     string          `json:"desc"`

	OpenCount  int64    `json:"open_count"`
	CloseCount int64    `json:"close_count"`
	ApplyCount int64    `json:"apply_count"`
	Deposit    Currency `json:"deposit"`

	TallyQuorum  Currency `json:"tally_quorum"`
	TallyApprove Currency `json:"tally_approve"`
	TallyReject  Currency `json:"tally_reject"`
}

type DraftEx struct {
	*DraftForQuery
	Votes []*VoteInfo `json:"votes"`
}
