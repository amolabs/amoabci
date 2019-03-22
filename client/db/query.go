package db

import (
	"github.com/paust-team/paust-db/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func Query(start, end uint64, owner, qualifier string) (*ctypes.ResultABCIQuery, error) {
	inputQueryObjs := client.InputQueryObj{
		Start:     start,
		End:       end,
		OwnerId:   owner,
		Qualifier: qualifier,
	}

	HTTPClient := client.NewHTTPClient(rpcRemote)
	return HTTPClient.Query(inputQueryObjs)
}
