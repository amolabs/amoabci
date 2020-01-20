package types

import (
	"encoding/binary"
	"encoding/json"
	"strconv"

	tm "github.com/tendermint/tendermint/libs/common"
)

func ConvIDFromHex(IDHex tm.HexBytes) (uint32, []byte, error) {
	var (
		IDStr       string
		IDUint      uint32
		IDByteArray []byte
	)

	err := json.Unmarshal(IDHex, &IDStr)
	if err != nil {
		return IDUint, IDByteArray, err
	}

	tmp, err := strconv.ParseUint(IDStr, 10, 32)
	if err != nil {
		return IDUint, IDByteArray, err
	}

	IDUint = uint32(tmp)

	IDByteArray = ConvIDFromUint(IDUint)

	return IDUint, IDByteArray, nil
}

func ConvIDFromUint(raw uint32) []byte {
	IDByteArray := make([]byte, 4)
	binary.BigEndian.PutUint32(IDByteArray, raw)

	return IDByteArray
}
