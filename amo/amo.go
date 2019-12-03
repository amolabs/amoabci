package amo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"
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
	AMOAppVersion      = "v1.1.0-dev"
	AMOProtocolVersion = 0x2
	// hard-coded configs
	defaultMaxValidators   = 100
	defaultWeightValidator = int64(2)
	defaultWeightDelegator = int64(1)

	defaultBlkReward = uint64(0)
	defaultTxReward  = uint64(types.OneAMOUint64 / 10)

	// TODO: not fixed default ratios yet
	defaultPenaltyRatioM = float64(0.3)
	defaultPenaltyRatioL = float64(0.3)

	defaultLazinessCounterSize  = int64(300)
	defaultLazinessCounterRatio = float64(0.8)

	defaultBlockBoundTxGracePeriod = uint64(1000)
	defaultLockupPeriod            = uint64(1000000)
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

type AMOAppConfig struct {
	MaxValidators   uint64 `json:"max_validators"`
	WeightValidator int64  `json:"weight_validator"`
	WeightDelegator int64  `json:"weight_delegator"`

	BlkReward uint64 `json:"blk_reward"`
	TxReward  uint64 `json:"tx_reward"`

	PenaltyRatioM float64 `json:"penalty_ratio_m"` // malicious validator
	PenaltyRatioL float64 `json:"penalty_ratio_l"` // lazy validators

	LazinessCounterSize  int64   `json:"laziness_counter_size"`
	LazinessCounterRatio float64 `json:"laziness_counter_ratio"`

	BlockBoundTxGracePeriod uint64 `json:"block_bound_tx_grace_period"`
	LockupPeriod            uint64 `json:"lockup_period"`
}

type AMOApp struct {
	// app scaffold
	abci.BaseApplication
	logger log.Logger

	// app config
	config AMOAppConfig

	// internal database
	merkleDB       tmdb.DB
	indexDB        tmdb.DB
	incentiveDB    tmdb.DB
	groupCounterDB tmdb.DB

	// state related variables
	stateFile *os.File
	state     State

	// abstraction of internal DBs to the outer world
	store *astore.Store

	// runtime temporary variables
	doValUpdate bool
	oldVals     abci.ValidatorUpdates

	// fee related variables
	staker          []byte
	feeAccumulated  types.Currency
	numDeliveredTxs int64

	pendingEvidences      []abci.Evidence
	pendingLazyValidators []crypto.Address

	lazinessCounter     blockchain.LazinessCounter
	blockBindingManager blockchain.BlockBindingManager
	replayPreventer     blockchain.ReplayPreventer
}

var _ abci.Application = (*AMOApp)(nil)

func NewAMOApp(stateFile *os.File, mdb, idxdb, incdb, gcdb tmdb.DB, l log.Logger) *AMOApp {
	if l == nil {
		l = log.NewNopLogger()
	}
	if mdb == nil {
		mdb = tmdb.NewMemDB()
	}
	if idxdb == nil {
		idxdb = tmdb.NewMemDB()
	}
	if incdb == nil {
		incdb = tmdb.NewMemDB()
	}
	if gcdb == nil {
		gcdb = tmdb.NewMemDB()
	}

	app := &AMOApp{
		logger: l,
		config: AMOAppConfig{ // TODO: read from config file
			defaultMaxValidators,
			defaultWeightValidator,
			defaultWeightDelegator,
			defaultBlkReward,
			defaultTxReward,
			defaultPenaltyRatioM,
			defaultPenaltyRatioL,
			defaultLazinessCounterSize,
			defaultLazinessCounterRatio,
			defaultBlockBoundTxGracePeriod,
			defaultLockupPeriod,
		},
		stateFile:      stateFile,
		state:          State{},
		merkleDB:       mdb,
		indexDB:        idxdb,
		incentiveDB:    incdb,
		groupCounterDB: gcdb,

		store: astore.NewStore(mdb, idxdb, incdb, gcdb),
	}

	// necessary to initialize state.json file
	app.save()

	// TODO: use something more elegant
	tx.ConfigLockupPeriod = app.config.LockupPeriod
	app.load()

	app.lazinessCounter = blockchain.NewLazinessCounter(
		app.store,
		app.state.LastHeight,
		app.state.CounterDue,
		app.config.LazinessCounterSize,
		app.config.LazinessCounterRatio,
	)

	app.blockBindingManager = blockchain.NewBlockBindingManager(
		app.config.BlockBoundTxGracePeriod,
		app.state.LastHeight,
	)

	app.replayPreventer = blockchain.NewReplayPreventer(
		app.store,
		app.config.BlockBoundTxGracePeriod,
		app.state.LastHeight,
	)

	return app
}

func (app *AMOApp) load() {
	err := app.state.LoadFrom(app.stateFile)
	if err != nil {
		panic(err)
	}
}

