package store

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/tendermint/tendermint/crypto"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/types"
)

var (
	prefixIncentiveHeight  = []byte("ba")
	prefixIncentiveAddress = []byte("ab")
)

type IncentiveInfo struct {
	BlockHeight int64           `json:"block_height"`
	Address     crypto.Address  `json:"address"`
	Amount      *types.Currency `json:"amount"`
}

func (s Store) AddIncentiveRecord(height int64, address crypto.Address, amount *types.Currency) error {
	if height < 0 {
		return errors.New("unavailable height")
	}

	if address == nil {
		return errors.New("address is nil")
	}

	if amount.Equals(new(types.Currency).Set(0)) {
		return errors.New("ignore recording 0 value")
	}

	baKey := makeHeightFirstKey(height, address)
	abKey := makeAddressFirstKey(address, height)
	amountValue := amount.Bytes()

	err := s.incentiveHeight.Set(baKey, amountValue)
	if err != nil {
		return err
	}
	s.incentiveAddress.Set(abKey, amountValue)
	if err != nil {
		return err
	}

	return nil
}

func (s Store) GetBlockIncentiveRecords(height int64) []IncentiveInfo {
	var (
		itr        tmdb.Iterator
		incentives []IncentiveInfo
	)

	hb := make([]byte, 8)
	binary.BigEndian.PutUint64(hb, uint64(height))

	itr, err := s.incentiveHeight.Iterator(hb, nil)
	if err != nil {
		s.logger.Error("Store", "GetBlockIncentiveRecords", err.Error())
		return []IncentiveInfo{}
	}
	defer itr.Close()

	for ; itr.Valid() && bytes.HasPrefix(itr.Key(), hb); itr.Next() {
		address := crypto.Address(itr.Key()[len(hb):])
		amount, err := new(types.Currency).SetBytes(itr.Value())
		if err != nil {
			s.logger.Error("Store", "GetBlockIncentiveRecords", err.Error())
			return []IncentiveInfo{}
		}

		incentive := IncentiveInfo{
			BlockHeight: height,
			Address:     address,
			Amount:      amount,
		}

		incentives = append(incentives, incentive)
	}

	return incentives
}

func (s Store) GetAddressIncentiveRecords(address crypto.Address) []IncentiveInfo {
	var (
		itr        tmdb.Iterator
		incentives []IncentiveInfo
	)

	ab := address.Bytes()

	itr, err := s.incentiveAddress.Iterator(ab, nil)
	if err != nil {
		s.logger.Error("Store", "GetAddressIncentiveRecords", err.Error())
		return []IncentiveInfo{}
	}
	defer itr.Close()

	for ; itr.Valid() && bytes.HasPrefix(itr.Key(), ab); itr.Next() {
		blockHeight := int64(binary.BigEndian.Uint64(itr.Key()[len(ab):]))
		amount, err := new(types.Currency).SetBytes(itr.Value())
		if err != nil {
			s.logger.Error("Store", "GetAddressIncentiveRecords", err.Error())
			return []IncentiveInfo{}
		}

		incentive := IncentiveInfo{
			BlockHeight: blockHeight,
			Address:     address,
			Amount:      amount,
		}

		incentives = append(incentives, incentive)
	}

	return incentives
}

func (s Store) GetIncentiveRecord(height int64, address crypto.Address) IncentiveInfo {
	ba := makeHeightFirstKey(height, address)

	value, err := s.incentiveHeight.Get(ba)
	if err != nil {
		return IncentiveInfo{}
	}
	if value == nil {
		return IncentiveInfo{}
	}

	amount, err := new(types.Currency).SetBytes(value)
	if err != nil {
		return IncentiveInfo{}
	}

	if amount.Equals(new(types.Currency).Set(0)) {
		return IncentiveInfo{}
	}

	return IncentiveInfo{
		BlockHeight: height,
		Address:     address,
		Amount:      amount,
	}
}

func makeHeightFirstKey(height int64, address crypto.Address) []byte {
	key := make([]byte, 8+20) // 64-bit height + 20-byte address

	hb := make([]byte, 8)
	binary.BigEndian.PutUint64(hb, uint64(height))

	copy(key[8-len(hb):], hb)
	copy(key[8:], address)

	return key
}

func makeAddressFirstKey(address crypto.Address, height int64) []byte {
	key := make([]byte, 20+8) // 20-byte address + 64-bit height

	hb := make([]byte, 8)
	binary.BigEndian.PutUint64(hb, uint64(height))

	copy(key[20-len(address):], address)
	copy(key[20:], hb)

	return key
}
