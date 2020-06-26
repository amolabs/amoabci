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

func (s Store) SetDIDDocument(did []byte, value *types.DIDDocument) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.set(makeParcelKey(did), b)
	return nil
}

func (s Store) GetDIDDocument(did []byte, committed bool) *types.DIDDocument {
	b := s.get(makeDIDKey(did), committed)
	if len(b) == 0 {
		return nil
	}
	var doc types.DIDDocument
	err := json.Unmarshal(b, &doc)
	if err != nil {
		return nil
	}
	return &doc
}

func (s Store) DeleteDIDDocument(did []byte) {
	s.remove(makeDIDKey(did))
}