func (app *AMOApp) save() {
	err := app.state.SaveTo(app.stateFile)
	if err != nil {
		panic(err)
	}
}

func (app *AMOApp) Info(req abci.RequestInfo) (resInfo abci.ResponseInfo) {
	return abci.ResponseInfo{
		Data:             fmt.Sprintf("%x", app.state.LastAppHash),
		Version:          AMOAppVersion,
		AppVersion:       AMOProtocolVersion,
		LastBlockHeight:  app.state.LastHeight,
		LastBlockAppHash: app.state.LastAppHash,
	}
}

func (app *AMOApp) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
	genAppState, err := ParseGenesisStateBytes(req.AppStateBytes)
	// TODO: use proper methods to inform error
	if err != nil {
		return abci.ResponseInitChain{}
	}
	// fill state db
	if FillGenesisState(app.store, genAppState) != nil {
		return abci.ResponseInitChain{}
	}

	hash, version, err := app.store.Save()
	if err != nil {
		return abci.ResponseInitChain{}
	}

	app.state.MerkleVersion = version
	app.state.LastHeight = 0
	app.state.LastAppHash = hash

	// apply config
	// TODO: need to do something to make it better
	b := app.store.GetAppConfig()
	if b == nil {
		app.config = AMOAppConfig{
			MaxValidators:           defaultMaxValidators,
			WeightValidator:         defaultWeightValidator,
			WeightDelegator:         defaultWeightDelegator,
			BlkReward:               defaultBlkReward,
			TxReward:                defaultTxReward,
			PenaltyRatioM:           defaultPenaltyRatioM,
			PenaltyRatioL:           defaultPenaltyRatioL,
			LazinessCounterSize:     defaultLazinessCounterSize,
			LazinessCounterRatio:    defaultLazinessCounterRatio,
			BlockBoundTxGracePeriod: defaultBlockBoundTxGracePeriod,
			LockupPeriod:            defaultLockupPeriod,
		}
	} else {
		json.Unmarshal(b, &app.config)
	}
	tx.ConfigLockupPeriod = app.config.LockupPeriod

	app.lazinessCounter = blockchain.NewLazinessCounter(
		app.store,
		app.state.LastHeight,
		app.state.CounterDue,
		app.config.LazinessCounterSize,
		app.config.LazinessCounterRatio,
	)

	app.blockBindingManager = blockchain.NewBlockBindingManager(
		app.config.BlockBoundTxGracePeriod,
		app.state.LastHeight,
	)

	app.blockBindingManager.Update()

	app.replayPreventer = blockchain.NewReplayPreventer(
		app.store,
		app.config.BlockBoundTxGracePeriod,
		app.state.LastHeight,
	)

	app.replayPreventer.Update()

	app.save()
	app.logger.Info("InitChain: new genesis app state applied.")

	return abci.ResponseInitChain{
		Validators: app.store.GetValidators(app.config.MaxValidators, false),
	}
}

// TODO: return proof also
func (app *AMOApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	switch reqQuery.Path {
	case "/app_config":
		resQuery = queryAppConfig(app.config)
	case "/balance":
		resQuery = queryBalance(app.store, reqQuery.Data)
	case "/stake":
		resQuery = queryStake(app.store, reqQuery.Data)
	case "/delegate":
		resQuery = queryDelegate(app.store, reqQuery.Data)
	case "/validator":
		resQuery = queryValidator(app.store, reqQuery.Data)
	case "/parcel":
		resQuery = queryParcel(app.store, reqQuery.Data)
	case "/request":
		resQuery = queryRequest(app.store, reqQuery.Data)
	case "/usage":
		resQuery = queryUsage(app.store, reqQuery.Data)
	case "/inc_block":
		resQuery = queryBlockIncentives(app.store, reqQuery.Data)
	case "/inc_address":
		resQuery = queryAddressIncentives(app.store, reqQuery.Data)
	case "/inc":
		resQuery = queryIncentive(app.store, reqQuery.Data)
	default:
		resQuery.Code = code.QueryCodeBadPath
	}

	app.logger.Debug("Query: "+reqQuery.Path, "query_data", reqQuery.Data,
		"query_response", resQuery.GetLog())

	return resQuery
}

