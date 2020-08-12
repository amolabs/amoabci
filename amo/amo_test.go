package amo

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/blockchain"
	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/tx"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

func makeAccAddr(seed string) crypto.Address {
	return p256.GenPrivKeyFromSecret([]byte(seed)).PubKey().Address()
}

func TestAppConfig(t *testing.T) {
	// test genesis app config
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
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

	tmp, err := new(types.Currency).SetString(defaultBlkReward, 10)
	assert.NoError(t, err)
	assert.Equal(t, *tmp, app.config.BlkReward)

	tmp, err = new(types.Currency).SetString(defaultTxReward, 10)
	assert.NoError(t, err)
	assert.Equal(t, *tmp, app.config.TxReward)

	assert.Equal(t, int64(2), app.config.LockupPeriod)
}

func TestInitChain(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
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
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	// query
	req := abci.RequestQuery{}
	req.Path = "/nostore"
	res := app.Query(req)
	assert.Equal(t, code.QueryCodeBadPath, res.Code)
}

func TestQueryAppConfig(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
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
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
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

func TestQueryStorage(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	storageID1 := uint32(123)
	storage := types.Storage{
		Owner: makeAccAddr("any"),
	}
	app.store.SetStorage(storageID1, &storage)
	app.store.Save()

	// query vars
	var req abci.RequestQuery
	var res abci.ResponseQuery
	var barr []byte

	// no key
	req = abci.RequestQuery{Path: "/storage"}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoKey, res.Code)

	// nonexistent storage id
	storageID2 := uint32(456)
	barr, _ = json.Marshal(storageID2)
	req = abci.RequestQuery{Path: "/storage", Data: []byte(barr)}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)
	assert.Equal(t, []byte(nil), res.Value)
	assert.Equal(t, "error: no such storage", res.Log)

	// valid match
	barr, _ = json.Marshal(storageID1)
	req = abci.RequestQuery{Path: "/storage", Data: []byte(barr)}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	barr, _ = json.Marshal(storage)
	assert.Equal(t, barr, res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(barr), res.Log)
}

func TestQueryParcel(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	_addrbin, _ := hex.DecodeString("FFECB223B976F27D77B0E03E95602DABCC28D876")
	_addr := crypto.Address(_addrbin)

	parcelID := tmbytes.HexBytes(tmrand.Bytes(32))
	queryjson, _ := json.Marshal(parcelID)

	parcel := types.Parcel{
		Owner:   addr,
		Custody: tmrand.Bytes(32),
	}

	request := types.Request{
		Payment: *new(types.Currency).Set(1),
	}

	parcelEx := types.ParcelEx{
		Parcel: &parcel,
		Requests: []*types.RequestEx{
			{
				Request: &request,
			},
		},
		Usages: []*types.UsageEx{},
	}

	app.store.SetParcel(parcelID, &parcel)
	app.store.SetRequest(_addr, parcelID, &request)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

	wrongParcelID := tmbytes.HexBytes(tmrand.Bytes(32))
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
	assert.Equal(t, "error: no such parcel", res.Log)

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
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := tmrand.Bytes(32)
	parcelID[31] = 0xFF

	request := types.Request{
		Payment: *new(types.Currency).Set(400),
	}

	requestEx := types.RequestEx{
		Request:   &request,
		Recipient: addr,
	}

	app.store.SetRequest(addr, parcelID, &request)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

	wrongParcelID := tmrand.Bytes(32)
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

	var keyMap = map[string]tmbytes.HexBytes{
		"recipient": addr,
		"target":    parcelID,
	}

	// query
	key, _ := json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/request", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(requestEx)
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)

	keyMap["recipient"] = wrongAddr
	key, _ = json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/request", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)

	delete(keyMap, "recipient")
	key, _ = json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/request", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
}

