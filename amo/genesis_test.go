package amo

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

const testRoot = "genesis_test"

const (
	addr0Json = "BC4BAF38355C6CCF8422DD3D273B3DBB83B2370B"
	t0json    = `{
	  "balances": [
		{
		  "owner": "BC4BAF38355C6CCF8422DD3D273B3DBB83B2370B",
		  "amount": "100"
		}
	  ]
	}`
	t1json = `{
	  "balances": [
		{
		  "owner": "BC4BAF38355C6CCF8422DD3D273B3DBB83B2370B",
		  "amount": "100"
		},
		{
		  "owner": "012F",
		  "amount": "10"
		}
	  ]
	}`
	t2json = `{
	  "balances": [
		{
		  "owner": "BC4BAF38355C6CCF8422DD3D273B3DBB83B2370B",
		  "amount": "100"
		},
		{
		  "owner": "012F",
		  "amount": "10"
		}
	  ],
	  "parcels": []
	}`
	s0json = `{
	  "stakes": [
		{
		  "holder": "BC4BAF38355C6CCF8422DD3D273B3DBB83B2370B",
		  "amount": "100",
		  "validator": "0cOwFQkn9/DTDo1BuqfargBy+1CAPdlQqZpWodbU2F8="
		}
	  ]
	}`
	valAddrJson = "7CECB223B976F27D77B0E03E95602DABCC28D876"
	valBytesHex = "D1C3B0150927F7F0D30E8D41BAA7DAAE0072FB50803DD950A99A56A1D6D4D85F"
)

func TestParseGenesisStateBytes(t *testing.T) {
	var bytes []byte

	stateBytes := []byte(t1json)
	genState, err := ParseGenesisStateBytes(stateBytes)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(genState.Balances))

	bytes, _ = hex.DecodeString(addr0Json)
	assert.Equal(t, crypto.Address(bytes), genState.Balances[0].Owner)
	assert.Equal(t, new(types.Currency).Set(100), &genState.Balances[0].Amount)

	// TODO: need to raise an error for this case
	bytes, _ = hex.DecodeString("012F")
	assert.Equal(t, crypto.Address(bytes), genState.Balances[1].Owner)
	assert.Equal(t, new(types.Currency).Set(10), &genState.Balances[1].Amount)

	// proper balances + garbage data
	stateBytes = []byte(t2json)
	genState, err = ParseGenesisStateBytes(stateBytes)
	// XXX: no member Parcels GenAmoAppState yet, but this should not raise an
	// error
	assert.NoError(t, err)

	bytes, _ = hex.DecodeString(addr0Json)
	assert.Equal(t, crypto.Address(bytes), genState.Balances[0].Owner)
	assert.Equal(t, new(types.Currency).Set(100), &genState.Balances[0].Amount)

	// stakes
	stateBytes = []byte(s0json)
	genState, err = ParseGenesisStateBytes(stateBytes)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(genState.Stakes))
}

func TestFillGenesisState(t *testing.T) {
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())

	// first fill the test store with some values
	addr1 := p256.GenPrivKey().PubKey().Address()
	addr2 := p256.GenPrivKey().PubKey().Address()
	s.SetBalance(addr1, new(types.Currency).Set(10))
	s.SetBalance(addr2, new(types.Currency).Set(20))

	assert.Equal(t, new(types.Currency).Set(10), s.GetBalance(addr1, false))

	genState, err := ParseGenesisStateBytes([]byte(t0json))
	assert.NoError(t, err)
	// this will purge previous data and fill with newly provided genesis state
	err = FillGenesisState(s, genState)
	assert.NoError(t, err)

	// check if the store has been purged prior to fill with genesis state
	assert.Equal(t, new(types.Currency).Set(0), s.GetBalance(addr1, false))
	assert.Equal(t, new(types.Currency).Set(0), s.GetBalance(addr2, false))

	// check if the genesis state is filled correctly
	addr0, _ := hex.DecodeString(addr0Json)
	assert.Equal(t, new(types.Currency).Set(100), s.GetBalance(addr0, false))

	///////////////////////////////
	// genesis with stakes
	genState, err = ParseGenesisStateBytes([]byte(s0json))
	assert.NoError(t, err)
	err = FillGenesisState(s, genState)
	assert.NoError(t, err)
	stake := s.GetStake(addr0, false)
	assert.NotNil(t, stake)
	assert.Equal(t, new(types.Currency).Set(100), &stake.Amount)
	valBytes, _ := hex.DecodeString(valBytesHex)
	assert.Equal(t, valBytes, []byte(stake.Validator[:]))
	valAddr, _ := hex.DecodeString(valAddrJson)
	assert.Equal(t, addr0, s.GetHolderByValidator(valAddr, false))
}
