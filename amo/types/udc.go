package types

import (
	"github.com/tendermint/tendermint/crypto"
)

type UDC struct {
	Owner     crypto.Address   `json:"owner"`     // required
	Desc      string           `json:"desc"`      // optional
	Operators []crypto.Address `json:"operators"` // optional
	Total     Currency         `json:"total"`     // required
}
