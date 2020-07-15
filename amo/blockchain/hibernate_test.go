package blockchain

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

func makeAccAddr(seed string) crypto.Address {
	return p256.GenPrivKeyFromSecret([]byte(seed)).PubKey().Address()
}

func makeValAddr(seed string) crypto.Address {
	priKey := ed25519.GenPrivKeyFromSecret([]byte(seed))
	pubKey := priKey.PubKey().(ed25519.PubKeyEd25519)
	return pubKey.Address()
}

func makeHibernate(start, end int64) *types.Hibernate {
	hib := types.Hibernate{
		Start: start,
		End:   end,
	}
	return &hib
}

func TestHibernate(t *testing.T) {
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	mr := MissRuns{
		store:              s,
		hibernateThreshold: 10,
		hibernatePeriod:    100,
		lazinessWindow:     10,
	}

	var start, length int64

	val1 := makeValAddr("val1")
	val2 := makeValAddr("val2")
	mvals := []crypto.Address{}
	mval10 := []crypto.Address{val1}
	mval12 := []crypto.Address{val1, val2}
	mval02 := []crypto.Address{val2}

	hib := s.GetHibernate(val1, false)
	assert.Nil(t, hib)

	_, _, err = mr.UpdateMissRuns(3, mvals)
	assert.NoError(t, err)
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(0), start)
	assert.Equal(t, int64(0), length)

	_, _, err = mr.UpdateMissRuns(4, mval10)
	assert.NoError(t, err)
	_, _, err = mr.UpdateMissRuns(5, mval12)
	assert.NoError(t, err)
	_, _, err = mr.UpdateMissRuns(6, mval12)
	assert.NoError(t, err)
	_, _, err = mr.UpdateMissRuns(7, mval02)
	assert.NoError(t, err)
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(4), start)
	assert.Equal(t, int64(3), length)
	start, length = mr.getLastMissRun(val2)
	assert.Equal(t, int64(5), start)
	assert.Equal(t, int64(0), length) // unfinished run has length zero

	// alternative history from height 6
	_, _, err = mr.UpdateMissRuns(6, mvals)
	assert.NoError(t, err)
	_, _, err = mr.UpdateMissRuns(7, mval02)
	assert.NoError(t, err)
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(4), start)
	assert.Equal(t, int64(2), length)
	start, length = mr.getLastMissRun(val2)
	assert.Equal(t, int64(7), start)
	assert.Equal(t, int64(0), length) // unfinished run has length zero

	// history gap does not break unfinished run
	_, _, err = mr.UpdateMissRuns(9, mval02)
	assert.NoError(t, err)
	start, length = mr.getLastMissRun(val2)
	assert.Equal(t, int64(7), start)
	assert.Equal(t, int64(0), length) // unfinished run has length zero

	// hibernate test
	// val came back before hibernation
	for h := int64(10); h < 19; h++ {
		_, _, err = mr.UpdateMissRuns(h, mval10)
		assert.NoError(t, err)
	}
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(10), start)
	assert.Equal(t, int64(0), length)
	_, _, err = mr.UpdateMissRuns(20, mvals) // empty missing validators
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(10), start)
	assert.Equal(t, int64(10), length)
	hib = s.GetHibernate(val1, false)
	assert.Nil(t, hib)

	// hibernate test
	// val goes to hibernate at height 19
	var evs []abci.Event
	for h := int64(10); h < 20; h++ {
		_, evs, err = mr.UpdateMissRuns(h, mval10)
		assert.NoError(t, err)
	}
	assert.Equal(t, "hibernate", evs[0].Type)
	assert.Equal(t, []byte("validator"), evs[0].Attributes[0].Key)
	assert.Equal(t, []byte("19"), evs[0].Attributes[1].Value)
	assert.Equal(t, []byte("119"), evs[0].Attributes[2].Value)
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(10), start)
	assert.Equal(t, int64(0), length)
	hib = s.GetHibernate(val1, false)
	assert.NotNil(t, hib)
	assert.Equal(t, int64(19), hib.Start)
	assert.Equal(t, int64(119), hib.End)
	// one more height to close the run:
	// Since val1 is set to hibernate at height 19, it will not be included in
	// the missed validators at height 20.
	_, _, err = mr.UpdateMissRuns(20, mvals)
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(10), start)
	assert.Equal(t, int64(10), length)
}

func TestMissStat(t *testing.T) {
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	mr := MissRuns{
		store:              s,
		hibernateThreshold: 10,
		hibernatePeriod:    100,
		lazinessWindow:     20, // dummy value
	}

	val1 := makeValAddr("val1")
	val2 := makeValAddr("val2")
	mval00 := []crypto.Address{}
	mval10 := []crypto.Address{val1}
	mval12 := []crypto.Address{val1, val2}
	mval02 := []crypto.Address{val2}

	mr.UpdateMissRuns(10, mval00)
	mr.UpdateMissRuns(11, mval10)
	mr.UpdateMissRuns(12, mval10)
	mr.UpdateMissRuns(13, mval10)
	mr.UpdateMissRuns(14, mval12)
	mr.UpdateMissRuns(15, mval12)
	mr.UpdateMissRuns(16, mval12)
	mr.UpdateMissRuns(17, mval02)
	mr.UpdateMissRuns(18, mval02)
	mr.UpdateMissRuns(19, mval12)
	mr.UpdateMissRuns(20, mval12)
	mr.UpdateMissRuns(21, mval02)
	mr.UpdateMissRuns(22, mval02)
	mr.UpdateMissRuns(23, mval02)
	mr.UpdateMissRuns(24, mval02)
	mr.UpdateMissRuns(25, mval12)
	mr.UpdateMissRuns(26, mval10)
	mr.UpdateMissRuns(27, mval10)
	mr.UpdateMissRuns(28, mval02)
	mr.UpdateMissRuns(29, mval12)
	mr.UpdateMissRuns(30, mval10)

	stat := mr.GetMissStat(13, 30)
	val, _ := hex.DecodeString(val1.String())
	assert.True(t, bytes.Equal(val1, val))
	assert.Equal(t, int64(11), stat[val1.String()])
	assert.Equal(t, int64(14), stat[val2.String()])
}

