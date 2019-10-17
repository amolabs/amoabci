package amo

import (
	"encoding/json"
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

// MERKLE TREE SCOPE (IMPORTANT)
//   Query related funcs must get data from committed(saved) tree
//   NOT from working tree as the users or clients
//   SHOULD see the data which are already commited by validators
//   So, it is mandatory to use 'true' for 'committed' arg input
//   to query data from merkle tree

func queryBalance(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
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

	bal := s.GetBalance(addr, true)
	jsonstr, _ := json.Marshal(bal)
	res.Log = string(jsonstr)
	// XXX: tendermint will convert this using base64 encoding
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryStake(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
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

	stake := s.GetStake(addr, true)
	if stake == nil {
		res.Code = code.QueryCodeNoMatch
	} else {
		res.Code = code.QueryCodeOK
	}

	stakeEx := types.StakeEx{stake, s.GetDelegatesByDelegatee(addr, true)}
	jsonstr, _ := json.Marshal(stakeEx)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Key = queryData

	return
}

func queryDelegate(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
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

	delegate := s.GetDelegate(addr, true)
	if delegate == nil {
		res.Code = code.QueryCodeNoMatch
	} else {
		res.Code = code.QueryCodeOK
	}

	jsonstr, _ := json.Marshal(delegate)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Key = queryData

	return
}

func queryValidator(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
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

	holder := s.GetHolderByValidator(addr, true)
	jsonstr, _ := json.Marshal(crypto.Address(holder))
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryParcel(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Code = code.QueryCodeNoKey
		return
	}

	// TODO: parse parcel id from queryData
	var id tm.HexBytes
	err := json.Unmarshal(queryData, &id)
	if err != nil {
		res.Code = code.QueryCodeBadKey
		return
	}

	parcel := s.GetParcel(id, true)
	if parcel == nil {
		res.Code = code.QueryCodeNoMatch
	} else {
		res.Code = code.QueryCodeOK
	}

	jsonstr, _ := json.Marshal(parcel)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Key = queryData

	return
}

func queryRequest(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
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

	// TODO: parse parcel id
	parcelID := keyMap["target"]

	request := s.GetRequest(addr, parcelID, true)
	if request == nil {
		res.Code = code.QueryCodeNoMatch
	} else {
		res.Code = code.QueryCodeOK
	}

	jsonstr, _ := json.Marshal(request)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Key = queryData

	return
}

func queryUsage(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
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

	// TODO: parse parcel id
	parcelID := keyMap["target"]

	usage := s.GetUsage(addr, parcelID, true)
	if usage == nil {
		res.Code = code.QueryCodeNoMatch
	} else {
		res.Code = code.QueryCodeOK
	}

	jsonstr, _ := json.Marshal(usage)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Key = queryData

	return
}

func queryBlockIncentives(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	var (
		incentives []store.IncentiveInfo
		height     int64
	)

	if len(queryData) == 0 {
		res.Code = code.QueryCodeNoKey
		return
	}

	err := json.Unmarshal(queryData, &height)
	if err != nil {
		res.Code = code.QueryCodeNoKey
		return
	}

	incentives = s.GetBlockIncentiveRecords(height)
	if len(incentives) == 0 {
		res.Code = code.QueryCodeNoMatch
		return
	}

	jsonstr, _ := json.Marshal(incentives)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Key = queryData

	return
}

func queryAddressIncentives(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	var (
		incentives []store.IncentiveInfo
		address    crypto.Address
	)

	if len(queryData) == 0 {
		res.Code = code.QueryCodeNoKey
		return
	}

	err := json.Unmarshal(queryData, &address)
	if err != nil {
		res.Code = code.QueryCodeNoKey
		return
	}

	incentives = s.GetAddressIncentiveRecords(address)
	if len(incentives) == 0 {
		res.Code = code.QueryCodeNoMatch
		return
	}

	jsonstr, _ := json.Marshal(incentives)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Key = queryData

	return
}

func queryIncentive(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	var (
		incentive store.IncentiveInfo
		height    int64
		address   crypto.Address
	)

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
	if _, ok := keyMap["height"]; !ok {
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["address"]; !ok {
		res.Code = code.QueryCodeBadKey
		return
	}

	height, err = strconv.ParseInt(string(keyMap["height"]), 10, 64)
	if err != nil {
		res.Code = code.QueryCodeBadKey
		return
	}

	address = keyMap["address"]

	incentive = s.GetIncentiveRecord(height, address)

	jsonstr, _ := json.Marshal(incentive)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Key = queryData

	return
}
