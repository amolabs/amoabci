package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/types"
)

func TestUDCSetGet(t *testing.T) {
	s := NewStore(
		tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NotNil(t, s)

	mycoin := &types.UDC{
		makeAccAddr("issuer"),
		"mycoin for test",
		[]crypto.Address{
			makeAccAddr("op1"),
			makeAccAddr("op2"),
		},
		*new(types.Currency).SetAMO(100),
	}
	assert.NotNil(t, mycoin)

	// save and load
	assert.NoError(t, s.SetUDC(123, mycoin))

	udc := s.GetUDC(123, true)
	assert.Nil(t, udc)

	udc = s.GetUDC(123, false)
	assert.NotNil(t, udc)
	assert.Equal(t, mycoin, udc)

	s.Save()

	udc = s.GetUDC(123, true)
	assert.NotNil(t, udc)
	assert.Equal(t, mycoin, udc)
}

func TestUDCBalance(t *testing.T) {
	s := NewStore(
		tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NotNil(t, s)

	udc := uint32(123)
	tester := makeAccAddr("tester")
	amo10 := new(types.Currency).SetAMO(10)
	amo0 := new(types.Currency)

	err := s.SetUDCBalance(udc, tester, amo10)
	assert.NoError(t, err)
	bal := s.GetUDCBalance(udc, tester, false)
	assert.NotNil(t, bal)
	assert.Equal(t, amo10, bal)

	err = s.SetUDCBalance(udc, tester, amo0)
	assert.NoError(t, err)
	bal = s.GetUDCBalance(udc, tester, false)
	assert.Equal(t, amo0, bal)
}
