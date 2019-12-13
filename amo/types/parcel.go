package types

import (
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type ParcelValue struct {
	Owner   crypto.Address `json:"owner"`
	Custody cmn.HexBytes   `json:"custody"`
	Info    cmn.HexBytes   `json:"info,omitempty"`
}

type ParcelValueEx struct {
	*ParcelValue
	Requests []*RequestValueEx `json:"requests,omitempty"`
	Usages   []*UsageValueEx   `json:"usages,omitempty"`
}
