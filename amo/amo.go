package amo

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/amolabs/amoabci/amo/code"
	adb "github.com/amolabs/amoabci/amo/db"
	"github.com/amolabs/amoabci/amo/operation"
	abci "github.com/amolabs/tendermint-amo/abci/types"
	dbm "github.com/amolabs/tendermint-amo/libs/db"
	"github.com/amolabs/tendermint-amo/version"
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
	state State
	store *adb.Store
}

var _ abci.Application = (*AMOApplication)(nil)

func NewAMOApplication(db dbm.DB) *AMOApplication {
	state := loadState(db)
	app := &AMOApplication{
		state: state,
		store: adb.NewStore("blockchain"),
	}
	b, _ := hex.DecodeString("E5DB787809EC89BBF972B0E6193D552A7D973AD7")
	app.store.SetBalance(b, 3000)
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
	resCode, tags := op.Execute(app.store, message.Signer)
	if resCode != code.TxCodeOK {
		return abci.ResponseDeliverTx{Code: resCode}
	}
	// TODO: change state
	switch message.Command {
	case operation.TxTransfer:
		app.state.Size += 1
	}
	return abci.ResponseDeliverTx{Code: resCode, Tags: tags}
}

func (app *AMOApplication) CheckTx(tx []byte) abci.ResponseCheckTx {
	message, op := operation.ParseTx(tx)
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
	if reqQuery.Prove {
		return
	} else {
		resQuery.Key = reqQuery.Data
		var value []byte
		switch len(resQuery.Key) {
		}
		if value != nil {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	}
}
