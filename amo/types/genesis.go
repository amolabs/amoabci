package types

import (
	"bytes"
	"fmt"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/types"
	"io/ioutil"
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
	for i, v := range genDoc.Owners {
		if v.Amount == 0 {
			return cmn.NewError("The genesis file cannot contain owner with no balance: %v", v)
		}
		if len(v.Address) > 0 && !bytes.Equal(v.PubKey.Address(), v.Address) {
			return cmn.NewError("Incorrect address for owner %v in the genesis file, should be %v", v, v.PubKey.Address())
		}
		if len(v.Address) == 0 {
			genDoc.Validators[i].Address = v.PubKey.Address()
		}
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

func (genDoc *AMOGenesisDoc) GenesisDocFromFile(genDocFile string) error {
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return cmn.ErrorWrap(err, "Couldn't read GenesisDoc file")
	}
	err = cdc.UnmarshalJSON(jsonBlob, genDoc)
	if err != nil {
		return cmn.ErrorWrap(err, fmt.Sprintf("Error reading GenesisDoc at %v", genDocFile))
	}
	return nil
}
