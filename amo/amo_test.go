package amo

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	tdb "github.com/tendermint/tendermint/libs/db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/operation"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

func TestInitChain(t *testing.T) {
	db := tdb.NewMemDB()
	app := NewAMOApplication(db, tdb.NewMemDB(), nil)
	req := abci.RequestInitChain{}
	req.AppStateBytes = []byte(`{ "balances": [ { "owner": "7CECB223B976F27D77B0E03E95602DABCC28D876", "amount": "100" } ] }`)
	res := app.InitChain(req)
	// TODO: need to check the contents of the response
	assert.Equal(t, abci.ResponseInitChain{}, res)

	// TODO: run series of app.Query() to check the genesis state
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	assert.Equal(t, new(types.Currency).Set(100), app.store.GetBalance(addr))
	//queryReq := abci.RequestQuery{}
	//queryRes := app.Query(queryReq)
}

func TestQueryDefault(t *testing.T) {
	db := tdb.NewMemDB()
	app := NewAMOApplication(db, tdb.NewMemDB(), nil)

	// query
	req := abci.RequestQuery{}
	req.Path = "/nostore"
	res := app.Query(req)
	assert.Equal(t, code.QueryCodeBadPath, res.Code)
}

func TestQueryBalance(t *testing.T) {
	db := tdb.NewMemDB()
	app := NewAMOApplication(db, tdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	queryjson, _ := json.Marshal(addr)
	app.store.SetBalanceUint64(addr, 100)

	_addrbin, _ := hex.DecodeString("FFECB223B976F27D77B0E03E95602DABCC28D876")
	_addr := crypto.Address(_addrbin)
	_queryjson, _ := json.Marshal(_addr)

	var req abci.RequestQuery
	var res abci.ResponseQuery
	var jsonstr []byte

	// errors
	req = abci.RequestQuery{Path: "/balance"}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoKey, res.Code)

	req = abci.RequestQuery{Path: "/balance", Data: []byte("f8das")}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)

	// XXX: current implementation returns zero balance for unknown address
	req = abci.RequestQuery{Path: "/balance", Data: []byte(_queryjson)}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(new(types.Currency).Set(0))
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)

	// query
	req = abci.RequestQuery{Path: "/balance", Data: []byte(queryjson)}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(new(types.Currency).Set(100))
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)
}

func TestQueryParcel(t *testing.T) {
	db := tdb.NewMemDB()
	app := NewAMOApplication(db, tdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := cmn.RandBytes(32)
	parcelID[31] = 0xFF

	parcel := types.ParcelValue{
		Owner:   addr,
		Custody: cmn.RandBytes(32),
		Info:    []byte("This is test parcel value"),
	}

	app.store.SetParcel(parcelID, &parcel)

	wrongParcelID := cmn.RandBytes(32)
	wrongParcelID[31] = 0xBB

	var req abci.RequestQuery
	var res abci.ResponseQuery
	var jsonstr []byte

	// errors
	req = abci.RequestQuery{Path: "/parcel"}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoKey, res.Code)

	// TODO: check this after parcel type implemented
	/*
		req = abci.RequestQuery{Path: "/parcel", Data: []byte("f8das")}
		res = app.Query(req)
		assert.Equal(t, code.QueryCodeBadKey, res.Code)
	*/

	req = abci.RequestQuery{Path: "/parcel", Data: wrongParcelID}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)
	assert.Nil(t, res.Value)
	assert.Len(t, res.Log, 0)

	// query
	req = abci.RequestQuery{Path: "/parcel", Data: parcelID}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(parcel)
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)
}

