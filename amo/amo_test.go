package amo

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/blockchain"
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

func TestAppConfig(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	// test genesis app config
	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	req := abci.RequestInitChain{}
	req.AppStateBytes = []byte(
		`{ "config": { "max_validators": 10, "lockup_period": 2 } }`)
	res := app.InitChain(req)
	// TODO: need to check the contents of the response
	assert.Equal(t, abci.ResponseInitChain{}, res)

	// check
	assert.Equal(t, uint64(10), app.config.MaxValidators)
	assert.Equal(t, defaultWeightValidator, app.config.WeightValidator)
	assert.Equal(t, defaultWeightDelegator, app.config.WeightDelegator)
	assert.Equal(t, defaultBlkReward, app.config.BlkReward)
	assert.Equal(t, defaultTxReward, app.config.TxReward)
	assert.Equal(t, uint64(2), app.config.LockupPeriod)
}

func TestInitChain(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	req := abci.RequestInitChain{}
	req.AppStateBytes = []byte(`{ "balances": [ { "owner": "7CECB223B976F27D77B0E03E95602DABCC28D876", "amount": "100" } ] }`)
	res := app.InitChain(req)
	// TODO: need to check the contents of the response
	assert.Equal(t, abci.ResponseInitChain{}, res)

	// TODO: run series of app.Query() to check the genesis state
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	assert.Equal(t, new(types.Currency).Set(100), app.store.GetBalance(addr, false))
	//queryReq := abci.RequestQuery{}
	//queryRes := app.Query(queryReq)
}

func TestQueryDefault(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	// query
	req := abci.RequestQuery{}
	req.Path = "/nostore"
	res := app.Query(req)
	assert.Equal(t, code.QueryCodeBadPath, res.Code)
}

func TestQueryAppConfig(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	config := app.config

	var (
		req     abci.RequestQuery
		res     abci.ResponseQuery
		jsonstr []byte
	)

	req = abci.RequestQuery{Path: "/config"}
	res = app.Query(req)

	jsonstr, _ = json.Marshal(config)

	assert.Equal(t, jsonstr, res.GetValue())
}

func TestQueryBalance(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
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

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	_addrbin, _ := hex.DecodeString("FFECB223B976F27D77B0E03E95602DABCC28D876")
	_addr := crypto.Address(_addrbin)

	parcelID := cmn.HexBytes(cmn.RandBytes(32))
	parcelID[31] = 0xFF
	queryjson, _ := json.Marshal(parcelID)

	parcel := types.Parcel{
		Owner:   addr,
		Custody: cmn.RandBytes(32),
	}

	request := types.Request{
		Payment: *new(types.Currency).Set(1),
	}

	parcelEx := types.ParcelEx{
		Parcel: &parcel,
		Requests: []*types.RequestEx{
			&types.RequestEx{
				Request: &request,
				Buyer:   _addr,
			},
		},
		Usages: []*types.UsageEx{},
	}

	app.store.SetParcel(parcelID, &parcel)
	app.store.SetRequest(_addr, parcelID, &request)

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
	assert.Equal(t, []byte(nil), res.Value)
	assert.Equal(t, "error: no parcel", res.Log)

	// query
	req = abci.RequestQuery{Path: "/parcel", Data: queryjson}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(parcelEx)
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)
}

