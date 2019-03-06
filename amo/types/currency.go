package types

import (
	"encoding/binary"
	"encoding/json"
	"errors"
)

type Currency uint64

func (c Currency) Serialize() ([]byte, error) {
	b := make([]byte, 64/8)
	binary.LittleEndian.PutUint64(b, uint64(c))
	return b, nil
}

func (c *Currency) Deserialize(data []byte) error {
	*c = Currency(binary.LittleEndian.Uint64(data))
	return nil
}

// TODO: UnmarshalJSON
func (c *Currency) UnmarshalJSON(data []byte) error {
	var number json.Number
	err := json.Unmarshal(data, &number)
	if err != nil {
		return errors.New("Currency should be represented as double-quoted integer or floating-point number")
	}

	tmp, err := number.Int64()
	if err != nil {
		return err
	}
	*c = Currency(tmp)

	return nil
}
