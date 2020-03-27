package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"
)

func TestTxIndexer(t *testing.T) {
	s, err := NewStore(nil, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)

	// bunch of txs
	h1 := [][]byte{[]byte("tx11"), []byte("tx12"), []byte("tx13")}
	h2 := [][]byte{[]byte("tx21"), []byte("tx22"), []byte("tx23")}

	txs := s.TxIndexerGetHash(1)
	assert.Nil(t, txs)

	txs = s.TxIndexerGetHash(2)
	assert.Nil(t, txs)

	// Add
	s.AddTxIndexer(1, h1)
	s.AddTxIndexer(2, h2)

	// GetHeight
	txs = s.TxIndexerGetHash(1)
	assert.ElementsMatch(t, h1, txs)
	txs = s.TxIndexerGetHash(2)
	assert.ElementsMatch(t, h2, txs)

	// GetHash
	height := s.TxIndexerGetHeight(h1[0])
	assert.Equal(t, int64(1), height)

	height = s.TxIndexerGetHeight(h1[1])
	assert.Equal(t, int64(1), height)

	height = s.TxIndexerGetHeight(h1[2])
	assert.Equal(t, int64(1), height)

	height = s.TxIndexerGetHeight(h2[0])
	assert.Equal(t, int64(2), height)

	height = s.TxIndexerGetHeight(h2[1])
	assert.Equal(t, int64(2), height)

	height = s.TxIndexerGetHeight(h2[2])
	assert.Equal(t, int64(2), height)

	// Delete
	s.TxIndexerDelete(1)

	txs = s.TxIndexerGetHash(1)
	assert.Nil(t, txs)

	height = s.TxIndexerGetHeight(h1[0])
	assert.Equal(t, int64(0), height)
	height = s.TxIndexerGetHeight(h1[1])
	assert.Equal(t, int64(0), height)
	height = s.TxIndexerGetHeight(h1[2])
	assert.Equal(t, int64(0), height)

	// Purge
	s.TxIndexerPurge()

	txs = s.TxIndexerGetHash(1)
	assert.Nil(t, txs)

	txs = s.TxIndexerGetHash(2)
	assert.Nil(t, txs)
}
