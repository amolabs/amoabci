package amo

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/amolabs/amoabci/amo/code"
	astore "github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/tx"
	"github.com/amolabs/amoabci/amo/types"
)

var (
	stateKey = []byte("stateKey") // TODO: remove this when applying merkle tree
)

const (
	// versions
	AMOAppVersion      = "v1.1.0-dev"
	AMOProtocolVersion = 0x2
	// hard-coded configs
	maxValidators = 100
	wValidator    = 2
	wDelegate     = 1
	blkRewardAMO  = uint64(0)
	txRewardAMO   = uint64(types.OneAMOUint64 / 10)
)

// Output are sorted by voting power.
func valUpdates(oldVals, newVals abci.ValidatorUpdates) abci.ValidatorUpdates {
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

// TODO: use 2-stage state
type State struct {
	Walk        int64 `json:"-"` // TODO: remove this
	height      int64 `json:"-"` // current block height
	appHash     []byte
	lastHeight  int64  `json:"last_height"`   // last completed block height
	lastAppHash []byte `json:"last_app_hash"` // TODO: use merkle tree
}

type AMOApp struct {
	abci.BaseApplication
	logger log.Logger

	stateDB dbm.DB
	indexDB dbm.DB
	state   State
	store   *astore.Store

	doValUpdate bool
	oldVals     abci.ValidatorUpdates
}

var _ abci.Application = (*AMOApp)(nil)

func NewAMOApp(sdb dbm.DB, idb dbm.DB, l log.Logger) *AMOApp {
	if l == nil {
		l = log.NewNopLogger()
	}
	if sdb == nil {
		sdb = dbm.NewMemDB()
	}
	if idb == nil {
		idb = dbm.NewMemDB()
	}
	app := &AMOApp{
		stateDB: sdb,
		indexDB: idb,
		store:   astore.NewStore(sdb, idb),
		logger:  l,
	}
	app.load()
	return app
}

func (app *AMOApp) load() {
	stateBytes := app.stateDB.Get(stateKey)
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &app.state)
		if err != nil {
			panic(err)
		}
	}
}

func (app *AMOApp) save() {
	stateBytes, err := json.Marshal(app.state)
	if err != nil {
		panic(err)
	}
	app.stateDB.Set(stateKey, stateBytes)
}

func (app *AMOApp) Info(req abci.RequestInfo) (resInfo abci.ResponseInfo) {
	return abci.ResponseInfo{
		Data:             fmt.Sprintf("{\"walk\":%v}", app.state.Walk),
		Version:          AMOAppVersion,
		AppVersion:       AMOProtocolVersion,
		LastBlockHeight:  app.state.lastHeight,
		LastBlockAppHash: app.state.lastAppHash,
	}
}

func (app *AMOApp) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
	genAppState, err := ParseGenesisStateBytes(req.AppStateBytes)
	// TODO: use proper methods to inform error
	if err != nil {
		return abci.ResponseInitChain{}
	}
	if FillGenesisState(app.store, genAppState) != nil {
		return abci.ResponseInitChain{}
	}
	app.state.Walk = 0 // TODO: Replace this with merkle tree
	b := make([]byte, 8)
	binary.PutVarint(b, app.state.Walk)
	app.state.lastHeight = 0
	app.state.lastAppHash = b

	app.save()
	app.logger.Info("InitChain: new genesis app state applied.")

	return abci.ResponseInitChain{
		Validators: app.store.GetValidators(maxValidators),
	}
}

// TODO: return proof also
func (app *AMOApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	switch reqQuery.Path {
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
	default:
		resQuery.Code = code.QueryCodeBadPath
	}

	app.logger.Debug("Query: "+reqQuery.Path, "query_data", reqQuery.Data)

	return resQuery
}

func (app *AMOApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.state.height = req.Header.Height
	app.doValUpdate = false
	app.oldVals = app.store.GetValidators(maxValidators)

	proposer := req.Header.GetProposerAddress()
	staker := app.store.GetHolderByValidator(proposer)
	numTxs := req.Header.GetNumTxs()

	// XXX no means to convey error to res
	app.DistributeReward(staker, numTxs)

	return res
}

// Invariant checks. Do not consider app's store.
// - check signature
// - check parameter format
func (app *AMOApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	message, op, _, err := tx.ParseTx(txBytes)
	if err != nil {
		return abci.ResponseCheckTx{
			Code:      code.TxCodeBadParam,
			Info:      err.Error(),
			Codespace: "amo",
		}
	}
	if !message.Verify() {
		return abci.ResponseCheckTx{
			Code:      code.TxCodeBadSignature,
			Info:      "Signature verification failed",
			Codespace: "amo",
		}
	}
	return abci.ResponseCheckTx{
		Code: op.Check(app.store, message.Sender),
	}
}

