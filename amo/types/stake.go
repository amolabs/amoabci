package types

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/bytes"
)

type Stake struct {
	Validator ed25519.PubKeyEd25519 `json:"validator"`
	Amount    Currency              `json:"amount"`
}

type StakeEx struct {
	*Stake
	Delegates []*DelegateEx `json:"delegates,omitempty"`
}

func (s StakeEx) MarshalJSON() ([]byte, error) {
	// The field type of Validator should be HexBytes, but it is not.
	// To marshal into hex-encoded string, we need to do this weird thing here.
	v := struct {
		Validator bytes.HexBytes `json:"validator"`
		Amount    Currency       `json:"amount"`
		Delegate  []*DelegateEx  `json:"delegates,omitempty"`
	}{
		Validator: s.Validator[:],
		Amount:    s.Amount,
		Delegate:  s.Delegates,
	}
	return json.Marshal(v)
}
