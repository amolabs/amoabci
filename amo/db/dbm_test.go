package db

import (
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/tendermint-amo/crypto/p256"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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

func TestStore(t *testing.T) {
	setUp(t)
	s := NewStore(testRoot)
	testAddr := p256.GenPrivKey().PubKey().Address()
	balance := types.Currency(100)
	s.SetBalance(testAddr, balance)
	assert.Equal(t, balance, s.GetBalance(testAddr))
	tearDown(t)
}