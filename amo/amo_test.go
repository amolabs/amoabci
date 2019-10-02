package amo

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/tx"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

var tmpFile *os.File

// setup and teardown
func setUpTest(t *testing.T) {
	file, err := ioutil.TempFile("", "state_*.json")
	assert.NoError(t, err)

	tmpFile = file
}

func tearDownTest(t *testing.T) {
	err := os.Remove(tmpFile.Name())
	assert.NoError(t, err)
}

func TestInitChain(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	req := abci.RequestInitChain{}
	req.AppStateBytes = []byte(`{ "balances": [ { "owner": "7CECB223B976F27D77B0E03E95602DABCC28D876", "amount": "100" } ] }`)
	res := app.InitChain(req)
	// TODO: need to check the contents of the response
	assert.Equal(t, abci.ResponseInitChain{}, res)

	// TODO: run series of app.Query() to check the genesis state
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	assert.Equal(t, new(types.Currency).Set(100), app.store.GetBalance(addr, fromStage))
	//queryReq := abci.RequestQuery{}
	//queryRes := app.Query(queryReq)
}

func TestQueryDefault(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	// query
	req := abci.RequestQuery{}
	req.Path = "/nostore"
	res := app.Query(req)
	assert.Equal(t, code.QueryCodeBadPath, res.Code)
}

func TestQueryBalance(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	queryjson, _ := json.Marshal(addr)
	app.store.SetBalanceUint64(addr, 100)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

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
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := cmn.HexBytes(cmn.RandBytes(32))
	parcelID[31] = 0xFF
	queryjson, _ := json.Marshal(parcelID)

	parcel := types.ParcelValue{
		Owner:   addr,
		Custody: cmn.RandBytes(32),
		Info:    []byte("This is test parcel value"),
	}

	app.store.SetParcel(parcelID, &parcel)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

	wrongParcelID := cmn.HexBytes(cmn.RandBytes(32))
	wrongParcelID[31] = 0xBB
	_queryjson, _ := json.Marshal(wrongParcelID)

	var req abci.RequestQuery
	var res abci.ResponseQuery
	var jsonstr []byte

	// errors
	req = abci.RequestQuery{Path: "/parcel"}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoKey, res.Code)

	// No bad key, for parcel id is an arbitrary length byte array.
	/*
		req = abci.RequestQuery{Path: "/parcel", Data: []byte("f8das")}
		res = app.Query(req)
		assert.Equal(t, code.QueryCodeBadKey, res.Code)
	*/

	req = abci.RequestQuery{Path: "/parcel", Data: []byte(_queryjson)}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)
	assert.Equal(t, []byte("null"), res.Value)
	assert.Equal(t, "null", res.Log)

	// query
	req = abci.RequestQuery{Path: "/parcel", Data: queryjson}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(parcel)
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)
}

func TestQueryRequest(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := cmn.RandBytes(32)
	parcelID[31] = 0xFF

	request := types.RequestValue{
		Payment: *new(types.Currency).Set(400),
	}

	app.store.SetRequest(addr, parcelID, &request)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

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
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := cmn.RandBytes(32)
	parcelID[31] = 0xFF

	usage := types.UsageValue{
		Custody: cmn.RandBytes(32),
	}

	app.store.SetUsage(addr, parcelID, &usage)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

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

func TestQueryValidator(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// stake holder
	priv := ed25519.GenPrivKey()
	validator, _ := priv.PubKey().(ed25519.PubKeyEd25519)
	valaddr := validator.Address()
	queryjson, _ := json.Marshal(valaddr)
	addrbin, _ := hex.DecodeString("BCECB223B976F27D77B0E03E95602DABCC28D876")
	holder := crypto.Address(addrbin)
	stake := types.Stake{
		Amount:    *new(types.Currency).Set(150),
		Validator: validator,
	}

	app.store.SetUnlockedStake(holder, &stake)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

	var req abci.RequestQuery
	var res abci.ResponseQuery
	var jsonstr []byte

	req = abci.RequestQuery{Path: "/validator"}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoKey, res.Code)

	req = abci.RequestQuery{Path: "/validator", Data: []byte("f8das")}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)

	req = abci.RequestQuery{Path: "/validator", Data: []byte(queryjson)}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(holder)
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)
}

func TestSignedTransactionTest(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	from := p256.GenPrivKeyFromSecret([]byte("alice"))

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	app.store.SetBalanceUint64(from.PubKey().Address(), 5000)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

	_tx := tx.TransferParam{
		To:     p256.GenPrivKeyFromSecret([]byte("bob")).PubKey().Address(),
		Amount: *new(types.Currency).Set(500),
	}
	payload, err := json.Marshal(_tx)
	assert.NoError(t, err)
	msg := tx.TxBase{
		Type:    "transfer",
		Payload: payload,
		Sender:  from.PubKey().Address(),
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
	}

	// not signed transaction
	rawMsg, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.Equal(t, code.TxCodeBadSignature, app.CheckTx(abci.RequestCheckTx{Tx: rawMsg}).Code)

	// signed transaction
	err = msg.Sign(from)
	assert.NoError(t, err)
	rawMsg, err = json.Marshal(msg)
	assert.NoError(t, err)
	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: rawMsg}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: rawMsg}).Code)
}

