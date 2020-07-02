package store

import (
	"encoding/json"
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/amo/types"
)

var (
	prefixHibernate = []byte("hibernate:")
)

func makeHibernateKey(valAddress []byte) []byte {
	return append(prefixHibernate, valAddress...)
}

func (s Store) SetHibernate(val crypto.Address, hib *types.Hibernate) error {
	key := makeHibernateKey(val)
	b, err := json.Marshal(hib)
	if err != nil {
		return fmt.Errorf("Invalid hibertate description")
	}

	s.set(key, b)

	return nil
}

func (s Store) GetHibernate(val crypto.Address, committed bool) *types.Hibernate {
	hib := types.Hibernate{}
	b := s.get(makeHibernateKey(val), committed)
	if len(b) == 0 {
		return nil
	}
	err := json.Unmarshal(b, &hib)
	if err != nil {
		return nil
	}
	return &hib
}

func (s Store) GetHibernates(committed bool) (vals []crypto.Address, hibs []*types.Hibernate) {
	prefixLen := len(prefixHibernate)
	start := prefixHibernate
	end := make([]byte, prefixLen)
	copy(end, start)
	end[prefixLen-1] = ';'
	s.merkleTree.IterateRange(start, end, true, func(k, v []byte) bool {
		val := k[prefixLen : prefixLen+crypto.AddressSize]
		var hib types.Hibernate
		err := json.Unmarshal(v, &hib)
		if err != nil {
			return true
		}
		vals = append(vals, val)
		hibs = append(hibs, &hib)
		return false
	})

	return
}

func (s Store) DeleteHibernate(val crypto.Address) {
	s.remove(makeHibernateKey(val))
}
