package store

import (
	"encoding/binary"

	"github.com/tendermint/tendermint/crypto"
)

type (
	Address        [crypto.AddressSize]byte
	LazyValidators map[Address]int64
)

// Record map[Address]int64
func (s Store) GroupCounterSet(candidates LazyValidators) {
	for address, count := range candidates {
		cb := make([]byte, 8)
		binary.BigEndian.PutUint64(cb, uint64(count))

		s.lazinessCounterDB.Set(address[:], cb)
	}
}

func (s Store) GroupCounterGetLazyValidators() LazyValidators {
	lazyValidators := LazyValidators{}

	itr := s.lazinessCounterDB.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		address := Address{}
		copy(address[:], itr.Key())
		count := int64(binary.BigEndian.Uint64(itr.Value()))

		lazyValidators[address] = count
	}

	return lazyValidators
}

func (s Store) GroupCounterPurge() {
	itr := s.lazinessCounterDB.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()

		s.lazinessCounterDB.Delete(k)
	}
}
