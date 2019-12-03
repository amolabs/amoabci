package blockchain

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/amo/store"
)

type (
	Address        = store.Address
	LazyValidators = store.LazyValidators
)

type LazinessCounter struct {
	store *store.Store

	Candidates LazyValidators `json:"lazy_validators"` // stored on store
	Height     int64          `json:"height"`

	Due   int64   `json:"due"`   // from state
	Size  int64   `json:"size"`  // from config
	Ratio float64 `json:"ratio"` // from config
}

// LazinessCounter
// - height: 0, due: 0, size: 4, ratio: 0.8 -> due: 4, limit: 3.2
//
// 0   4   8
// |---|---|--->
//  1234
//  OXXX(3) => Gotcha!
//      5678
//      XOOX(2) => No prob!
//          ...

func NewLazinessCounter(store *store.Store, height, due, size int64, ratio float64) LazinessCounter {
	if due == 0 {
		due = height + size
	}

	lc := LazinessCounter{
		store:  store,
		Height: height,
		Due:    due,
		Size:   size,
		Ratio:  ratio,
	}

	lc.Candidates = lc.store.GroupCounterGetLazyValidators()

	return lc
}

func (lc *LazinessCounter) Investigate(height int64, commitInfo abci.LastCommitInfo) ([]crypto.Address, int64) {
	var lazyValidators []crypto.Address

	if lc.checkEnd() {
		lazyValidators = lc.get()
		lc.purge()
		lc.Due = lc.Height + lc.Size
	}

	votes := commitInfo.GetVotes()
	for _, vote := range votes {
		if !vote.GetSignedLastBlock() {
			validator := vote.GetValidator()
			lc.add(validator)
		}
	}

	// update height
	lc.Height = height

	// to decrease db set overload
	if len(votes) > 0 {
		lc.store.GroupCounterSet(lc.Candidates)
	}

	return lazyValidators, lc.Due
}

func (lc *LazinessCounter) add(validator abci.Validator) {
	var address Address

	// convert slice to array
	copy(address[:], validator.Address)

	_, exists := lc.Candidates[address]
	if !exists {
		lc.Candidates[address] = 0
	}

	lc.Candidates[address] += 1
}

func (lc *LazinessCounter) get() []crypto.Address {
	lazyValidators := make([]crypto.Address, 0, len(lc.Candidates))
	limit := int64(float64(lc.Size) * lc.Ratio)

	// copy data
	for key, value := range lc.Candidates {
		if value >= limit {
			lazyValidator := key // copy of a key array by value
			lazyValidators = append(lazyValidators, lazyValidator[:])
		}
	}

	return lazyValidators
}

func (lc *LazinessCounter) purge() {
	for key, _ := range lc.Candidates {
		delete(lc.Candidates, key)
	}

	lc.store.GroupCounterPurge()
}

func (lc *LazinessCounter) checkEnd() bool {
	if lc.Height == lc.Due {
		return true
	}
	return false
}
