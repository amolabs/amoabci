package types

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/crypto/p256"
)

type Request struct {
	Payment         Currency        `json:"payment"`
	RecipientPubKey p256.PubKeyP256 `json:"recipient_pubkey"`
	Dealer          crypto.Address  `json:"dealer,omitempty"`
	DealerFee       Currency        `json:"dealer_fee,omitempty"`
	Extra           Extra           `json:"extra,omitempty"`
}

type RequestEx struct {
	*Request
	Buyer crypto.Address `json:"buyer"`
}