func TestQueryRequest(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := cmn.RandBytes(32)
	parcelID[31] = 0xFF

	request := types.Request{
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
	assert.Equal(t, "error: unmarshal", res.Log)

	jsonstr, _ = json.Marshal(parcelID)
	req = abci.RequestQuery{Path: "/request", Data: jsonstr}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
	assert.Nil(t, res.Value)
	assert.Equal(t, "error: unmarshal", res.Log)

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

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := cmn.RandBytes(32)
	parcelID[31] = 0xFF

	usage := types.Usage{
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
	assert.Equal(t, "error: unmarshal", res.Log)

	jsonstr, _ = json.Marshal(parcelID)
	req = abci.RequestQuery{Path: "/usage", Data: jsonstr}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
	assert.Nil(t, res.Value)
	assert.Equal(t, "error: unmarshal", res.Log)

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

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

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

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.blockBindingManager.Update()

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
		Type:       "transfer",
		Payload:    payload,
		Sender:     from.PubKey().Address(),
		Fee:        *new(types.Currency).Set(0),
		LastHeight: "1",
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

func TestPenaltyEvidence(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// setup
	//
	// node composition
	// val - staker (convict)
	//     - delegator1
	//     - delegator2

	val, _ := ed25519.GenPrivKeyFromSecret([]byte("val")).PubKey().(ed25519.PubKeyEd25519)
	staker := p256.GenPrivKeyFromSecret([]byte("staker"))
	app.store.SetBalance(staker.PubKey().Address(), new(types.Currency).Set(1000))
	app.store.SetUnlockedStake(staker.PubKey().Address(), &types.Stake{
		Amount:    *new(types.Currency).Set(1000),
		Validator: val,
	})

	delegator1 := p256.GenPrivKeyFromSecret([]byte("delegator1"))
	app.store.SetBalance(delegator1.PubKey().Address(), new(types.Currency).Set(500))
	app.store.SetDelegate(delegator1.PubKey().Address(), &types.Delegate{
		Amount:    *new(types.Currency).Set(500),
		Delegatee: staker.PubKey().Address(),
	})

	delegator2 := p256.GenPrivKeyFromSecret([]byte("delegator2"))
	app.store.SetBalance(delegator2.PubKey().Address(), new(types.Currency).Set(500))
	app.store.SetDelegate(delegator2.PubKey().Address(), &types.Delegate{
		Amount:    *new(types.Currency).Set(500),
		Delegatee: staker.PubKey().Address(),
	})

	app.store.Save()

	// imitate target convict: val - validator
	evidences := []abci.Evidence{}
	evidences = append(evidences, abci.Evidence{
		Validator: abci.Validator{Address: val.Address()},
		Height:    int64(2),
	})

	app.BeginBlock(abci.RequestBeginBlock{ByzantineValidators: evidences})

	// before effective stake
	stakerbes := app.store.GetEffStake(staker.PubKey().Address(), false)
	stakerbesf := new(big.Float).SetInt(&stakerbes.Amount.Int)

	app.EndBlock(abci.RequestEndBlock{})

	// after effective stake
	stakeraes := app.store.GetEffStake(staker.PubKey().Address(), false)
	stakeraesf := new(big.Float).SetInt(&stakeraes.Amount.Int)

	// slashing effective stake calculation
	// ces = bes * (1 - m)
	prf := new(big.Float).SetFloat64(app.config.PenaltyRatioM)
	tmpf := new(big.Float).SetInt64(1)
	tmpf.Sub(tmpf, prf)

	// candidate effective stake
	stakercesf := stakerbesf.Mul(stakerbesf, tmpf)

	// to ignore last three digits
	divisor := new(big.Float).SetInt64(1000)

	ces, _ := new(big.Float).Quo(stakercesf, divisor).Int64()
	aes, _ := new(big.Float).Quo(stakeraesf, divisor).Int64()

	// compare aes == ces
	assert.Equal(t, ces, aes)
}

func TestPenaltyLazyValidators(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	app.lazinessCounter = blockchain.NewLazinessCounter(
		app.store,
		app.state.Height,
		app.state.CounterDue,
		int64(4),
		float64(0.5),
	)

	// setup
	//
	// node composition
	// val - staker (convict)
	//     - delegator1
	//     - delegator2

	val, _ := ed25519.GenPrivKeyFromSecret([]byte("val")).PubKey().(ed25519.PubKeyEd25519)
	staker := p256.GenPrivKeyFromSecret([]byte("staker"))
	app.store.SetBalance(staker.PubKey().Address(), new(types.Currency).Set(1000))
	app.store.SetUnlockedStake(staker.PubKey().Address(), &types.Stake{
		Amount:    *new(types.Currency).Set(1000),
		Validator: val,
	})

	delegator1 := p256.GenPrivKeyFromSecret([]byte("delegator1"))
	app.store.SetBalance(delegator1.PubKey().Address(), new(types.Currency).Set(500))
	app.store.SetDelegate(delegator1.PubKey().Address(), &types.Delegate{
		Amount:    *new(types.Currency).Set(500),
		Delegatee: staker.PubKey().Address(),
	})

	delegator2 := p256.GenPrivKeyFromSecret([]byte("delegator2"))
	app.store.SetBalance(delegator2.PubKey().Address(), new(types.Currency).Set(500))
	app.store.SetDelegate(delegator2.PubKey().Address(), &types.Delegate{
		Amount:    *new(types.Currency).Set(500),
		Delegatee: staker.PubKey().Address(),
	})

	app.store.Save()

	lastCommitInfo := abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{
				Validator: abci.Validator{
					Address: val.Address(),
				},
				SignedLastBlock: false,
			},
		},
	}

	// before effective stake
	stakerbes := app.store.GetEffStake(staker.PubKey().Address(), false)
	stakerbesf := new(big.Float).SetInt(&stakerbes.Amount.Int)

	app.BeginBlock(abci.RequestBeginBlock{
		Header:         abci.Header{Height: 1},
		LastCommitInfo: lastCommitInfo,
	})
	app.EndBlock(abci.RequestEndBlock{})
	// lazinessCounter height -> 1
	// 				   candidates -> val: 1

	app.BeginBlock(abci.RequestBeginBlock{
		Header:         abci.Header{Height: 2},
		LastCommitInfo: lastCommitInfo,
	})
	app.EndBlock(abci.RequestEndBlock{})
	// lazinessCounter height -> 2
	// 				   candidates -> val: 2

	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{
				Validator: abci.Validator{
					Address: val.Address(),
				},
				SignedLastBlock: true,
			},
		},
	}

	app.BeginBlock(abci.RequestBeginBlock{
		Header:         abci.Header{Height: 3},
		LastCommitInfo: lastCommitInfo,
	})
	app.EndBlock(abci.RequestEndBlock{})
	// lazinessCounter height -> 3
	// 				   candidates -> val: 2

	app.BeginBlock(abci.RequestBeginBlock{
		Header:         abci.Header{Height: 4},
		LastCommitInfo: lastCommitInfo,
	})
	app.EndBlock(abci.RequestEndBlock{})
	// lazinessCounter height -> 4
	// 				   candidates -> val: 2

	app.BeginBlock(abci.RequestBeginBlock{
		Header:         abci.Header{Height: 5},
		LastCommitInfo: lastCommitInfo,
	})
	app.EndBlock(abci.RequestEndBlock{})
	// lazinessCounter height -> 5
	// 				   candidates -> nil

	// after effective stake
	stakeraes := app.store.GetEffStake(staker.PubKey().Address(), false)
	stakeraesf := new(big.Float).SetInt(&stakeraes.Amount.Int)

	// slashing effective stake calculation
	// ces = bes * (1 - m)
	prf := new(big.Float).SetFloat64(app.config.PenaltyRatioL)
	tmpf := new(big.Float).SetInt64(1)
	tmpf.Sub(tmpf, prf)

	// candidate effective stake
	stakercesf := stakerbesf.Mul(stakerbesf, tmpf)

	// to ignore last three digits
	divisor := new(big.Float).SetInt64(1000)

	ces, _ := new(big.Float).Quo(stakercesf, divisor).Int64()
	aes, _ := new(big.Float).Quo(stakeraesf, divisor).Int64()

	// compare aes == ces
	assert.Equal(t, ces, aes)
}

