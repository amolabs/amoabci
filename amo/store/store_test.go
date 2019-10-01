package store

import (
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

const testRoot = "store_test"

type dummyTx struct {
	Key   []byte
	Value []byte
}

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

func makeParcel(seed string, custody []byte) *types.ParcelValue {
	return &types.ParcelValue{
		Owner:   makeAccAddr(seed),
		Custody: custody,
		Info:    []byte("dummy parcel"),
	}
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

func TestPurge(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	sdb, err := NewDBProxy("state", testRoot)
	idb, err := NewDBProxy("index", testRoot)
	assert.NoError(t, err)

	s := NewStore(sdb, idb)
	assert.NotNil(t, s)

	err = s.Purge()
	assert.NoError(t, err)
}

func TestBalance(t *testing.T) {
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	testAddr := p256.GenPrivKey().PubKey().Address()
	balance := new(types.Currency).Set(1000)
	s.SetBalance(testAddr, balance)
	assert.Equal(t, balance, s.GetBalance(testAddr))
}

func TestParcel(t *testing.T) {
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
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
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
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
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
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
	var err error
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())

	holder1 := makeAccAddr("holder1")
	holder2 := makeAccAddr("holder2")
	val1 := makeValAddr("val1")
	val2 := makeValAddr("val2")
	stake1 := makeStake("val1", 100)
	stake2 := makeStake("val2", 100)
	stake10 := makeStake("val1", 0)
	stake20 := makeStake("val2", 0)

	stake := s.GetStake(makeAccAddr("nobody"))
	assert.Nil(t, stake)
	stake = s.GetStakeByValidator(makeValAddr("none"))
	assert.Nil(t, stake)

	err = s.SetUnlockedStake(holder1, stake1)
	assert.NoError(t, err)
	stake = s.GetStake(holder1)
	assert.NotNil(t, stake)
	assert.Equal(t, stake1, stake)

	err = s.SetUnlockedStake(holder2, stake1)
	assert.Error(t, err) // conflict val1

	err = s.SetUnlockedStake(holder2, stake2)
	assert.NoError(t, err)
	stake = s.GetStake(holder2)
	assert.NotNil(t, stake)
	assert.Equal(t, stake2, stake)

	stake = s.GetStakeByValidator(val1)
	assert.NotNil(t, stake)
	assert.Equal(t, stake1, stake)

	stake = s.GetStakeByValidator(val2)
	assert.NotNil(t, stake)
	assert.Equal(t, stake2, stake)

	err = s.SetUnlockedStake(holder1, stake10)
	assert.NoError(t, err)
	stake = s.GetStake(holder1)
	assert.Nil(t, stake)

	err = s.SetUnlockedStake(holder2, stake20)
	assert.Error(t, err) // LastValidator error
}

func TestLockedStake(t *testing.T) {
	// setup
	var err error
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())

	holder1 := makeAccAddr("holder1")
	holder2 := makeAccAddr("holder2")
	val1 := makeValAddr("val1")
	//val2 := makeValAddr("val2")
	stake11 := makeStake("val1", 100)
	stake12 := makeStake("val1", 200)
	stake13 := makeStake("val1", 300)
	stake10 := makeStake("val1", 0)
	stake2 := makeStake("val2", 100)

	// basic interface test

	stake := s.GetStake(holder1)
	assert.Nil(t, stake)
	stake = s.GetStakeByValidator(val1)
	assert.Nil(t, stake)

	//// test static case

	err = s.SetLockedStake(holder1, stake11, 1)
	assert.NoError(t, err)

	// height does not matter here
	err = s.SetLockedStake(holder2, stake11, 1)
	// conflict: holder mismatch
	assert.Equal(t, code.TxErrPermissionDenied, err)

	// height does not matter here
	err = s.SetLockedStake(holder1, stake2, 1)
	// conflict: validator mismatch
	assert.Equal(t, code.TxErrBadValidator, err)

	// height DOES matter here
	err = s.SetLockedStake(holder1, stake12, 1)
	// conflict: height already taken
	assert.Equal(t, code.TxErrHeightTaken, err)

	stake = s.GetStake(holder1)
	assert.NotNil(t, stake)
	assert.Equal(t, stake11, stake)

	stake = s.GetStakeByValidator(val1)
	assert.NotNil(t, stake)
	assert.Equal(t, stake11, stake)

	err = s.SetLockedStake(holder1, stake12, 10)
	assert.NoError(t, err)

	stake = s.GetStake(holder1)
	assert.NotNil(t, stake)
	assert.Equal(t, stake13, stake)

	stake = s.GetStakeByValidator(val1)
	assert.NotNil(t, stake)
	assert.Equal(t, stake13, stake)

	//// test unlocking

	// stakes locked at height 1 will be unlocked
	s.LoosenLockedStakes()

	stake = s.GetUnlockedStake(holder1)
	assert.NotNil(t, stake)
	assert.Equal(t, stake11, stake)

	stake = s.GetStake(holder1)
	assert.NotNil(t, stake)
	assert.Equal(t, stake13, stake)

	// delete unlocked stake
	err = s.SetUnlockedStake(holder1, stake10)
	assert.NoError(t, err)

	stake = s.GetStake(holder1)
	assert.NotNil(t, stake)
	assert.Equal(t, stake12, stake)
}

