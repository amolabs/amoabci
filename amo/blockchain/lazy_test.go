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
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	lc := NewLazinessCounter(s, 0, 0, 4, 0.5)

	val1 := abci.Validator{Address: makeTestAddress([]byte("val1"))}
	val2 := abci.Validator{Address: makeTestAddress([]byte("val2"))}
	val3 := abci.Validator{Address: makeTestAddress([]byte("val3"))}

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
	// height -> 1
	// candidates -> val1: 1

	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: false},
			{Validator: val3, SignedLastBlock: true},
		},
	}

	lv, due = lc.Investigate(2, lastCommitInfo)
	assert.Nil(t, lv)
	// height -> 2
	// candidates -> val1: 1
	//               val2: 1

	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: false},
			{Validator: val2, SignedLastBlock: false},
			{Validator: val3, SignedLastBlock: true},
		},
	}

	lv, due = lc.Investigate(3, lastCommitInfo)
	assert.Nil(t, lv)
	// height -> 3
	// candidates -> val1: 2
	//               val2: 2

	// imitate down of amod
	lc_new := NewLazinessCounter(s, 3, due, 4, 0.5)

	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: true},
			{Validator: val3, SignedLastBlock: true},
		},
	}

	lv, due = lc_new.Investigate(4, lastCommitInfo)
	assert.Nil(t, lv)
	// height -> 4
	// candidates -> val1: 2
	//               val2: 2

	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: true},
			{Validator: val3, SignedLastBlock: false},
		},
	}

	lv, due = lc_new.Investigate(5, lastCommitInfo)
	// height -> 5
	// candidates -> val3: 1

	assert.Equal(t, int64(8), due)
	assert.Equal(t, 2, len(lv))

	lv = lc_new.get()

	assert.Equal(t, 0, len(lv))
}

func makeTestAddress(seed []byte) crypto.Address {
	pubkey := ed25519.GenPrivKeyFromSecret(seed).PubKey().(ed25519.PubKeyEd25519)
	return pubkey.Address()
}
