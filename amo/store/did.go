package store

import (
	"encoding/json"

	"github.com/amolabs/amoabci/amo/types"
)

var (
	prefixDID = []byte("did:")
	prefixVC  = []byte("vc:")
)

// did

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

// Verifiable Credential

func makeVCKey(vcId string) []byte {
	return append(prefixVC, []byte(vcId)...)
}

func (s Store) SetVCEntry(vcId string, value *types.VCEntry) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.set(makeVCKey(vcId), b)
	return nil
}

func (s Store) GetVCEntry(vcId string, committed bool) *types.VCEntry {
	b := s.get(makeVCKey(vcId), committed)
	if len(b) == 0 {
		return nil
	}
	var vc types.VCEntry
	err := json.Unmarshal(b, &vc)
	if err != nil {
		return nil
	}
	return &vc
}

func (s Store) DeleteVCEntry(vcId string) {
	s.remove(makeVCKey(vcId))
}
