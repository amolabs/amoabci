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
func (s Store) SetBalance(addr types.Address, balance atypes.Currency) {
	b, _ := balance.Serialize()
	s.store.Set(append(prefixBalance, addr.Bytes()...), b)
}

func (s Store) GetBalance(addr types.Address) atypes.Currency {
	var c atypes.Currency
	balance := s.store.Get(append(prefixBalance, addr.Bytes()...))
	if len(balance) == 0 {
		return 0
	}
	err := binary.Deserialize(balance, &c)
	if err != nil {
		panic(err)
	}
	return c
}

// Parcel store
func parcelKey(parcelID []byte) []byte {
	return append(prefixParcel, parcelID...)
}

func (s Store) SetParcel(parcelID []byte, value *dtypes.ParcelValue) {
	b, err := value.Serialize()
	if err != nil {
		panic(err)
	}
	s.store.Set(parcelKey(parcelID), b)
}

func (s Store) GetParcel(parcelID []byte) *dtypes.ParcelValue {
	b := s.store.Get(parcelKey(parcelID))
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

func (s Store) DeleteParcel(parcelID []byte) {
	s.store.DeleteSync(parcelKey(parcelID))
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

func (s Store) DeleteRequest(buyer crypto.Address, parcelID []byte) {
	s.store.DeleteSync(requestKey(buyer, parcelID))
}

// Usage store
func usageKey(buyer crypto.Address, parcelID []byte) []byte {
	return append(prefixUsage, append(append(buyer, ':'), parcelID...)...)
}

func (s Store) SetUsage(buyer crypto.Address, parcelID []byte, value *dtypes.UsageValue) {
	b, err := value.Serialize()
	if err != nil {
		panic(err)
	}
	s.store.Set(usageKey(buyer, parcelID), b)
}

func (s Store) GetUsage(buyer crypto.Address, parcelID []byte) *dtypes.UsageValue {
	b := s.store.Get(usageKey(buyer, parcelID))
	if len(b) == 0 {
		return nil
	}
	var usage dtypes.UsageValue
	err := binary.Deserialize(b, &usage)
	if err != nil {
		panic(err)
	}
	return &usage
}

func (s Store) DeleteUsage(buyer crypto.Address, parcelID []byte) {
	s.store.DeleteSync(usageKey(buyer, parcelID))
}