func TestQueryUsage(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// populate db store
	addrbin, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrbin)
	parcelID := tmrand.Bytes(32)
	parcelID[31] = 0xFF

	usage := types.Usage{
		Custody: tmrand.Bytes(32),
	}

	usageEx := types.UsageEx{
		Usage:     &usage,
		Recipient: addr,
	}

	app.store.SetUsage(addr, parcelID, &usage)

	_, _, err := app.store.Save()
	assert.NoError(t, err)

	wrongParcelID := tmrand.Bytes(32)
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

	var keyMap = map[string]tmbytes.HexBytes{
		"recipient": addr,
		"target":    parcelID,
	}

	// query
	key, _ := json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/usage", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	jsonstr, _ = json.Marshal(usageEx)
	assert.Equal(t, []byte(jsonstr), res.Value)
	assert.Equal(t, req.Data, res.Key)
	assert.Equal(t, string(jsonstr), res.Log)

	keyMap["recipient"] = wrongAddr
	key, _ = json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/usage", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)

	delete(keyMap, "recipient")
	key, _ = json.Marshal(keyMap)
	req = abci.RequestQuery{Path: "/usage", Data: key}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeBadKey, res.Code)
}

func TestQueryValidator(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

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

	req = abci.RequestQuery{Path: "/stake", Data: []byte("\"BCECB223B976F27D77B0E03E95602DABCC28D876\"")}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
}

func TestSignedTransactionTest(t *testing.T) {
	from := p256.GenPrivKeyFromSecret([]byte("alice"))

	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion

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

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 1}})

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
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion

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
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion
	app.config.LazinessWindow = 4
	app.config.LazinessThreshold = 2

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

	app.BeginBlock(abci.RequestBeginBlock{
		Header:         abci.Header{Height: 2},
		LastCommitInfo: lastCommitInfo,
	})
	app.EndBlock(abci.RequestEndBlock{})

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

	app.BeginBlock(abci.RequestBeginBlock{
		Header:         abci.Header{Height: 4},
		LastCommitInfo: lastCommitInfo,
	})
	app.EndBlock(abci.RequestEndBlock{})

	app.BeginBlock(abci.RequestBeginBlock{
		Header:         abci.Header{Height: 5},
		LastCommitInfo: lastCommitInfo,
	})
	app.EndBlock(abci.RequestEndBlock{})

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
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion

	// setup
	tx.ConfigAMOApp.LockupPeriod = 1                               // manipulate
	tx.ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(100) // manipulate
	priv1 := p256.GenPrivKeyFromSecret([]byte("staker1"))
	app.store.SetBalance(priv1.PubKey().Address(), new(types.Currency).Set(500))
	priv2 := p256.GenPrivKeyFromSecret([]byte("staker2"))
	app.store.SetBalance(priv2.PubKey().Address(), new(types.Currency).Set(500))

	// immitate initChain() function call
	_, _, err := app.store.Save()
	assert.NoError(t, err)

	// begin block
	blkRequest := abci.RequestBeginBlock{Header: abci.Header{Height: 1}}
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
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion
	tx.ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(50)

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

	balD1 := app.store.GetBalance(d1Priv.PubKey().Address(), true)
	balD2 := app.store.GetBalance(d2Priv.PubKey().Address(), true)
	balS := app.store.GetBalance(sPriv.PubKey().Address(), true)

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

	res := app.EndBlock(abci.RequestEndBlock{Height: 1})

	app.Commit()
	tx.ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(50)

	// check incentive records
	events := res.GetEvents()
	assert.Equal(t, 1, len(events))

	balS = app.store.GetBalance(sPriv.PubKey().Address(), true).Sub(balS)

	amountS := new(types.Currency).Set(0)
	err := json.Unmarshal(events[0].Attributes[1].GetValue(), amountS)
	assert.NoError(t, err)

	amountS = DivCurrency(amountS, divisor)
	balS = DivCurrency(balS, divisor)

	assert.Equal(t, amountS, balS)

	balD1 = app.store.GetBalance(d1Priv.PubKey().Address(), true)
	balD2 = app.store.GetBalance(d2Priv.PubKey().Address(), true)
	balS = app.store.GetBalance(sPriv.PubKey().Address(), true)

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

	res = app.EndBlock(abci.RequestEndBlock{Height: 2})

	app.Commit()
	tx.ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(50)

	// check incentive records
	events = res.GetEvents()
	assert.Equal(t, 3, len(events))

	sort.Slice(events, func(i, j int) bool {
		tmpI := new(types.Currency).Set(0)
		err = json.Unmarshal(events[i].Attributes[1].GetValue(), tmpI)
		tmpJ := new(types.Currency).Set(0)
		err = json.Unmarshal(events[j].Attributes[1].GetValue(), tmpJ)
		return tmpI.LessThan(tmpJ)
	})

	balD1 = app.store.GetBalance(d1Priv.PubKey().Address(), true).Sub(balD1)
	balD2 = app.store.GetBalance(d2Priv.PubKey().Address(), true).Sub(balD2)
	balS = app.store.GetBalance(sPriv.PubKey().Address(), true).Sub(balS)

	amountD1 := new(types.Currency).Set(0)
	err = json.Unmarshal(events[0].Attributes[1].GetValue(), amountD1)
	assert.NoError(t, err)
	t.Logf("%s", events[0].Attributes[1].GetValue())
	amountD2 := new(types.Currency).Set(0)
	err = json.Unmarshal(events[1].Attributes[1].GetValue(), amountD2)
	assert.NoError(t, err)
	t.Logf("%s", events[1].Attributes[1].GetValue())
	amountS = new(types.Currency).Set(0)
	err = json.Unmarshal(events[2].Attributes[1].GetValue(), amountS)
	assert.NoError(t, err)
	t.Logf("%s", events[2].Attributes[1].GetValue())

	amountD1 = DivCurrency(amountD1, divisor)
	amountD2 = DivCurrency(amountD2, divisor)
	amountS = DivCurrency(amountS, divisor)

	balD1 = DivCurrency(balD1, divisor)
	balD2 = DivCurrency(balD2, divisor)
	balS = DivCurrency(balS, divisor)

	assert.Equal(t, amountD1, balD1)
	assert.Equal(t, amountD2, balD2)
	assert.Equal(t, amountS, balS)
}

