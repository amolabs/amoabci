package amo

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
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
	state  State
	store  *astore.Store
	logger log.Logger
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
		var pub ed25519.PubKeyEd25519
		switch message.Type {
		case operation.TxStake:
			stake := op.(*operation.Stake)
			copy(pub[:], stake.Validator)
		case operation.TxWithdraw:
			pub = app.store.GetStake(message.Sig.PubKey.Address()).Validator
		case operation.TxDelegate:
			pub = app.store.GetStake(op.(*operation.Delegate).To).Validator
		case operation.TxRetract:
			pub = app.store.GetStake(op.(*operation.Retract).From).Validator
		}
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

func (app *AMOApplication) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
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
