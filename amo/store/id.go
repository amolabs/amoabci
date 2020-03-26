package store

import (
	"encoding/binary"
	"encoding/json"
	"strconv"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

func ConvIDFromHex(IDHex tmbytes.HexBytes) (uint32, []byte, error) {
	var (
		IDStr       string
		IDUint      uint32
		IDByteSlice []byte
	)

	err := json.Unmarshal(IDHex, &IDStr)
	if err != nil {
		return IDUint, IDByteSlice, err
	}

	IDByteSlice, err = ConvIDFromStr(IDStr)
	if err != nil {
		return IDUint, IDByteSlice, err
	}

	return IDUint, IDByteSlice, nil
}

func ConvIDFromStr(raw string) ([]byte, error) {
	tmp, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return nil, err
	}

	IDUint := uint32(tmp)

	return ConvIDFromUint(IDUint), nil
}

func ConvIDFromUint(raw uint32) []byte {
	IDByteSlice := make([]byte, 4)
	binary.BigEndian.PutUint32(IDByteSlice, raw)

	return IDByteSlice
}
