package amo

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/blockchain"
	"github.com/amolabs/amoabci/amo/types"
)

func makeValAddr(seed string) crypto.Address {
	priKey := ed25519.GenPrivKeyFromSecret([]byte(seed))
	pubKey := priKey.PubKey().(ed25519.PubKeyEd25519)
	return pubKey.Address()
}

func makeStake(seed string, amount uint64) *types.Stake {
	valPriKey := ed25519.GenPrivKeyFromSecret([]byte(seed))
	valPubKey := valPriKey.PubKey().(ed25519.PubKeyEd25519)
	coins := new(types.Currency).Set(amount)

	stake := types.Stake{
		Validator: valPubKey,
		Amount:    *coins,
	}
	return &stake
}

func makeBB(height int64, seed string, signed bool) abci.RequestBeginBlock {
	addr := makeValAddr(seed)
	lci := abci.LastCommitInfo{}
	if len(seed) > 0 {
		lci = abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       abci.Validator{Address: addr},
				SignedLastBlock: signed,
			}},
		}
	}
	return abci.RequestBeginBlock{
		Header:         abci.Header{Height: height},
		LastCommitInfo: lci,
	}
}

func makeEB(height int64) abci.RequestEndBlock {
	return abci.RequestEndBlock{Height: height}
}

func TestHibernate(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	app := NewAMOApp(tmpFile, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), nil)
	assert.NotNil(t, app)
	app.state.ProtocolVersion = AMOProtocolVersion // tweak
	memDB := tmdb.NewMemDB()
	app.missRuns = blockchain.NewMissRuns(app.store, memDB, 10, 10)
	// to test ValidatorUpdates
	app.store.SetUnlockedStake(makeAccAddr("val1"), makeStake("val1", 100))

	var reqBB abci.RequestBeginBlock
	var reqEB abci.RequestEndBlock
	for i := int64(0); i < 9; i++ {
		reqBB = makeBB(10+i, "val1", false)
		app.BeginBlock(reqBB)
		assert.Equal(t, []crypto.Address{makeValAddr("val1")}, app.missingVals)

		reqEB = makeEB(10 + i)
		res := app.EndBlock(reqEB)
		assert.Equal(t, 0, len(res.ValidatorUpdates))
	}

	reqBB = makeBB(19, "val1", false)
	app.BeginBlock(reqBB)
	assert.Equal(t, []crypto.Address{makeValAddr("val1")}, app.missingVals)

	reqEB = makeEB(19)
	res := app.EndBlock(reqEB)
	hib := app.store.GetHibernate(makeValAddr("val1"), false)
	assert.NotNil(t, hib)
	assert.Equal(t, int64(19), hib.Start)
	assert.Equal(t, int64(29), hib.End)
	assert.Equal(t, 1, len(res.ValidatorUpdates))
	assert.Equal(t, makeValAddr("val1"),
		crypto.AddressHash(res.ValidatorUpdates[0].GetPubKey().Data))
	assert.Equal(t, int64(0), res.ValidatorUpdates[0].GetPower())

	reqBB = makeBB(20, "", false) // XXX we need to check this
	reqEB = makeEB(20)
	app.BeginBlock(reqBB)
	app.EndBlock(reqEB)

	// wake up hibernating validator
	reqBB = makeBB(29, "", false) // XXX we need to check this
	reqEB = makeEB(29)
	app.BeginBlock(reqBB)
	res = app.EndBlock(reqEB)
	// check ev
	assert.Equal(t, "wakeup", res.Events[0].Type)
	var addrBytes crypto.Address
	err := json.Unmarshal(res.Events[0].Attributes[0].Value, &addrBytes)
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(makeValAddr("val1"), addrBytes))
	// check hib
	hib = app.store.GetHibernate(makeValAddr("val1"), false)
	assert.Nil(t, hib)
	assert.Equal(t, 1, len(res.ValidatorUpdates))
	assert.Equal(t, makeValAddr("val1"),
		crypto.AddressHash(res.ValidatorUpdates[0].GetPubKey().Data))
	assert.Equal(t, int64(100), res.ValidatorUpdates[0].GetPower())
}
