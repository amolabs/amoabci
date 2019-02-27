package types

import (
	"encoding/json"
	"github.com/amolabs/tendermint-amo/types"
	"strings"
)

const (
	TxTransfer = "transfer"
	TxPurchase = "purchase"
)

type Message struct {
	Command   string          `json:"command"`
	Timestamp int64           `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type Transfer struct {
	From   types.Address `json:"from"`
	To     types.Address `json:"to"`
	Amount Currency  `json:"amount"`
}

type Purchase struct {
	From     types.Address `json:"from"`
	FileHash Hash    `json:"file_hash"`
}

func ParseTx(tx []byte) (Message, interface{}) {
	var message Message

	err := json.Unmarshal(tx, &message)
	if err != nil {
		panic(err)
	}

	message.Command = strings.ToLower(message.Command)

	var payload interface{}
	switch message.Command {
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
