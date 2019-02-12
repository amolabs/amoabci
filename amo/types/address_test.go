package types

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
	"testing"
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
