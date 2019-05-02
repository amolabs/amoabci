package rpc

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"
)

func QueryBalance(address crypto.Address) ([]byte, error) {
	bytes, err := json.Marshal(address)
	if err != nil {
		return nil, err
	}

	result, err := RPCABCIQuery("/balance", tm.HexBytes(bytes))
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}

func QueryStake(address crypto.Address) ([]byte, error) {
	bytes, err := json.Marshal(address)
	if err != nil {
		return nil, err
	}

	result, err := RPCABCIQuery("/stake", tm.HexBytes(bytes))
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}

func QueryDelegate(address crypto.Address) ([]byte, error) {
	bytes, err := json.Marshal(address)
	if err != nil {
		return nil, err
	}

	result, err := RPCABCIQuery("/delegate", tm.HexBytes(bytes))
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}

func QueryParcel(parcelID []byte) ([]byte, error) {
	result, err := RPCABCIQuery("/parcel", parcelID)
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}

func QueryRequest(buyer tm.HexBytes, target tm.HexBytes) ([]byte, error) {
	keyMap := make(map[string]tm.HexBytes)

	keyMap["buyer"] = buyer
	keyMap["target"] = target

	keyMapJSON, err := json.Marshal(keyMap)
	if err != nil {
		return nil, err
	}

	result, err := RPCABCIQuery("/request", keyMapJSON)
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}

func QueryUsage(buyer tm.HexBytes, target tm.HexBytes) ([]byte, error) {
	keyMap := make(map[string]tm.HexBytes)

	keyMap["buyer"] = buyer
	keyMap["target"] = target

	keyMapJSON, err := json.Marshal(keyMap)
	if err != nil {
		return nil, err
	}

	result, err := RPCABCIQuery("/usage", keyMapJSON)
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}
