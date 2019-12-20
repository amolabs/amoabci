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

type ParcelValue struct {
	Owner        crypto.Address `json:"owner"`
	Custody      cmn.HexBytes   `json:"custody"`
	ProxyAccount crypto.Address `json:"proxy_account,omitempty"`

	Extra `json:"extra,omitempty"`
}

type ParcelValueEx struct {
	*ParcelValue
	Requests []*RequestValueEx `json:"requests,omitempty"`
	Usages   []*UsageValueEx   `json:"usages,omitempty"`
}
