package rpc

import (
	"encoding/json"

	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"

	"github.com/amolabs/amoabci/amo/operation"
	"github.com/amolabs/amoabci/client/keys"
	"github.com/amolabs/amoabci/crypto/p256"
)

var (
	RpcRemote     = "tcp://0.0.0.0:26657"
	rpcWsEndpoint = "/websocket"
)

// MakeMessage handles making tx message
func MakeMessage(t string, nonce uint32, payload interface{}, key keys.Key) (types.Tx, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var privKey p256.PrivKeyP256
	copy(privKey[:], key.PrivKey)

	msg := operation.Message{
		Type:    t,
		Payload: raw,
	}

	err = msg.Sign(privKey)
	if err != nil {
		return nil, err
	}

	tx, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// RPCBroadcastTxCommit handles sending transactions
func RPCBroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	cli := client.NewHTTP(RpcRemote, rpcWsEndpoint)
	return cli.BroadcastTxCommit(tx)
}

// RPCABCIQuery handles querying
func RPCABCIQuery(path string, data cmn.HexBytes) (*ctypes.ResultABCIQuery, error) {
	cli := client.NewHTTP(RpcRemote, rpcWsEndpoint)
	return cli.ABCIQuery(path, data)
}

// RPCStatus handle querying the status
func RPCStatus() (*ctypes.ResultStatus, error) {
	cli := client.NewHTTP(RpcRemote, rpcWsEndpoint)
	return cli.Status()
}

func RPCBlock(height int64) (*ctypes.ResultBlock, error) {
	cli := client.NewHTTP(RpcRemote, rpcWsEndpoint)
	return cli.Block(&height)
}