func TestCleanUpFinishedRuns(t *testing.T) {
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// test case for lazinessWindow > hibernateThreshold
	mr := MissRuns{
		store:              s,
		hibernateThreshold: 10,
		hibernatePeriod:    100,
		lazinessWindow:     12,
	}
	val1 := makeValAddr("val1")
	val2 := makeValAddr("val2")
	mval00 := []crypto.Address{}
	mval10 := []crypto.Address{val1}
	mval12 := []crypto.Address{val1, val2}
	// mval02 := []crypto.Address{val2}

	mr.UpdateMissRuns(10, mval00)

	stat := mr.GetMissStat(10, 10)
	assert.Empty(t, stat)

	// val1: hibernating
	// val2: non-hibernating
	mr.UpdateMissRuns(11, mval10)
	mr.UpdateMissRuns(12, mval12)
	mr.UpdateMissRuns(13, mval12)
	mr.UpdateMissRuns(14, mval12)
	mr.UpdateMissRuns(15, mval12)
	mr.UpdateMissRuns(16, mval12)
	mr.UpdateMissRuns(17, mval12)
	mr.UpdateMissRuns(18, mval12)
	mr.UpdateMissRuns(19, mval12)
	mr.UpdateMissRuns(20, mval12) // val1 hibernates

	stat = mr.GetMissStat(10, 20)
	assert.Equal(t, int64(10), stat[val1.String()])
	assert.Equal(t, int64(9), stat[val2.String()])

	mr.UpdateMissRuns(21, mval00)
	stat = mr.GetMissStat(10, 21)
	_, exist := stat[val1.String()]
	assert.True(t, exist)
	_, exist = stat[val2.String()]
	assert.True(t, exist)

	mr.UpdateMissRuns(22, mval00)
	stat = mr.GetMissStat(10, 22)
	_, exist = stat[val1.String()]
	assert.True(t, exist)
	_, exist = stat[val2.String()]
	assert.True(t, exist)

	mr.UpdateMissRuns(23, mval00) // val1's missRun gets removed
	stat = mr.GetMissStat(10, 23)
	_, exist = stat[val1.String()]
	assert.False(t, exist)
	_, exist = stat[val2.String()]
	assert.True(t, exist)

	mr.UpdateMissRuns(24, mval00) // val2's missRun gets removed
	stat = mr.GetMissStat(10, 24)
	_, exist = stat[val1.String()]
	assert.False(t, exist)
	_, exist = stat[val2.String()]
	assert.False(t, exist)
	assert.Empty(t, stat)

	// test case for lazinessWindow < hibernateThreshold
	mr = MissRuns{
		store:              s,
		hibernateThreshold: 12,
		hibernatePeriod:    100,
		lazinessWindow:     10,
	}

	mr.UpdateMissRuns(10, mval00)

	stat = mr.GetMissStat(10, 10)
	assert.Empty(t, stat)

	// val1: hibernating
	// val2: non-hibernating
	mr.UpdateMissRuns(11, mval10)
	mr.UpdateMissRuns(12, mval12)
	mr.UpdateMissRuns(13, mval12)
	mr.UpdateMissRuns(14, mval12)
	mr.UpdateMissRuns(15, mval12)
	mr.UpdateMissRuns(16, mval12)
	mr.UpdateMissRuns(17, mval12)
	mr.UpdateMissRuns(18, mval12)
	mr.UpdateMissRuns(19, mval12)
	mr.UpdateMissRuns(20, mval12)

	stat = mr.GetMissStat(10, 20)
	assert.Equal(t, int64(10), stat[val1.String()])
	assert.Equal(t, int64(9), stat[val2.String()])

	mr.UpdateMissRuns(21, mval12)

	stat = mr.GetMissStat(10, 21)
	_, exist = stat[val1.String()]
	assert.True(t, exist)
	_, exist = stat[val2.String()]
	assert.True(t, exist)

	mr.UpdateMissRuns(22, mval12) // val1 hibernates

	stat = mr.GetMissStat(10, 22)
	_, exist = stat[val1.String()]
	assert.True(t, exist)
	_, exist = stat[val2.String()]
	assert.True(t, exist)

	mr.UpdateMissRuns(23, mval00)

	stat = mr.GetMissStat(10, 23)
	_, exist = stat[val1.String()]
	assert.True(t, exist)
	_, exist = stat[val2.String()]
	assert.True(t, exist)

	mr.UpdateMissRuns(24, mval00) // val1, val2's missRuns get removed

	stat = mr.GetMissStat(10, 24)
	_, exist = stat[val1.String()]
	assert.False(t, exist)
	_, exist = stat[val2.String()]
	assert.False(t, exist)
	assert.Empty(t, stat)
}
