package tx

import (
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/cmd/amocli/util"
)

// QueryAddressInfo is ..
func QueryAddressInfo(target atypes.Address) ([]byte, error) {
	result, err := util.RPCABCIQuery("", target[:])
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}
