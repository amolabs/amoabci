package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestLazinessCounter(t *testing.T) {
	lc := NewLazinessCounter(4, 0.5)

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

	lv := lc.Investigate(lastCommitInfo)
	assert.Nil(t, lv)
	// height -> 1
	// candidates -> val1: 1

	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: false},
			{Validator: val3, SignedLastBlock: true},
		},
	}

	lv = lc.Investigate(lastCommitInfo)
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

	lv = lc.Investigate(lastCommitInfo)
	assert.Nil(t, lv)
	// height -> 3
	// candidates -> val1: 2
	//               val2: 2

	lastCommitInfo = abci.LastCommitInfo{
		Votes: []abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: true},
			{Validator: val3, SignedLastBlock: true},
		},
	}

	lv = lc.Investigate(lastCommitInfo)
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

	lv = lc.Investigate(lastCommitInfo)
	// height -> 1
	// candidates -> val3: 1

	assert.Equal(t, 2, len(lv))
	assert.Equal(t, val1.Address, lv[0].Bytes())
	assert.Equal(t, val2.Address, lv[1].Bytes())

	lv = lc.get()

	assert.Equal(t, 0, len(lv))
}

func makeTestAddress(seed []byte) crypto.Address {
	pubkey := ed25519.GenPrivKeyFromSecret(seed).PubKey().(ed25519.PubKeyEd25519)
	return pubkey.Address()
}
