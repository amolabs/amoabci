package types

import (
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"github.com/amolabs/tendermint-amo/crypto"
)

type ParcelValue struct {
	Owner   crypto.Address
	Custody []byte
	Info    []byte `json:"info,omitempty"`
}

func (value ParcelValue) Serialize() ([]byte, error) {
	length := crypto.AddressSize + 1 + len(value.Custody) + len(value.Info)
	data := make([]byte, 0, length)
	data = append(data, value.Owner[:]...)
	data = append(data, byte(len(value.Custody)))
	data = append(data, value.Custody...)
	if len(value.Info) != 0 {
		data = append(data, value.Info...)
	}
	return data, nil
}

func (value *ParcelValue) Deserialize(data []byte) error {
	ind := crypto.AddressSize
	owner := crypto.Address(data[0:ind])
	custodyLen := int(data[ind])
	custody := make([]byte, custodyLen)
	ind += 1
	copy(custody[:], data[ind:ind+custodyLen])
	ind += custodyLen
	*value = ParcelValue{
		Owner: owner,
		Custody: custody,
	}
	if len(data) != ind {
		value.Info = make([]byte, len(data)-ind)
		copy(value.Info[:], data[ind:])
	}
	return nil
}

var _ binary.Serializer = (*ParcelValue)(nil)
var _ binary.Deserializer = (*ParcelValue)(nil)
