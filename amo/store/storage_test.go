package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/types"
)

func TestStorageSetGet(t *testing.T) {
	s, err := NewStore(nil, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	mysto := &types.Storage{
		Owner:           makeAccAddr("provider"),
		Url:             "http://need_to_check_url_format",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(1),
		Active:          true,
	}
	assert.NotNil(t, mysto)

	// save and load
	assert.NoError(t, s.SetStorage(123, mysto))

	sto := s.GetStorage(123, true)
	assert.Nil(t, sto)

	sto = s.GetStorage(123, false)
	assert.NotNil(t, sto)
	assert.Equal(t, mysto, sto)

	s.Save()

	sto = s.GetStorage(123, true)
	assert.NotNil(t, sto)
	assert.Equal(t, mysto, sto)
}
