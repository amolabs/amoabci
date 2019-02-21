package types

import (
	"encoding/json"
	"github.com/amolabs/amoabci/amo/encoding/base58"
	"github.com/amolabs/tendermint-amo/crypto"
)

const (
	AddressSize    = 40
	AddressVersion = byte(0x0)
)

type Address [AddressSize]byte

func (addr Address) MarshalJSON() ([]byte, error) {
	data := make([]byte, AddressSize+2)
	data[0] = '"'
	data[len(data)-1] = '"'
	copy(data[1:AddressSize+1], addr[:])
	return data, nil
}

func (addr *Address) UnmarshalJSON(data []byte) error {
	*addr = *NewAddress(data[1 : AddressSize+1])
	return nil
}

func NewAddress(bAddr []byte) *Address {
	addr := Address{}
	copy(addr[:], bAddr)
	return &addr
}

func (addr Address) String() string {
	return string(addr[:])
}

func GenAddress(pubKey crypto.PubKey) *Address {
	r160 := crypto.Ripemd160(crypto.Sha256(pubKey.Bytes()))
	er160 := make([]byte, 1+160/8)
	er160[0] = AddressVersion
	copy(er160[1:], r160)
	checksum := crypto.Sha256(crypto.Sha256(r160))[:4]
	address := append(er160, checksum...)
	encoded := base58.Encode(address)
	return NewAddress([]byte(encoded))
}

var _ json.Marshaler = (*Address)(nil)
var _ json.Unmarshaler = (*Address)(nil)
