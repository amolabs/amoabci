package p256

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func TestSignAndVerifyP256(t *testing.T) {
	privKey := GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestP256JSON(t *testing.T) {
	privKey := GenPrivKey()
	pubKey := privKey.PubKey()

	bPrivKey, err := json.Marshal(privKey)
	require.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("\"%s\"", privKey), string(bPrivKey))
	var RprivKey PrivKeyP256
	err = json.Unmarshal(bPrivKey, &RprivKey)
	require.Nil(t, err)
	assert.True(t, privKey.Equals(RprivKey))

	bPubKey, err := json.Marshal(pubKey)
	require.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("\"%s\"", pubKey), string(bPubKey))
	var RpubKey PubKeyP256
	err = json.Unmarshal(bPubKey, &RpubKey)
	require.Nil(t, err)
	assert.True(t, pubKey.Equals(RpubKey))
}
