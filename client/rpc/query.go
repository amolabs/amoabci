package rpc

import (
	"github.com/amolabs/tendermint-amo/crypto"
)

// QueryAddressInfo is ..
func QueryAddressInfo(target crypto.Address) ([]byte, error) {
	result, err := RPCABCIQuery("", target[:])
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}
