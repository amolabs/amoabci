package types

import (
	"encoding/binary"
	"errors"
	"strconv"
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

func (c Currency) MarshalJSON() ([]byte, error) {
	// TODO: big number
	s := strconv.FormatUint(uint64(c), 10)
	data := []byte("\"" + s + "\"")
	return data, nil
}

func (c *Currency) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return errors.New(
			"Currency should be represented as double-quoted string.")
	}
	s = s[1 : len(s)-1]
	if len(s) == 0 {
		*c = Currency(0)
		return nil
	}

	// TODO: big number
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*c = Currency(n)
	return nil
}
