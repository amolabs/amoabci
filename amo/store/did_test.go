package store

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/types"
)

func TestDIDGetSet(t *testing.T) {
	s, err := NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)

	var jsonDoc = []byte(`{"jsonkey":"jsonvalue"}`)

	entry := s.GetDIDEntry("myid", false)
	assert.Nil(t, entry)
	entry = &types.DIDEntry{Owner: makeAccAddr("me"), Document: jsonDoc}
	err = s.SetDIDEntry("myid", entry)
	assert.NoError(t, err)
	_entry := s.GetDIDEntry("myid", false)
	assert.NotNil(t, _entry)
	assert.Equal(t, entry, _entry)
	assert.Equal(t, makeAccAddr("me"), _entry.Owner)
	assert.True(t, bytes.Equal(jsonDoc, _entry.Document))
}
