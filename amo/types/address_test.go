package types

import (
	"encoding/hex"
	"github.com/amolabs/amoabci/amo/crypto/p256"
	"testing"
)

const (
	testPriKey = "7e86b1729aa04fd8563fbe09587366d0e646c280677c1a5bd55769a62d589c866d94b84063700bd987d2de8b3aad7c3afaec329d542343019ee093103c7244b4"
)

var (
	secret, _ = hex.DecodeString(testPriKey)
	priKey    = p256.GenPrivKeyFromSecret(secret)
	priKey2   = p256.GenPrivKey()
)

var (
	testAddr  = GenTestAddress(priKey.PubKey()) // B2F18D445ADD140711B64E7370C8AD44DA083EEB
	testAddr2 = GenTestAddress(priKey2.PubKey())
)

func TestGenAddress(t *testing.T) {
	key := priKey.PubKey()
	t.Log(GenTestAddress(key))
	t.Log(GenMainAddress(key))
}

func TestGenRandomAddress(t *testing.T) {
	pri := p256.GenPrivKey()
	pub := pri.PubKey()
	t.Log(GenTestAddress(pub))
}
