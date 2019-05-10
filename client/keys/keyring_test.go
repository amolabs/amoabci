package keys

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/amolabs/amoabci/crypto/p256"
)

const (
	testfile = "test_keyring.json"
)

func _tearDown() {
	os.RemoveAll(testfile)
}

func TestGetKeyRing(t *testing.T) {
	kr, err := GetKeyRing(testfile)
	assert.NoError(t, err)
	assert.NotNil(t, kr)

	_tearDown()
}

func TestGenKey(t *testing.T) {
	kr, err := GetKeyRing(testfile)
	assert.NoError(t, err)
	assert.NotNil(t, kr)
	assert.Equal(t, 0, len(kr.keyList))

	key, err := kr.GenerateNewKey("test", []byte("pass"), true, "test")
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, 40, len(key.Address))
	assert.Equal(t, 65, len(key.PubKey)) // XXX: really?
	assert.True(t, key.Encrypted)
	key2 := kr.GetKey("test")
	assert.NotNil(t, key2)
	assert.Equal(t, key, key2)

	// check if the actual file was updated
	err = kr.Load()
	assert.NoError(t, err)
	key2 = kr.GetKey("test")
	assert.NotNil(t, key2)
	assert.Equal(t, key, key2)

	// test remove
	err = kr.RemoveKey("test")
	assert.NoError(t, err)
	key2 = kr.GetKey("test")
	assert.Nil(t, key2)

	err = kr.Load()
	assert.NoError(t, err)
	key2 = kr.GetKey("test")
	assert.Nil(t, key2)

	// test genkey without enc
	key, err = kr.GenerateNewKey("test", nil, false, "test")
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, 40, len(key.Address))
	assert.Equal(t, 65, len(key.PubKey)) // XXX: really?
	assert.False(t, key.Encrypted)

	_tearDown()
}

func TestImportKey(t *testing.T) {
	kr, err := GetKeyRing(testfile)
	assert.NoError(t, err)
	assert.NotNil(t, kr)
	assert.Equal(t, 0, len(kr.keyList))

	// test import
	testKey := p256.GenPrivKeyFromSecret([]byte("test"))
	testKeyBytes := testKey.RawBytes()
	testPubKey, _ := testKey.PubKey().(p256.PubKeyP256)
	wrongBytes := testKeyBytes[:len(testKey)-1]

	key, err := kr.ImportPrivKey(wrongBytes, "test", []byte("pass"), true)
	assert.Error(t, err)
	assert.Nil(t, key)

	key, err = kr.ImportPrivKey(testKeyBytes, "test", []byte("pass"), true)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, 40, len(key.Address))
	assert.Equal(t, testPubKey.RawBytes(), key.PubKey)
	assert.True(t, key.Encrypted)
	key2 := kr.GetKey("test")
	assert.NotNil(t, key2)
	assert.Equal(t, key, key2)

	// check if the actual file was updated
	err = kr.Load()
	assert.NoError(t, err)
	key2 = kr.GetKey("test")
	assert.NotNil(t, key2)
	assert.Equal(t, key, key2)

	_tearDown()
}
