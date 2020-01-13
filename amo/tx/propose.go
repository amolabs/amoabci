package tx

import (
	"encoding/binary"
	"encoding/json"
	"strconv"

	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type ProposeParam struct {
	DraftID tm.HexBytes        `json:"draft_id"`
	Config  types.AMOAppConfig `json:"config"`
	Desc    json.RawMessage    `json:"desc"`
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

	draftIDInt, draftIDByteArray, err := ConvDraftIDFromHex(txParam.DraftID)
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	if draftIDInt != StateNextDraftID {
		return code.TxCodeImproperDraftID, "improper draft ID", nil
	}

	latestDraftIDByteArray := ConvDraftIDFromUint(StateNextDraftID - 1)
	latestDraft := store.GetDraft(latestDraftIDByteArray, false)
	if latestDraft != nil {
		if !(latestDraft.DraftOpenCount == 0 &&
			latestDraft.DraftCloseCount == 0 &&
			latestDraft.DraftApplyCount == 0) {
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
	ok, cfg := ConfigAMOApp.Check(t.getPayload())
	if !ok {
		return code.TxCodeImproperDraftConfig, "improper config to apply", nil
	}

	// set draft
	store.SetDraft(draftIDByteArray, &types.Draft{
		Proposer: t.GetSender(),
		Config:   cfg,
		Desc:     t.Param.Desc,

		DraftOpenCount:  ConfigAMOApp.DraftOpenCount,
		DraftCloseCount: ConfigAMOApp.DraftCloseCount,
		DraftApplyCount: ConfigAMOApp.DraftApplyCount,
		DraftDeposit:    *deposit,

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
