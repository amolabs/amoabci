package types

import (
	"encoding/hex"
	"encoding/json"
	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	HashSize   = 32
	HelloWorld = "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
)

type Hash [HashSize]byte

func NewHash(data []byte) *Hash {
	if len(data) != HashSize {
		panic(cmn.NewError("hash: wrong hash size"))
	}
	var h Hash
	copy(h[:], data)
	return &h
}

func NewHashByHexString(hexString string) *Hash {
	if len(hexString) != HashSize<<1 {
		panic(cmn.NewError("hash: wrong hash size"))
	}
	var h Hash
	hash, err := hex.DecodeString(hexString)
	if err != nil {
		panic(err)
	}
	copy(h[:], hash)
	return &h
}

func NewHashByHexBytes(hexBytes []byte) *Hash {
	if len(hexBytes) != HashSize<<1 {
		panic(cmn.NewError("hash: wrong hash size"))
	}
	var h Hash
	_, err := hex.Decode(h[:], hexBytes)
	if err != nil {
		panic(err)
	}
	return &h
}

func (h Hash) MarshalJSON() ([]byte, error) {
	data := make([]byte, HashSize<<1+2)
	data[0] = '"'
	data[len(data)-1] = '"'
	copy(data[1:HashSize<<1+1], []byte(hex.EncodeToString(h[:])))
	return data, nil
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	*h = *NewHashByHexBytes(data[1 : HashSize<<1+1])
	return nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

var _ json.Marshaler = (*Hash)(nil)
var _ json.Unmarshaler = (*Hash)(nil)