func TestFuncValUpdates(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	val1 := abci.ValidatorUpdate{
		PubKey: abci.PubKey{Type: "anything", Data: []byte("0001")},
		Power:  1,
	}
	val2 := abci.ValidatorUpdate{
		PubKey: abci.PubKey{Type: "anything", Data: []byte("0002")},
		Power:  2,
	}
	val22 := abci.ValidatorUpdate{
		PubKey: abci.PubKey{Type: "anything", Data: []byte("0002")},
		Power:  22,
	}
	val3 := abci.ValidatorUpdate{
		PubKey: abci.PubKey{Type: "anything", Data: []byte("0003")},
		Power:  3,
	}
	uold := abci.ValidatorUpdates{val1, val2, val3}
	unew := abci.ValidatorUpdates{val22, val3}
	assert.Equal(t, 3, len(uold))
	assert.Equal(t, 2, len(unew))
	updates := findValUpdates(uold, unew)
	assert.Equal(t, 3, len(updates))
	assert.Equal(t, int64(22), updates[0].Power)
	assert.Equal(t, int64(3), updates[1].Power)
	assert.Equal(t, int64(0), updates[2].Power)
	assert.Equal(t, []byte("0001"), updates[2].PubKey.Data)
}

func makeTxStake(priv p256.PrivKeyP256, val string, amount uint64) []byte {
	validator, _ := ed25519.GenPrivKeyFromSecret([]byte(val)).
		PubKey().(ed25519.PubKeyEd25519)
	param := tx.StakeParam{
		Amount:    *new(types.Currency).Set(amount),
		Validator: validator[:],
	}
	payload, _ := json.Marshal(param)
	_tx := tx.TxBase{
		Type:    "stake",
		Payload: payload,
		Sender:  priv.PubKey().Address(),
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
	}
	_tx.Sign(priv)
	rawTx, _ := json.Marshal(_tx)
	return rawTx
}

func makeTxWithdraw(priv p256.PrivKeyP256, amount uint64) []byte {
	param := tx.WithdrawParam{
		Amount: *new(types.Currency).Set(amount),
	}
	payload, _ := json.Marshal(param)
	_tx := tx.TxBase{
		Type:    "withdraw",
		Payload: payload,
		Sender:  priv.PubKey().Address(),
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
	}
	_tx.Sign(priv)
	rawTx, _ := json.Marshal(_tx)
	return rawTx
}

func TestEndBlock(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// setup
	tx.ConfigLockupPeriod = 1 // manipulate
	priv1 := p256.GenPrivKeyFromSecret([]byte("staker1"))
	app.store.SetBalance(priv1.PubKey().Address(), new(types.Currency).Set(500))
	priv2 := p256.GenPrivKeyFromSecret([]byte("staker2"))
	app.store.SetBalance(priv2.PubKey().Address(), new(types.Currency).Set(500))

	// immitate initChain() function call
	_, _, err := app.store.Save()
	assert.NoError(t, err)

	// begin block
	blkRequest := abci.RequestBeginBlock{}
	app.BeginBlock(blkRequest) // we need this

	// deliver stake tx
	rawTx := makeTxStake(priv1, "val1", 100)
	resDeliver := app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resDeliver.Code)

	rawTx = makeTxStake(priv2, "val1", 200)
	resCheck := app.CheckTx(abci.RequestCheckTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resCheck.Code)
	resDeliver = app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodePermissionDenied, resDeliver.Code)

	rawTx = makeTxStake(priv2, "val2", 200)
	resDeliver = app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resDeliver.Code)

	// deliver withdraw tx. this should fail
	rawTx = makeTxWithdraw(priv2, 200)
	resDeliver = app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeStakeLocked, resDeliver.Code)

	// end block
	endRequest := abci.RequestEndBlock{Height: 1}
	validators := app.EndBlock(endRequest).ValidatorUpdates
	assert.Equal(t, 2, len(validators))

	assert.Equal(t, int64(200), validators[0].Power)
	assert.Equal(t, int64(100), validators[1].Power)

	// deliver withdraw tx. this should succeed now.
	rawTx = makeTxWithdraw(priv2, 200)
	resDeliver = app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resDeliver.Code)
}

func TestBlockReward(t *testing.T) {
	// setup
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// stake holder
	priv := ed25519.GenPrivKey()
	validator, _ := priv.PubKey().(ed25519.PubKeyEd25519)
	addrbin, _ := hex.DecodeString("BCECB223B976F27D77B0E03E95602DABCC28D876")
	holder := crypto.Address(addrbin)
	stake := types.Stake{
		Amount:    *new(types.Currency).Set(150),
		Validator: validator,
	}

	app.store.SetUnlockedStake(holder, &stake)

	// delegated stake holders
	daddr1 := p256.GenPrivKeyFromSecret([]byte("d1")).PubKey().Address()
	daddr2 := p256.GenPrivKeyFromSecret([]byte("d2")).PubKey().Address()
	delegate1 := types.Delegate{
		Delegatee: holder,
		Amount:    *new(types.Currency).Set(100),
	}
	delegate2 := types.Delegate{
		Delegatee: holder,
		Amount:    *new(types.Currency).Set(200),
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
	var delta int64
	var bal, ass *types.Currency

	bal = app.store.GetBalance(holder, fromStage)
	ass = new(types.Currency).Set(uint64(types.OneAMOUint64 * float64(0.2/2)))
	delta = bal.Int.Sub(&bal.Int, &ass.Int).Int64()
	assert.True(t, delta < 10 && delta > -10)

	bal = app.store.GetBalance(daddr1, fromStage)
	ass = new(types.Currency).Set(uint64(types.OneAMOUint64 * float64(0.2/6)))
	delta = bal.Int.Sub(&bal.Int, &ass.Int).Int64()
	assert.True(t, delta < 10 && delta > -10)

	bal = app.store.GetBalance(daddr2, fromStage)
	ass = new(types.Currency).Set(uint64(types.OneAMOUint64 * float64(0.2/3)))
	delta = bal.Int.Sub(&bal.Int, &ass.Int).Int64()
	assert.True(t, delta < 10 && delta > -10)
}
