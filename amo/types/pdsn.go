package types

import (
	"errors"
)

type PDSNMetaData struct {
	FileHash Hash    `json:"file_hash"`
	Price    uint64  `json:"price"`
	Owner    Address `json:"owner"`
}

var (
	H1            = *NewHashFromHexString(HelloWorld)
	SampleAddress = NewAddress([]byte("B2F18D445ADD140711B64E7370C8AD44DA083EEB"))
)

var FileHashes = map[Hash]PDSNMetaData{
	H1: {
		FileHash: H1,
		Price:    100,
		Owner:    *SampleAddress,
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