func TestDelegate(t *testing.T) {
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	// staker will be the delegatee of holders(delegators)
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

	s.SetUnlockedStake(staker, &stake)
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
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())

	vals := s.GetValidators(100)
	assert.Equal(t, 0, len(vals))

	s.SetUnlockedStake(newStake("1000000000000000000"))
	s.SetUnlockedStake(newStake("10000000000000000"))
	s.SetUnlockedStake(newStake("100000000000000000"))

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
	s.SetUnlockedStake(newStake("1152921504606846975")) // 0xfffffffffffffff
	vals = s.GetValidators(100)
	assert.Equal(t, int64(0x7ffffffffffffff), vals[0].Power)

	s.SetUnlockedStake(newStake("1"))
	vals = s.GetValidators(100)
	// The second staker's power shall be adjusted to be zero,
	// so it shall not be returned as valid validator.
	assert.Equal(t, 1, len(vals))
	assert.Equal(t, int64(0x7ffffffffffffff), vals[0].Power)

	s.SetUnlockedStake(newStake("47389214732891473289147321"))
	s.SetUnlockedStake(newStake("98327483195748293743892147"))
	s.SetUnlockedStake(newStake("64738214738918483219483177"))
	s.SetUnlockedStake(newStake("10239481297483914839120049"))
	s.SetUnlockedStake(newStake("10239481297483914839120049"))

	var sum int64
	vals = s.GetValidators(100)
	for _, val := range vals {
		sum += val.Power
	}
	assert.True(t, sum <= MaxTotalVotingPower)
	assert.Equal(t, vals[3].Power, vals[4].Power)

	//
	s.Purge()
	s.SetUnlockedStake(newStake("10000000000000000000"))
	s.SetUnlockedStake(newStake("10000000000000000000"))
	s.SetUnlockedStake(newStake("10000000000000000000"))
	sum = 0
	vals = s.GetValidators(100)
	for _, val := range vals {
		sum += val.Power
	}
	assert.True(t, sum <= MaxTotalVotingPower)

	//
	s.Purge()
	s.SetUnlockedStake(newStake("1000000000000000000"))
	s.SetUnlockedStake(newStake("1000000000000000000"))
	s.SetUnlockedStake(newStake("1000000000000000000"))
	sum = 0
	vals = s.GetValidators(100)
	for _, val := range vals {
		sum += val.Power
	}
	assert.True(t, sum <= MaxTotalVotingPower)
}

func TestMerkleTree(t *testing.T) {
	// suppose merkleTree is already defined in Store structure
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())

	// make transactions to put into merkleTree
	hash := "8019daa70792ac5db4f4418add9d7504c8c4d63b06f8bbea02cc4bfbbd2f77fe"
	expectedHash, err := hex.DecodeString(hash)
	assert.NoError(t, err)

	ok := s.merkleTree.IsEmpty()
	assert.True(t, ok)

	t1acc := makeAccAddr("t1")
	t1bal := new(types.Currency).Set(100)
	t2acc := makeAccAddr("t2")
	t2stake := makeStake("t2val", 100)
	parcel := []byte{0xC, 0xC, 0xC, 0xC}
	parcelVal := makeParcel("t3", []byte{0xA, 0xA, 0xA, 0xA})

	s.SetBalance(t1acc, t1bal)
	s.SetUnlockedStake(t2acc, t2stake)
	s.SetParcel(parcel, parcelVal)

	resultHash, version, err := s.Save()
	assert.NoError(t, err)

	imt, err := s.merkleTree.GetImmutable(version)
	assert.NoError(t, err)

	// check if nodes are put into the merkle tree
	assert.Equal(t, int64(3), imt.Size())

	assert.True(t, imt.Has(getBalanceKey(t1acc)))
	assert.True(t, imt.Has(makeStakeKey(t2acc)))
	assert.True(t, imt.Has(getParcelKey(parcel)))

	// compare expected root hash to generated one
	assert.Equal(t, expectedHash, resultHash)
}

func TestMutableTree(t *testing.T) {
	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())

	key := []byte("alice")
	value := []byte("1")

	s.set(key, value)
	assert.True(t, s.merkleTree.Has(key))
	assert.NotEqual(t, []byte("2"), s.get(key))
	assert.Equal(t, value, s.get(key))

	s.remove(key)
	assert.False(t, s.merkleTree.Has(key))
	assert.Nil(t, s.get(key))

	s.set(key, value)
	assert.True(t, s.merkleTree.Has(key))

	workingHash := s.Root()

	savedHash, version, err := s.Save()
	assert.NoError(t, err)

	assert.Equal(t, int64(1), version)
	assert.Equal(t, workingHash, savedHash)
}
