package rpc

import (
	"encoding/json"
	//"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	amino "github.com/tendermint/go-amino"
	//ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmrpc "github.com/tendermint/tendermint/rpc/lib/types"
)

func TestQueryJsonFormat(t *testing.T) {
	// This test actually does not ensure anything in amoabci codes. The
	// purpose of this chunk of codes is to provide a full example of JSON
	// formatted request message.

	testcdc := amino.NewCodec()

	req, _ := tmrpc.MapToRequest(testcdc,
		tmrpc.JSONRPCStringID("test-client"),
		"abci_query",
		map[string]interface{}{"path": "/balance", "data": "rawdata"})
	reqJson, _ := json.Marshal(req)
	//fmt.Println("req =", string(reqJson))

	// XXX Current implementation of MapToRequest changes the order of "data"
	// and "path", but the order does not matter in real operation env.
	assertJson := []byte(`{"jsonrpc":"2.0","id":"test-client","method":"abci_query","params":{"data":"rawdata","path":"/balance"}}`)
	var assertReq tmrpc.RPCRequest
	_ = json.Unmarshal(assertJson, &assertReq)

	assert.Equal(t, assertReq, req)
	assert.Equal(t, string(assertJson), string(reqJson))
}
