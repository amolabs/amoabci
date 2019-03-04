package types

import (
	"github.com/amolabs/tendermint-amo/crypto/p256"
	"testing"
)

func TestGenRandomAddress(t *testing.T) {
	pri := p256.GenPrivKey()
	pub := pri.PubKey()
	t.Log(GenAddress(pub))
}
