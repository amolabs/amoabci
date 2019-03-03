package tx

import (
	"github.com/amolabs/amoabci/cmd/amocli/util"
	"github.com/amolabs/tendermint-amo/crypto"
)

// QueryAddressInfo is ..
func QueryAddressInfo(target crypto.Address) ([]byte, error) {
	result, err := util.RPCABCIQuery("", target[:])
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}
