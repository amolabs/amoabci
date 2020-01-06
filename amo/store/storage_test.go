package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/types"
)

func TestStorageSetGet(t *testing.T) {
	s := NewStore(
		tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
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
	assert.NoError(t, s.SetStorage([]byte("mystorage"), mysto))

	sto := s.GetStorage([]byte("mystorage"), true)
	assert.Nil(t, sto)

	sto = s.GetStorage([]byte("mystorage"), false)
	assert.NotNil(t, sto)
	assert.Equal(t, mysto, sto)

	s.Save()

	sto = s.GetStorage([]byte("mystorage"), true)
	assert.NotNil(t, sto)
	assert.Equal(t, mysto, sto)
}