func TestEndBlock(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.blockBindingManager.Update()

	// setup
	tx.ConfigLockupPeriod = 1       // manipulate
	tx.ConfigMinStakingUnit = "100" // manipulate
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
	rawTx := makeTxStake(priv1, "val1", 100, "1")
	resDeliver := app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resDeliver.Code)

	rawTx = makeTxStake(priv2, "val1", 200, "1")
	resCheck := app.CheckTx(abci.RequestCheckTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resCheck.Code)
	resDeliver = app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodePermissionDenied, resDeliver.Code)

	rawTx = makeTxStake(priv2, "val2", 200, "1")
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

func DivCurrency(origin *types.Currency, divisor *types.Currency) *types.Currency {
	return new(types.Currency).Set(origin.Div(&origin.Int, &divisor.Int).Uint64())
}

func TestIncentive(t *testing.T) {
	// setup
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	tx.ConfigMinStakingUnit = "50"

	validator, _ := ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)

	d1Priv := p256.GenPrivKeyFromSecret([]byte("delegate1"))
	d2Priv := p256.GenPrivKeyFromSecret([]byte("delegate2"))
	sPriv := p256.GenPrivKeyFromSecret([]byte("stake"))

	app.store.SetBalance(d1Priv.PubKey().Address(), new(types.Currency).Set(200))
	app.store.SetBalance(d2Priv.PubKey().Address(), new(types.Currency).Set(400))
	app.store.SetBalance(sPriv.PubKey().Address(), new(types.Currency).Set(150))

	stake := types.Stake{
		Amount:    *new(types.Currency).Set(150),
		Validator: validator,
	}

	app.store.SetUnlockedStake(sPriv.PubKey().Address(), &stake)
	app.store.Save()

	// to ignore last three digits
	divisor := new(types.Currency).Set(1000)

	bald1 := app.store.GetBalance(d1Priv.PubKey().Address(), true)
	bald2 := app.store.GetBalance(d2Priv.PubKey().Address(), true)
	bals := app.store.GetBalance(sPriv.PubKey().Address(), true)

	app.BeginBlock(abci.RequestBeginBlock{
		Header: abci.Header{
			ProposerAddress: validator.Address(),
			Height:          1,
		},
	})

	rawTx := makeTxDelegate(d1Priv, sPriv.PubKey().Address(), 100)
	resDeliver := app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resDeliver.Code)

	rawTx = makeTxDelegate(d2Priv, sPriv.PubKey().Address(), 200)
	resDeliver = app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resDeliver.Code)

	app.EndBlock(abci.RequestEndBlock{Height: 1})

	app.Commit()
	tx.ConfigMinStakingUnit = "50"

	// check incentive records
	bir := app.store.GetBlockIncentiveRecords(1)
	assert.Equal(t, 1, len(bir))

	bals = app.store.GetBalance(sPriv.PubKey().Address(), true).Sub(bals)

	bir[0].Amount = DivCurrency(bir[0].Amount, divisor)
	bals = DivCurrency(bals, divisor)

	assert.Equal(t, bir[0].Amount, bals)

	bald1 = app.store.GetBalance(d1Priv.PubKey().Address(), true)
	bald2 = app.store.GetBalance(d2Priv.PubKey().Address(), true)
	bals = app.store.GetBalance(sPriv.PubKey().Address(), true)

	app.BeginBlock(abci.RequestBeginBlock{
		Header: abci.Header{
			ProposerAddress: validator.Address(),
			Height:          2,
		},
	})

	rawTx = makeTxDelegate(d1Priv, sPriv.PubKey().Address(), 100)
	resDeliver = app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resDeliver.Code)

	rawTx = makeTxDelegate(d2Priv, sPriv.PubKey().Address(), 200)
	resDeliver = app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, resDeliver.Code)

	app.EndBlock(abci.RequestEndBlock{Height: 2})

	app.Commit()
	tx.ConfigMinStakingUnit = "50"

	bir = app.store.GetBlockIncentiveRecords(2)
	assert.Equal(t, 3, len(bir))

	sort.Slice(bir, func(i, j int) bool {
		return bir[i].Amount.LessThan(bir[j].Amount)
	})

	bald1 = app.store.GetBalance(d1Priv.PubKey().Address(), true).Sub(bald1)
	bald2 = app.store.GetBalance(d2Priv.PubKey().Address(), true).Sub(bald2)
	bals = app.store.GetBalance(sPriv.PubKey().Address(), true).Sub(bals)

	bir[0].Amount = DivCurrency(bir[0].Amount, divisor)
	bir[1].Amount = DivCurrency(bir[1].Amount, divisor)
	bir[2].Amount = DivCurrency(bir[2].Amount, divisor)

	bald1 = DivCurrency(bald1, divisor)
	bald2 = DivCurrency(bald2, divisor)
	bals = DivCurrency(bals, divisor)

	assert.Equal(t, bir[0].Amount, bald1)
	assert.Equal(t, bir[1].Amount, bald2)
	assert.Equal(t, bir[2].Amount, bals)
}

