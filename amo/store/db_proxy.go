// +build !cleveldb

package store

import tmdb "github.com/tendermint/tm-db"

func NewDBProxy(name, dir string) (tmdb.DB, error) {
	return tmdb.NewGoLevelDB(name, dir)
}
