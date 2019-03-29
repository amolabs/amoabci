package store

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"

	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
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
	s := NewStore(db.NewMemDB(), db.NewMemDB())
	testAddr := p256.GenPrivKey().PubKey().Address()
	balance := new(types.Currency).Set(1000)
	s.SetBalance(testAddr, balance)
	assert.Equal(t, balance, s.GetBalance(testAddr))
}

func TestParcel(t *testing.T) {
	s := NewStore(db.NewMemDB(), db.NewMemDB())
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
}

func TestRequest(t *testing.T) {
	s := NewStore(db.NewMemDB(), db.NewMemDB())
	testAddr := p256.GenPrivKey().PubKey().Address()
	parcelID := cmn.RandBytes(32)
	exp := time.Now().UTC()
	exp = exp.Add(100 * time.Minute)
	requestInput := types.RequestValue{
		Payment: *new(types.Currency).Set(100),
		Exp:     exp,
	}
	s.SetRequest(testAddr, parcelID, &requestInput)
	requestOutput := s.GetRequest(testAddr, parcelID)
	assert.Equal(t, requestInput.Payment, (*requestOutput).Payment)
	assert.Equal(t, requestInput.Exp.Unix(), (*requestOutput).Exp.Unix())
	assert.False(t, requestOutput.IsExpired())
	t.Log(requestInput)
	t.Log(*requestOutput)
}

func TestUsage(t *testing.T) {
	s := NewStore(db.NewMemDB(), db.NewMemDB())
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
}

func TestStake(t *testing.T) {
	s := NewStore(db.NewMemDB(), db.NewMemDB())
	holder := p256.GenPrivKeyFromSecret([]byte("holder")).PubKey().Address()
	valKey := ed25519.GenPrivKeyFromSecret([]byte("holder")).PubKey().(ed25519.PubKeyEd25519)
	validator := valKey.Address()
	stake := types.Stake{
		Amount:    *new(types.Currency).Set(100),
		Validator: valKey,
	}
	s.SetStake(holder, &stake)

	assert.NotNil(t, s.GetStake(holder))
	assert.Equal(t, stake, *s.GetStake(holder))
	assert.NotNil(t, s.GetStakeByValidator(validator))
	assert.Equal(t, stake, *s.GetStakeByValidator(validator))
}

func TestDelegate(t *testing.T) {
	s := NewStore(db.NewMemDB(), db.NewMemDB())
	// staker will be the delegator of holders
	staker := p256.GenPrivKeyFromSecret([]byte("staker")).PubKey().Address()
	valkey, _ := ed25519.GenPrivKeyFromSecret([]byte("val")).PubKey().(ed25519.PubKeyEd25519)
	holder1 := p256.GenPrivKeyFromSecret([]byte("holder1")).PubKey().Address()
	holder2 := p256.GenPrivKeyFromSecret([]byte("holder2")).PubKey().Address()
	stake := types.Stake{
		Amount:    *new(types.Currency).Set(100),
		Validator: valkey,
	}
	delegate1 := &types.Delegate{
		Holder:    holder1,
		Amount:    *new(types.Currency).Set(101),
		Delegator: staker,
	}
	delegate2 := &types.Delegate{
		Holder:    holder2,
		Amount:    *new(types.Currency).Set(102),
		Delegator: staker,
	}

	// staker must have his own stake in order to be a delegator.
	assert.Error(t, s.SetDelegate(holder1, delegate1))

	s.SetStake(staker, &stake)
	s.SetDelegate(holder1, delegate1)
	s.SetDelegate(holder2, delegate2)

	assert.Equal(t, delegate1, s.GetDelegate(holder1))
	assert.Equal(t, delegate2, s.GetDelegate(holder2))

	// test delegator search index
	ds := s.GetDelegatesByDelegator(staker)
	assert.Equal(t, 2, len(ds))
	assert.Equal(t, delegate1, ds[0])
	assert.Equal(t, delegate2, ds[1])

	es := *new(types.Currency)
	es.Add(&stake.Amount)
	es.Add(&delegate1.Amount)
	es.Add(&delegate2.Amount)
	assert.Equal(t, *new(types.Currency).Set(303), es)
	es = s.GetEffStake(staker).Amount
	assert.Equal(t, *new(types.Currency).Set(303), es)

	// test effective stake cache
	ts := s.GetTopStakes(10)
	assert.Equal(t, 1, len(ts))
	assert.Equal(t, s.GetEffStake(staker), ts[0])
}
