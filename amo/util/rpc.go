package util

import (
	"encoding/json"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/tendermint/tendermint/rpc/core"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/lib/server"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func addRoutes() {
	routes := core.Routes
	routes["transfer"] = rpcserver.NewRPCFunc(rpcTransfer, "from,to,amount")
	routes["purchase"] = rpcserver.NewRPCFunc(rpcPurchase, "from,file_hash")
}

func rpcTransfer(from atypes.Address, to atypes.Address, amount *uint64) (*ctypes.ResultBroadcastTxCommit, error) {
	return core.BroadcastTxCommit(makeMessage(atypes.TxTransfer, atypes.Transfer{
		From:   from,
		To:     to,
		Amount: *amount,
	}))
}

func rpcPurchase(from atypes.Address, fileHash atypes.Hash) (*ctypes.ResultBroadcastTxCommit, error) {
	return core.BroadcastTxCommit(makeMessage(atypes.TxPurchase, atypes.Purchase{
		From:     from,
		FileHash: fileHash,
	}))
}

func makeMessage(t string, payload interface{}) types.Tx {
	raw, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	msg := atypes.Message{
		Type:      t,
		Timestamp: tmtime.Now().Unix(),
		Payload:   raw,
	}
	tx, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return tx
}
