package amo

import (
	"encoding/hex"
	"testing"

	abci "github.com/amolabs/tendermint-amo/abci/types"
	"github.com/amolabs/tendermint-amo/crypto"
	tdb "github.com/amolabs/tendermint-amo/libs/db"
	"github.com/stretchr/testify/assert"

	"github.com/amolabs/amoabci/amo/types"
)

func TestInitChain(t *testing.T) {
	db := tdb.NewMemDB()
	app := NewAMOApplication(db)
	req := abci.RequestInitChain{}
	req.AppStateBytes = []byte(`{ "balances": [ { "owner": "7CECB223B976F27D77B0E03E95602DABCC28D876", "amount": "100" } ] }`)
	res := app.InitChain(req)
	// TODO: need to check the contents of the response
	assert.Equal(t, abci.ResponseInitChain{}, res)

	// TODO: run series of app.Query() to check the genesis state
	addrBytes, _ := hex.DecodeString("7CECB223B976F27D77B0E03E95602DABCC28D876")
	addr := crypto.Address(addrBytes)
	assert.Equal(t, types.Currency(100), app.store.GetBalance(addr))
	//queryReq := abci.RequestQuery{}
	//queryRes := app.Query(queryReq)
}
