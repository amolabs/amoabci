package db

import (
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"github.com/amolabs/tendermint-amo/crypto"
)

type ParcelValue struct {
	Owner crypto.Address
	Custody crypto.PubKey
	Info []byte `json:"info,omitempty"`
}

func (value ParcelValue) Serialize() ([]byte, error) {
	return nil, nil
}

func (value *ParcelValue) Deserialize(data []byte) error {
	return nil
}

var _ binary.Serializer = (*ParcelValue)(nil)
var _ binary.Deserializer = (*ParcelValue)(nil)

type RequestValue struct {

}

func (value RequestValue) Serialize() ([]byte, error) {
	return nil, nil
}

func (value *RequestValue) Deserialize(data []byte) error {
	return nil
}


var _ binary.Serializer = (*RequestValue)(nil)
var _ binary.Deserializer = (*RequestValue)(nil)

type UsageValue struct {
}

func (value UsageValue) Serialize() ([]byte, error) {
	return nil, nil
}

func (value *UsageValue) Deserialize(data []byte) error {
	return nil
}


var _ binary.Serializer = (*UsageValue)(nil)
var _ binary.Deserializer = (*UsageValue)(nil)