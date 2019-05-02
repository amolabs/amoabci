package parcel

import (
	"time"

	"github.com/paust-team/paust-db/client"
	cmn "github.com/tendermint/tendermint/libs/common"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func Upload(owner string, data cmn.HexBytes, qualifier string) (*ctypes.ResultBroadcastTxCommit, error) {
	inputDataObjs := []client.InputDataObj{{
		Timestamp: uint64(time.Now().UnixNano()),
		OwnerId:   owner,
		Qualifier: qualifier,
		Data:      data,
	}}

	HTTPClient := client.NewHTTPClient(rpcRemote)

	return HTTPClient.Put(inputDataObjs)
}
