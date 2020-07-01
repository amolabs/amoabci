package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	mr := MissRuns{
		store:              s,
		runDB:              tmdb.NewMemDB(),
		hibernateThreshold: 10,
		hibernatePeriod:    100,
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

	_, err = mr.UpdateMissRuns(3, mvals)
	assert.NoError(t, err)
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(0), start)
	assert.Equal(t, int64(0), length)

	_, err = mr.UpdateMissRuns(4, mval10)
	assert.NoError(t, err)
	_, err = mr.UpdateMissRuns(5, mval12)
	assert.NoError(t, err)
	_, err = mr.UpdateMissRuns(6, mval12)
	assert.NoError(t, err)
	_, err = mr.UpdateMissRuns(7, mval02)
	assert.NoError(t, err)
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(4), start)
	assert.Equal(t, int64(3), length)
	start, length = mr.getLastMissRun(val2)
	assert.Equal(t, int64(5), start)
	assert.Equal(t, int64(0), length) // unfinished run has length zero

	// alternative history from height 6
	_, err = mr.UpdateMissRuns(6, mvals)
	assert.NoError(t, err)
	_, err = mr.UpdateMissRuns(7, mval02)
	assert.NoError(t, err)
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(4), start)
	assert.Equal(t, int64(2), length)
	start, length = mr.getLastMissRun(val2)
	assert.Equal(t, int64(7), start)
	assert.Equal(t, int64(0), length) // unfinished run has length zero

	// history gap does not break unfinished run
	_, err = mr.UpdateMissRuns(9, mval02)
	assert.NoError(t, err)
	start, length = mr.getLastMissRun(val2)
	assert.Equal(t, int64(7), start)
	assert.Equal(t, int64(0), length) // unfinished run has length zero

	// hibernate test
	// val came back before hibernation
	for h := int64(10); h < 19; h++ {
		_, err = mr.UpdateMissRuns(h, mval10)
		assert.NoError(t, err)
	}
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(10), start)
	assert.Equal(t, int64(0), length)
	_, err = mr.UpdateMissRuns(20, mvals) // empty missing validators
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(10), start)
	assert.Equal(t, int64(10), length)
	hib = s.GetHibernate(val1, false)
	assert.Nil(t, hib)

	// hibernate test
	// val goes to hibernate at height 19
	for h := int64(10); h < 20; h++ {
		_, err = mr.UpdateMissRuns(h, mval10)
		assert.NoError(t, err)
	}
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
	_, err = mr.UpdateMissRuns(20, mvals)
	start, length = mr.getLastMissRun(val1)
	assert.Equal(t, int64(10), start)
	assert.Equal(t, int64(10), length)
}
