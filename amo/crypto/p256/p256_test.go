package p256

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenPrivKey(t *testing.T) {
	r := require.New(t)
	privKey := PrivKeyP256{}
	privKeyBytes, _ := hex.DecodeString("CDDB6810F12C7713B97D685316EE56086C463449E03BBB6A256CC6547B342A70")
	privKey.SetBytes(privKeyBytes)
	pubKey := privKey.PubKey()
	pubKeyBytes, _ := hex.DecodeString("047E10BF3D00C47403FD73AAAAA46B9CF9E680A80E259658E5FA0699FA3658F712ED4025E469F3CD0E9B872DBCA44158A450A04CAEA48005B52C541EAD92FCF581")
	r.Equal(pubKey.Bytes(), pubKeyBytes)
	msg := []byte("aaa")
	sig, err := privKey.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	r.True(pubKey.VerifyBytes(msg, sig))
}