func TestIncentiveNoTouch(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// setup
	validator, _ := ed25519.GenPrivKeyFromSecret([]byte("test")).
		PubKey().(ed25519.PubKeyEd25519)
	priv := p256.GenPrivKeyFromSecret([]byte("test"))

	app.store.SetLockedStake(priv.PubKey().Address(),
		&types.Stake{Validator: validator, Amount: *new(types.Currency).Set(500)}, 1)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

	prevBalance := app.store.GetBalance(priv.PubKey().Address(), true)

	prevHash, _, err := app.store.Save()
	assert.NoError(t, err)

	err = blockchain.DistributeIncentive(
		app.store,
		app.logger,
		app.config.WeightValidator, app.config.WeightDelegator,
		app.config.BlkReward, app.config.TxReward,
		app.state.Height, app.numDeliveredTxs,
		app.staker,
		app.feeAccumulated,
	)
	assert.NoError(t, err)

	balance := app.store.GetBalance(priv.PubKey().Address(), true)

	hash, _, err := app.store.Save()
	assert.NoError(t, err)

	assert.Equal(t, prevBalance, balance)
	assert.Equal(t, prevHash, hash)
}

func TestEmptyBlock(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// setup
	tx.ConfigLockupPeriod = 2       // manipulate
	tx.ConfigMinStakingUnit = "100" // manipulate
	priv := p256.GenPrivKeyFromSecret([]byte("test"))
	app.store.SetBalance(priv.PubKey().Address(), new(types.Currency).Set(500))

	// init chain
	app.InitChain(abci.RequestInitChain{})

	// begin block
	app.BeginBlock(abci.RequestBeginBlock{})

	rawTx := makeTxStake(priv, "test", 500, "1")
	app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})

	// end block
	app.EndBlock(abci.RequestEndBlock{})

	// commit
	app.Commit()

	stakes := app.store.GetLockedStakes(makeTestAddress("test"), true)
	assert.Equal(t, 1, len(stakes))

	// get hash to compare
	prevHash := app.state.LastAppHash

	// simulate no txs to process in this block
	app.BeginBlock(abci.RequestBeginBlock{})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// get hash to compare
	hash := app.state.LastAppHash

	// should not equal as the stake lock-up remains
	assert.NotEqual(t, prevHash, hash)

	// stake lock-up should end here
	stakes = app.store.GetLockedStakes(makeTestAddress("test"), true)
	assert.Equal(t, 0, len(stakes))

	prevHash = app.state.LastAppHash

	// simulate no txs to process in this block
	app.BeginBlock(abci.RequestBeginBlock{})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// get hash to compare
	hash = app.state.LastAppHash

	// should equal as the stake lock-up ends
	assert.Equal(t, prevHash, hash)
}

