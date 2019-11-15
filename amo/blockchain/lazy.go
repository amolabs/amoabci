package blockchain

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

type (
	Address        [crypto.AddressSize]byte
	LazyValidators map[Address]int64
)

type LazinessCounter struct {
	Candidates LazyValidators `json:"lazy_validators"`
	Height     int64          `json:"height"`
	Ratio      float64        `json:"ratio"`
	Size       int64          `json:"size"`
}

func NewLazinessCounter(size int64, ratio float64) LazinessCounter {
	return LazinessCounter{
		Candidates: make(LazyValidators),
		Height:     int64(0),
		Ratio:      ratio,
		Size:       size,
	}
}

func (lc *LazinessCounter) Investigate(commitInfo abci.LastCommitInfo) []crypto.Address {
	var lazyValidators []crypto.Address

	votes := commitInfo.GetVotes()

	if lc.checkEnd() {
		lazyValidators = lc.get()
		lc.purge()
	}

	for _, vote := range votes {
		if !vote.GetSignedLastBlock() {
			lc.add(vote.GetValidator())
		}
	}

	lc.Height += 1

	return lazyValidators
}

func (lc *LazinessCounter) add(validator abci.Validator) {
	var address Address
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

	lc.Height = 0
}

func (lc *LazinessCounter) checkEnd() bool {
	if lc.Height == lc.Size {
		return true
	}
	return false
}
