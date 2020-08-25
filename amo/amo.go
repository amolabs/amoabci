package amo

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/kv"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/blockchain"
	"github.com/amolabs/amoabci/amo/code"
	astore "github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/tx"
	"github.com/amolabs/amoabci/amo/types"
)

const (
	// versions
	AMOAppVersion      = "v1.7.0"
	AMOProtocolVersion = uint64(0x4)
)

// Output are sorted by voting power.
func findValUpdates(oldVals, newVals abci.ValidatorUpdates) abci.ValidatorUpdates {
	sort.Slice(oldVals, func(i, j int) bool {
		return bytes.Compare(oldVals[i].PubKey.Data, oldVals[j].PubKey.Data) < 0
	})
	sort.Slice(newVals, func(i, j int) bool {
		return bytes.Compare(newVals[i].PubKey.Data, newVals[j].PubKey.Data) < 0
	})

	// extract updates
	i := 0
	j := 0
	updates := abci.ValidatorUpdates{}
	for i < len(oldVals) && j < len(newVals) {
		comp := bytes.Compare(oldVals[i].PubKey.Data, newVals[j].PubKey.Data)
		if comp < 0 {
			updates = append(updates, abci.ValidatorUpdate{
				PubKey: oldVals[i].PubKey, Power: 0})
			i++
		} else if comp == 0 {
			updates = append(updates, newVals[j])
			i++
			j++
		} else {
			updates = append(updates, newVals[j])
			j++
		}
	}

	for ; i < len(oldVals); i++ {
		updates = append(updates, abci.ValidatorUpdate{
			PubKey: oldVals[i].PubKey, Power: 0})
	}

	for ; j < len(newVals); j++ {
		updates = append(updates, newVals[j])
	}

	sort.Slice(updates, func(i, j int) bool {
		// reverse order
		return updates[i].Power > updates[j].Power
	})
	return updates
}

type AMOApp struct {
	// app scaffold
	abci.BaseApplication
	logger log.Logger

	// app config
	config types.AMOAppConfig

	// state related variables
	stateFile *os.File
	state     State

	// abstraction of internal DBs to the outer world
	store               *astore.Store
	checkpoint_interval int64 // NOTE: this is a tentative workaround

	// runtime temporary variables
	doValUpdate bool
	oldVals     abci.ValidatorUpdates

	// fee related variables
	staker          []byte
	feeAccumulated  types.Currency
	numDeliveredTxs int64

	pendingEvidences      []abci.Evidence
	pendingLazyValidators []crypto.Address
	missingVals           []crypto.Address

	replayPreventer blockchain.ReplayPreventer
	missRuns        *blockchain.MissRuns
}

var _ abci.Application = (*AMOApp)(nil)

func NewAMOApp(checkpoint_interval int64, mdb, idxdb tmdb.DB, l log.Logger) *AMOApp {
	if l == nil {
		l = log.NewNopLogger()
	}
	if mdb == nil {
		mdb = tmdb.NewMemDB()
	}
	if idxdb == nil {
		idxdb = tmdb.NewMemDB()
	}

	s, err := astore.NewStore(l, checkpoint_interval, mdb, idxdb)
	if err != nil {
		panic(err)
	}

	app := &AMOApp{
		logger:              l,
		state:               State{},
		store:               s,
		checkpoint_interval: checkpoint_interval,
	}

	// load state, db and config
	app.load()

	// TODO: use something more elegant
	tx.ConfigAMOApp = app.config
	tx.StateNextDraftID = app.state.NextDraftID
	tx.StateProtocolVersion = app.state.ProtocolVersion

	app.missRuns = blockchain.NewMissRuns(
		app.store,
		tmdb.NewMemDB(),
		app.config.HibernateThreshold,
		app.config.HibernatePeriod,
		app.config.LazinessWindow,
	)

	app.replayPreventer = blockchain.NewReplayPreventer(
		app.store,
		app.state.LastHeight,
		app.config.BlockBindingWindow,
	)

	return app
}

