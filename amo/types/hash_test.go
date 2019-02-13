package types

import (
	"crypto/sha256"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHashType(t *testing.T) {
	r := require.New(t)
	var h Hash
	hash := sha256.New()
	hash.Write([]byte("hello world"))
	r1 := copy(h[:], hash.Sum(nil))
	r.Equal(32, r1)
	r2 := h.String()
	r.Equal(HelloWorld, r2)
}
