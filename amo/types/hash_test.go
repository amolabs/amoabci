package types

import (
	"crypto/sha256"
	"testing"
)

func TestHashType(t *testing.T) {
	var h Hash
	hash := sha256.New()
	hash.Write([]byte("hello world"))
	if result := copy(h[:], hash.Sum(nil)); result != 32 {
		t.Logf("excepted: %d actual: %d", 32, result)
		t.Fail()
	}
	if result := h.String(); HelloWorld != result {
		t.Logf("excepted: %s actual: %s", HelloWorld, result)
		t.Fail()
	}
}
