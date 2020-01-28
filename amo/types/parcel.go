package types

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type Extra struct {
	Register json.RawMessage `json:"register,omitempty"`
	Request  json.RawMessage `json:"request,omitempty"`
	Grant    json.RawMessage `json:"grant,omitempty"`
}

type Parcel struct {
	Owner        crypto.Address `json:"owner"`
	Custody      cmn.HexBytes   `json:"custody"`
	ProxyAccount crypto.Address `json:"proxy_account,omitempty"`
	Extra        Extra          `json:"extra,omitempty"`
	OnSale       bool           `json:"on_sale"`
}

type ParcelEx struct {
	*Parcel
	Requests []*RequestEx `json:"requests,omitempty"`
	Usages   []*UsageEx   `json:"usages,omitempty"`
}
