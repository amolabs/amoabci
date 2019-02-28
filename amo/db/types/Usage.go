package types

import "github.com/amolabs/amoabci/amo/encoding/binary"

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
