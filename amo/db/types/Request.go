package types

import (
	gobinary "encoding/binary"
	"github.com/amolabs/amoabci/amo/encoding/binary"
	atypes "github.com/amolabs/amoabci/amo/types"
	"time"
)

type RequestValue struct {
	Payment atypes.Currency
	Exp     time.Time
}

func (value RequestValue) Serialize() ([]byte, error) {
	data := make([]byte, 64/8+64/8)
	payment, _ := value.Payment.Serialize()
	copy(data, payment)
	gobinary.LittleEndian.PutUint64(data[64/8:], uint64(value.Exp.Unix()))
	return data, nil
}

func (value *RequestValue) Deserialize(data []byte) error {
	var payment atypes.Currency
	err := binary.Deserialize(data[0:64/8], &payment)
	if err != nil {
		return err
	}
	exp := time.Unix(int64(gobinary.LittleEndian.Uint64(data[64/8:])), 0)
	if err != nil {
		return err
	}
	*value = RequestValue{
		Payment: payment,
		Exp: exp,
	}
	return nil
}

func (value RequestValue) IsExpired() bool {
	return value.Exp.Before(time.Now())
}

var _ binary.Serializer = (*RequestValue)(nil)
var _ binary.Deserializer = (*RequestValue)(nil)