const (
	// hard-coded configs
	defaultMaxValidators   = uint64(100)
	defaultWeightValidator = float64(2)
	defaultWeightDelegator = float64(1)

	defaultMinStakingUnit = "1000000000000000000000000"

	defaultBlkReward = "0"
	defaultTxReward  = "10000000000000000000"

	// TODO: not fixed default ratios yet
	defaultPenaltyRatioM = float64(0.3)
	defaultPenaltyRatioL = float64(0.3)

	defaultLazinessWindow     = int64(10000)
	defaultLazinessThreshold  = int64(8000)
	defaultHibernateThreshold = int64(100)
	defaultHibernatePeriod    = int64(10000)

	defaultBlockBindingWindow = int64(10000)
	defaultLockupPeriod       = int64(1000000)

	defaultDraftOpenCount  = int64(10000)
	defaultDraftCloseCount = int64(10000)
	defaultDraftApplyCount = int64(10000)
	defaultDraftDeposit    = "1000000000000000000000000"
	defaultDraftQuorumRate = float64(0.3)
	defaultDraftPassRate   = float64(0.51)
	defaultDraftRefundRate = float64(0.2)

	defaultUpgradeProtocolHeight  = int64(1)
	defaultUpgradeProtocolVersion = AMOProtocolVersion
)

func (app *AMOApp) loadAppConfig() error {
	cfg := types.AMOAppConfig{
		MaxValidators:          defaultMaxValidators,
		WeightValidator:        defaultWeightValidator,
		WeightDelegator:        defaultWeightDelegator,
		PenaltyRatioM:          defaultPenaltyRatioM,
		PenaltyRatioL:          defaultPenaltyRatioL,
		LazinessWindow:         defaultLazinessWindow,
		LazinessThreshold:      defaultLazinessThreshold,
		HibernateThreshold:     defaultHibernateThreshold,
		HibernatePeriod:        defaultHibernatePeriod,
		BlockBindingWindow:     defaultBlockBindingWindow,
		LockupPeriod:           defaultLockupPeriod,
		DraftOpenCount:         defaultDraftOpenCount,
		DraftCloseCount:        defaultDraftCloseCount,
		DraftApplyCount:        defaultDraftApplyCount,
		DraftQuorumRate:        defaultDraftQuorumRate,
		DraftPassRate:          defaultDraftPassRate,
		DraftRefundRate:        defaultDraftRefundRate,
		UpgradeProtocolHeight:  defaultUpgradeProtocolHeight,
		UpgradeProtocolVersion: defaultUpgradeProtocolVersion,
	}

	tmp, err := new(types.Currency).SetString(defaultMinStakingUnit, 10)
	if err != nil {
		return err
	}
	cfg.MinStakingUnit = *tmp

	tmp, err = new(types.Currency).SetString(defaultBlkReward, 10)
	if err != nil {
		return err
	}
	cfg.BlkReward = *tmp

	tmp, err = new(types.Currency).SetString(defaultTxReward, 10)
	if err != nil {
		return err
	}
	cfg.TxReward = *tmp

	tmp, err = new(types.Currency).SetString(defaultDraftDeposit, 10)
	if err != nil {
		return err
	}
	cfg.DraftDeposit = *tmp

	b := app.store.GetAppConfig()

	// if config exists
	if len(b) > 0 {
		// TODO: remove these lines at v1.7.2
		if cfg.UpgradeProtocolHeight != defaultUpgradeProtocolHeight &&
			app.state.Height == cfg.UpgradeProtocolHeight {
			var bCfg struct {
				MaxValidators          uint64         `json:"max_validators"`
				WeightValidator        float64        `json:"weight_validator"`
				WeightDelegator        float64        `json:"weight_delegator"`
				MinStakingUnit         types.Currency `json:"min_staking_unit"`
				BlkReward              types.Currency `json:"blk_reward"`
				TxReward               types.Currency `json:"tx_reward"`
				PenaltyRatioM          float64        `json:"penalty_ratio_m"`
				PenaltyRatioL          float64        `json:"penalty_ratio_l"`
				LazinessCounterWindow  int64          `json:"laziness_counter_window"`
				LazinessThreshold      float64        `json:"laziness_threshold"`
				BlockBindingWindow     int64          `json:"block_binding_window"`
				LockupPeriod           int64          `json:"lockup_period"`
				DraftOpenCount         int64          `json:"draft_open_count"`
				DraftCloseCount        int64          `json:"draft_close_count"`
				DraftApplyCount        int64          `json:"draft_apply_count"`
				DraftDeposit           types.Currency `json:"draft_deposit"`
				DraftQuorumRate        float64        `json:"draft_quorum_rate"`
				DraftPassRate          float64        `json:"draft_pass_rate"`
				DraftRefundRate        float64        `json:"draft_refund_rate"`
				UpgradeProtocolHeight  int64          `json:"upgrade_protocol_height"`
				UpgradeProtocolVersion uint64         `json:"upgrade_protocol_version"`
			}
			err = json.Unmarshal(b, &bCfg)
			if err != nil {
				return err
			}

			// explict config change process
			cfg.LazinessWindow = bCfg.LazinessCounterWindow
			cfg.LazinessThreshold = int64(float64(bCfg.LazinessCounterWindow) * bCfg.LazinessThreshold)
			cfg.HibernateThreshold = defaultHibernateThreshold
			cfg.HibernatePeriod = defaultHibernatePeriod

			cfg.MaxValidators = bCfg.MaxValidators
			cfg.WeightValidator = bCfg.WeightValidator
			cfg.WeightDelegator = bCfg.WeightDelegator
			cfg.MinStakingUnit = bCfg.MinStakingUnit
			cfg.BlkReward = bCfg.BlkReward
			cfg.TxReward = bCfg.TxReward
			cfg.PenaltyRatioM = bCfg.PenaltyRatioM
			cfg.PenaltyRatioL = bCfg.PenaltyRatioL
			cfg.BlockBindingWindow = bCfg.BlockBindingWindow
			cfg.LockupPeriod = bCfg.LockupPeriod
			cfg.DraftOpenCount = bCfg.DraftOpenCount
			cfg.DraftCloseCount = bCfg.DraftCloseCount
			cfg.DraftApplyCount = bCfg.DraftApplyCount
			cfg.DraftDeposit = bCfg.DraftDeposit
			cfg.DraftQuorumRate = bCfg.DraftQuorumRate
			cfg.DraftPassRate = bCfg.DraftPassRate
			cfg.DraftRefundRate = bCfg.DraftRefundRate
			cfg.UpgradeProtocolHeight = bCfg.UpgradeProtocolHeight
			cfg.UpgradeProtocolVersion = bCfg.UpgradeProtocolVersion
		}
	}

	app.config = cfg

	return nil
}

