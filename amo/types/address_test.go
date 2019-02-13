package types

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
	"testing"
)

var (
	testAddr  = GenAddress(priKey.PubKey()) // B2F18D445ADD140711B64E7370C8AD44DA083EEB
	testAddr2 = GenAddress(priKey2.PubKey())
)

func TestGenAddress(t *testing.T) {
	key := priKey.PubKey()
	t.Log(GenAddress(key))
}

func TestGenRandomAddress(t *testing.T) {
	pri := ed25519.GenPrivKey()
	pub := pri.PubKey()
	t.Log(GenAddress(pub))
}
