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
	batch := s.laziCache.NewBatch()
	defer batch.Close()

	for address, count := range candidates {
		ab := make([]byte, len(address))
		cb := make([]byte, 8)
		copy(ab, address[:]) // deep copy
		binary.BigEndian.PutUint64(cb, uint64(count))

		batch.Set(ab, cb)
	}

	err := batch.Write()
	if err != nil {
		s.logger.Error("Store", "GroupCounterSet", err.Error())
	}
}

func (s Store) GroupCounterGetLazyValidators() LazyValidators {
	lazyValidators := LazyValidators{}

	itr, err := s.laziCache.Iterator(nil, nil)
	if err != nil {
		s.logger.Error("Store", "GroupCounterGetLazyValidators", err.Error())
		return LazyValidators{}
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		address := Address{}
		copy(address[:], itr.Key())
		count := int64(binary.BigEndian.Uint64(itr.Value()))

		lazyValidators[address] = count
	}

	return lazyValidators
}

func (s Store) GroupCounterReset() {
	itr, err := s.laziCache.Iterator(nil, nil)
	if err != nil {
		s.logger.Error("Store", "GroupCounterPurge", err.Error())
		return
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()

		err := s.laziCache.Delete(k)
		if err != nil {
			s.logger.Error("Store", "GroupCounterPurge", err.Error())
			continue
		}
	}
}
