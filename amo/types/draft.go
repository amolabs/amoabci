package types

import (
	"encoding/binary"
	"encoding/json"
	"strconv"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"
)

type Draft struct {
	Proposer crypto.Address  `json:"proposer"`
	Config   AMOAppConfig    `json:"config"`
	Desc     json.RawMessage `json:"desc"`

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
	Votes []*VoteInfo `josn:"votes"`
}

func ConvDraftIDFromHex(raw tm.HexBytes) (uint32, []byte, error) {
	var (
		draftIDStr       string
		draftIDUint      uint32
		draftIDByteArray []byte
	)

	err := json.Unmarshal(raw, &draftIDStr)
	if err != nil {
		return draftIDUint, draftIDByteArray, err
	}

	tmp, err := strconv.ParseUint(draftIDStr, 10, 32)
	if err != nil {
		return draftIDUint, draftIDByteArray, err
	}

	draftIDUint = uint32(tmp)

	draftIDByteArray = ConvDraftIDFromUint(draftIDUint)

	return draftIDUint, draftIDByteArray, nil
}

func ConvDraftIDFromUint(raw uint32) []byte {
	draftIDByteArray := make([]byte, 4)
	binary.BigEndian.PutUint32(draftIDByteArray, raw)

	return draftIDByteArray
}
