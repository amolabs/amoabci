package rpc

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/types"
)

func QueryBalance(address crypto.Address) (types.Currency, error) {
	bytes, err := json.Marshal(address)
	result, err := RPCABCIQuery("/balance", tm.HexBytes(bytes))
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	var balance types.Currency
	json.Unmarshal(result.Response.Value, &balance)
	return balance, nil
}

func QueryParcel(parcelID []byte) (types.ParcelValue, error) {
	var parcelValue = types.ParcelValue{}

	result, err := RPCABCIQuery("/parcel", parcelID)
	if err != nil {
		return parcelValue, err
	}

	json.Unmarshal(result.Response.Value, &parcelValue)
	return parcelValue, nil
}

func QueryRequest(buyer tm.HexBytes, target tm.HexBytes) (types.RequestValue, error) {
	var requestValue = types.RequestValue{}

	keyMap := make(map[string]tm.HexBytes)

	keyMap["buyer"] = buyer
	keyMap["target"] = target

	keyMapJSON, err := json.Marshal(keyMap)
	if err != nil {
		return requestValue, err
	}

	result, err := RPCABCIQuery("/request", keyMapJSON)
	if err != nil {
		return requestValue, err
	}

	json.Unmarshal(result.Response.Value, requestValue)
	return requestValue, err
}

func QueryUsage(buyer tm.HexBytes, target tm.HexBytes) (types.UsageValue, error) {
	var usageValue = types.UsageValue{}

	keyMap := make(map[string]tm.HexBytes)

	keyMap["buyer"] = buyer
	keyMap["target"] = target

	keyMapJSON, err := json.Marshal(keyMap)
	if err != nil {
		return usageValue, err
	}

	result, err := RPCABCIQuery("/usage", keyMapJSON)
	if err != nil {
		return usageValue, err
	}

	json.Unmarshal(result.Response.Value, usageValue)
	return usageValue, err
}
