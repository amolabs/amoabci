package types

import (
	"encoding/json"
	"time"

	"github.com/tendermint/tendermint/crypto"
)

type RequestValue struct {
	Payment Currency  `json:"payment"`
	Exp     time.Time `json:"exp"`

	Register json.RawMessage `json:"register"`
	Request  json.RawMessage `json:"request"`
}

type RequestValueEx struct {
	*RequestValue
	Buyer crypto.Address `json:"buyer"`
}

func (value RequestValue) IsExpired() bool {
	return value.Exp.Before(time.Now())
}
