package types

import (
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
)

const ParcelAminoName = "amo/ParcelValue"

type ParcelValue struct {
	Owner   crypto.Address `json:"owner"`
	Custody cmn.HexBytes   `json:"custody"`
	Info    cmn.HexBytes   `json:"info,omitempty"`
}

type ParcelValueEx struct {
	*ParcelValue
	Requests []*RequestValueEx `json:"requests,omitempty"`
	Usages   []*UsageValueEx   `json:"usages,omitempty"`
}

func (value ParcelValue) Serialize() ([]byte, error) {
	return cdc.MarshalBinaryBare(value)
}

func (value *ParcelValue) Deserialize(data []byte) error {
	return cdc.UnmarshalBinaryBare(data, value)
}

var _ binary.Serializer = (*ParcelValue)(nil)
var _ binary.Deserializer = (*ParcelValue)(nil)
