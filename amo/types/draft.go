package types

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
)

type Draft struct {
	Proposer crypto.Address  `json:"proposer"`
	Config   AMOAppConfig    `json:"config"`
	Desc     json.RawMessage `json:"desc"`

	DraftOpenCount  uint64   `json:"draft_open_count"`
	DraftCloseCount uint64   `json:"draft_close_count"`
	DraftApplyCount uint64   `json:"draft_apply_count"`
	DraftDeposit    Currency `json:"draft_deposit"`

	TallyQuorum  Currency `json:"tally_quorum"`
	TallyApprove Currency `json:"tally_approve"`
	TallyReject  Currency `json:"tally_reject"`
}
