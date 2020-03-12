package tx

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type VoteParam struct {
	DraftID uint32 `json:"draft_id"`
	Approve bool   `json:"approve"`
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

func (t *TxVote) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseVoteParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	stakes := store.GetTopStakes(ConfigAMOApp.MaxValidators, t.GetSender(), false)
	if len(stakes) == 0 {
		return code.TxCodePermissionDenied, "no permission to vote", nil
	}

	draft := store.GetDraft(txParam.DraftID, false)
	if draft == nil {
		return code.TxCodeNonExistingDraft, "non-existing draft", nil
	}

	if bytes.Equal(draft.Proposer, t.GetSender()) {
		return code.TxCodeSelfTransaction, "proposer cannot vote on own draft", nil
	}

	if !(draft.OpenCount == 0 &&
		draft.CloseCount > 0 &&
		draft.ApplyCount > 0) {
		return code.TxCodeVoteNotOpened, "vote is not opened", nil
	}

	vote := store.GetVote(txParam.DraftID, t.GetSender(), false)
	if vote != nil {
		return code.TxCodeAlreadyVoted, "already voted", nil
	}

	store.SetVote(txParam.DraftID, t.GetSender(), &types.Vote{
		Approve: txParam.Approve,
	})

	return code.TxCodeOK, "ok", nil
}
