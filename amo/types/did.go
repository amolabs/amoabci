package types

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
)

type DIDEntry struct {
	Owner    crypto.Address  `json:"owner"`
	Document json.RawMessage `json:"document"`
}