func TestIncentiveNoTouch(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion

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

	_, err = blockchain.DistributeIncentive(
		app.store,
		app.logger,
		app.config.WeightValidator, app.config.WeightDelegator,
		app.config.BlkReward, app.config.TxReward,
		app.numDeliveredTxs,
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
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion

	// init chain
	app.InitChain(abci.RequestInitChain{})

	// setup
	tx.ConfigAMOApp.LockupPeriod = 2                               // manipulate
	tx.ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(100) // manipulate
	priv := p256.GenPrivKeyFromSecret([]byte("test"))
	app.store.SetBalance(priv.PubKey().Address(), new(types.Currency).Set(500))

	// begin block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 1}})

	rawTx := makeTxStake(priv, "test", 500, "1")
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: rawTx})
	assert.Equal(t, code.TxCodeOK, res.Code)

	// end block
	app.EndBlock(abci.RequestEndBlock{Height: 1})

	// commit
	app.Commit()

	stakes := app.store.GetLockedStakes(makeTestAddress("test"), true)
	assert.Equal(t, 1, len(stakes))

	// get hash to compare
	prevHash := app.state.LastAppHash

	// simulate no txs to process in this block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	app.EndBlock(abci.RequestEndBlock{Height: 2})
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
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 3}})
	app.EndBlock(abci.RequestEndBlock{Height: 3})
	app.Commit()

	// get hash to compare
	hash = app.state.LastAppHash

	// should equal as the stake lock-up ends
	assert.Equal(t, prevHash, hash)
}

