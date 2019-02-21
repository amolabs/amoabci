package types

import "encoding/binary"

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

func (c *Currency) Add(op *Currency) {
	*c += *op
}
