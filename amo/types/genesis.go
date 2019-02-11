package types

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/types"
)

var cdc = amino.NewCodec()

const (
	GenesisOwnerAminoName = "amo/GenesisOwner"
)

func init() {
	cryptoAmino.RegisterAmino(cdc)
	cdc.RegisterConcrete(GenesisOwner{}, GenesisOwnerAminoName, nil)
}

const (
	ChainID = "amo-testnet"
)

type GenesisOwner struct {
	Address types.Address `json:"address"`
	PubKey  crypto.PubKey `json:"pub_key"`
	Amount  uint64        `json:"amount"`
}

type AMOGenesisDoc struct {
	types.GenesisDoc
	Owners []GenesisOwner `json:"owners,omitempty"`
}

func (genDoc AMOGenesisDoc) ValidateAndComplete() error {
	if err := genDoc.GenesisDoc.ValidateAndComplete(); err != nil {
		return err
	}
	return nil
}

func (genDoc AMOGenesisDoc) SaveAs(file string) error {
	genDocBytes, err := cdc.MarshalJSONIndent(genDoc, "", "  ")
	if err != nil {
		return err
	}
	return cmn.WriteFile(file, genDocBytes, 0644)
}