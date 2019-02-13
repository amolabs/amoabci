package p256

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenPrivKey(t *testing.T) {
	r := require.New(t)
	privKey := genPribKeyP256FromHexString("00f39137aca2c70ff798ac96d65f06d052484397d9b05d68c54d7efaef12856e")
	pubKey := privKey.PubKey()
	pubKeyBytes, _ := hex.DecodeString("04b3c6735ab65653fea7e18dd4fc9de654c9ff9c3d20b2f38dd120d5df2540c7cfbd9002bf43a03df92b123f19042d424859b9fa88f08ba5340c6e8cc3cb035d2b")
	r.Equal(pubKey.Bytes(), pubKeyBytes)
	msg := []byte("aaa")
	sig, err := privKey.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	sigBytes, _ := hex.DecodeString("3046022100aa5b81a1e0b065e5d1bd21d80abc03066de4722642450eeed7831662d40c30eb0221009e79dafbe04a47c1d49b263b59e534df000f6860f2694c51a72db1e64dafaead")
	// FIXME
	r.Equal(sig, sigBytes)
	// FIXME
	r.True(pubKey.VerifyBytes(msg, sigBytes))
}
