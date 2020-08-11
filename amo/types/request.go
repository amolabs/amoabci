package types

import (
	"github.com/tendermint/tendermint/crypto"
)

type Request struct {
	Agency    crypto.Address `json:"agency"`
	Recipient crypto.Address `json:"recipient"`
	Payment   Currency       `json:"payment"`
	Dealer    crypto.Address `json:"dealer,omitempty"`
	DealerFee Currency       `json:"dealer_fee,omitempty"`
	Extra     Extra          `json:"extra,omitempty"`
}

type RequestEx struct {
	*Request
}
