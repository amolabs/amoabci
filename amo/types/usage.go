package types

import (
	"time"

	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/encoding/binary"
)

const UsageAminoName = "amo/UsageValue"

type UsageValue struct {
	Custody cmn.HexBytes `json:"custody"`
	Exp     time.Time    `json:"exp"`
}

type UsageValueEx struct {
	*UsageValue
	Buyer crypto.Address `json:"buyer"`
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
