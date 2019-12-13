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
		[]byte("mycoin"),
		makeAccAddr("issuer"),
		[]crypto.Address{
			makeAccAddr("op1"),
			makeAccAddr("op2"),
		},
		"mycoin for test",
		*new(types.Currency).SetAMO(100),
	}
	assert.NotNil(t, mycoin)

	// save and load
	assert.NoError(t, s.SetUDC(mycoin.Id, mycoin))

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
