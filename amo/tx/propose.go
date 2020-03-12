package tx

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type ProposeParam struct {
	DraftID uint32          `json:"draft_id"`
	Config  json.RawMessage `json:"config,omitempty"`
	Desc    string          `json:"desc"`
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

func (t *TxPropose) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseProposeParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	stakes := store.GetTopStakes(ConfigAMOApp.MaxValidators, t.GetSender(), false)
	if len(stakes) == 0 {
		return code.TxCodePermissionDenied, "no permission to propose a draft", nil
	}

	if txParam.DraftID != StateNextDraftID {
		return code.TxCodeImproperDraftID, "improper draft ID", nil
	}

	latestDraftID := StateNextDraftID - uint32(1)
	latestDraft := store.GetDraft(latestDraftID, false)
	if latestDraft != nil {
		if !(latestDraft.OpenCount == 0 &&
			latestDraft.CloseCount == 0 &&
			latestDraft.ApplyCount == 0) {
			return code.TxCodeAnotherDraftInProcess, "another draft in process", nil
		}
	}

	draft := store.GetDraft(txParam.DraftID, false)
	if draft != nil {
		return code.TxCodeProposedDraft, "already proposed draft", nil
	}

	balance := store.GetBalance(t.GetSender(), false)
	if balance.LessThan(&ConfigAMOApp.DraftDeposit) {
		return code.TxCodeNotEnoughBalance, "not enough balance", nil
	}
	balance.Sub(&ConfigAMOApp.DraftDeposit)

	// config check
	cfg, err := ConfigAMOApp.Check(StateBlockHeight, StateProtocolVersion, t.Param.Config)
	if err != nil {
		return code.TxCodeImproperDraftConfig, err.Error(), nil
	}

	// set draft
	store.SetDraft(txParam.DraftID, &types.Draft{
		Proposer: t.GetSender(),
		Config:   cfg,
		Desc:     t.Param.Desc,

		OpenCount:  ConfigAMOApp.DraftOpenCount,
		CloseCount: ConfigAMOApp.DraftCloseCount,
		ApplyCount: ConfigAMOApp.DraftApplyCount,
		Deposit:    ConfigAMOApp.DraftDeposit,

		TallyQuorum:  *types.Zero,
		TallyApprove: *types.Zero,
		TallyReject:  *types.Zero,
	})

	// set sender balance
	store.SetBalance(t.GetSender(), balance)

	return code.TxCodeOK, "ok", nil
}