func TestReplayAttack(t *testing.T) {
	t1 := p256.GenPrivKeyFromSecret([]byte("test1"))
	tx1 := makeTxStake(t1, "test1", 10000, "1")
	tx2 := makeTxStake(t1, "test1", 10000, "1")
	tx3 := makeTxStake(t1, "test1", 10000, "1")

	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion
	app.config.BlockBindingWindow = int64(3)
	app.replayPreventer = blockchain.NewReplayPreventer(
		app.store,
		app.state.LastHeight,
		app.config.BlockBindingWindow,
	)

	tx.ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(100) // manipulate

	app.store.SetBalance(t1.PubKey().Address(), new(types.Currency).Set(40000))

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 1}})

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)

	assert.Equal(t, code.TxCodeImproperTx, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeImproperTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 1})

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})

	assert.Equal(t, code.TxCodeImproperTx, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeImproperTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx2}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx2}).Code)

	assert.Equal(t, code.TxCodeImproperTx, app.CheckTx(abci.RequestCheckTx{Tx: tx2}).Code)
	assert.Equal(t, code.TxCodeImproperTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx2}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 2})

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 3}})

	assert.Equal(t, code.TxCodeImproperTx, app.CheckTx(abci.RequestCheckTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeImproperTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx1}).Code)
	assert.Equal(t, code.TxCodeImproperTx, app.CheckTx(abci.RequestCheckTx{Tx: tx2}).Code)
	assert.Equal(t, code.TxCodeImproperTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx2}).Code)

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx3}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx3}).Code)

	assert.Equal(t, code.TxCodeImproperTx, app.CheckTx(abci.RequestCheckTx{Tx: tx3}).Code)
	assert.Equal(t, code.TxCodeImproperTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx3}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 3})
}

func TestBindingBlock(t *testing.T) {
	t1 := p256.GenPrivKeyFromSecret([]byte("test1"))
	tx1 := makeTxStake(t1, "test1", 10000, "1")
	tx2 := makeTxStake(t1, "test1", 10000, "2")
	tx3 := makeTxStake(t1, "test1", 10000, "3")
	tx4 := makeTxStake(t1, "test1", 10000, "1")
	tx5 := makeTxStake(t1, "test1", 10000, "2")

	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion
	app.config.BlockBindingWindow = int64(3)
	app.replayPreventer = blockchain.NewReplayPreventer(
		app.store,
		app.state.LastHeight,
		app.config.BlockBindingWindow,
	)

	tx.ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(100) // manipulate

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

	assert.Equal(t, code.TxCodeImproperTx, app.CheckTx(abci.RequestCheckTx{Tx: tx4}).Code)
	assert.Equal(t, code.TxCodeImproperTx, app.DeliverTx(abci.RequestDeliverTx{Tx: tx4}).Code)

	assert.Equal(t, code.TxCodeOK, app.CheckTx(abci.RequestCheckTx{Tx: tx5}).Code)
	assert.Equal(t, code.TxCodeOK, app.DeliverTx(abci.RequestDeliverTx{Tx: tx5}).Code)

	app.EndBlock(abci.RequestEndBlock{Height: 4})
}

