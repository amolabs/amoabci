package types

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
)

type DIDDocument struct {
	Owner    crypto.Address  `json:"owner"`
	Document json.RawMessage `json:"document"`
}
