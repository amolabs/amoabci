package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/store"
)

func TestReplayPreventer(t *testing.T) {
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())

	rp := NewReplayPreventer(s, 3, 0)

	// bunch of txs
	h1 := [][]byte{[]byte("tx11"), []byte("tx12"), []byte("tx13")}
	h2 := [][]byte{[]byte("tx21"), []byte("tx22"), []byte("tx23")}
	h3 := [][]byte{[]byte("tx31"), []byte("tx32"), []byte("tx33")}

	// r: 3, f: 1, t: 1
	rp.Update()

	ok := rp.Check(h1[0])
	assert.True(t, ok)
	ok = rp.Check(h1[1])
	assert.True(t, ok)
	ok = rp.Check(h1[2])
	assert.True(t, ok)

	ok = rp.Append(h1[0])
	assert.True(t, ok)
	ok = rp.Append(h1[1])
	assert.True(t, ok)
	ok = rp.Append(h1[2])
	assert.True(t, ok)

	ok = rp.Check(h1[0])
	assert.False(t, ok)
	ok = rp.Check(h1[1])
	assert.False(t, ok)
	ok = rp.Check(h1[2])
	assert.False(t, ok)

	// before:
	//   txBucket: ["tx11", "tx12", "tx13"]
	//   txIndexer: []
	rp.Index()
	// after:
	//   txBucket: []
	//   txIndexer: ["tx11", "tx12", "tx13"]

	// r: 3, f: 1, t: 2
	rp.Update()

	ok = rp.Check(h1[0])
	assert.False(t, ok)
	ok = rp.Check(h1[1])
	assert.False(t, ok)
	ok = rp.Check(h1[2])
	assert.False(t, ok)
	ok = rp.Check(h2[0])
	assert.True(t, ok)
	ok = rp.Check(h2[1])
	assert.True(t, ok)
	ok = rp.Check(h2[2])
	assert.True(t, ok)

	ok = rp.Append(h1[0])
	assert.False(t, ok)
	ok = rp.Append(h1[1])
	assert.False(t, ok)
	ok = rp.Append(h1[2])
	assert.False(t, ok)
	ok = rp.Append(h2[0])
	assert.True(t, ok)
	ok = rp.Append(h2[1])
	assert.True(t, ok)
	ok = rp.Append(h2[2])
	assert.True(t, ok)

	ok = rp.Check(h1[0])
	assert.False(t, ok)
	ok = rp.Check(h1[1])
	assert.False(t, ok)
	ok = rp.Check(h1[2])
	assert.False(t, ok)
	ok = rp.Check(h2[0])
	assert.False(t, ok)
	ok = rp.Check(h2[1])
	assert.False(t, ok)
	ok = rp.Check(h2[2])
	assert.False(t, ok)

	// before:
	//   txBucket: ["tx21", "tx22", "tx23"]
	//   txIndexer: ["tx11", "tx12", "tx13"]
	rp.Index()
	// after:
	//   txBucket: []
	//   txIndexer: ["tx11", "tx12", "tx13", "tx21", "tx22", "tx23"]

	// r: 3, f: 1, t: 3
	rp.Update()

	ok = rp.Check(h1[0])
	assert.False(t, ok)
	ok = rp.Check(h1[1])
	assert.False(t, ok)
	ok = rp.Check(h1[2])
	assert.False(t, ok)
	ok = rp.Check(h2[0])
	assert.False(t, ok)
	ok = rp.Check(h2[1])
	assert.False(t, ok)
	ok = rp.Check(h2[2])
	assert.False(t, ok)
	ok = rp.Check(h3[0])
	assert.True(t, ok)
	ok = rp.Check(h3[1])
	assert.True(t, ok)
	ok = rp.Check(h3[2])
	assert.True(t, ok)

	ok = rp.Append(h1[0])
	assert.False(t, ok)
	ok = rp.Append(h1[1])
	assert.False(t, ok)
	ok = rp.Append(h1[2])
	assert.False(t, ok)
	ok = rp.Append(h2[0])
	assert.False(t, ok)
	ok = rp.Append(h2[1])
	assert.False(t, ok)
	ok = rp.Append(h2[2])
	assert.False(t, ok)
	ok = rp.Append(h3[0])
	assert.True(t, ok)
	ok = rp.Append(h3[1])
	assert.True(t, ok)
	ok = rp.Append(h3[2])
	assert.True(t, ok)

	ok = rp.Check(h1[0])
	assert.False(t, ok)
	ok = rp.Check(h1[1])
	assert.False(t, ok)
	ok = rp.Check(h1[2])
	assert.False(t, ok)
	ok = rp.Check(h2[0])
	assert.False(t, ok)
	ok = rp.Check(h2[1])
	assert.False(t, ok)
	ok = rp.Check(h2[2])
	assert.False(t, ok)
	ok = rp.Check(h3[0])
	assert.False(t, ok)
	ok = rp.Check(h3[1])
	assert.False(t, ok)
	ok = rp.Check(h3[2])
	assert.False(t, ok)

	// before:
	//   txBucket: ["tx31", "tx32", "tx33"]
	//   txIndexer: ["tx11", "tx12", "tx13", "tx21", "tx22", "tx23"]
	rp.Index()
	// after:
	//   txBucket: []
	//   txIndexer: ["tx21", "tx22", "tx23", "tx31", "tx32", "tx33"]

	// r: 3, f: 2, t: 4
	// txs indexed at height 1 are removed
	rp.Update()

	ok = rp.Check(h1[0])
	assert.True(t, ok)
	ok = rp.Check(h1[1])
	assert.True(t, ok)
	ok = rp.Check(h1[2])
	assert.True(t, ok)
	ok = rp.Check(h2[0])
	assert.False(t, ok)
	ok = rp.Check(h2[1])
	assert.False(t, ok)
	ok = rp.Check(h2[2])
	assert.False(t, ok)
	ok = rp.Check(h3[0])
	assert.False(t, ok)
	ok = rp.Check(h3[1])
	assert.False(t, ok)
	ok = rp.Check(h3[2])
	assert.False(t, ok)

	txs := rp.store.TxIndexerGetHash(int64(1))
	assert.Equal(t, 0, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(2))
	assert.Equal(t, 3, len(txs))
	txs = rp.store.TxIndexerGetHash(int64(3))
	assert.Equal(t, 3, len(txs))
}
