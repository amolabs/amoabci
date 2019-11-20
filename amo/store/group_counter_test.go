package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	tmdb "github.com/tendermint/tm-db"
)

func TestGroupCounter(t *testing.T) {
	val1 := AddrSliceToArray(makeValAddr("val1"))
	val2 := AddrSliceToArray(makeValAddr("val2"))
	val3 := AddrSliceToArray(makeValAddr("val3"))

	lazyValidators := LazyValidators{
		val1: 1,
		val2: 2,
		val3: 3,
	}

	s := NewStore(tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())

	s.SetGroupCounter(lazyValidators)
	output := s.GetGroupCounter()

	assert.Equal(t, lazyValidators, output)

	s.PurgeGroupCounter()
	output = s.GetGroupCounter()

	assert.Equal(t, 0, len(output))
}

func AddrSliceToArray(sAddr crypto.Address) Address {
	var aAddr Address
	copy(aAddr[:], sAddr)

	return aAddr
}
