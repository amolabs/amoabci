package rpc

import (
	"encoding/json"

	"github.com/amolabs/tendermint-amo/crypto"
	tm "github.com/amolabs/tendermint-amo/libs/common"

	"github.com/amolabs/amoabci/amo/types"
)

func QueryBalance(address crypto.Address) (types.Currency, error) {
	bytes, err := json.Marshal(address)
	result, err := RPCABCIQuery("/balance", tm.HexBytes(bytes))
	if err != nil {
		return types.Currency(0), err
	}

	var balance types.Currency
	json.Unmarshal(result.Response.Value, &balance)
	return balance, nil
}
