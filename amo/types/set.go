package types

import (
	"bytes"
	"encoding/json"
	"github.com/amolabs/amoabci/amo/encoding/binary"
)

type Iterator interface {
	Next(c chan interface{})
}

type HashSet map[Hash]bool

type AddressSet map[Address]bool

func Next(iter Iterator, c chan interface{}) {

}

func (set AddressSet) MarshalJSON() ([]byte, error) {
	d := make([]Address, len(set))
	i := 0
	for k := range set {
		d[i] = k
		i += 1
	}
	return json.Marshal(d)
}

func (set *AddressSet) UnmarshalJSON(data []byte) error {
	*set = make(AddressSet)
	if len(data) < 3 {
		return nil
	}
	addresses := bytes.Split(data[1:len(data)-1], []byte(","))
	for _, address := range addresses {
		(*set)[*NewAddress(address[1:len(address)-1])] = true
	}
	return nil
}

func (set AddressSet) Serialize() ([]byte, error) {
	s := make([]byte, AddressSize*len(set))
	i := 0
	for k := range set {
		copy(s[i*AddressSize:(i+1)*AddressSize], k[:])
		i += 1
	}
	return s, nil
}

func (set *AddressSet) Deserialize(data []byte) error {
	length := len(data) / AddressSize
	*set = make(map[Address]bool, length)
	for i := 0; i < length; i++ {
		(*set)[*NewAddress(data[i*AddressSize:(i+1)*AddressSize])] = true
	}
	return nil
}

func (set HashSet) MarshalJSON() ([]byte, error) {
	data := make([]Hash, len(set))
	i := 0
	for k := range set {
		data[i] = k
		i += 1
	}
	return json.Marshal(data)
}

func (set *HashSet) UnmarshalJSON(data []byte) error {
	*set = make(HashSet)
	if len(data) < 3 {
		return nil
	}
	hashes := bytes.Split(data[1:len(data)-2], []byte(","))
	for i, hash := range hashes {
		if i == 0 {
			(*set)[*NewHashByHexBytes(hash[1:])] = true
		} else {
			(*set)[*NewHashByHexBytes(hash[1 : len(hash)-1])] = true
		}
	}
	return nil
}

func (set HashSet) Serialize() ([]byte, error) {
	s := make([]byte, HashSize*len(set))
	i := 0
	for k := range set {
		copy(s[i*HashSize:(i+1)*HashSize], k[:])
		i += 1
	}
	return s, nil
}

func (set *HashSet) Deserialize(data []byte) error {
	length := len(data) / HashSize
	*set = make(map[Hash]bool, length)
	for i := 0; i < length; i++ {
		(*set)[*NewHash(data[i*HashSize : (i+1)*HashSize])] = true
	}
	return nil
}

var _ json.Marshaler = (*HashSet)(nil)
var _ json.Unmarshaler = (*HashSet)(nil)
var _ json.Marshaler = (*AddressSet)(nil)
var _ json.Unmarshaler = (*AddressSet)(nil)

var _ binary.Serializer = (*AddressSet)(nil)
var _ binary.Deserializer = (*AddressSet)(nil)
var _ binary.Serializer = (*HashSet)(nil)
var _ binary.Deserializer = (*HashSet)(nil)