func (app *AMOApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {
	message, op, isStake, err := tx.ParseTx(txBytes)
	if err != nil {
		return abci.ResponseDeliverTx{
			Code:      code.TxCodeBadParam,
			Info:      err.Error(),
			Codespace: "amo",
		}
	}

	if !message.Verify() { // move to CheckTx()
		return abci.ResponseDeliverTx{
			Code:      code.TxCodeBadSignature,
			Info:      "Signature verification failed",
			Codespace: "amo",
		}
	}
	defTags := []tm.KVPair{
		{Key: []byte("tx.type"), Value: []byte(message.Type)},
		{Key: []byte("tx.sender"), Value: []byte(message.Sender.String())},
	}
	resCode, opTags := op.Execute(app.store, message.Sender)
	if resCode != code.TxCodeOK {
		return abci.ResponseDeliverTx{
			Code: resCode,
		}
	}
	if isStake {
		app.doValUpdate = true
	}
	app.state.Walk++
	tags := append(defTags, opTags...)
	return abci.ResponseDeliverTx{
		Code: resCode,
		Tags: tags,
	}
}

// TODO: use req.Height
func (app *AMOApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	if app.doValUpdate {
		app.doValUpdate = false
		newVals := app.store.GetValidators(maxValidators)
		res.ValidatorUpdates = valUpdates(app.oldVals, newVals)
	}
	// update appHash
	// TODO: use merkle tree
	app.state.appHash = make([]byte, 8)
	binary.PutVarint(app.state.appHash, app.state.Walk)
	return res
}

func (app *AMOApp) Commit() abci.ResponseCommit {
	app.state.lastAppHash = app.state.appHash
	app.state.lastHeight = app.state.height
	app.save()
	return abci.ResponseCommit{Data: app.state.lastAppHash}
}

/////////////////////////////////////

func (app *AMOApp) DistributeReward(staker crypto.Address, numTxs int64) error {
	stake := app.store.GetStake(staker)
	if stake == nil {
		return errors.New("No stake, no reward.")
	}
	ds := app.store.GetDelegatesByDelegatee(staker)

	var tmp, tmp2 types.Currency

	// total reward
	var rTotal, rTx types.Currency
	rTotal.Set(blkRewardAMO)
	rTx.Set(txRewardAMO)
	tmp.SetInt64(numTxs)
	tmp.Mul(&tmp.Int, &rTx.Int)
	rTotal.Add(&tmp)

	// weighted sum
	var wsum, w big.Int
	w.SetInt64(wValidator)
	wsum.Mul(&w, &stake.Amount.Int)
	w.SetInt64(wDelegate)
	for _, d := range ds {
		tmp.Mul(&w, &d.Amount.Int)
		wsum.Add(&wsum, &tmp.Int)
	}
	// individual rewards
	tmp.Set(0) // subtotal for delegate holders
	for _, d := range ds {
		tmp2 = *partialReward(wDelegate, &d.Amount.Int, &wsum, &rTotal)
		if !tmp2.Equals(new(types.Currency).Set(0)) {
			app.state.Walk++
		}
		tmp.Add(&tmp2)
		b := app.store.GetBalance(d.Delegator).Add(&tmp2)
		app.store.SetBalance(d.Delegator, b)
		app.logger.Debug("Block reward",
			"delegate", hex.EncodeToString(d.Delegator), "reward", tmp2.Int64())
	}
	tmp2.Int.Sub(&rTotal.Int, &tmp.Int)
	if !tmp2.Equals(new(types.Currency).Set(0)) {
		app.state.Walk++
	}
	b := app.store.GetBalance(staker).Add(&tmp2)
	app.store.SetBalance(staker, b)
	app.logger.Debug("Block reward",
		"proposer", hex.EncodeToString(staker)[:20], "reward", tmp2.Int64())

	return nil
}

/////////////////////////////////////

// r = (weight * stake / total) * base
// TODO: eliminate ambiguity in float computation
func partialReward(weight int64, stake, total *big.Int, base *types.Currency) *types.Currency {
	var wf, t1f, t2f big.Float
	wf.SetInt64(weight)
	t1f.SetInt(stake)
	t1f.Mul(&wf, &t1f)
	t2f.SetInt(total)
	t1f.Quo(&t1f, &t2f)
	t2f.SetInt(&base.Int)
	t1f.Mul(&t1f, &t2f)
	r := types.Currency{}
	t1f.Int(&r.Int)
	return &r
}
