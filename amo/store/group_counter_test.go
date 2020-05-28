package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	tmdb "github.com/tendermint/tm-db"
)

func TestGroupCounter(t *testing.T) {
	s, err := NewStore(nil, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	lazyValidators := LazyValidators{
		AddrSliceToArray(makeValAddr("val1")): 1,
		AddrSliceToArray(makeValAddr("val2")): 2,
		AddrSliceToArray(makeValAddr("val3")): 3,
	}

	s.GroupCounterPurge()
	output := s.GroupCounterGetLazyValidators()
	assert.Equal(t, 0, len(output))

	s.GroupCounterSet(lazyValidators)
	output = s.GroupCounterGetLazyValidators()
	assert.Equal(t, lazyValidators, output)

	s.GroupCounterPurge()
	output = s.GroupCounterGetLazyValidators()
	assert.Equal(t, 0, len(output))
}

func AddrSliceToArray(sAddr crypto.Address) Address {
	var aAddr Address
	copy(aAddr[:], sAddr)

	return aAddr
}
