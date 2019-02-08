package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

const (
	TxTransfer = "transfer"
	TxPurchase = "purchase"
)

const HashSize = 32

type Address string
type Hash [HashSize]byte
type HashSet map[Hash]bool
type AddressSet map[Address]bool

func (set AddressSet) MarshalJSON() ([]byte, error) {
	data := make([]Address, len(set))
	i := 0
	for k := range set {
		data[i] = k
		i += 1
	}
	return json.Marshal(data)
}

func (set *AddressSet) UnmarshalJSON(data []byte) error {
	*set = make(AddressSet)
	if len(data) < 3 {
		return nil
	}
	addresses := bytes.Split(data[1:len(data)-1], []byte(","))
	for _, address := range addresses {
		(*set)[Address(string(address[1:len(address)-1]))] = true
	}
	return nil
}

func (set HashSet) MarshalJSON() ([]byte, error) {
	data := make([]Hash, len(set))
	i := 0
	for k := range set {
		data[i] = k
		i += 1
	}
	return json.Marshal(data)
}

func (set *HashSet) UnmarshalJSON(data []byte) error {
	*set = make(HashSet)
	if len(data) < 3 {
		return nil
	}
	hashes := bytes.Split(data[1:len(data)-2], []byte(","))
	for i, hash := range hashes {
		if i == 0 {
			(*set)[*NewHashByHexBytes(hash[1:])] = true
		} else {
			(*set)[*NewHashByHexBytes(hash[1:len(hash)-1])] = true
		}
	}
	return nil
}

func NewHash(data []byte) *Hash {
	if len(data) != HashSize {
		panic(errors.New("hash: wrong hash size"))
	}
	var h Hash
	copy(h[:], data)
	return &h
}

func NewHashByHexString(hexString string) *Hash {
	if len(hexString) != HashSize<<1 {
		panic(errors.New("hash: wrong hash size"))
	}
	var h Hash
	hash, err := hex.DecodeString(hexString)
	if err != nil {
		panic(err)
	}
	copy(h[:], hash)
	return &h
}

func NewHashByHexBytes(hexBytes []byte) *Hash {
	if len(hexBytes) != HashSize<<1 {
		panic(errors.New("hash: wrong hash size"))
	}
	var h Hash
	_, err := hex.Decode(h[:], hexBytes)
	if err != nil {
		panic(err)
	}
	return &h
}

func (h Hash) MarshalJSON() ([]byte, error) {
	data := make([]byte, HashSize<<1+2)
	data[0] = '"'
	data[len(data)-1] = '"'
	copy(data[1:HashSize<<1+1], []byte(hex.EncodeToString(h[:])))
	return data, nil
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	*h = *NewHashByHexBytes(data[1:HashSize<<1+1])
	return nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

type Account struct {
	Address        `json:"address"`
	Balance        uint64 `json:"balance"`
	PurchasedFiles HashSet `json:"purchased_files"`
}

type Message struct {
	Type      string          `json:"type"`
	Timestamp int64          `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type Transfer struct {
	From   Address `json:"from"`
	To     Address `json:"to"`
	Amount uint64  `json:"amount"`
}

type Purchase struct {
	From     Address `json:"from"`
	FileHash Hash    `json:"file_hash"`
}

func ParseTx(tx []byte) (Message, interface{}) {
	var message Message

	err := json.Unmarshal(tx, &message)
	if err != nil {
		panic(err)
	}

	message.Type = strings.ToLower(message.Type)

	var payload interface{}
	switch message.Type {
	case TxTransfer:
		payload = new(Transfer)
	case TxPurchase:
		payload = new(Purchase)
	}

	err = json.Unmarshal(message.Payload, &payload)
	if err != nil {
		panic(err)
	}

	return message, payload
}

var _ json.Marshaler = (*Hash)(nil)
var _ json.Unmarshaler = (*Hash)(nil)
var _ json.Marshaler = (*HashSet)(nil)
var _ json.Unmarshaler = (*HashSet)(nil)
var _ json.Marshaler = (*AddressSet)(nil)
var _ json.Unmarshaler = (*AddressSet)(nil)