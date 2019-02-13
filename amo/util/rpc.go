package util

import (
	"encoding/json"
	"fmt"
	atypes "github.com/amolabs/amoabci/amo/types"
	"github.com/tendermint/tendermint/libs/common"
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
	routes["dump_blocks"] = rpcserver.NewRPCFunc(rpcDumpBlocks, "start,size")
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

type ResultDumpBlocks struct {
	Blocks []ctypes.ResultBlock
}

func rpcDumpBlocks(start, size int64) (*ResultDumpBlocks, error) {
	status, _ := core.Status()
	lastHeight := status.SyncInfo.LatestBlockHeight
	if start > lastHeight {
		if start <= 0 {
			return nil, fmt.Errorf("Height must be greater than 0")
		}
		if start > lastHeight {
			return nil, fmt.Errorf("Height must be less than or equal to the current blockchain height")
		}
	}
	length := common.MinInt64(lastHeight-start+1, size)
	result := ResultDumpBlocks{
		Blocks: make([]ctypes.ResultBlock, length),
	}
	for i := int64(0); i < length; i++ {
		height := int64(i + start)
		block, err := core.Block(&height)
		if err != nil {
			return nil, err
		}
		result.Blocks[i] = *block
	}
	return &result, nil
}