func TestReplayAttack(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	t1 := p256.GenPrivKeyFromSecret([]byte("test1"))
	tx1 := makeTxStake(t1, "test1", 10000, "1")
	tx2 := makeTxStake(t1, "test1", 10000, "1")
	tx3 := makeTxStake(t1, "test1", 10000, "1")

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.replayPreventer = blockchain.NewReplayPreventer(
		app.store,
		3,
		app.state.LastHeight,
	)

	tx.ConfigMinStakingUnit = "100" // manipulate

	app.store.SetBalance(t1.PubKey().Address(), new(types.Currency).Set(40000))

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 1}})

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)

	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 1})

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})

	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx2}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx2}).Code)

	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.CheckTx(abci.RequestCheckTx{Tx: tx2}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx2}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 2})

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 3}})

	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.CheckTx(abci.RequestCheckTx{Tx: tx2}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx2}).Code)

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx3}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx3}).Code)

	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.CheckTx(abci.RequestCheckTx{Tx: tx3}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx3}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 3})

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 4}})

	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.CheckTx(abci.RequestCheckTx{Tx: tx2}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx2}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.CheckTx(abci.RequestCheckTx{Tx: tx3}).Code)
	assert.Equal(t, code.TxCodeAlreadyProcessedTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx3}).Code)

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 4})
}

