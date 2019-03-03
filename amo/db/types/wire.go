package types

import "github.com/tendermint/go-amino"

var cdc *amino.Codec

func init() {
	cdc = amino.NewCodec()
	cdc.RegisterConcrete(ParcelValue{}, ParcelAminoName, nil)
	cdc.RegisterConcrete(RequestValue{}, RequestAminoName, nil)
	cdc.RegisterConcrete(UsageValue{}, UsageAminoName, nil)
}