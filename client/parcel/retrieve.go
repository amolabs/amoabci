package parcel

import (
	"github.com/paust-team/paust-db/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func Retrieve(parcelID []byte) (*ctypes.ResultABCIQuery, error) {
	inputFetchObj := client.InputFetchObj{Ids: [][]byte{parcelID}}

	HTTPClient := client.NewHTTPClient(rpcRemote)
	return HTTPClient.Fetch(inputFetchObj)
}
