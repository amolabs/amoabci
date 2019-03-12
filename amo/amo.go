package amo

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/version"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/operation"
	astore "github.com/amolabs/amoabci/amo/store"
)

var (
	stateKey                         = []byte("stateKey")
	ProtocolVersion version.Protocol = 0x1
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

func NewAMOApplication(db dbm.DB, l log.Logger) *AMOApplication {
	state := loadState(db)
	if l == nil {
		l = log.NewNopLogger()
	}
	app := &AMOApplication{
		state:  state,
		store:  astore.NewStore(db),
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
	message, op := operation.ParseTx(tx)
	if !message.Verify() {
		return abci.ResponseDeliverTx{
			Code: code.TxCodeBadSignature,
			Tags: nil,
		}
	}
	resCode, tags := op.Execute(app.store, message.Signer)
	if resCode != code.TxCodeOK {
		return abci.ResponseDeliverTx{
			Code: resCode,
		}
	}
	// TODO: change state
	switch message.Command {
	case operation.TxTransfer:
		app.state.Size += 1
	}
	return abci.ResponseDeliverTx{
		Code: resCode,
		Tags: tags,
	}
}

func (app *AMOApplication) CheckTx(tx []byte) abci.ResponseCheckTx {
	message, op := operation.ParseTx(tx)
	if !message.Verify() {
		return abci.ResponseCheckTx{
			Code: code.TxCodeBadSignature,
		}
	}
	// TODO: implement signature verify logic
	return abci.ResponseCheckTx{
		Code: op.Check(app.store, message.Signer),
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
	resQuery.Key = reqQuery.Data

	switch reqQuery.Path {
	case "/balance":
		if len(reqQuery.Data) == 0 {
			resQuery.Code = code.QueryCodeNoKey
			break
		}

		var addr crypto.Address
		err := json.Unmarshal(reqQuery.Data, &addr)
		if err != nil {
			resQuery.Code = code.QueryCodeBadKey
			break
		}

		bal := app.store.GetBalance(addr)
		jsonstr, _ := json.Marshal(bal)
		resQuery.Log = string(jsonstr)
		// XXX: tendermint will convert this using base64 encoding
		resQuery.Value = []byte(jsonstr)
		resQuery.Code = code.QueryCodeOK
	case "/parcel":
		if len(reqQuery.Data) == 0 {
			resQuery.Code = code.QueryCodeNoKey
			break
		}

		// TODO: check parcel id
		parcel := app.store.GetParcel(reqQuery.Data)
		if parcel == nil {
			resQuery.Code = code.QueryCodeNoMatch
			break
		}

		jsonstr, _ := json.Marshal(parcel)
		resQuery.Log = string(jsonstr)
		resQuery.Value = []byte(jsonstr)
		resQuery.Code = code.QueryCodeOK
	case "/request":
		if len(reqQuery.Data) == 0 {
			resQuery.Code = code.QueryCodeNoKey
			break
		}

		keyMap := make(map[string]common.HexBytes)
		err := json.Unmarshal(reqQuery.Data, &keyMap)
		if err != nil {
			resQuery.Code = code.QueryCodeBadKey
			break
		}
		if _, ok := keyMap["buyer"]; !ok {
			resQuery.Code = code.QueryCodeBadKey
			break
		}
		if _, ok := keyMap["target"]; !ok {
			resQuery.Code = code.QueryCodeBadKey
			break
		}
		addr := crypto.Address(keyMap["buyer"])
		if len(addr) != crypto.AddressSize {
			resQuery.Code = code.QueryCodeBadKey
			break
		}

		// TODO: check parcel id
		parcelID := keyMap["target"]

		request := app.store.GetRequest(addr, parcelID)
		if request == nil {
			resQuery.Code = code.QueryCodeNoMatch
			break
		}
		jsonstr, _ := json.Marshal(request)
		resQuery.Log = string(jsonstr)
		resQuery.Value = []byte(jsonstr)
		resQuery.Code = code.QueryCodeOK
	case "/usage":
		if len(reqQuery.Data) == 0 {
			resQuery.Code = code.QueryCodeNoKey
			break
		}

		keyMap := make(map[string]common.HexBytes)
		err := json.Unmarshal(reqQuery.Data, &keyMap)
		if err != nil {
			resQuery.Code = code.QueryCodeBadKey
			break
		}
		if _, ok := keyMap["buyer"]; !ok {
			resQuery.Code = code.QueryCodeBadKey
			break
		}
		if _, ok := keyMap["target"]; !ok {
			resQuery.Code = code.QueryCodeBadKey
			break
		}
		addr := crypto.Address(keyMap["buyer"])
		if len(addr) != crypto.AddressSize {
			resQuery.Code = code.QueryCodeBadKey
			break
		}

		// TODO: check parcel id
		parcelID := keyMap["target"]

		request := app.store.GetUsage(addr, parcelID)
		if request == nil {
			resQuery.Code = code.QueryCodeNoMatch
			break
		}
		jsonstr, _ := json.Marshal(request)
		resQuery.Log = string(jsonstr)
		resQuery.Value = []byte(jsonstr)
		resQuery.Code = code.QueryCodeOK
	default:
		resQuery.Code = code.QueryCodeBadPath
	}

	app.logger.Debug("Query: "+reqQuery.Path, "query_data", reqQuery.Data) // debug

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
