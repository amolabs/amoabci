package amo

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	adb "github.com/amolabs/amoabci/amo/db"
	"github.com/amolabs/amoabci/amo/types"
	abci "github.com/amolabs/tendermint-amo/abci/types"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	dbm "github.com/amolabs/tendermint-amo/libs/db"
	"github.com/amolabs/tendermint-amo/version"
	"strconv"
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
	message, payload := types.ParseTx(tx)
	var tags []cmn.KVPair
	var resCode = TxCodeOK

	switch message.Command {
	case types.TxTransfer:
		transfer, _ := payload.(*types.Transfer)
		resCode, tags = app.procTransfer(transfer)
	case types.TxPurchase:
		purchase, _ := payload.(*types.Purchase)
		resCode, tags = app.procPurchase(purchase)
	}
	return abci.ResponseDeliverTx{Code: resCode, Tags: tags}
}

func (app *AMOApplication) procTransfer(transfer *types.Transfer) (uint32, []cmn.KVPair) {
	fromBalance := *app.store.GetBalance(transfer.From)
	toBalance := *app.store.GetBalance(transfer.To)
	fromBalance -= transfer.Amount
	toBalance += transfer.Amount
	app.store.SetBalance(transfer.From, fromBalance)
	app.store.SetBalance(transfer.To, toBalance)
	app.state.Size += 1
	tags := []cmn.KVPair{
		{Key: transfer.From, Value: []byte(strconv.FormatUint(uint64(fromBalance), 10))},
		{Key: transfer.To, Value: []byte(strconv.FormatUint(uint64(toBalance), 10))},
	}
	return TxCodeOK, tags
}

func (app *AMOApplication) procPurchase(purchase *types.Purchase) (uint32, []cmn.KVPair) {
	var metaData types.PDSNMetaData
	err := types.RequestMetaData(purchase.FileHash, &metaData)
	if err != nil {
		panic(err)
	}
	fromBalance := *app.store.GetBalance(purchase.From)
	fromBalance -= metaData.Price
	// TODO: modify ownership
	result, err := json.Marshal(metaData)
	if err != nil {
		panic(err)
	}
	tags := []cmn.KVPair{
		{Key: []byte(hex.EncodeToString(metaData.FileHash[:])), Value: result},
		{Key: purchase.From[:], Value: []byte(strconv.FormatUint(uint64(fromBalance), 10))},
	}
	return TxCodeOK, tags
}

func (app *AMOApplication) CheckTx(tx []byte) abci.ResponseCheckTx {
	message, payload := types.ParseTx(tx)
	var resCode = TxCodeOK

	switch message.Command {
	case types.TxTransfer:
		transfer, _ := payload.(*types.Transfer)
		fromBalance := *app.store.GetBalance(transfer.From)
		if fromBalance < transfer.Amount {
			resCode = TxCodeNotEnoughBalance
			break
		}
		if bytes.Equal(transfer.From, transfer.To) {
			resCode = TxCodeSelfTransaction
			break
		}
	case types.TxPurchase:
		purchase, _ := payload.(*types.Purchase)
		var metaData types.PDSNMetaData
		err := types.RequestMetaData(purchase.FileHash, &metaData)
		if err != nil {
			panic(err)
		}
		fromBalance := *app.store.GetBalance(purchase.From)
		if fromBalance < metaData.Price {
			resCode = TxCodeNotEnoughBalance
			break
		}
		// TODO: TxCodeAlreadyBought
		if bytes.Equal(purchase.From, metaData.Owner) {
			resCode = TxCodeSelfTransaction
			break
		}
	}
	return abci.ResponseCheckTx{Code: resCode}
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
		case types.AddressSize:
			// TODO
			break
		case types.HashSize << 1:
			// TODO
			break
		}
		if value != nil {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	}
}
