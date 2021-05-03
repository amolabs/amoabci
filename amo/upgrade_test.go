package amo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
)

func TestGetProtocolVersion(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	assert.Nil(t, app.proto)

	// app.store is in near-genesis state
	ver := app.state.ProtocolVersion
	assert.Equal(t, AMOGenesisProtocolVersion, ver) // which is 0x3

	// manipulate
	// NOTE: This can be done since we are writing a SW in retrospective
	// manner. That is, we already observed a state DB which holds data
	// produced via protocol greater than 3.
	app.config.UpgradeProtocolHeight = 3
	app.config.UpgradeProtocolVersion = 4
	jsonStr, _ := json.Marshal(app.config)
	app.store.SetAppConfig(jsonStr)
	app.store.Save() // emulate Save in InitChain

	app.store.Save() // save height 1
	// assume restart took place here
	app.load()
	assert.Equal(t, int64(2), app.store.GetMerkleVersion())
	// at height 1, still in protocol version 3
	assert.Equal(t, uint64(0x3), app.state.ProtocolVersion)

	app.store.Save() // save height 2
	app.store.Save() // save height 3
	// assume restart took place here
	app.load()
	assert.Equal(t, int64(4), app.store.GetMerkleVersion())
	assert.Equal(t, int64(3), app.state.Height)
	// at height 2, should be in protocol version 4
	assert.Equal(t, uint64(0x4), app.state.ProtocolVersion)
}

func TestProtocolUpgrade(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	assert.Equal(t, AMOGenesisProtocolVersion, app.state.ProtocolVersion)
	assert.Equal(t, int64(0), app.config.UpgradeProtocolHeight)
	assert.Equal(t, uint64(0), app.config.UpgradeProtocolVersion)
	assert.Nil(t, app.proto)

	// schedule protocol upgrade
	app.state.LastHeight = 8
	app.state.ProtocolVersion = 0x4
	app.config.UpgradeProtocolHeight = 10
	app.config.UpgradeProtocolVersion = 0x5
	b, err := json.Marshal(app.config)
	assert.NoError(t, err)
	err = app.store.SetAppConfig(b)
	assert.NoError(t, err)

	// protocol version 4
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 9}})
	// no change in protocol version
	assert.Equal(t, uint64(0x4), app.state.ProtocolVersion)
	assert.NotNil(t, app.proto)
	assert.Equal(t, uint64(0x4), app.proto.Version())
	// transfer v5 tx: should be rejected
	tx1 := []byte(`{"type":"transfer","sender":"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","payload":{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","parcel":"00000010EFEF"},"signature":{"pubkey":"0485FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B185FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B1EF1D55E9B1EF1D","sig_bytes":"FFFFFFFF"}}`)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: tx1})
	assert.Equal(t, code.TxCodeInvalidAmount, res.Code)
	//
	app.EndBlock(abci.RequestEndBlock{Height: 9})
	app.Commit()

	// protocol 4 -> 5
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 10}})
	// now protocol version 5
	assert.Equal(t, uint64(0x5), app.state.ProtocolVersion)
	assert.NotNil(t, app.proto)
	assert.Equal(t, uint64(0x5), app.proto.Version())
	// transfer v5 tx again: should be accepted
	// (but rejected since the parcel is not found)
	tx2 := []byte(`{"type":"transfer","sender":"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","payload":{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","parcel":"00000010EFEF"},"signature":{"pubkey":"0485FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B185FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B1EF1D55E9B1EF1D","sig_bytes":"AFFFFFFF"}}`)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: tx2})
	assert.Equal(t, code.TxCodeParcelNotFound, res.Code)
	//
	app.EndBlock(abci.RequestEndBlock{Height: 10})
	app.Commit()

	app.config.UpgradeProtocolHeight = 11
	app.config.UpgradeProtocolVersion = 0x6

	// The following will panic, so we will use a different testing point.
	//b, err = json.Marshal(app.config)
	//assert.NoError(t, err)
	//err = app.store.SetAppConfig(b)
	//assert.NoError(t, err)
	//app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 11}})
	//app.EndBlock(abci.RequestEndBlock{Height: 11})
	//app.Commit()
	app.state.Height = 11
	app.upgradeProtocol()

	assert.Equal(t, uint64(0x6), app.state.ProtocolVersion)
	assert.Nil(t, app.proto)
	err = checkProtocolVersion(app.state.ProtocolVersion)
	assert.Error(t, err) // protocol version 6 is not supported
}