func TestGovernance(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	app.state.ProtocolVersion = AMOProtocolVersion

	// manipulate InitChain() func
	app.state.NextDraftID = uint32(1)

	err := app.loadAppConfig()
	assert.NoError(t, err)

	// manipulate draft related configs
	app.config.DraftDeposit = *new(types.Currency).Set(1000)
	app.config.DraftQuorumRate = float64(0.6)
	app.config.DraftPassRate = float64(0.51)
	app.config.DraftRefundRate = float64(0.25)

	tx.ConfigAMOApp = app.config

	// prepare validator set
	p := prepForGov(app.store, "p", 1000)
	v1 := prepForGov(app.store, "v1", 1000)
	v2 := prepForGov(app.store, "v2", 1000)
	v3 := prepForGov(app.store, "v3", 1000)
	v4 := prepForGov(app.store, "v4", 1000)
	v5 := prepForGov(app.store, "v5", 1000)
	v6 := prepForGov(app.store, "v6", 1000)
	v7 := prepForGov(app.store, "v7", 1000)
	v8 := prepForGov(app.store, "v8", 1000)
	v9 := prepForGov(app.store, "v9", 1000)
	v10 := prepForGov(app.store, "v10", 1000)
	v11 := prepForGov(app.store, "v11", 1000)
	v12 := prepForGov(app.store, "v12", 1000)
	v13 := prepForGov(app.store, "v13", 1000)
	v14 := prepForGov(app.store, "v14", 1000)

	// test for draft being approved, deposit returned to proposer
	// total: 15, voters: 10(yay: 6, nay: 4), non-voters: 5

	// check target value before draft application
	tmp, err := new(types.Currency).SetString(defaultTxReward, 10)
	assert.NoError(t, err)
	assert.Equal(t, *tmp, app.config.TxReward)
	assert.Equal(t, defaultLockupPeriod, app.config.LockupPeriod)

	// proposer propose a draft in height 1
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 1}})

	draftID := uint32(1)
	cfg := app.config
	cfg.TxReward = *types.Zero
	cfg.LockupPeriod = 10000
	desc := "I want others to get no reward and stay locked shorter"

	// imitate 'propose' tx
	app.store.SetDraft(draftID, &types.Draft{
		Proposer:     p,
		Config:       cfg,
		Desc:         desc,
		OpenCount:    int64(1),
		CloseCount:   int64(1),
		ApplyCount:   int64(1),
		Deposit:      app.config.DraftDeposit,
		TallyQuorum:  *types.Zero,
		TallyApprove: *types.Zero,
		TallyReject:  *types.Zero,
	})

	// withdraw draft deposit from proposer's balance
	balance := app.store.GetBalance(p, false)
	balance.Sub(&app.config.DraftDeposit)
	app.store.SetBalance(p, balance)
	assert.Equal(t, types.Zero, app.store.GetBalance(p, false))

	// imitate a job done after 'propose' tx is successfully processed
	app.state.NextDraftID += uint32(1)

	app.EndBlock(abci.RequestEndBlock{Height: 1})

	// check if draft is properly stored
	draft := app.store.GetDraft(draftID, false)
	assert.Equal(t, int64(0), draft.OpenCount)
	assert.Equal(t, int64(1), draft.CloseCount)
	assert.Equal(t, int64(1), draft.ApplyCount)

	// voters vote for the draft in height 2
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})

	app.store.SetVote(draftID, v1, &types.Vote{Approve: true})
	app.store.SetVote(draftID, v2, &types.Vote{Approve: true})
	app.store.SetVote(draftID, v3, &types.Vote{Approve: true})
	app.store.SetVote(draftID, v4, &types.Vote{Approve: true})
	app.store.SetVote(draftID, v5, &types.Vote{Approve: true})
	app.store.SetVote(draftID, v6, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v7, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v8, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v9, &types.Vote{Approve: false})

	app.EndBlock(abci.RequestEndBlock{Height: 2})

	// check if vote is closed and tally_* values are properly calculated
	draft = app.store.GetDraft(draftID, false)
	assert.Equal(t, int64(0), draft.OpenCount)
	assert.Equal(t, int64(0), draft.CloseCount)
	assert.Equal(t, int64(1), draft.ApplyCount)
	assert.Equal(t, *new(types.Currency).Set(6000), draft.TallyApprove)
	assert.Equal(t, *new(types.Currency).Set(4000), draft.TallyReject)

	// check if draft deposit is returned to proposer
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(p, false))

	// check if draft deposit is not distributed to voters
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v1, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v2, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v3, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v4, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v5, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v6, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v7, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v8, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v9, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v10, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v11, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v12, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v13, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v14, false))

	// wait for draft to get applied for 1 at height 3
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 3}})
	app.EndBlock(abci.RequestEndBlock{Height: 3})

	// check if draft's counts are proper
	draft = app.store.GetDraft(draftID, false)
	assert.Equal(t, int64(0), draft.OpenCount)
	assert.Equal(t, int64(0), draft.CloseCount)
	assert.Equal(t, int64(0), draft.ApplyCount)

	// imitate Commit() to load new app config
	_, _, err = app.store.Save()
	assert.NoError(t, err)
	err = app.loadAppConfig()
	assert.NoError(t, err)

	// after target
	assert.Equal(t, *types.Zero, app.config.TxReward)
	assert.Equal(t, int64(10000), app.config.LockupPeriod)

	// test for draft being rejected, deposit distributed to voters
	// total: 15, voters: 10(yay: 2, nay: 8), non-voters: 5

	// check target value before draft application
	assert.Equal(t, defaultBlockBindingWindow, app.config.BlockBindingWindow)

	// proposer propose a draft in height 4
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 4}})

	draftID = uint32(2)
	cfg = app.config
	cfg.BlockBindingWindow = int64(100000000)
	desc = "block_bound_tx_grace_period should be longer for no reason"

	// imitate 'propose' tx
	app.store.SetDraft(draftID, &types.Draft{
		Proposer:     p,
		Config:       cfg,
		Desc:         desc,
		OpenCount:    int64(1),
		CloseCount:   int64(1),
		ApplyCount:   int64(1),
		Deposit:      app.config.DraftDeposit,
		TallyQuorum:  *types.Zero,
		TallyApprove: *types.Zero,
		TallyReject:  *types.Zero,
	})

	// withdraw draft deposit from proposer's balance
	balance = app.store.GetBalance(p, false)
	balance.Sub(&app.config.DraftDeposit)
	app.store.SetBalance(p, balance)
	assert.Equal(t, types.Zero, app.store.GetBalance(p, false))

	// imitate a job done after 'propose' tx is successfully processed
	app.state.NextDraftID += uint32(1)

	app.EndBlock(abci.RequestEndBlock{Height: 4})

	// check if draft is properly stored
	draft = app.store.GetDraft(draftID, false)
	assert.Equal(t, int64(0), draft.OpenCount)
	assert.Equal(t, int64(1), draft.CloseCount)
	assert.Equal(t, int64(1), draft.ApplyCount)

	// voters vote for the draft in height 5
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 5}})

	app.store.SetVote(draftID, v1, &types.Vote{Approve: true})
	app.store.SetVote(draftID, v2, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v3, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v4, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v5, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v6, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v7, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v8, &types.Vote{Approve: false})
	app.store.SetVote(draftID, v9, &types.Vote{Approve: false})

	app.EndBlock(abci.RequestEndBlock{Height: 5})

	// check if vote is closed completely and tally_* values are properly calculated
	draft = app.store.GetDraft(draftID, false)
	assert.Equal(t, int64(0), draft.OpenCount)
	assert.Equal(t, int64(0), draft.CloseCount)
	assert.Equal(t, int64(0), draft.ApplyCount)
	assert.Equal(t, *new(types.Currency).Set(2000), draft.TallyApprove)
	assert.Equal(t, *new(types.Currency).Set(8000), draft.TallyReject)

	// check if draft deposit is not returned to proposer
	assert.Equal(t, types.Zero, app.store.GetBalance(p, false))

	// check if draft deposit is distributed to voters, not to non-voters
	assert.Equal(t, new(types.Currency).Set(1111), app.store.GetBalance(v1, false))
	assert.Equal(t, new(types.Currency).Set(1111), app.store.GetBalance(v2, false))
	assert.Equal(t, new(types.Currency).Set(1111), app.store.GetBalance(v3, false))
	assert.Equal(t, new(types.Currency).Set(1111), app.store.GetBalance(v4, false))
	assert.Equal(t, new(types.Currency).Set(1111), app.store.GetBalance(v5, false))
	assert.Equal(t, new(types.Currency).Set(1111), app.store.GetBalance(v6, false))
	assert.Equal(t, new(types.Currency).Set(1111), app.store.GetBalance(v7, false))
	assert.Equal(t, new(types.Currency).Set(1111), app.store.GetBalance(v8, false))
	assert.Equal(t, new(types.Currency).Set(1111), app.store.GetBalance(v9, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v10, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v11, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v12, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v13, false))
	assert.Equal(t, new(types.Currency).Set(1000), app.store.GetBalance(v14, false))

	// after target: should be same as befor drafte
	assert.Equal(t, defaultBlockBindingWindow, app.config.BlockBindingWindow)
}

