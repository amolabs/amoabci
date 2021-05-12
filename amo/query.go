package amo

import (
	"encoding/json"
	"sort"
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bytes"

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

func queryVersion(app *AMOApp) (res abci.ResponseQuery) {
	var r struct {
		AppVersion           string   `json:"app_version,omitempty"`
		AppProtocolVersions  []uint64 `json:"app_protocol_versions,omitempty"`
		StateProtocolVersion uint64   `json:"state_protocol_version,omitempty"`
		AppProtocolVersion   uint64   `json:"app_protocol_version,omitempty"`
	}
	r.AppVersion = AMOAppVersion
	protoVersions := make([]uint64, 0, len(AMOProtocolVersions))
	for k := range AMOProtocolVersions {
		protoVersions = append(protoVersions, k)
	}
	sort.Slice(protoVersions, func(i, j int) bool {
		return protoVersions[i] < protoVersions[j]
	})
	r.AppProtocolVersions = protoVersions
	r.StateProtocolVersion = app.state.ProtocolVersion
	r.AppProtocolVersion = app.proto.Version()
	jsonstr, _ := json.Marshal(r)
	res.Log = string(jsonstr)
	res.Key = []byte("version")
	res.Value = jsonstr
	res.Code = code.QueryCodeOK

	return
}

func queryAppConfig(config types.AMOAppConfig) (res abci.ResponseQuery) {
	jsonstr, _ := json.Marshal(config)
	res.Log = string(jsonstr)
	res.Key = []byte("config")
	res.Value = jsonstr
	res.Code = code.QueryCodeOK

	return
}

func queryBalance(s *store.Store, udc string, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var addr crypto.Address
	err := json.Unmarshal(queryData, &addr)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	udcID := uint32(0)
	if udc != "" {
		tmp, err := strconv.ParseInt(udc, 10, 32)
		if err != nil {
			res.Log = "error: cannot convert udc id"
			res.Code = code.QueryCodeBadKey
			return
		}
		udcID = uint32(tmp)
	}

	bal := s.GetUDCBalance(udcID, addr, true)

	jsonstr, _ := json.Marshal(bal)
	res.Log = string(jsonstr)
	// XXX: tendermint will convert this using base64 encoding
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryUDC(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var udcID uint32
	err := json.Unmarshal(queryData, &udcID)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	udc := s.GetUDC(udcID, true)

	jsonstr, _ := json.Marshal(udc)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryUDCLock(s *store.Store, udc string, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var addr crypto.Address
	err := json.Unmarshal(queryData, &addr)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	var udcID uint32
	tmp, err := strconv.ParseInt(udc, 10, 32)
	if err != nil {
		res.Log = "error: cannot convert udc id"
		res.Code = code.QueryCodeBadKey
		return
	}
	udcID = uint32(tmp)

	udcLock := s.GetUDCLock(udcID, addr, true)

	jsonstr, _ := json.Marshal(udcLock)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryStake(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var addr crypto.Address
	err := json.Unmarshal(queryData, &addr)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	stake := s.GetStake(addr, true)
	if stake == nil {
		res.Log = "error: no stake"
		res.Code = code.QueryCodeNoMatch
		return
	}

	stakeEx := types.StakeEx{stake, s.GetDelegatesByDelegatee(addr, true)}
	jsonstr, _ := json.Marshal(stakeEx)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryDelegate(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var addr crypto.Address
	err := json.Unmarshal(queryData, &addr)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	delegate := s.GetDelegate(addr, true)
	if delegate == nil {
		res.Log = "error: no delegate"
		res.Code = code.QueryCodeNoMatch
		return
	}

	jsonstr, _ := json.Marshal(delegate)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryValidator(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var addr crypto.Address
	err := json.Unmarshal(queryData, &addr)
	if err != nil {
		res.Log = "error: unmarshal"
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

func queryHibernate(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var addr crypto.Address
	err := json.Unmarshal(queryData, &addr)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	hib := s.GetHibernate(addr, true)
	if hib == nil {
		res.Code = code.QueryCodeNoMatch
		res.Key = queryData
		return
	}
	jsonstr, _ := json.Marshal(hib)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryStorage(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var storageID uint32
	err := json.Unmarshal(queryData, &storageID)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	storage := s.GetStorage(storageID, true)
	if storage == nil {
		res.Log = "error: no such storage"
		res.Code = code.QueryCodeNoMatch
		return
	}

	jsonstr, _ := json.Marshal(storage)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryDraft(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var draftID uint32
	err := json.Unmarshal(queryData, &draftID)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	draft := s.GetDraftForQuery(draftID, true)
	if draft == nil {
		res.Log = "error: no draft"
		res.Code = code.QueryCodeNoMatch
		return
	}

	draftEx := types.DraftEx{
		DraftForQuery: draft,
		Votes:         s.GetVotes(draftID, true),
	}

	jsonstr, err := json.Marshal(draftEx)
	if err != nil {
		res.Log = "error: marshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryVote(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var param struct {
		DraftID uint32         `json:"draft_id"`
		Voter   crypto.Address `json:"voter"`
	}

	err := json.Unmarshal(queryData, &param)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	if len(param.Voter) != crypto.AddressSize {
		res.Log = "error: not avaiable address"
		res.Code = code.QueryCodeBadKey
		return
	}

	vote := s.GetVote(param.DraftID, param.Voter, true)
	if vote == nil {
		res.Log = "error: no vote"
		res.Code = code.QueryCodeNoMatch
		return
	}

	voteInfo := types.VoteInfo{
		Voter: param.Voter,
		Vote:  vote,
	}

	jsonstr, _ := json.Marshal(voteInfo)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryParcel(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var id bytes.HexBytes
	err := json.Unmarshal(queryData, &id)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	parcel := s.GetParcel(id, true)
	if parcel == nil {
		res.Log = "error: no such parcel"
		res.Code = code.QueryCodeNoMatch
		return
	}

	parcelEx := types.ParcelEx{
		Parcel:   parcel,
		Requests: s.GetRequests(id, true),
		Usages:   s.GetUsages(id, true),
	}

	jsonstr, _ := json.Marshal(parcelEx)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryRequest(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	keyMap := make(map[string]bytes.HexBytes)
	err := json.Unmarshal(queryData, &keyMap)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["recipient"]; !ok {
		res.Log = "error: recipient is missing"
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["target"]; !ok {
		res.Log = "error: target is missing"
		res.Code = code.QueryCodeBadKey
		return
	}
	addr := crypto.Address(keyMap["recipient"])
	if len(addr) != crypto.AddressSize {
		res.Log = "error: not avaiable address"
		res.Code = code.QueryCodeBadKey
		return
	}

	// TODO: parse parcel id
	parcelID := keyMap["target"]

	request := s.GetRequest(addr, parcelID, true)
	if request == nil {
		res.Log = "error: no request"
		res.Code = code.QueryCodeNoMatch
		return
	}

	requestEx := types.RequestEx{
		Request:   request,
		Recipient: addr,
	}

	jsonstr, _ := json.Marshal(requestEx)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryUsage(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	keyMap := make(map[string]bytes.HexBytes)
	err := json.Unmarshal(queryData, &keyMap)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["recipient"]; !ok {
		res.Log = "error: recipient is missing"
		res.Code = code.QueryCodeBadKey
		return
	}
	if _, ok := keyMap["target"]; !ok {
		res.Log = "error: target is missing"
		res.Code = code.QueryCodeBadKey
		return
	}
	addr := crypto.Address(keyMap["recipient"])
	if len(addr) != crypto.AddressSize {
		res.Log = "error: not avaiable address"
		res.Code = code.QueryCodeBadKey
		return
	}

	// TODO: parse parcel id
	parcelID := keyMap["target"]

	usage := s.GetUsage(addr, parcelID, true)
	if usage == nil {
		res.Log = "error: no usage"
		res.Code = code.QueryCodeNoMatch
		return
	}

	usageEx := types.UsageEx{
		Usage:     usage,
		Recipient: addr,
	}

	jsonstr, _ := json.Marshal(usageEx)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}

func queryDIDEntry(s *store.Store, queryData []byte) (res abci.ResponseQuery) {
	if len(queryData) == 0 {
		res.Log = "error: no query_data"
		res.Code = code.QueryCodeNoKey
		return
	}

	var id string
	err := json.Unmarshal(queryData, &id)
	if err != nil {
		res.Log = "error: unmarshal"
		res.Code = code.QueryCodeBadKey
		return
	}

	entry := s.GetDIDEntry(id, true)
	if entry == nil {
		res.Log = "error: no such did entry"
		res.Code = code.QueryCodeNoMatch
		return
	}

	jsonstr, _ := json.Marshal(entry)
	res.Log = string(jsonstr)
	res.Value = jsonstr
	res.Code = code.QueryCodeOK
	res.Key = queryData

	return
}
