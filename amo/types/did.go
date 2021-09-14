package types

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
)

type DIDEntry struct {
	Owner    crypto.Address  `json:"owner,omitempty"` // obsolete
	Document json.RawMessage `json:"document"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

type VCEntry struct {
	Credential json.RawMessage `json:"credential"`
	Meta       json.RawMessage `json:"meta,omitempty"`
}
