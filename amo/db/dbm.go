package db

import (
	"encoding/binary"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/types"
	"path"
)

const (
	dbBalanceName = "balance"
	dbParcelName = "parcel"
	dbRequestName = "request"
	dbUsageName = "usage"
)

type Store struct {
	balance db.DB
	parcel db.DB
	request db.DB
	usage db.DB
}

func getGoLevelDB(name, dir string) *db.GoLevelDB {
	leveldb, err := db.NewGoLevelDB(name, dir)
	if err != nil {
		panic(err)
	}
	return leveldb
}

func NewStore(root string) *Store {
	store := Store{
		balance: getGoLevelDB(dbBalanceName, path.Join(root, dbBalanceName)),
		parcel: getGoLevelDB(dbParcelName, path.Join(root, dbParcelName)),
		request: getGoLevelDB(dbRequestName, path.Join(root, dbRequestName)),
		usage: getGoLevelDB(dbUsageName, path.Join(root, dbUsageName)),
	}
	return &store
}

// Balance store
func (s Store) SetBalance(addr types.Address, balance uint64) {
	b := make([]byte, 64/8)
	binary.LittleEndian.PutUint64(b, balance)
 	s.balance.Set(addr.Bytes(), b)
}

func (s Store) GetBalance(addr types.Address) uint64 {
	return binary.LittleEndian.Uint64(s.balance.Get(addr.Bytes()))
}

// Parcel store

// Request store

// Usage store