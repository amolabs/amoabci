package types

import (
	gobinary "encoding/binary"
	"github.com/amolabs/amoabci/amo/encoding/binary"
)

type Account struct {
	Balance        Currency  `json:"balance"`
	PurchasedFiles HashSet `json:"purchased_files"`
}

func (acc Account) Serialize() ([]byte, error) {
	data := make([]byte, 64/8+len(acc.PurchasedFiles)*HashSize)
	gobinary.PutUvarint(data, uint64(acc.Balance))
	hset, _ := binary.Serialize(acc.PurchasedFiles)
	copy(data[64/8:], hset)
	return data, nil
}

func (acc *Account) Deserialize(data []byte) error {
	*acc = Account{}
	balance, _ := gobinary.Uvarint(data)
	acc.Balance = Currency(balance)
	err := binary.Deserialize(data[64/8:], &acc.PurchasedFiles)
	if err != nil {
		panic(nil)
	}
	return nil
}

var _ binary.Serializer = (*Account)(nil)
var _ binary.Deserializer = (*Account)(nil)