package amo

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tm "github.com/tendermint/tendermint/types"
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
	blkRewardAMO  = types.OneAMOUint64
	txRewardAMO   = uint64(types.OneAMOUint64 / 10)
)

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
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
	state := loadState(db)
	if l == nil {
		l = log.NewNopLogger()
	}
	app := &AMOApplication{
		state:  state,
		store:  astore.NewStore(db, index),
		logger: l,
	}
	return app
}

func (app *AMOApplication) Info(req abci.RequestInfo) (resInfo abci.ResponseInfo) {
	return abci.ResponseInfo{
		Data:       fmt.Sprintf("{\"size\":%v}", app.state.Size),
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
	resCode, tags := op.Execute(app.store, message.Sender)
	if resCode != code.TxCodeOK {
		return abci.ResponseDeliverTx{
			Code: resCode,
		}
	}
	// TODO: change state
	switch message.Type {
	case operation.TxTransfer:
		app.state.Size += 1
	}
	if isStake {
		app.flagValUpdate = true
	}
	return abci.ResponseDeliverTx{
		Code: resCode,
		Tags: tags,
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
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height += 1
	saveState(app.state)
	return abci.ResponseCommit{Data: appHash}
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

func (app *AMOApplication) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	app.flagValUpdate = false

	proposer := req.Header.GetProposerAddress()
	staker := app.store.GetHolderByValidator(proposer)
	stake := app.store.GetStake(staker)
	if stake == nil {
		return abci.ResponseBeginBlock{}
	}
	ds := app.store.GetDelegatesByDelegator(staker)

	var tmp, tmp2 types.Currency

	// total reward
	numTxs := req.Header.GetNumTxs()
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
	tmp.Set(0)
	for _, d := range ds {
		tmp2 = *partialReward(wDelegate, &d.Amount.Int, &wsum, &rTotal)
		tmp.Add(&tmp2)
		b := app.store.GetBalance(d.Holder).Add(&tmp2)
		app.store.SetBalance(d.Holder, b)
		app.logger.Debug("Block reward",
			"delegate", hex.EncodeToString(d.Holder), "reward", tmp2.Int64())
	}
	tmp2.Int.Sub(&rTotal.Int, &tmp.Int)
	b := app.store.GetBalance(staker).Add(&tmp2)
	app.store.SetBalance(staker, b)
	app.logger.Debug("Block reward",
		"proposer", hex.EncodeToString(staker), "reward", tmp2.Int64())

	return abci.ResponseBeginBlock{}
}

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

func (app *AMOApplication) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	if app.flagValUpdate {
		app.flagValUpdate = false

		var vals abci.ValidatorUpdates
		stakes := app.store.GetTopStakes(maxValidators)
		adjFactor := calcAdjustFactor(stakes)
		for _, stake := range stakes {
			key := abci.PubKey{ // TODO
				Type: "ed25519",
				Data: stake.Validator[:],
			}
			var power big.Int
			power.Rsh(&stake.Amount.Int, adjFactor)
			val := abci.ValidatorUpdate{
				PubKey: key,
				Power:  power.Int64(),
			}
			vals = append(vals, val)
		}
		res.ValidatorUpdates = vals
	}
	return res
}

func calcAdjustFactor(stakes []*types.Stake) uint {
	var vp big.Int
	max := (tm.MaxTotalVotingPower)
	var vps int64 = 0
	var shifts uint = 0
	for _, stake := range stakes {
		vp = stake.Amount.Int
		vp.Rsh(&vp, shifts)
		for !vp.IsInt64() {
			vp.Rsh(&vp, 1)
			shifts++
		}
		vpi := vp.Int64()
		tmp := vps + vpi
		if tmp < vps || tmp > max {
			vps >>= 1
			vpi >>= 1
			shifts++
			tmp = vps + vpi
		}
	}
	return shifts
}
