package store

import (
	"encoding/json"

	"github.com/amolabs/amoabci/amo/types"
)

var (
	prefixDID = []byte("did:")
)

func makeDIDKey(did string) []byte {
	return append(prefixDID, []byte(did)...)
}

func (s Store) SetDIDEntry(id string, value *types.DIDEntry) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.set(makeDIDKey(id), b)
	return nil
}

func (s Store) GetDIDEntry(id string, committed bool) *types.DIDEntry {
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

func (s Store) DeleteDIDEntry(id string) {
	s.remove(makeDIDKey(id))
}
