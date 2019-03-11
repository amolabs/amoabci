package rpc

import (
	"encoding/hex"
	"encoding/json"

	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/p256"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"github.com/amolabs/tendermint-amo/rpc/client"
	ctypes "github.com/amolabs/tendermint-amo/rpc/core/types"
	"github.com/amolabs/tendermint-amo/types"

	"github.com/amolabs/amoabci/amo/operation"
	"github.com/amolabs/amoabci/client/keys"
)

var (
	rpcRemote     = "tcp://0.0.0.0:26657"
	rpcWsEndpoint = "/websocket"
)

// MakeMessage handles making tx message
func MakeMessage(t string, nonce uint32, payload interface{}, key keys.Key) (types.Tx, error) {
	var (
		signer        = crypto.Address{}
		signingPubKey = p256.PubKeyP256{}
		signature     = cmn.HexBytes{}
	)

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	privKey := p256.GenPrivKeyFromSecret(key.PrivKey)
	signerHex, err := hex.DecodeString(key.Address)
	if err != nil {
		return nil, err
	}

	signer = crypto.Address(signerHex)
	copy(signingPubKey[:], key.PubKey)
	signature, err = privKey.Sign(raw)
	if err != nil {
		return nil, err
	}

	msg := operation.Message{
		Command:       t,
		Signer:        signer,
		SigningPubKey: signingPubKey,
		Signature:     signature,
		Payload:       raw,
		Nonce:         nonce,
	}

	tx, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// RPCBroadcastTxCommit handles sending transactions
func RPCBroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	cli := client.NewHTTP(rpcRemote, rpcWsEndpoint)
	return cli.BroadcastTxCommit(tx)
}

// RPCABCIQuery handles querying
func RPCABCIQuery(path string, data cmn.HexBytes) (*ctypes.ResultABCIQuery, error) {
	cli := client.NewHTTP(rpcRemote, rpcWsEndpoint)
	return cli.ABCIQuery(path, data)
}

// RPCStatus handle querying the status
func RPCStatus() (*ctypes.ResultStatus, error) {
	cli := client.NewHTTP(rpcRemote, rpcWsEndpoint)
	return cli.Status()
}
