package store

import (
	"encoding/json"

	"github.com/amolabs/amoabci/amo/types"
)

var (
	prefixStorage = []byte("storage:")
)

func getStorageKey(id uint32) []byte {
	return append(prefixStorage, ConvIDFromUint(id)...)
}

func (s Store) SetStorage(id uint32, sto *types.Storage) error {
	b, err := json.Marshal(sto)
	if err != nil {
		return err
	}
	// TODO: consider return value 'updated'
	s.set(getStorageKey(id), b)
	return nil
}

func (s Store) GetStorage(id uint32, committed bool) *types.Storage {
	b := s.get(getStorageKey(id), committed)
	if len(b) == 0 {
		return nil
	}
	var sto types.Storage
	err := json.Unmarshal(b, &sto)
	if err != nil {
		return nil
	}
	return &sto
}
