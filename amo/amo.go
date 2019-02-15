package amo

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/amolabs/amoabci/amo/types"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/version"
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
}

var _ abci.Application = (*AMOApplication)(nil)

func NewAMOApplication(db dbm.DB) *AMOApplication {
	state := loadState(db)
	app := &AMOApplication{state: state}
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
	from := app.GetAccount(transfer.From)
	to := app.GetAccount(transfer.To)
	from.Balance -= transfer.Amount
	to.Balance += transfer.Amount
	app.SetAccount(transfer.From, from)
	app.SetAccount(transfer.To, to)
	app.state.Size += 1
	tags := []cmn.KVPair{
		{Key: transfer.From[:], Value: []byte(strconv.FormatUint(uint64(from.Balance), 10))},
		{Key: transfer.To[:], Value: []byte(strconv.FormatUint(uint64(to.Balance), 10))},
	}
	return TxCodeOK, tags
}

func (app *AMOApplication) procPurchase(purchase *types.Purchase) (uint32, []cmn.KVPair) {
	var metaData types.PDSNMetaData
	err := types.RequestMetaData(purchase.FileHash, &metaData)
	if err != nil {
		panic(err)
	}
	from := app.GetAccount(purchase.From)
	from.Balance -= metaData.Price
	from.PurchasedFiles[metaData.FileHash] = true
	app.SetAccount(purchase.From, from)
	buyer := app.GetBuyer(metaData.FileHash)
	(*buyer)[purchase.From] = true
	app.SetBuyer(metaData.FileHash, buyer)
	result, err := json.Marshal(metaData)
	if err != nil {
		panic(err)
	}
	tags := []cmn.KVPair{
		{Key: []byte(hex.EncodeToString(metaData.FileHash[:])), Value: result},
		{Key: purchase.From[:], Value: []byte(strconv.FormatUint(uint64(from.Balance), 10))},
	}
	return TxCodeOK, tags
}

func (app *AMOApplication) CheckTx(tx []byte) abci.ResponseCheckTx {
	message, payload := types.ParseTx(tx)
	var resCode = TxCodeOK

	switch message.Command {
	case types.TxTransfer:
		transfer, _ := payload.(*types.Transfer)
		from := app.GetAccount(transfer.From)
		if from.Balance < transfer.Amount {
			resCode = TxCodeNotEnoughBalance
			break
		}
		if transfer.From == transfer.To {
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
		from := app.GetAccount(purchase.From)
		if from.Balance < metaData.Price {
			resCode = TxCodeNotEnoughBalance
			break
		}
		if _, ok := (*app.GetBuyer(purchase.FileHash))[purchase.From]; ok {
			resCode = TxCodeAlreadyBought
			break
		}
		if purchase.From == metaData.Owner {
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
			value, _ = json.Marshal(app.GetAccount(*types.NewAddress(reqQuery.Data)))
			resQuery.Value = value
		case types.HashSize << 1:
			value, _ = json.Marshal(app.GetBuyer(*types.NewHashFromHexBytes(reqQuery.Data)))
			resQuery.Value = value
		}
		if value != nil {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	}
}