func (app *AMOApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.state.Height = req.Header.Height
	app.blockBindingManager.Update()
	app.replayPreventer.Update()

	app.doValUpdate = false
	app.oldVals = app.store.GetValidators(app.config.MaxValidators, false)

	proposer := req.Header.GetProposerAddress()

	app.staker = app.store.GetHolderByValidator(proposer, false)
	app.feeAccumulated = *new(types.Currency).Set(0)
	app.numDeliveredTxs = int64(0)

	app.pendingEvidences = req.GetByzantineValidators()
	app.pendingLazyValidators, app.state.CounterDue = app.lazinessCounter.Investigate(app.state.Height, req.GetLastCommitInfo())

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

	if !app.blockBindingManager.Check(t.GetLastHeight()) {
		return abci.ResponseCheckTx{
			Code:      code.TxCodeTooOldTx,
			Log:       "Binding tx to block failed",
			Info:      "Binding tx to block failed",
			Codespace: "amo",
		}
	}

	if !app.replayPreventer.Check(req.Tx) {
		return abci.ResponseCheckTx{
			Code:      code.TxCodeAlreadyProcessedTx,
			Log:       "Tx was processed before",
			Info:      "Tx was processed before",
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

	if !app.blockBindingManager.Check(t.GetLastHeight()) {
		return abci.ResponseDeliverTx{
			Code:      code.TxCodeTooOldTx,
			Log:       "Binding tx to block failed",
			Info:      "Binding tx to block failed",
			Codespace: "amo",
		}
	}

	if !app.replayPreventer.Append(req.Tx) {
		return abci.ResponseDeliverTx{
			Code:      code.TxCodeAlreadyProcessedTx,
			Log:       "Tx was processed before",
			Info:      "Tx was processed before",
			Codespace: "amo",
		}
	}

	tags := []tm.KVPair{
		{Key: []byte("tx.type"), Value: []byte(t.GetType())},
		{Key: []byte("tx.sender"), Value: []byte(t.GetSender().String())},
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

	rc, info, opTags := t.Execute(app.store)

	// if the operation was not successful,
	// change nothing and rollback the fee
	if rc == code.TxCodeOK {
		if t.GetType() == "stake" || t.GetType() == "withdraw" ||
			t.GetType() == "delegate" || t.GetType() == "retract" {
			app.doValUpdate = true
		}
		tags = append(tags, opTags...)
		app.numDeliveredTxs += 1
	} else {
		app.feeAccumulated.Sub(&fee)
		app.store.SetBalance(t.GetSender(), balance)
	}

	return abci.ResponseDeliverTx{
		Code: rc,
		Log:  info,
		Info: info,
		Events: []abci.Event{abci.Event{
			Type:       "default",
			Attributes: tags,
		}},
		Codespace: "amo",
	}
}

// TODO: use req.Height
func (app *AMOApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	// XXX no means to convey error to res
	blockchain.DistributeIncentive(
		app.store,
		app.logger,
		app.config.WeightValidator, app.config.WeightDelegator,
		app.config.BlkReward, app.config.TxReward,
		app.state.Height, app.numDeliveredTxs,
		app.staker,
		app.feeAccumulated,
	)

	if app.doValUpdate {
		app.doValUpdate = false
		newVals := app.store.GetValidators(app.config.MaxValidators, false)
		res.ValidatorUpdates = findValUpdates(app.oldVals, newVals)
	}

	app.store.LoosenLockedStakes(false)

	blockchain.PenalizeConvicts(
		app.store,
		app.logger,
		app.pendingEvidences,
		app.pendingLazyValidators,
		app.config.WeightValidator, app.config.WeightDelegator,
		app.config.PenaltyRatioM, app.config.PenaltyRatioL,
	)

	app.replayPreventer.Index()

	// update appHash
	hash := app.store.Root()
	if hash == nil {
		return abci.ResponseEndBlock{}
	}

	app.state.AppHash = hash

	return res
}

func (app *AMOApp) Commit() abci.ResponseCommit {
	hash, version, err := app.store.Save()
	if err != nil {
		return abci.ResponseCommit{}
	}

	// check if there are no state changes between EndBlock() and Commit()
	ok := bytes.Equal(hash, app.state.AppHash)
	if !ok {
		return abci.ResponseCommit{}
	}

	app.state.MerkleVersion = version
	app.state.LastAppHash = app.state.AppHash
	app.state.LastHeight = app.state.Height

	// apply config
	// TODO: need to do something to make it better
	b := app.store.GetAppConfig()
	if b == nil {
		app.config = AMOAppConfig{
			MaxValidators:           defaultMaxValidators,
			WeightValidator:         defaultWeightValidator,
			WeightDelegator:         defaultWeightDelegator,
			BlkReward:               defaultBlkReward,
			TxReward:                defaultTxReward,
			PenaltyRatioM:           defaultPenaltyRatioM,
			PenaltyRatioL:           defaultPenaltyRatioL,
			BlockBoundTxGracePeriod: defaultBlockBoundTxGracePeriod,
			LockupPeriod:            defaultLockupPeriod,
		}
	} else {
		json.Unmarshal(b, &app.config)
	}
	tx.ConfigLockupPeriod = app.config.LockupPeriod
	app.save()

	return abci.ResponseCommit{Data: app.state.LastAppHash}
}