func (app *AMOApp) load() {
	_, err := app.store.Load()
	if err != nil {
		panic(err)
	}

	err = app.loadAppConfig()
	if err != nil {
		panic(err)
	}

	app.state.LoadFrom(app.store, app.config)

	app.store.RebuildIndex()
}

func (app *AMOApp) upgradeProtocol() []abci.Event {
	events := []abci.Event{}
	if app.state.Height != app.config.UpgradeProtocolHeight {
		return events
	}
	app.state.ProtocolVersion = app.config.UpgradeProtocolVersion
	versionJson, _ := json.Marshal(app.state.ProtocolVersion)
	events = append(events, abci.Event{
		Type: "protocol_upgrade",
		Attributes: []kv.Pair{
			{Key: []byte("version"), Value: versionJson},
		},
	})

	return events
}

func (app *AMOApp) checkProtocolVersion() error {
	if app.state.ProtocolVersion != AMOProtocolVersion {
		return fmt.Errorf("software protocol version(%d) doesn't "+
			"match state protocol version(%d)",
			AMOProtocolVersion, app.state.ProtocolVersion)
	}

	return nil
}

func (app *AMOApp) Info(req abci.RequestInfo) (resInfo abci.ResponseInfo) {
	return abci.ResponseInfo{
		Data:             fmt.Sprintf("%x", app.state.LastAppHash),
		Version:          AMOAppVersion,
		AppVersion:       0, // TODO: would get updated if tendermint supports it
		LastBlockHeight:  app.state.LastHeight,
		LastBlockAppHash: app.state.LastAppHash,
	}
}

