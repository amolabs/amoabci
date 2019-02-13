package types

import (
	"encoding/hex"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"testing"
)

const (
	testPriKey = "7e86b1729aa04fd8563fbe09587366d0e646c280677c1a5bd55769a62d589c866d94b84063700bd987d2de8b3aad7c3afaec329d542343019ee093103c7244b4"
)

var (
	secret, _ = hex.DecodeString(testPriKey)
	priKey    = ed25519.GenPrivKeyFromSecret(secret)
	priKey2   = ed25519.GenPrivKey()
)

func TestAMOGenesisDoc(t *testing.T) {
	key := priKey.PubKey()
	genDoc := AMOGenesisDoc{
		GenesisDoc: types.GenesisDoc{
			ChainID:         ChainID,
			GenesisTime:     tmtime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
		},
	}
	genDoc.Validators = []types.GenesisValidator{{
		Address: key.Address(),
		PubKey:  key,
		Power:   10,
	}}
	genDoc.Owners = []GenesisOwner{{
		Address: key.Address(),
		PubKey:  key,
		Amount:  5000,
	}}
	if err := genDoc.ValidateAndComplete(); err != nil {
		t.Fatal(err)
	}
	t.Log(genDoc)
	//genDoc.SaveAs("test1.json")
	//genDoc.GenesisDoc.SaveAs("test2.json")
}
