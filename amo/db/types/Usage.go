package types

import (
	gobinary "encoding/binary"
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"time"
)

type UsageValue struct {
	Custody []byte
	Exp     time.Time
}

func (value UsageValue) Serialize() ([]byte, error) {
	data := make([]byte, len(value.Custody)+64/8)
	gobinary.LittleEndian.PutUint64(data, uint64(value.Exp.Unix()))
	copy(data[64/8:], value.Custody)
	return data, nil
}

func (value *UsageValue) Deserialize(data []byte) error {
	exp := time.Unix(int64(gobinary.LittleEndian.Uint64(data)), 0)
	*value = UsageValue{
		Custody: data[64/8:],
		Exp: exp,
	}
	return nil
}

func (value UsageValue) IsExpired() bool {
	return value.Exp.Before(time.Now())
}

var _ binary.Serializer = (*UsageValue)(nil)
var _ binary.Deserializer = (*UsageValue)(nil)
