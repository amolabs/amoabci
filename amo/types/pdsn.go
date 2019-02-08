package types

import (
	"errors"
)

type PDSNMetaData struct {
	FileHash Hash    `json:"file_hash"`
	Price    uint64  `json:"price"`
	Owner    Address `json:"owner"`
}

var h1 = *NewHashByHexString("b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9")

var FileHashes = map[Hash]PDSNMetaData{
	h1: {
		FileHash: h1,
		Price: 100,
		Owner: "aaaaa",
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