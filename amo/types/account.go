package types

import (
	gobinary "encoding/binary"
	"github.com/amolabs/amoabci/amo/encoding/binary"
)

type Account struct {
	Address        `json:"address"`
	Balance        uint64  `json:"balance"`
	PurchasedFiles HashSet `json:"purchased_files"`
}

func (acc Account) Serialize() ([]byte, error) {
	data := make([]byte, AddressSize+64/8+len(acc.PurchasedFiles)*HashSize)
	copy(data, acc.Address)
	gobinary.PutUvarint(data[AddressSize:], acc.Balance)
	hset, _ := binary.Serialize(acc.PurchasedFiles)
	copy(data[AddressSize+64/8:], hset)
	return data, nil
}

func (acc *Account) Deserialize(data []byte) error {
	*acc = Account{}
	acc.Address = Address(data[0:AddressSize])
	acc.Balance, _ = gobinary.Uvarint(data[AddressSize:])
	err := binary.Deserialize(data[AddressSize+64/8:], &acc.PurchasedFiles)
	if err != nil {
		panic(nil)
	}
	return nil
}

var _ binary.Serializer = (*Account)(nil)
var _ binary.Deserializer = (*Account)(nil)
