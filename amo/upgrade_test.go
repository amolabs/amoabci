package amo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
)

func TestProtocolUpgrade(t *testing.T) {
	app := NewAMOApp(1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	assert.Nil(t, app.proto)

	// app.store is in near-genesis state
	ver := app.state.ProtocolVersion
	assert.Equal(t, AMOGenesisProtocolVersion, ver) // which is 0x3

	app.store.Save() // emulate Save in InitChain
	app.store.Save() // save height 1

	// save protocol 3 config
	var configV3 struct {
		LazinessCounterWindow  int64  `json:"laziness_counter_window"`
		UpgradeProtocolHeight  int64  `json:"upgrade_protocol_height"`
		UpgradeProtocolVersion uint64 `json:"upgrade_protocol_version"`
	}
	configV3.LazinessCounterWindow = 100
	configV3.UpgradeProtocolHeight = 5
	configV3.UpgradeProtocolVersion = 4
	jsonStr, _ := json.Marshal(configV3)
	app.store.SetAppConfig(jsonStr)
	app.store.Save() // save height 2

	app.load() // assume restart took place here
	assert.Equal(t, int64(3), app.store.GetMerkleVersion())
	// at height 1, still in protocol version 3
	assert.Equal(t, uint64(0x3), app.state.ProtocolVersion)

	// NOTE: This sw cannot deal with protocol v3, so just save configV4 to
	// simulate protocol upgrade from v3 to v4.
	// save protocol 3 config
	var configV4 struct {
		LazinessWindow  int64  `json:"laziness_window"`
		UpgradeProtocolHeight  int64  `json:"upgrade_protocol_height"`
		UpgradeProtocolVersion uint64 `json:"upgrade_protocol_version"`
	}
	configV4.LazinessWindow = 100
	configV4.UpgradeProtocolHeight = 10
	configV4.UpgradeProtocolVersion = 5
	jsonStr, _ = json.Marshal(configV4)
	app.store.SetAppConfig(jsonStr)
	app.store.Save() // save height 3

	app.load() // assume restart took place here
	assert.Equal(t, int64(4), app.store.GetMerkleVersion())
	assert.Equal(t, int64(3), app.state.Height)
	// at height 3, should be in protocol version 4
	assert.Equal(t, uint64(0x4), app.state.ProtocolVersion)

	app.store.Save() // save height 4
	app.store.Save() // save height 5
	app.store.Save() // save height 6
	app.store.Save() // save height 7
	app.store.Save() // save height 8
	app.store.Save() // save height 9

	app.load() // assume restart took place here
	assert.Equal(t, int64(10), app.store.GetMerkleVersion())
	assert.Equal(t, int64(9), app.state.Height)
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 10}})
	assert.Equal(t, int64(10), app.state.Height)
	// at height 10, should be in protocol version 4
	assert.Equal(t, uint64(0x5), app.state.ProtocolVersion)
}

func TestProtocolDifference(t *testing.T) {
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