func (app *AMOApp) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
	genAppState, err := ParseGenesisStateBytes(req.AppStateBytes)
	// TODO: use proper methods to inform error
	if err != nil {
		panic(err)
	}
	// fill state db
	err = FillGenesisState(&app.state, app.store, genAppState)
	if err != nil {
		panic(err)
	}

	hash, version, err := app.store.Save()
	if err != nil {
		panic(err)
	}

	app.state.LastHeight = version - 1
	app.state.LastAppHash = hash
	app.state.NextDraftID = uint32(1)

	err = app.loadAppConfig()
	if err != nil {
		panic(err)
	}

	tx.ConfigAMOApp = app.config
	tx.StateNextDraftID = app.state.NextDraftID
	tx.StateProtocolVersion = app.state.ProtocolVersion

	app.missRuns = blockchain.NewMissRuns(
		app.store,
		tmdb.NewMemDB(),
		app.config.HibernateThreshold,
		app.config.HibernatePeriod,
		app.config.LazinessWindow,
	)

	app.replayPreventer = blockchain.NewReplayPreventer(
		app.store,
		app.state.LastHeight,
		app.config.BlockBindingWindow,
	)

	app.logger.Info("InitChain: new genesis app state applied.")

	return abci.ResponseInitChain{
		Validators: app.store.GetValidators(app.config.MaxValidators, false),
	}
}

// TODO: return proof also
func (app *AMOApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	reqs := strings.Split(reqQuery.Path, "/")
	if len(reqs) > 1 {
		reqs = append(reqs[:0], reqs[1:]...) // remove empty string
	}

	if len(reqs) == 0 || len(reqs) > 2 {
		resQuery.Code = code.QueryCodeBadPath
		return resQuery
	}

	switch reqs[0] {
	case "config":
		resQuery = queryAppConfig(app.config)
	case "balance":
		switch len(reqs) {
		case 1:
			resQuery = queryBalance(app.store, "", reqQuery.Data)
		case 2:
			resQuery = queryBalance(app.store, reqs[1], reqQuery.Data)
		default:
			resQuery.Code = code.QueryCodeBadPath
			return resQuery
		}
	case "udc":
		resQuery = queryUDC(app.store, reqQuery.Data)
	case "udclock":
		if len(reqs) != 2 {
			resQuery.Code = code.QueryCodeBadPath
			return resQuery
		}
		resQuery = queryUDCLock(app.store, reqs[1], reqQuery.Data)
	case "stake":
		resQuery = queryStake(app.store, reqQuery.Data)
	case "delegate":
		resQuery = queryDelegate(app.store, reqQuery.Data)
	case "validator":
		resQuery = queryValidator(app.store, reqQuery.Data)
	case "hibernate":
		resQuery = queryHibernate(app.store, reqQuery.Data)
	case "storage":
		resQuery = queryStorage(app.store, reqQuery.Data)
	case "draft":
		resQuery = queryDraft(app.store, reqQuery.Data)
	case "vote":
		resQuery = queryVote(app.store, reqQuery.Data)
	case "parcel":
		resQuery = queryParcel(app.store, reqQuery.Data)
	case "request":
		resQuery = queryRequest(app.store, reqQuery.Data)
	case "usage":
		resQuery = queryUsage(app.store, reqQuery.Data)
	case "did":
		resQuery = queryDIDEntry(app.store, reqQuery.Data)
	default:
		resQuery.Code = code.QueryCodeBadPath
		return resQuery
	}

	app.logger.Debug("Query: "+reqQuery.Path, "query_data", reqQuery.Data,
		"query_response", resQuery.GetLog())

	return resQuery
}

