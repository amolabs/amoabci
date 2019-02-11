package types

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/types"
)

const ChainID = "amo-testnet"

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
