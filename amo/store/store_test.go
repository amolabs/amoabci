package store

import (
	"os"
	"testing"
	"time"

	"github.com/amolabs/tendermint-amo/crypto/p256"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"github.com/amolabs/tendermint-amo/libs/db"
	"github.com/stretchr/testify/assert"

	"github.com/amolabs/amoabci/amo/types"
)

const testRoot = "store_test"

func setUp(t *testing.T) {
	err := cmn.EnsureDir(testRoot, 0700)
	if err != nil {
		t.Fatal(err)
	}
}

func tearDown(t *testing.T) {
	err := os.RemoveAll(testRoot)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBalance(t *testing.T) {
	setUp(t)
	s := NewStore(db.NewMemDB())
	testAddr := p256.GenPrivKey().PubKey().Address()
	balance := types.Currency(100)
	s.SetBalance(testAddr, balance)
	assert.Equal(t, balance, s.GetBalance(testAddr))
	tearDown(t)
}

func TestParcel(t *testing.T) {
	setUp(t)
	s := NewStore(db.NewMemDB())
	testAddr := p256.GenPrivKey().PubKey().Address()
	custody := cmn.RandBytes(32)
	parcelInput := types.ParcelValue{
		Owner:   testAddr,
		Custody: custody,
		Info:    []byte("test"),
	}
	parcelID := cmn.RandBytes(32)
	s.SetParcel(parcelID, &parcelInput)
	parcelOutput := s.GetParcel(parcelID)
	assert.Equal(t, parcelInput, *parcelOutput)
	t.Log(parcelInput)
	t.Log(*parcelOutput)
	tearDown(t)
}

func TestRequest(t *testing.T) {
	setUp(t)
	s := NewStore(db.NewMemDB())
	testAddr := p256.GenPrivKey().PubKey().Address()
	parcelID := cmn.RandBytes(32)
	exp := time.Now().UTC()
	exp = exp.Add(100 * time.Minute)
	requestInput := types.RequestValue{
		Payment: types.Currency(100),
		Exp:     exp,
	}
	s.SetRequest(testAddr, parcelID, &requestInput)
	requestOutput := s.GetRequest(testAddr, parcelID)
	assert.Equal(t, requestInput.Payment, (*requestOutput).Payment)
	assert.Equal(t, requestInput.Exp.Unix(), (*requestOutput).Exp.Unix())
	assert.False(t, requestOutput.IsExpired())
	t.Log(requestInput)
	t.Log(*requestOutput)
	tearDown(t)
}

func TestUsage(t *testing.T) {
	setUp(t)
	s := NewStore(db.NewMemDB())
	testAddr := p256.GenPrivKey().PubKey().Address()
	parcelID := cmn.RandBytes(32)
	custody := cmn.RandBytes(32)
	exp := time.Now().UTC()
	exp = exp.Add(100 * time.Minute)
	usageInput := types.UsageValue{
		Custody: custody,
		Exp:     exp,
	}
	s.SetUsage(testAddr, parcelID, &usageInput)
	usageOutput := s.GetUsage(testAddr, parcelID)
	assert.Equal(t, usageInput.Custody, (*usageOutput).Custody)
	assert.Equal(t, usageInput.Exp.Unix(), (*usageOutput).Exp.Unix())
	assert.False(t, usageOutput.IsExpired())
	t.Log(usageInput)
	t.Log(*usageOutput)
	tearDown(t)
}