func (app *AMOApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.state.Height = req.Header.Height
	tx.StateBlockHeight = app.state.Height
	tx.StateProtocolVersion = app.state.ProtocolVersion

	// upgrade protocol version
	evs := app.upgradeProtocol()
	res.Events = append(res.Events, evs...)

	// check if app's protocol version matches supported version
	err := app.checkProtocolVersion()
	if err != nil {
		panic(err)
	}

	// migrate to 4
	app.MigrateTo4()

	app.doValUpdate = false
	app.oldVals = app.store.GetValidators(app.config.MaxValidators, false)

	proposer := req.Header.GetProposerAddress()

	app.staker = app.store.GetHolderByValidator(proposer, false)
	app.feeAccumulated = *new(types.Currency).Set(0)
	app.numDeliveredTxs = int64(0)

	// blockchain modules
	app.replayPreventer.Update(app.state.Height, app.config.BlockBindingWindow)
	app.pendingEvidences = req.GetByzantineValidators()

	lci := req.GetLastCommitInfo()
	app.missingVals = []crypto.Address{}
	for _, v := range lci.GetVotes() {
		if !v.GetSignedLastBlock() {
			app.missingVals = append(app.missingVals, v.Validator.Address)
		}
	}

	return res
}

// Invariant checks. Do not consider app's store.
// - check signature
// - check parameter format
// - check availability of binding tx to block
// - check replay attack of txs which were processed before
func (app *AMOApp) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	t, err := tx.ParseTx(req.Tx)
	if err != nil {
		return abci.ResponseCheckTx{
			Code:      code.TxCodeBadParam,
			Log:       err.Error(),
			Info:      err.Error(),
			Codespace: "amo",
		}
	}

	if req.Type == abci.CheckTxType_New {
		if !t.Verify() {
			return abci.ResponseCheckTx{
				Code:      code.TxCodeBadSignature,
				Log:       "Signature verification failed",
				Info:      "Signature verification failed",
				Codespace: "amo",
			}
		}
	}

	err = app.replayPreventer.Check(req.Tx, t.GetLastHeight(), app.state.Height)
	if err != nil {
		return abci.ResponseCheckTx{
			Code:      code.TxCodeImproperTx,
			Log:       err.Error(),
			Info:      err.Error(),
			Codespace: "amo",
		}
	}

	rc, info := t.Check()

	return abci.ResponseCheckTx{
		Code:      rc,
		Log:       info,
		Info:      info,
		Codespace: "amo",
	}
}

func (app *AMOApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	t, err := tx.ParseTx(req.Tx)
	if err != nil {
		return abci.ResponseDeliverTx{
			Code:      code.TxCodeBadParam,
			Log:       err.Error(),
			Info:      err.Error(),
			Codespace: "amo",
		}
	}

	err = app.replayPreventer.Append(req.Tx, t.GetLastHeight(), app.state.Height)
	if err != nil {
		return abci.ResponseDeliverTx{
			Code:      code.TxCodeImproperTx,
			Log:       err.Error(),
			Info:      err.Error(),
			Codespace: "amo",
		}
	}

	typeJson, _ := json.Marshal(t.GetType())
	senderJson, _ := json.Marshal(t.GetSender())
	events := []abci.Event{
		{
			Type: "tx",
			Attributes: []kv.Pair{
				{Key: []byte("type"), Value: typeJson},
				{Key: []byte("sender"), Value: senderJson},
			},
		},
	}

	fee := t.GetFee()
	balance := app.store.GetBalance(t.GetSender(), false)

	if balance.LessThan(&fee) {
		return abci.ResponseDeliverTx{
			Code:      code.TxCodeNotEnoughBalance,
			Log:       "not enough balance to pay fee",
			Info:      "not enough balance to pay fee",
			Codespace: "amo",
		}
	}

	app.store.SetBalance(t.GetSender(), balance.Sub(&fee))
	app.feeAccumulated.Add(&fee)

	rc, info, opEvents := t.Execute(app.store)

	// if the operation was not successful,
	// change nothing and rollback the fee
	if rc == code.TxCodeOK {
		if t.GetType() == "stake" || t.GetType() == "withdraw" ||
			t.GetType() == "delegate" || t.GetType() == "retract" {
			app.doValUpdate = true
		}

		if t.GetType() == "propose" {
			app.state.NextDraftID += uint32(1)
		}

		events = append(events, opEvents...)
		app.numDeliveredTxs += 1

	} else {
		app.feeAccumulated.Sub(&fee)
		app.store.SetBalance(t.GetSender(), balance)
	}

	return abci.ResponseDeliverTx{
		Code:      rc,
		Log:       info,
		Info:      info,
		Events:    events,
		Codespace: "amo",
	}
}

