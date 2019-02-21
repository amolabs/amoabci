package db

import (
	"github.com/amolabs/amoabci/amo/encoding/binary"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/tendermint-amo/libs/db"
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

func (s Store) SetBalance(addr *atypes.Address, balance *atypes.Currency) {
 	s.setBalance(addr[:], balance)
}

func (s Store) GetBalance(addr *atypes.Address) *atypes.Currency {
	var c atypes.Currency
	_ = binary.Deserialize(s.getBalance(addr[:]), &c)
	return &c
}

// Parcel store
func (s Store) setParcel(key []byte, value *ParcelValue) {

}

func (s Store) getParcel(key []byte) *ParcelValue {
	return nil
}

// Request store

// Usage store