func TestQueryRequest(t *testing.T) {
	db := tdb.NewMemDB()
	app := NewAMOApplication(db, tdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := cmn.RandBytes(32)
	parcelID[31] = 0xFF

	request := types.RequestValue{
		Payment: *new(types.Currency).Set(400),
	}

	app.store.SetRequest(addr, parcelID, &request)

	wrongParcelID := cmn.RandBytes(32)
	wrongParcelID[31] = 0xBB
	wrongAddr := p256.GenPrivKey().PubKey().Address()
	wrongAddr[19] = 0xFF

	var req abci.RequestQuery
	var res abci.ResponseQuery
	var jsonstr []byte

	// errors
	req = abci.RequestQuery{Path: "/request"}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoKey, res.Code)

	// TODO: check this after parcel type implemented
	/*
		req = abci.RequestQuery{Path: "/parcel", Data: []byte("f8das")}
		res = app.Query(req)
		assert.Equal(t, code.QueryCodeBadKey, res.Code)
	*/

	jsonstr, _ = json.Marshal(addr)
	req = abci.RequestQuery{Path: "/request", Data: jsonstr}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
	assert.Nil(t, res.Value)
	assert.Len(t, res.Log, 0)

	jsonstr, _ = json.Marshal(parcelID)
	req = abci.RequestQuery{Path: "/request", Data: jsonstr}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
	assert.Nil(t, res.Value)
	assert.Len(t, res.Log, 0)

	var keyMap = map[string]cmn.HexBytes{
		"buyer":  addr,
		"target": parcelID,
	}

	// query
	key, _ := json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/request", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(request)
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)

	keyMap["buyer"] = wrongAddr
	key, _ = json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/request", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)

	delete(keyMap, "buyer")
	key, _ = json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/request", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
}

func TestQueryUsage(t *testing.T) {
	db := tdb.NewMemDB()
	app := NewAMOApplication(db, tdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := cmn.RandBytes(32)
	parcelID[31] = 0xFF

	usage := types.UsageValue{
		Custody: cmn.RandBytes(32),
	}

	app.store.SetUsage(addr, parcelID, &usage)

	wrongParcelID := cmn.RandBytes(32)
	wrongParcelID[31] = 0xBB
	wrongAddr := p256.GenPrivKey().PubKey().Address()
	wrongAddr[19] = 0xFF

	var req abci.RequestQuery
	var res abci.ResponseQuery
	var jsonstr []byte

	// errors
	req = abci.RequestQuery{Path: "/usage"}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoKey, res.Code)

	// TODO: check this after parcel type implemented
	/*
		req = abci.RequestQuery{Path: "/parcel", Data: []byte("f8das")}
		res = app.Query(req)
		assert.Equal(t, code.QueryCodeBadKey, res.Code)
	*/

	jsonstr, _ = json.Marshal(addr)
	req = abci.RequestQuery{Path: "/usage", Data: jsonstr}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
	assert.Nil(t, res.Value)
	assert.Len(t, res.Log, 0)

	jsonstr, _ = json.Marshal(parcelID)
	req = abci.RequestQuery{Path: "/usage", Data: jsonstr}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
	assert.Nil(t, res.Value)
	assert.Len(t, res.Log, 0)

	var keyMap = map[string]cmn.HexBytes{
		"buyer":  addr,
		"target": parcelID,
	}

	// query
	key, _ := json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/usage", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(usage)
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)

	keyMap["buyer"] = wrongAddr
	key, _ = json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/usage", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)

	delete(keyMap, "buyer")
	key, _ = json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/usage", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
}

func TestSignedTransactionTest(t *testing.T) {
	from := p256.GenPrivKeyFromSecret([]byte("alice"))

	db := tdb.NewMemDB()
	app := NewAMOApplication(db, tdb.NewMemDB(), nil)
	app.store.SetBalanceUint64(from.PubKey().Address(), 5000)

	tx := operation.Transfer{
		To:     p256.GenPrivKeyFromSecret([]byte("bob")).PubKey().Address(),
		Amount: *new(types.Currency).Set(500),
	}
	payload, err := json.Marshal(tx)
	assert.NoError(t, err)
	msg := operation.Message{
		Type:   operation.TxTransfer,
		Params: payload,
	}

	// not signed transaction
	rawMsg, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.Equal(t, code.TxCodeBadSignature, app.CheckTx(rawMsg).Code)

	// signed transaction
	err = msg.Sign(from)
	assert.NoError(t, err)
	rawMsg, err = json.Marshal(msg)
	assert.NoError(t, err)
	assert.Equal(t, code.TxCodeOK, app.CheckTx(rawMsg).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(rawMsg).Code)
}

