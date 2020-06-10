package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/store"
)

func TestLazinessCounter(t *testing.T) {
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	lc := NewLazinessCounter(s, 0, 0, 4, 0.5)

	val1 := abci.Validator{Address: makeTestAddress([]byte("val1"))}
	val2 := abci.Validator{Address: makeTestAddress([]byte("val2"))}
	val3 := abci.Validator{Address: makeTestAddress([]byte("val3"))}

	// height -> 1
	// candidates -> val1: 1
	lastCommitInfo := abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: false},
			{Validator: val2, SignedLastBlock: true},
			{Validator: val3, SignedLastBlock: true},
		},
	}

	lv, due := lc.Investigate(1, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(4), due)

	// height -> 2
	// candidates -> val1: 1
	//               val2: 1
	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: false},
			{Validator: val3, SignedLastBlock: true},
		},
	}

	lv, due = lc.Investigate(2, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(4), due)

	// height -> 3
	// candidates -> val1: 2
	//               val2: 2
	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: false},
			{Validator: val2, SignedLastBlock: false},
			{Validator: val3, SignedLastBlock: true},
		},
	}

	lv, due = lc.Investigate(3, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(4), due)

	// imitate down of amod
	lc_new := NewLazinessCounter(s, 3, due, 4, 0.5)

	// height -> 4
	// candidates -> val1: 2
	//               val2: 2
	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: true},
			{Validator: val3, SignedLastBlock: true},
		},
	}

	lv, due = lc_new.Investigate(4, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(4), due)

	// height -> 5
	// candidates -> val3: 1
	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: true},
			{Validator: val3, SignedLastBlock: false},
		},
	}

	lv, due = lc_new.Investigate(5, lastCommitInfo)
	assert.Equal(t, 2, len(lv))
	assert.Equal(t, int64(8), due)

	lv = lc_new.get()

	assert.Equal(t, 0, len(lv))

	// pending_size: 6, pending_ratio: 0.8
	lc_new.Set(6, 0.8)

	lv, due = lc_new.Investigate(6, lastCommitInfo)
	// height -> 6
	// candidates -> val3: 2
	assert.Nil(t, lv)
	assert.Equal(t, int64(8), due)

	// height -> 7
	// candidates -> val3: 3
	lv, due = lc_new.Investigate(7, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(8), due)

	// height -> 8
	// candidates -> val3: 4
	lv, due = lc_new.Investigate(8, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(8), due)

	// size: 4, ratio: 0.5 -> size: 6, ratio: 0.8 (limit: 4.8)
	// height -> 9
	// candidates -> val2: 1
	//               val3: 1
	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: false},
			{Validator: val3, SignedLastBlock: false},
		},
	}

	lv, due = lc_new.Investigate(9, lastCommitInfo)
	assert.Equal(t, 1, len(lv))
	assert.Equal(t, int64(14), due)

	// height -> 10
	// candidates -> val2: 1
	//               val3: 2
	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: true},
			{Validator: val3, SignedLastBlock: false},
		},
	}

	lv, due = lc_new.Investigate(10, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(14), due)

	// height -> 11
	// candidates -> val2: 1
	//               val3: 3
	lv, due = lc_new.Investigate(11, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(14), due)

	// height -> 12
	// candidates -> val2: 1
	//               val3: 4
	lv, due = lc_new.Investigate(12, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(14), due)

	// height -> 13
	// candidates -> val2: 1
	//               val3: 5
	lv, due = lc_new.Investigate(13, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(14), due)

	// height -> 14
	// candidates -> val2: 1
	//               val3: 5
	lv, due = lc_new.Investigate(14, lastCommitInfo)
	assert.Nil(t, lv)
	assert.Equal(t, int64(14), due)

	// height -> 15
	// candidates -> val3: 1
	lv, due = lc_new.Investigate(15, lastCommitInfo)
	assert.Equal(t, 1, len(lv))
	assert.Equal(t, int64(20), due)
}

func makeTestAddress(seed []byte) crypto.Address {
	pubkey := ed25519.GenPrivKeyFromSecret(seed).PubKey().(ed25519.PubKeyEd25519)
	return pubkey.Address()
}
