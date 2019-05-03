package types

import (
	"time"

	"github.com/amolabs/amoabci/amo/encoding/binary"
)

const RequestAminoName = "amo/RequestValue"

type RequestValue struct {
	Payment Currency  `json:"payment"`
	Exp     time.Time `json:"exp"`
}

func (value RequestValue) Serialize() ([]byte, error) {
	return cdc.MarshalBinaryBare(value)
}

func (value *RequestValue) Deserialize(data []byte) error {
	return cdc.UnmarshalBinaryBare(data, value)
}

func (value RequestValue) IsExpired() bool {
	return value.Exp.Before(time.Now())
}

var _ binary.Serializer = (*RequestValue)(nil)
var _ binary.Deserializer = (*RequestValue)(nil)
