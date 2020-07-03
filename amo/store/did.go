package store

import (
	"encoding/json"

	"github.com/amolabs/amoabci/amo/types"
)

var (
	prefixDID = []byte("did:")
)

func makeDIDKey(did []byte) []byte {
	return append(prefixDID, did...)
}

func (s Store) SetDIDEntry(id []byte, value *types.DIDEntry) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.set(makeDIDKey(id), b)
	return nil
}

func (s Store) GetDIDEntry(id []byte, committed bool) *types.DIDEntry {
	b := s.get(makeDIDKey(id), committed)
	if len(b) == 0 {
		return nil
	}
	var entry types.DIDEntry
	err := json.Unmarshal(b, &entry)
	if err != nil {
		return nil
	}
	return &entry
}

func (s Store) DeleteDIDEntry(id []byte) {
	s.remove(makeDIDKey(id))
}