func makeTxStake(priv p256.PrivKeyP256, amount uint64) []byte {
	//staker := priv.PubKey().Address()
	validator, _ := ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)
	op := operation.Stake{
		Amount:    *new(types.Currency).Set(amount),
		Validator: validator[:],
	}
	payload, _ := json.Marshal(op)
	tx := operation.Message{
		Type:   operation.TxStake,
		Params: payload,
	}
	tx.Sign(priv)
	rawTx, _ := json.Marshal(tx)
	return rawTx
}

func TestValidatorUpdates(t *testing.T) {
	db := tdb.NewMemDB()
	app := NewAMOApplication(db, tdb.NewMemDB(), nil)

	// setup
	priv := p256.GenPrivKeyFromSecret([]byte("staker"))
	app.store.SetBalance(priv.PubKey().Address(), new(types.Currency).Set(200))

	// begin block
	blkRequest := abci.RequestBeginBlock{}
	app.BeginBlock(blkRequest) // does nothing here

	// deliver stake tx
	rawTx := makeTxStake(priv, 100)
	resDeliver := app.DeliverTx(rawTx)
	assert.Equal(t, code.TxCodeOK, resDeliver.Code)

	// end block
	endRequest := abci.RequestEndBlock{Height: 1}
	validators := app.EndBlock(endRequest).ValidatorUpdates
	assert.Equal(t, 1, len(validators))

	// TODO: test voting power calculcation
	assert.Equal(t, int64(100), validators[0].Power)
}

func TestBlockReward(t *testing.T) {
	// setup
	app := NewAMOApplication(tdb.NewMemDB(), tdb.NewMemDB(), nil)

	// stake holder
	priv := ed25519.GenPrivKey()
	validator, _ := priv.PubKey().(ed25519.PubKeyEd25519)
	addrbin, _ := hex.DecodeString("BCECB223B976F27D77B0E03E95602DABCC28D876")
	holder := crypto.Address(addrbin)
	stake := types.Stake{
		Amount:    *new(types.Currency).Set(150),
		Validator: validator,
	}
	app.store.SetStake(holder, &stake)

	// delegated stake holders
	daddr1 := p256.GenPrivKeyFromSecret([]byte("d1")).PubKey().Address()
	daddr2 := p256.GenPrivKeyFromSecret([]byte("d2")).PubKey().Address()
	delegate1 := types.Delegate{
		Holder:    daddr1,
		Amount:    *new(types.Currency).Set(100),
		Delegator: holder,
	}
	delegate2 := types.Delegate{
		Holder:    daddr2,
		Amount:    *new(types.Currency).Set(200),
		Delegator: holder,
	}
	app.store.SetDelegate(daddr1, &delegate1)
	app.store.SetDelegate(daddr2, &delegate2)

	// execute
	req := abci.RequestBeginBlock{
		Header: abci.Header{
			NumTxs:          2,
			ProposerAddress: validator.Address(),
		},
	}
	_ = app.BeginBlock(req)

	// check distributed rewards
	bal := app.store.GetBalance(holder)
	assert.Equal(t,
		new(types.Currency).Set(uint64(types.OneAMOUint64*1.2/2)),
		bal)
	bal = app.store.GetBalance(daddr1)
	assert.Equal(t,
		new(types.Currency).Set(uint64(types.OneAMOUint64*1.2/6)),
		bal)
	bal = app.store.GetBalance(daddr2)
	assert.Equal(t,
		new(types.Currency).Set(uint64(types.OneAMOUint64*1.2/3)),
		bal)
}
