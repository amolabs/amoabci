package store

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"

	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

const testRoot = "store_test"

// utils

func makeAccAddr(seed string) crypto.Address {
	return p256.GenPrivKeyFromSecret([]byte(seed)).PubKey().Address()
}

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

// setup and teardown

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

// tests

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
	// setup
	s := NewStore(db.NewMemDB(), db.NewMemDB())

	stake1 := makeStake("val1", 100)
	stake2 := makeStake("val2", 100)
	s.SetStake(makeAccAddr("holder1"), stake1)
	s.SetStake(makeAccAddr("holder2"), stake2)

	test := s.GetStake(makeAccAddr("nobody"))
	assert.Nil(t, test)
	test = s.GetStakeByValidator(makeValAddr("none"))
	assert.Nil(t, test)

	test = s.GetStake(makeAccAddr("holder1"))
	assert.NotNil(t, test)
	assert.Equal(t, stake1, test)

	test = s.GetStakeByValidator(makeValAddr("val1"))
	assert.NotNil(t, test)
	assert.Equal(t, stake1, test)

	test = s.GetStake(makeAccAddr("holder2"))
	assert.NotNil(t, test)
	assert.Equal(t, stake2, test)

	test = s.GetStakeByValidator(makeValAddr("val2"))
	assert.NotNil(t, test)
	assert.Equal(t, stake2, test)
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
		Delegatee: staker,
		Amount:    *new(types.Currency).Set(101),
	}
	delegate2 := &types.Delegate{
		Delegatee: staker,
		Amount:    *new(types.Currency).Set(102),
	}

	// staker must have his own stake in order to be a delegator.
	assert.Error(t, s.SetDelegate(holder1, delegate1))

	s.SetStake(staker, &stake)
	s.SetDelegate(holder1, delegate1)
	s.SetDelegate(holder2, delegate2)

	assert.Equal(t, delegate1, s.GetDelegate(holder1))
	assert.Equal(t, delegate2, s.GetDelegate(holder2))

	// test delegator search index
	ds := s.GetDelegatesByDelegatee(staker)
	assert.Equal(t, 2, len(ds))
	assert.Equal(t, delegate1, ds[0].Delegate)
	assert.Equal(t, delegate2, ds[1].Delegate)

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

func newStake(amount string) (crypto.Address, *types.Stake) {
	priv := ed25519.GenPrivKey()
	validator, _ := priv.PubKey().(ed25519.PubKeyEd25519)
	holder := p256.GenPrivKey().PubKey().Address()
	coins, _ := new(types.Currency).SetString(amount, 10)
	stake := types.Stake{
		Amount:    *coins,
		Validator: validator,
	}
	return holder, &stake
}

func TestVotingPowerCalc(t *testing.T) {
	s := NewStore(db.NewMemDB(), db.NewMemDB())

	vals := s.GetValidators(100)
	assert.Equal(t, 0, len(vals))

	s.SetStake(newStake("1000000000000000000"))
	s.SetStake(newStake("10000000000000000"))
	s.SetStake(newStake("100000000000000000"))

	vals = s.GetValidators(1)
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, int64(500000000000000000), vals[0].Power)

	vals = s.GetValidators(100)
	assert.Equal(t, 3, len(vals))
	assert.Equal(t, int64(500000000000000000), vals[0].Power)
	assert.Equal(t, int64(50000000000000000), vals[1].Power)
	assert.Equal(t, int64(5000000000000000), vals[2].Power)

	// test voting power adjustment
	s.Purge()
	s.SetStake(newStake("1152921504606846975")) // 0xfffffffffffffff
	vals = s.GetValidators(100)
	assert.Equal(t, int64(0x7ffffffffffffff), vals[0].Power)

	s.SetStake(newStake("1"))
	vals = s.GetValidators(100)
	// The second staker's power shall be adjusted to be zero,
	// so it shall not be returned as valid validator.
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, int64(0x7ffffffffffffff), vals[0].Power)

	s.SetStake(newStake("47389214732891473289147321"))
	s.SetStake(newStake("98327483195748293743892147"))
	s.SetStake(newStake("64738214738918483219483177"))
	s.SetStake(newStake("10239481297483914839120049"))
	s.SetStake(newStake("10239481297483914839120049"))

	var sum int64
	vals = s.GetValidators(100)
	for _, val := range vals {
		sum += val.Power
	}
	assert.True(t, sum <= MaxTotalVotingPower)
	assert.Equal(t, vals[3].Power, vals[4].Power)

	//
	s.Purge()
	s.SetStake(newStake("10000000000000000000"))
	s.SetStake(newStake("10000000000000000000"))
	s.SetStake(newStake("10000000000000000000"))
	sum = 0
	vals = s.GetValidators(100)
	for _, val := range vals {
		sum += val.Power
	}
	assert.True(t, sum <= MaxTotalVotingPower)

	//
	s.Purge()
	s.SetStake(newStake("1000000000000000000"))
	s.SetStake(newStake("1000000000000000000"))
	s.SetStake(newStake("1000000000000000000"))
	sum = 0
	vals = s.GetValidators(100)
	for _, val := range vals {
		sum += val.Power
	}
	assert.True(t, sum <= MaxTotalVotingPower)
}
