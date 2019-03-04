package types

import (
	"github.com/amolabs/amoabci/amo/encoding/binary"
	atypes "github.com/amolabs/amoabci/amo/types"
	"time"
)

const RequestAminoName = "amo/RequestValue"

type RequestValue struct {
	Payment atypes.Currency
	Exp     time.Time
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
