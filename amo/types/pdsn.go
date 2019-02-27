package types

import (
	"errors"
	"github.com/amolabs/tendermint-amo/types"
)

type PDSNMetaData struct {
	FileHash Hash     `json:"file_hash"`
	Price    Currency `json:"price"`
	Owner    types.Address  `json:"owner"`
}

var (
	H1            = *NewHashFromHexString(HelloWorld)
	SampleAddress = NewAddress([]byte("B2F18D445ADD140711B64E7370C8AD44DA083EEB"))
)

var FileHashes = map[Hash]PDSNMetaData{
	H1: {
		FileHash: H1,
		Price:    100,
		Owner:    []byte("B2F18D445ADD140711B64E7370C8AD44DA083EEB"),
	},
}

func RequestMetaData(fileHash Hash, metaData *PDSNMetaData) error {
	if data, ok := FileHashes[fileHash]; !ok {
		return errors.New("fail to find metadata")
	} else {
		*metaData = data
		return nil
	}
}
