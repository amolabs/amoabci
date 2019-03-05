package keys

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/amolabs/amoabci/cmd/amocli/util"
)

const (
	testKeyFile = "key_test/keys.json"

	tester1 = "alice"
	tester2 = "bob"
)

var passphrase = map[string]string{
	tester1: "hello i'm alice!",
	tester2: "this is my passprhase",
}

func testInit(t *testing.T) {
	err := util.EnsureFile(testKeyFile)
	assert.NoError(t, err)
}

func testEnd(t *testing.T) {
	err := os.RemoveAll(testKeyFile)
	assert.NoError(t, err)
}

func TestKey(t *testing.T) {
	testInit(t)

	// tester1(alice), encryption=true
	err := Generate(tester1, []byte(passphrase[tester1]), true, testKeyFile)
	assert.NoError(t, err)

	keyStatus := Check(tester1, testKeyFile)
	assert.Equal(t, Encrypted, keyStatus)

	// Generate a key with existing nickname -> error
	err = Generate(tester1, []byte(passphrase[tester1]), true, testKeyFile)
	assert.Error(t, err)

	// tester2(bob), encryption=false
	err = Generate(tester2, nil, false, testKeyFile)
	assert.NoError(t, err)

	keyStatus = Check(tester2, testKeyFile)
	assert.Equal(t, Exists, keyStatus)

	// Remove tester1(alice)'s key with tester2(bob)'s passphrase -> error
	err = Remove(tester1, []byte(passphrase[tester2]), testKeyFile)
	assert.Error(t, err)

	err = Remove(tester1, []byte(passphrase[tester1]), testKeyFile)
	assert.NoError(t, err)

	keyStatus = Check(tester1, testKeyFile)
	assert.Equal(t, NoExists, keyStatus)

	// Remove already removed tester1(alice)'s key -> eror
	err = Remove(tester1, nil, testKeyFile)
	assert.Error(t, err)

	err = Remove(tester2, nil, testKeyFile)
	assert.NoError(t, err)

	keyStatus = Check(tester2, testKeyFile)
	assert.Equal(t, NoExists, keyStatus)

	testEnd(t)
}

func TestKeyGenerateWithEncryption(t *testing.T) {
	testInit(t)

	err := Generate(tester1, []byte(passphrase[tester1]), true, testKeyFile)
	assert.NoError(t, err)

	keyStatus := Check(tester1, testKeyFile)
	assert.Equal(t, Encrypted, keyStatus)

	testEnd(t)
}

func TestKeyGenerate(t *testing.T) {
	testInit(t)

	err := Generate(tester1, nil, false, testKeyFile)
	assert.NoError(t, err)

	KeyStatus := Check(tester1, testKeyFile)
	assert.Equal(t, Exists, KeyStatus)

	testEnd(t)
}

func TestKeyRemoveWithEncryption(t *testing.T) {
	testInit(t)

	err := Generate(tester1, []byte(passphrase[tester1]), true, testKeyFile)
	assert.NoError(t, err)

	keyStatus := Check(tester1, testKeyFile)
	assert.Equal(t, Encrypted, keyStatus)

	err = Remove(tester1, []byte(passphrase[tester2]), testKeyFile)
	assert.Error(t, err)

	err = Remove(tester1, []byte(passphrase[tester1]), testKeyFile)
	assert.NoError(t, err)

	keyStatus = Check(tester1, testKeyFile)
	assert.Equal(t, NoExists, keyStatus)

	testEnd(t)
}

func TestKeyRemove(t *testing.T) {
	testInit(t)

	err := Generate(tester1, nil, false, testKeyFile)
	assert.NoError(t, err)

	keyStatus := Check(tester1, testKeyFile)
	assert.Equal(t, Exists, keyStatus)

	err = Remove(tester1, nil, testKeyFile)
	assert.NoError(t, err)

	keyStatus = Check(tester1, testKeyFile)
	assert.Equal(t, NoExists, keyStatus)

	testEnd(t)
}
