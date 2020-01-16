package tx

import (
	"encoding/json"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type VoteParam struct {
	DraftID tm.HexBytes `json:"draft_id"`
	Approve bool        `json:"approve"`
}

func parseVoteParam(raw []byte) (VoteParam, error) {
	var param VoteParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxVote struct {
	TxBase
	Param VoteParam `json:"-"`
}

var _ Tx = &TxVote{}

func (t *TxVote) Check() (uint32, string) {
	_, err := parseVoteParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}
	return code.TxCodeOK, "ok"
}

func (t *TxVote) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseVoteParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	stakes := store.GetTopStakes(ConfigAMOApp.MaxValidators, t.GetSender(), false)
	if len(stakes) == 0 {
		return code.TxCodePermissionDenied, "no permission to vote", nil
	}

	_, draftIDByteArray, err := types.ConvDraftIDFromHex(txParam.DraftID)
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	draft := store.GetDraft(draftIDByteArray, false)
	if draft == nil {
		return code.TxCodeNonExistingDraft, "non-existing draft", nil
	}

	if !(draft.OpenCount == 0 &&
		draft.CloseCount > 0 &&
		draft.ApplyCount > 0) {
		return code.TxCodeVoteNotOpened, "vote is not opened", nil
	}

	vote := store.GetVote(draftIDByteArray, t.GetSender(), false)
	if vote != nil {
		return code.TxCodeAlreadyVoted, "already voted", nil
	}

	store.SetVote(draftIDByteArray, t.GetSender(), &types.Vote{
		Approve: t.Param.Approve,
	})

	return code.TxCodeOK, "ok", nil
}
