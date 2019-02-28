package types

import "github.com/amolabs/amoabci/amo/encoding/binary"

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