func TestBindingBlock(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	t1 := p256.GenPrivKeyFromSecret([]byte("test1"))
	tx1 := makeTxStake(t1, "test1", 10000, "1")
	tx2 := makeTxStake(t1, "test1", 10000, "2")
	tx3 := makeTxStake(t1, "test1", 10000, "3")
	tx4 := makeTxStake(t1, "test1", 10000, "1")
	tx5 := makeTxStake(t1, "test1", 10000, "2")

	app := NewAMOApp(tmpFile, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.blockBindingManager = blockchain.NewBlockBindingManager(
		3,
		app.state.LastHeight,
	)

	tx.ConfigMinStakingUnit = "100" // manipulate

	app.store.SetBalance(t1.PubKey().Address(), new(types.Currency).Set(50000))

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 1}})

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 1})

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx2}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx2}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 2})

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 3}})

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx3}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx3}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 3})

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 4}})

	assert.Equal(t, code.TxCodeTooOldTx, app.CheckTx(abci.RequestCheckTx{Tx: tx4}).Code)
	assert.Equal(t, code.TxCodeTooOldTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx4}).Code)

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx5}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx5}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 4})
}

func makeTxStake(priv p256.PrivKeyP256, val string, amount uint64, lastHeight string) []byte {
	validator, _ := ed25519.GenPrivKeyFromSecret([]byte(val)).
		PubKey().(ed25519.PubKeyEd25519)
	param := tx.StakeParam{
		Amount:    *new(types.Currency).Set(amount),
		Validator: validator[:],
	}
	payload, _ := json.Marshal(param)
	_tx := tx.TxBase{
		Type:       "stake",
		Payload:    payload,
		Sender:     priv.PubKey().Address(),
		Fee:        *new(types.Currency).Set(0),
		LastHeight: lastHeight,
	}
	_tx.Sign(priv)
	rawTx, _ := json.Marshal(_tx)
	return rawTx
}

func makeTxDelegate(priv p256.PrivKeyP256, to crypto.Address, amount uint64) []byte {
	param := tx.DelegateParam{
		To:     to,
		Amount: *new(types.Currency).Set(amount),
	}
	payload, _ := json.Marshal(param)
	_tx := tx.TxBase{
		Type:       "delegate",
		Payload:    payload,
		Sender:     priv.PubKey().Address(),
		Fee:        *new(types.Currency).Set(0),
		LastHeight: "1",
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
		Type:       "withdraw",
		Payload:    payload,
		Sender:     priv.PubKey().Address(),
		Fee:        *new(types.Currency).Set(0),
		LastHeight: "1",
	}
	_tx.Sign(priv)
	rawTx, _ := json.Marshal(_tx)
	return rawTx
}

func makeTestAddress(seed string) crypto.Address {
	privKey := p256.GenPrivKeyFromSecret([]byte(seed))
	addr := privKey.PubKey().Address()
	return addr
}
