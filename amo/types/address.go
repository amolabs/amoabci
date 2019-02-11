package types

import (
	"github.com/amolabs/amoabci/amo/encoding/base58"
	"github.com/tendermint/tendermint/crypto"
)

const (
	AddressSize = 5
	AddressVersion = byte(0x0)
)

type Address string

func doubleHash(b []byte) []byte {
	return crypto.Sha256(crypto.Sha256(b))
}

func GenAddress(pubKey crypto.PubKey) crypto.Address {
	r160 := crypto.Ripemd160(doubleHash(pubKey.Bytes()))
	er160 := make([]byte, 1 + 160/8)
	er160[0] = AddressVersion
	copy(er160[1:], r160)
	checksum := doubleHash(doubleHash(r160))[:4]
	address := append(er160, checksum...)
	encoded := base58.Encode(address)
	return []byte(encoded)
}