package genesis

import (
	"github.com/amolabs/amoabci/amo/crypto/p256"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
)

var cdc = amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	p256.RegisterAmino(cdc)
	cdc.RegisterConcrete(GenesisOwner{}, GenesisOwnerAminoName, nil)
}
