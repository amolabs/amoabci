package store

import (
	"encoding/json"

	"github.com/amolabs/amoabci/amo/types"
)

var (
	prefixUDC = []byte("udc:")
)

func getUDCKey(id []byte) []byte {
	return append(prefixUDC, id...)
}

func (s Store) SetUDC(id []byte, udc *types.UDC) error {
	b, err := json.Marshal(udc)
	if err != nil {
		return err
	}
	// TODO: consider return value 'updated'
	s.set(getUDCKey(id), b)
	return nil
}

func (s Store) GetUDC(id []byte, committed bool) *types.UDC {
	b := s.get(getUDCKey(id), committed)
	if len(b) == 0 {
		return nil
	}
	var udc types.UDC
	err := json.Unmarshal(b, &udc)
	if err != nil {
		return nil
	}
	return &udc
}
