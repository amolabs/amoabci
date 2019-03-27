package amo

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

func queryBalance(store *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Code = code.QueryCodeNoKey
		return
	}

	var addr crypto.Address
	err := json.Unmarshal(queryData, &addr)
	if err != nil {
		res.Code = code.QueryCodeBadKey
		return
	}

	bal := store.GetBalance(addr)
	jsonstr, _ := json.Marshal(bal)
	res.Log = string(jsonstr)
	// XXX: tendermint will convert this using base64 encoding
	res.Value = []byte(jsonstr)
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryStake(store *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Code = code.QueryCodeNoKey
		return
	}

	var addr crypto.Address
	err := json.Unmarshal(queryData, &addr)
	if err != nil {
		res.Code = code.QueryCodeBadKey
		return
	}

	stake := store.GetStake(addr)
	jsonstr, _ := json.Marshal(stake)
	res.Log = string(jsonstr)
	res.Value = []byte(jsonstr)
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryDelegate(store *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Code = code.QueryCodeNoKey
		return
	}
	res.Key = queryData

	keyMap := make(map[string]tm.HexBytes)
	err := json.Unmarshal(queryData, &keyMap)
	if err != nil {
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["holder"]; !ok {
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["delegator"]; !ok {
		res.Code = code.QueryCodeBadKey
		return
	}
	holder := crypto.Address(keyMap["holder"])
	if len(holder) != crypto.AddressSize {
		res.Code = code.QueryCodeBadKey
		return
	}
	delegator := crypto.Address(keyMap["delegator"])
	if len(delegator) != crypto.AddressSize {
		res.Code = code.QueryCodeBadKey
		return
	}

	delegate := store.GetDelegate(holder)
	if delegate == nil {
		res.Code = code.QueryCodeNoMatch
		return
	}
	jsonstr, _ := json.Marshal(delegate)
	res.Value = jsonstr
	res.Log = string(jsonstr)
	res.Code = code.QueryCodeOK

	return
}

func queryParcel(store *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Code = code.QueryCodeNoKey
		return
	}

	// TODO: check parcel id
	parcel := store.GetParcel(queryData)
	if parcel == nil {
		res.Code = code.QueryCodeNoMatch
		return
	}

	jsonstr, _ := json.Marshal(parcel)
	res.Log = string(jsonstr)
	res.Value = []byte(jsonstr)
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryRequest(store *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Code = code.QueryCodeNoKey
		return
	}

	keyMap := make(map[string]tm.HexBytes)
	err := json.Unmarshal(queryData, &keyMap)
	if err != nil {
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["buyer"]; !ok {
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["target"]; !ok {
		res.Code = code.QueryCodeBadKey
		return
	}
	addr := crypto.Address(keyMap["buyer"])
	if len(addr) != crypto.AddressSize {
		res.Code = code.QueryCodeBadKey
		return
	}

	// TODO: check parcel id
	parcelID := keyMap["target"]

	request := store.GetRequest(addr, parcelID)
	if request == nil {
		res.Code = code.QueryCodeNoMatch
		return
	}
	jsonstr, _ := json.Marshal(request)
	res.Log = string(jsonstr)
	res.Value = []byte(jsonstr)
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryUsage(store *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Code = code.QueryCodeNoKey
		return
	}

	keyMap := make(map[string]tm.HexBytes)
	err := json.Unmarshal(queryData, &keyMap)
	if err != nil {
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["buyer"]; !ok {
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["target"]; !ok {
		res.Code = code.QueryCodeBadKey
		return
	}
	addr := crypto.Address(keyMap["buyer"])
	if len(addr) != crypto.AddressSize {
		res.Code = code.QueryCodeBadKey
		return
	}

	// TODO: check parcel id
	parcelID := keyMap["target"]

	request := store.GetUsage(addr, parcelID)
	if request == nil {
		res.Code = code.QueryCodeNoMatch
		return
	}
	jsonstr, _ := json.Marshal(request)
	res.Log = string(jsonstr)
	res.Value = []byte(jsonstr)
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}
