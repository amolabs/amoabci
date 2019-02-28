package db

import (
	dtypes "github.com/amolabs/amoabci/amo/db/types"
	"github.com/amolabs/amoabci/amo/encoding/binary"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/libs/db"
	"github.com/amolabs/tendermint-amo/types"
	"path"
)

var (
	prefixBalance = []byte("balance:")
	prefixParcel = []byte("parcel:")
	prefixRequest = []byte("request:")
	prefixUsage = []byte("usage:")
)

type Store struct {
	store db.DB
}

func getGoLevelDB(name, dir string) *db.GoLevelDB {
	leveldb, err := db.NewGoLevelDB(name, dir)
	if err != nil {
		panic(err)
	}
	return leveldb
}

func NewStore(root string) *Store {
	return &Store{getGoLevelDB("store", path.Join(root, "store"))}
}

// Balance store
func (s Store) setBalance(key []byte, balance *atypes.Currency) {
	b, _ := balance.Serialize()
	s.store.Set(append(prefixBalance, key...), b)
}

func (s Store) getBalance(key []byte) []byte {
	return s.store.Get(append(prefixBalance, key...))
}

func (s Store) SetBalance(addr types.Address, balance atypes.Currency) {
 	s.setBalance(addr.Bytes(), &balance)
}

func (s Store) GetBalance(addr types.Address) *atypes.Currency {
	var c atypes.Currency
	err := binary.Deserialize(s.getBalance(addr.Bytes()), &c)
	if err != nil {
		return nil
	}
	return &c
}

// Parcel store
func (s Store) setParcel(key []byte, value *dtypes.ParcelValue) {
	b, err := value.Serialize()
	if err != nil {
		panic(err)
	}
	s.store.Set(append(prefixParcel, key...), b)
}

func (s Store) getParcel(key []byte) *dtypes.ParcelValue {
	b := s.store.Get(append(prefixParcel, key...))
	if len(b) == 0 {
		return nil
	}
	var parcel dtypes.ParcelValue
	err := binary.Deserialize(b, &parcel)
	if err != nil {
		panic(err)
	}
	return &parcel
}

func (s Store) SetParcel(parcelID []byte, value *dtypes.ParcelValue) {
	s.setParcel(parcelID, value)
}

func (s Store) GetParcel(parcelID []byte) *dtypes.ParcelValue {
	return s.getParcel(parcelID)
}

// Request store
func requestKey(buyer crypto.Address, parcelID []byte) []byte {
	return append(prefixRequest, append(append(buyer, ':'), parcelID...)...)
}

func (s Store) SetRequest(buyer crypto.Address, parcelID []byte, value *dtypes.RequestValue) {
	b, err := value.Serialize()
	if err != nil {
		panic(err)
	}
	s.store.Set(requestKey(buyer, parcelID), b)
}

func (s Store) GetRequest(buyer crypto.Address, parcelID []byte) *dtypes.RequestValue {
	b := s.store.Get(requestKey(buyer, parcelID))
	if len(b) == 0 {
		return nil
	}
	var request dtypes.RequestValue
	err := binary.Deserialize(b, &request)
	if err != nil {
		panic(err)
	}
	return &request
}

// Usage store