func TestProtocolUpgrade(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	// manipulate
	app.state.LastHeight = 8
	app.state.ProtocolVersion = AMOProtocolVersion
	app.config.UpgradeProtocolHeight = 10
	app.config.UpgradeProtocolVersion = AMOProtocolVersion + 1
	b, err := json.Marshal(app.config)
	assert.NoError(t, err)
	err = app.store.SetAppConfig(b)
	assert.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 9}})
	app.EndBlock(abci.RequestEndBlock{Height: 9})
	app.Commit()

	assert.Equal(t, uint64(AMOProtocolVersion), app.state.ProtocolVersion)

	// manipulate to avoid panic
	app.state.Height = 10
	app.upgradeProtocol()

	assert.Equal(t, uint64(AMOProtocolVersion+1), app.state.ProtocolVersion)

	err = app.checkProtocolVersion()
	assert.Error(t, err)
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

func prepForGov(s *store.Store, seed string, amount uint64) crypto.Address {
	validator := ed25519.GenPrivKeyFromSecret([]byte(seed))
	holder := p256.GenPrivKeyFromSecret([]byte(seed))

	s.SetBalance(holder.PubKey().Address(), new(types.Currency).Set(amount))
	s.SetUnlockedStake(holder.PubKey().Address(), &types.Stake{
		Validator: validator.PubKey().(ed25519.PubKeyEd25519),
		Amount:    *new(types.Currency).Set(amount),
	})

	return holder.PubKey().Address()
}

