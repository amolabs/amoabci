package types

import (
	"github.com/amolabs/amoabci/amo/encoding/binary"
	cmn "github.com/tendermint/tendermint/libs/common"
	"time"
)

const UsageAminoName = "amo/UsageValue"

type UsageValue struct {
	Custody cmn.HexBytes
	Exp     time.Time
}

func (value UsageValue) Serialize() ([]byte, error) {
	return cdc.MarshalBinaryBare(value)
}

func (value *UsageValue) Deserialize(data []byte) error {
	return cdc.UnmarshalBinaryBare(data, value)
}

func (value UsageValue) IsExpired() bool {
	return value.Exp.Before(time.Now())
}

var _ binary.Serializer = (*UsageValue)(nil)
var _ binary.Deserializer = (*UsageValue)(nil)
