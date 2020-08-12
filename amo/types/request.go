package types

import (
	"github.com/tendermint/tendermint/crypto"
)

type Request struct {
	Payment   Currency       `json:"payment"`
	Agency    crypto.Address `json:"agency,omitempty"`
	Dealer    crypto.Address `json:"dealer,omitempty"`
	DealerFee Currency       `json:"dealer_fee,omitempty"`
	Extra     Extra          `json:"extra,omitempty"`
}

type RequestEx struct {
	*Request
	Recipient crypto.Address `json:"recipient"`
}