func (app *AMOApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	// XXX no means to convey error to res

	// update miss runs
	doValUpdate, evs, err := app.missRuns.UpdateMissRuns(app.state.Height, app.missingVals)
	if err != nil {
		app.logger.Error(err.Error())
	}
	app.doValUpdate = app.doValUpdate || doValUpdate
	res.Events = append(res.Events, evs...)

	// wake up hibernating validators
	vals, hibs := app.store.GetHibernates(false)
	for i, hib := range hibs {
		if hib.End <= app.state.Height {
			app.store.DeleteHibernate(vals[i])
			addressJson, _ := json.Marshal(vals[i])
			ev := abci.Event{
				Type: "wakeup",
				Attributes: []kv.Pair{
					{Key: []byte("validator"), Value: addressJson},
				},
			}
			res.Events = append(res.Events, ev)
			app.doValUpdate = true
		}
	}

	evs, _ = blockchain.DistributeIncentive(
		app.store,
		app.logger,
		app.config.WeightValidator, app.config.WeightDelegator,
		app.config.BlkReward, app.config.TxReward,
		app.numDeliveredTxs,
		app.staker,
		app.feeAccumulated,
	)
	res.Events = append(res.Events, evs...)

	evs = app.store.LoosenLockedStakes(false)
	res.Events = append(res.Events, evs...)

	// get lazy validators
	lazyValidators := []crypto.Address{}
	if app.state.Height%app.config.LazinessWindow == 0 {
		missStat := app.missRuns.GetMissStat(
			app.state.Height-app.config.LazinessWindow+1,
			app.state.Height)
		for valString, count := range missStat {
			b, err := hex.DecodeString(valString)
			if err != nil {
				continue
			}
			val := crypto.Address(b)
			if count >= app.config.LazinessThreshold {
				lazyValidators = append(lazyValidators, val)
			}
		}
	}

	// penalize
	doValUpdate, evs, _ = blockchain.PenalizeConvicts(
		app.store,
		app.logger,
		app.pendingEvidences,
		lazyValidators,
		app.config.WeightValidator, app.config.WeightDelegator,
		app.config.PenaltyRatioM, app.config.PenaltyRatioL,
	)
	res.Events = append(res.Events, evs...)
	app.doValUpdate = app.doValUpdate || doValUpdate

	if app.doValUpdate {
		app.doValUpdate = false
		newVals := app.store.GetValidators(app.config.MaxValidators, false)
		res.ValidatorUpdates = findValUpdates(app.oldVals, newVals)
	}

	app.replayPreventer.Index(app.state.Height)

	evs = app.store.ProcessDraftVotes(
		app.state.NextDraftID-uint32(1),
		app.config.MaxValidators,
		app.config.DraftQuorumRate,
		app.config.DraftPassRate,
		app.config.DraftRefundRate,
		false,
	)
	res.Events = append(res.Events, evs...)

	return res
}

func (app *AMOApp) Commit() abci.ResponseCommit {
	hash, version, err := app.store.Save()
	if err != nil {
		return abci.ResponseCommit{}
	}

	app.state.LastAppHash = hash
	app.state.LastHeight = version - 1

	err = app.loadAppConfig()
	if err != nil {
		return abci.ResponseCommit{}
	}

	tx.ConfigAMOApp = app.config
	tx.StateNextDraftID = app.state.NextDraftID
	tx.StateProtocolVersion = app.state.ProtocolVersion

	return abci.ResponseCommit{Data: app.state.LastAppHash}
}

func (app *AMOApp) Close() {
	app.store.Close()
}
