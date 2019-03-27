package rpc

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/types"
)

func QueryBalance(address crypto.Address) (types.Currency, error) {
	var balance types.Currency

	bytes, err := json.Marshal(address)
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	result, err := RPCABCIQuery("/balance", tm.HexBytes(bytes))
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	err = json.Unmarshal(result.Response.Value, &balance)
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	return balance, nil
}

func QueryStake(address crypto.Address) (types.Currency, error) {
	var stake types.Currency

	bytes, err := json.Marshal(address)
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	result, err := RPCABCIQuery("/stake", tm.HexBytes(bytes))
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	err = json.Unmarshal(result.Response.Value, &stake)
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	return stake, nil
}

func QueryDelegate(holder crypto.Address, delegator crypto.Address) (types.Currency, error) {
	var amount types.Currency

	keyMap := make(map[string]tm.HexBytes)

	keyMap["holder"] = holder
	keyMap["delegator"] = delegator

	keyMapJSON, err := json.Marshal(keyMap)
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	result, err := RPCABCIQuery("/delegate", keyMapJSON)
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	err = json.Unmarshal(result.Response.Value, &amount)
	if err != nil {
		return *new(types.Currency).Set(0), err
	}

	return amount, nil
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

	err = json.Unmarshal(result.Response.Value, requestValue)
	if err != nil {
		return requestValue, err
	}

	return requestValue, nil
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

	err = json.Unmarshal(result.Response.Value, usageValue)
	if err != nil {
		return usageValue, err
	}

	return usageValue, nil
}
