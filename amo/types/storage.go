package types

import (
	"github.com/tendermint/tendermint/crypto"
)

const StorageIDLen = 4

type Storage struct {
	Owner           crypto.Address `json:"owner"`
	Url             string         `json:"url"`
	RegistrationFee Currency       `json:"registration_fee"`
	HostingFee      Currency       `json:"hosting_fee"`
	Active          bool           `json:"active"`
}
