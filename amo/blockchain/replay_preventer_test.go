package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/store"
)

func TestReplayPreventer(t *testing.T) {
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())

	rp := NewReplayPreventer(s, 0, 3)

	// bunch of txs
	h1 := [][]byte{[]byte("tx11"), []byte("tx12"), []byte("tx13")}
	h2 := [][]byte{[]byte("tx21"), []byte("tx22"), []byte("tx23")}
	h3 := [][]byte{[]byte("tx31"), []byte("tx32"), []byte("tx33")}
	h4 := [][]byte{[]byte("tx41"), []byte("tx42"), []byte("tx43")}

	// r: 3, f: 1, t: 1
	rp.Update(1, 3)

	_, err := rp.Check(h1[0], 1, 1)
	assert.NoError(t, err)
	_, err = rp.Check(h1[1], 1, 1)
	assert.NoError(t, err, 1, 1)
	_, err = rp.Check(h1[2], 1, 1)
	assert.NoError(t, err, 1, 1)

	err = rp.Append(h1[0], 1, 1)
	assert.NoError(t, err, 1, 1)
	err = rp.Append(h1[1], 1, 1)
	assert.NoError(t, err, 1, 1)
	err = rp.Append(h1[2], 1, 1)
	assert.NoError(t, err, 1, 1)

	_, err = rp.Check(h1[0], 1, 1)
	assert.Error(t, err, 1, 1)
	_, err = rp.Check(h1[1], 1, 1)
	assert.Error(t, err, 1, 1)
	_, err = rp.Check(h1[2], 1, 1)
	assert.Error(t, err, 1, 1)

	// before:
	//   txBucket: ["tx11", "tx12", "tx13"]
	//   txIndexer: []
	rp.Index(1)
	// after:
	//   txBucket: []
	//   txIndexer: ["tx11", "tx12", "tx13"]

	// r: 3, f: 1, t: 2
	rp.Update(2, 3)

	_, err = rp.Check(h1[0], 2, 2)
	assert.Error(t, err)
	_, err = rp.Check(h1[1], 2, 2)
	assert.Error(t, err)
	_, err = rp.Check(h1[2], 2, 2)
	assert.Error(t, err)
	_, err = rp.Check(h2[0], 2, 2)
	assert.NoError(t, err)
	_, err = rp.Check(h2[1], 2, 2)
	assert.NoError(t, err)
	_, err = rp.Check(h2[2], 2, 2)
	assert.NoError(t, err)

	err = rp.Append(h1[0], 2, 2)
	assert.Error(t, err)
	err = rp.Append(h1[1], 2, 2)
	assert.Error(t, err)
	err = rp.Append(h1[2], 2, 2)
	assert.Error(t, err)
	err = rp.Append(h2[0], 2, 2)
	assert.NoError(t, err)
	err = rp.Append(h2[1], 2, 2)
	assert.NoError(t, err)
	err = rp.Append(h2[2], 2, 2)
	assert.NoError(t, err)

	_, err = rp.Check(h1[0], 2, 2)
	assert.Error(t, err)
	_, err = rp.Check(h1[1], 2, 2)
	assert.Error(t, err)
	_, err = rp.Check(h1[2], 2, 2)
	assert.Error(t, err)
	_, err = rp.Check(h2[0], 2, 2)
	assert.Error(t, err)
	_, err = rp.Check(h2[1], 2, 2)
	assert.Error(t, err)
	_, err = rp.Check(h2[2], 2, 2)
	assert.Error(t, err)

	// before:
	//   txBucket: ["tx21", "tx22", "tx23"]
	//   txIndexer: ["tx11", "tx12", "tx13"]
	rp.Index(2)
	// after:
	//   txBucket: []
	//   txIndexer: ["tx11", "tx12", "tx13", "tx21", "tx22", "tx23"]

	// r: 3, f: 1, t: 3
	rp.Update(3, 3)

	_, err = rp.Check(h1[0], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h1[1], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h1[2], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h2[0], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h2[1], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h2[2], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h3[0], 3, 3)
	assert.NoError(t, err)
	_, err = rp.Check(h3[1], 3, 3)
	assert.NoError(t, err)
	_, err = rp.Check(h3[2], 3, 3)
	assert.NoError(t, err)

	err = rp.Append(h1[0], 3, 3)
	assert.Error(t, err)
	err = rp.Append(h1[1], 3, 3)
	assert.Error(t, err)
	err = rp.Append(h1[2], 3, 3)
	assert.Error(t, err)
	err = rp.Append(h2[0], 3, 3)
	assert.Error(t, err)
	err = rp.Append(h2[1], 3, 3)
	assert.Error(t, err)
	err = rp.Append(h2[2], 3, 3)
	assert.Error(t, err)
	err = rp.Append(h3[0], 3, 3)
	assert.NoError(t, err)
	err = rp.Append(h3[1], 3, 3)
	assert.NoError(t, err)
	err = rp.Append(h3[2], 3, 3)
	assert.NoError(t, err)

	_, err = rp.Check(h1[0], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h1[1], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h1[2], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h2[0], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h2[1], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h2[2], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h3[0], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h3[1], 3, 3)
	assert.Error(t, err)
	_, err = rp.Check(h3[2], 3, 3)
	assert.Error(t, err)

	// before:
	//   txBucket: ["tx31", "tx32", "tx33"]
	//   txIndexer: ["tx11", "tx12", "tx13", "tx21", "tx22", "tx23"]
	rp.Index(3)
	// txs indexed at height 1 are removed
	// after:
	//   txBucket: []
	//   txIndexer: ["tx21", "tx22", "tx23", "tx31", "tx32", "tx33"]

	// r: 3, f: 2, t: 4
	rp.Update(4, 3)

	_, err = rp.Check(h1[0], 4, 4)
	assert.NoError(t, err)
	_, err = rp.Check(h1[1], 4, 4)
	assert.NoError(t, err)
	_, err = rp.Check(h1[2], 4, 4)
	assert.NoError(t, err)
	_, err = rp.Check(h2[0], 4, 4)
	assert.Error(t, err)
	_, err = rp.Check(h2[1], 4, 4)
	assert.Error(t, err)
	_, err = rp.Check(h2[2], 4, 4)
	assert.Error(t, err)
	_, err = rp.Check(h3[0], 4, 4)
	assert.Error(t, err)
	_, err = rp.Check(h3[1], 4, 4)
	assert.Error(t, err)
	_, err = rp.Check(h3[2], 4, 4)
	assert.Error(t, err)

	txs := rp.store.TxIndexerGetHash(int64(1))
	assert.Equal(t, 0, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(2))
	assert.Equal(t, 3, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(3))
	assert.Equal(t, 3, len(txs))

	_, err = rp.Check(h4[0], 4, 4)
	assert.NoError(t, err)
	_, err = rp.Check(h4[1], 4, 4)
	assert.NoError(t, err)
	_, err = rp.Check(h4[2], 4, 4)
	assert.NoError(t, err)

	err = rp.Append(h4[0], 4, 4)
	assert.NoError(t, err)
	err = rp.Append(h4[1], 4, 4)
	assert.NoError(t, err)
	err = rp.Append(h4[2], 4, 4)
	assert.NoError(t, err)

	_, err = rp.Check(h4[0], 4, 4)
	assert.Error(t, err)
	_, err = rp.Check(h4[1], 4, 4)
	assert.Error(t, err)
	_, err = rp.Check(h4[2], 4, 4)
	assert.Error(t, err)

	// before:
	//   txBucket: ["tx41", "tx42", "tx43"]
	//   txIndexer: ["tx21", "tx22", "tx23", "tx31", "tx32", "tx33"]
	rp.Index(4)
	// txs indexed at height 2 are removed
	// after:
	//   txBucket: []
	//   txIndexer: ["tx31", "tx32", "tx33", "tx41", "tx42", "tx43"]

	txs = rp.store.TxIndexerGetHash(int64(1))
	assert.Equal(t, 0, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(2))
	assert.Equal(t, 0, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(3))
	assert.Equal(t, 3, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(4))
	assert.Equal(t, 3, len(txs))

	// before:
	//   txBucket: []
	//   txIndexer: ["tx31", "tx32", "tx33", "tx41", "tx42", "tx43"]
	// r: 3 -> 2
	rp.Update(5, 2)
	// r: 2, f: 4, t: 5
	// orphan tx indexed at height 3 are removed
	// after:
	//   txBucket: []
	//   txIndexer: ["tx41", "tx42", "tx43"]

	txs = rp.store.TxIndexerGetHash(int64(1))
	assert.Equal(t, 0, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(2))
	assert.Equal(t, 0, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(3))
	assert.Equal(t, 0, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(4))
	assert.Equal(t, 3, len(txs))
}
