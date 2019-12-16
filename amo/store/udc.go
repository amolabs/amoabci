package store

import (
	"encoding/json"
	"errors"

	tm "github.com/tendermint/tendermint/types"

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

// UDC Balance store
func getUDCBalanceKey(udc []byte, addr tm.Address) []byte {
	key := append(prefixBalance, udc...)
	key = append(key, ':')
	key = append(key, addr.Bytes()...)
	return key
}

func (s Store) SetUDCBalance(udc []byte,
	addr tm.Address, balance *types.Currency) error {
	zero := new(types.Currency).Set(0)
	balanceKey := getUDCBalanceKey(udc, addr)

	if balance.LessThan(zero) {
		return errors.New("negative balance")
	}

	// pre-process for setting zero balance, just remove corresponding key
	if s.has(balanceKey) && balance.Equals(zero) {
		s.remove(balanceKey)
		return nil
	}

	b, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	s.set(balanceKey, b)

	return nil
}

func (s Store) GetUDCBalance(udc []byte,
	addr tm.Address, committed bool) *types.Currency {
	c := types.Currency{}
	balance := s.get(getUDCBalanceKey(udc, addr), committed)
	if len(balance) == 0 {
		return &c
	}
	err := json.Unmarshal(balance, &c)
	if err != nil {
		return &c
	}
	return &c
}
