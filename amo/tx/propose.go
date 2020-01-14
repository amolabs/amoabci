package tx

import (
	"encoding/json"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type ProposeParam struct {
	DraftID tm.HexBytes     `json:"draft_id"`
	Config  json.RawMessage `json:"config,omitempty"`
	Desc    json.RawMessage `json:"desc"`
}

func parseProposeParam(raw []byte) (ProposeParam, error) {
	var param ProposeParam

	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}

	return param, nil
}

type TxPropose struct {
	TxBase
	Param ProposeParam `json:"-"`
}

var _ Tx = &TxPropose{}

func (t *TxPropose) Check() (uint32, string) {
	_, err := parseProposeParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxPropose) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	txParam, err := parseProposeParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	stakes := store.GetTopStakes(ConfigAMOApp.MaxValidators, t.GetSender(), false)
	if len(stakes) == 0 {
		return code.TxCodePermissionDenied, "no permission to propose a draft", nil
	}

	draftIDInt, draftIDByteArray, err := types.ConvDraftIDFromHex(txParam.DraftID)
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if draftIDInt != StateNextDraftID {
		return code.TxCodeImproperDraftID, "improper draft ID", nil
	}

	latestDraftIDByteArray := types.ConvDraftIDFromUint(StateNextDraftID - 1)
	latestDraft := store.GetDraft(latestDraftIDByteArray, false)
	if latestDraft != nil {
		if !(latestDraft.OpenCount == 0 &&
			latestDraft.CloseCount == 0 &&
			latestDraft.ApplyCount == 0) {
			return code.TxCodeAnotherDraftInProcess, "another draft in process", nil
		}
	}

	draft := store.GetDraft(draftIDByteArray, false)
	if draft != nil {
		return code.TxCodeProposedDraft, "already proposed draft", nil
	}

	deposit, err := new(types.Currency).SetString(ConfigAMOApp.DraftDeposit, 10)
	if err != nil {
		return code.TxCodeImproperDraftDeposit, err.Error(), nil
	}

	balance := store.GetBalance(t.GetSender(), false)
	if balance.LessThan(deposit) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}

	balance.Sub(deposit)

	// config check
	ok, cfg := ConfigAMOApp.Check(t.Param.Config)
	if !ok {
		return code.TxCodeImproperDraftConfig, "improper config to apply", nil
	}

	// set draft
	store.SetDraft(draftIDByteArray, &types.Draft{
		Proposer: t.GetSender(),
		Config:   cfg,
		Desc:     t.Param.Desc,

		OpenCount:  ConfigAMOApp.DraftOpenCount,
		CloseCount: ConfigAMOApp.DraftCloseCount,
		ApplyCount: ConfigAMOApp.DraftApplyCount,
		Deposit:    *deposit,

		TallyQuorum:  *new(types.Currency).Set(0),
		TallyApprove: *new(types.Currency).Set(0),
		TallyReject:  *new(types.Currency).Set(0),
	})

	// set sender balance
	store.SetBalance(t.GetSender(), balance)

	// sender approve draft as proposer
	store.SetVote(draftIDByteArray, t.GetSender(), &types.Vote{
		Approve: true,
		Power:   *new(types.Currency).Set(0),
	})

	return code.TxCodeOK, "ok", nil
}
