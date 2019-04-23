package amo

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/version"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/operation"
	astore "github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

var (
	stateKey                         = []byte("stateKey")
	ProtocolVersion version.Protocol = 0x1
)

const (
	maxValidators = 100
	wValidator    = 2
	wDelegate     = 1
	blkRewardAMO  = uint64(0)
	txRewardAMO   = uint64(types.OneAMOUint64 / 10)
)

type State struct {
	db      dbm.DB
	Walk    int64  `json:"walk"`
	AppHash []byte `json:"app_hash"`
}

func loadState(db dbm.DB) State {
	stateBytes := db.Get(stateKey)
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

type AMOApplication struct {
	abci.BaseApplication

	state         State
	store         *astore.Store
	logger        log.Logger
	flagValUpdate bool
}

var _ abci.Application = (*AMOApplication)(nil)

func NewAMOApplication(db dbm.DB, index dbm.DB, l log.Logger) *AMOApplication {
	if l == nil {
		l = log.NewNopLogger()
	}
	if db == nil {
		db = dbm.NewMemDB()
	}
	if index == nil {
		index = dbm.NewMemDB()
	}
	app := &AMOApplication{
		state:  loadState(db),
		store:  astore.NewStore(db, index),
		logger: l,
	}
	return app
}

func (app *AMOApplication) Info(req abci.RequestInfo) (resInfo abci.ResponseInfo) {
	return abci.ResponseInfo{
		Data:       fmt.Sprintf("{\"walk\":%v}", app.state.Walk),
		Version:    version.ABCIVersion,
		AppVersion: ProtocolVersion.Uint64(),
	}
}

func (app *AMOApplication) DeliverTx(tx []byte) abci.ResponseDeliverTx {
	message, op, isStake := operation.ParseTx(tx)
	if !message.Verify() {
		return abci.ResponseDeliverTx{
			Code: code.TxCodeBadSignature,
			Tags: nil,
		}
	}
	resCode := op.Execute(app.store, message.Sender)
	if resCode != code.TxCodeOK {
		return abci.ResponseDeliverTx{
			Code: resCode,
		}
	}
	if isStake {
		app.flagValUpdate = true
	}
	app.state.Walk++
	return abci.ResponseDeliverTx{
		Code: resCode,
		Tags: []tm.KVPair{
			{Key: []byte("all"), Value: []byte("true")},
			{Key: []byte("tx.type"), Value: []byte(message.Type)},
			{Key: []byte("tx.sender"), Value: []byte(message.Sender.String())},
		},
	}
}

func (app *AMOApplication) CheckTx(tx []byte) abci.ResponseCheckTx {
	message, op, _ := operation.ParseTx(tx)
	if !message.Verify() {
		return abci.ResponseCheckTx{
			Code: code.TxCodeBadSignature,
		}
	}
	// TODO: implement signature verify logic
	return abci.ResponseCheckTx{
		Code: op.Check(app.store, message.Sender),
	}
}

func (app *AMOApplication) Commit() abci.ResponseCommit {
	b := make([]byte, 8)
	binary.PutVarint(b, app.state.Walk)
	app.state.AppHash = b

	saveState(app.state)
	return abci.ResponseCommit{Data: app.state.AppHash}
}

func (app *AMOApplication) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	switch reqQuery.Path {
	case "/balance":
		resQuery = queryBalance(app.store, reqQuery.Data)
	case "/stake":
		resQuery = queryStake(app.store, reqQuery.Data)
	case "/delegate":
		resQuery = queryDelegate(app.store, reqQuery.Data)
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

func (app *AMOApplication) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
	genAppState, err := ParseGenesisStateBytes(req.AppStateBytes)
	// TODO: use proper methods to inform error
	if err != nil {
		return abci.ResponseInitChain{}
	}
	if FillGenesisState(app.store, genAppState) != nil {
		return abci.ResponseInitChain{}
	}
	app.logger.Info("InitChain: new genesis app state applied.")

	return abci.ResponseInitChain{}
}

func (app *AMOApplication) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.flagValUpdate = false

	proposer := req.Header.GetProposerAddress()
	staker := app.store.GetHolderByValidator(proposer)
	numTxs := req.Header.GetNumTxs()

	// XXX no means to convey error to res
	app.DistributeReward(staker, numTxs)

	return res
}

func (app *AMOApplication) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	if app.flagValUpdate {
		app.flagValUpdate = false
		res.ValidatorUpdates = app.store.GetValidatorUpdates(maxValidators)
	}
	return res
}

/////////////////////////////////////

func (app *AMOApplication) DistributeReward(staker crypto.Address, numTxs int64) error {
	stake := app.store.GetStake(staker)
	if stake == nil {
		return errors.New("No stake to calculate reward.")
	}
	ds := app.store.GetDelegatesByDelegator(staker)

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
		b := app.store.GetBalance(d.Holder).Add(&tmp2)
		app.store.SetBalance(d.Holder, b)
		app.logger.Debug("Block reward",
			"delegate", hex.EncodeToString(d.Holder), "reward", tmp2.Int64())
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