func TestQueryDID(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	var req abci.RequestQuery
	var res abci.ResponseQuery
	var jsonstr []byte

	req = abci.RequestQuery{Path: "/did"}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoKey, res.Code)

	var jsonDoc = []byte(`{"jsonkey":"jsonvalue"}`)
	entry := &types.DIDEntry{Owner: makeAccAddr("me"), Document: jsonDoc}
	app.store.SetDIDEntry("myid", entry)

	req = abci.RequestQuery{Path: "/did", Data: []byte(`"myid"`)}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)
	app.store.Save()
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	assert.Equal(t, []byte(`"myid"`), res.Key)
	jsonstr, _ = json.Marshal(entry)
	assert.Equal(t, jsonstr, res.Value)

	app.store.DeleteDIDEntry("myid")
	app.store.Save()

	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)
}

func TestQueryHibernate(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)

	var req abci.RequestQuery
	var res abci.ResponseQuery
	var jsonstr []byte

	req = abci.RequestQuery{Path: "/hibernate"}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoKey, res.Code)

	hib := &types.Hibernate{Start: 100, End: 200}
	app.store.SetHibernate(makeValAddr("val1"), hib)

	jsonstr, _ = json.Marshal(makeValAddr("val1"))
	req = abci.RequestQuery{Path: "/hibernate", Data: jsonstr}
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)

	app.store.Save()
	res = app.Query(req)
	assert.Equal(t, code.QueryCodeOK, res.Code)
	assert.Equal(t, jsonstr, res.Key)
	jsonstr, _ = json.Marshal(hib)
	assert.Equal(t, jsonstr, res.Value)

	app.store.DeleteHibernate(makeValAddr("val1"))
	app.store.Save()

	res = app.Query(req)
	assert.Equal(t, code.QueryCodeNoMatch, res.Code)
}
