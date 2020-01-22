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
	assert.NoError(t, s.SetUDC([]byte("mycoin"), mycoin))

	udc := s.GetUDC([]byte("mycoin"), true)
	assert.Nil(t, udc)

	udc = s.GetUDC([]byte("mycoin"), false)
	assert.NotNil(t, udc)
	assert.Equal(t, mycoin, udc)

	s.Save()

	udc = s.GetUDC([]byte("mycoin"), true)
	assert.NotNil(t, udc)
	assert.Equal(t, mycoin, udc)
}

func TestUDCBalance(t *testing.T) {
	s := NewStore(
		tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NotNil(t, s)

	udc := []byte("mycoin")
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

func TestUDCLock(t *testing.T) {
	s := NewStore(
		tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NotNil(t, s)

	udcid := []byte("mycoin")
	holder := makeAccAddr("holder")
	amo10 := new(types.Currency).SetAMO(10)

	assert.NoError(t, s.SetUDCLock(udcid, holder, amo10))
	assert.Equal(t, amo10, s.GetUDCLock(udcid, holder, false))
	k := append([]byte("udclock:mycoin:"), holder.Bytes()...)
	assert.NotNil(t, s.get(k, false))
}
