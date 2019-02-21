package privval

import (
	"github.com/amolabs/amoabci/amo/crypto/p256"
	"github.com/tendermint/go-amino"
	cryptoAmino "github.com/amolabs/tendermint-amo/crypto/encoding/amino"
)

var cdc = amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	RegisterRemoteSignerMsg(cdc)
	p256.RegisterAmino(cdc)
}
