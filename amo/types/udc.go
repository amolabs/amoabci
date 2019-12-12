package types

import (
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type UDC struct {
	Id        cmn.HexBytes     `json:"id"`        // required // TODO: remove
	Issuer    crypto.Address   `json:"owner"`     // required
	Operators []crypto.Address `json:"operators"` // optional
	Desc      string           `json:"desc"`      // optional
	Total     Currency         `json:"total"`     // required
}
