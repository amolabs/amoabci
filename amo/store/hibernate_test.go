package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/types"
)

func makeHibernate(start, end int64) *types.Hibernate {
	hib := types.Hibernate{
		Start: start,
		End: end,
	}
	return &hib
}

func TestGetSetHibernate(t *testing.T) {
	s, err := NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	hib := s.GetHibernate(makeValAddr("val1"), false)
	assert.Nil(t, hib)

	hib = makeHibernate(1, 10)
	err = s.SetHibernate(makeValAddr("val1"), hib)
	assert.NoError(t, err)

	hib2 := s.GetHibernate(makeValAddr("val1"), false)
	assert.NotNil(t, hib2)
	assert.Equal(t, hib, hib2)

	s.DeleteHibernate(makeValAddr("val1"))
	hib2 = s.GetHibernate(makeValAddr("val1"), false)
	assert.Nil(t, hib2)
}

func TestValidatorSelection(t *testing.T) {
	s, err := NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	s1 := makeStake("val1", 1000000000000000000)
	s2 := makeStake("val2", 10000000000000000)
	s3 := makeStake("val3", 100000000000000000)
	a1 := makeAccAddr("val1")
	a2 := makeAccAddr("val2")
	a3 := makeAccAddr("val3")

	s.SetUnlockedStake(a1, s1)
	s.SetUnlockedStake(a2, s2)
	s.SetUnlockedStake(a3, s3)
	vals := s.GetValidators(10, false)
	assert.Equal(t, 3, len(vals))

	hib := makeHibernate(10, 100)
	s.SetHibernate(s2.Validator.Address(), hib)
	vals = s.GetValidators(10, false)
	assert.Equal(t, 2, len(vals))

	s.DeleteHibernate(s2.Validator.Address())
	vals = s.GetValidators(10, false)
	assert.Equal(t, 3, len(vals))
}
