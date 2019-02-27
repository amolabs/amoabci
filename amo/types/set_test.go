package types

import (
	"encoding/json"
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"github.com/amolabs/tendermint-amo/crypto/p256"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

var (
	testAddr  = GenAddress(p256.GenPrivKey().PubKey()) // B2F18D445ADD140711B64E7370C8AD44DA083EEB
	testAddr2 = GenAddress(p256.GenPrivKey().PubKey())
)

func TestAddressSetJSON(t *testing.T) {
	r := require.New(t)
	set := make(AddressSet)
	key := testAddr
	set[*key] = true
	b1, err := json.Marshal(set)
	r.NoError(err)
	t.Log(string(b1))
	set2 := make(AddressSet)
	err = json.Unmarshal(b1, &set2)
	r.NoError(err)
	r.True(reflect.DeepEqual(set, set2))
}

func TestAddressSetBinary(t *testing.T) {
	r := require.New(t)
	set := make(AddressSet)
	set[*testAddr] = true
	set[*testAddr2] = true
	b, err := binary.Serialize(set)
	r.NoError(err)
	var set2 AddressSet
	err = binary.Deserialize(b, &set2)
	r.NoError(err)
	r.True(reflect.DeepEqual(set, set2))
}

func TestHashSetJSON(t *testing.T) {
	r := require.New(t)
	set := make(HashSet)
	key := NewHashFromHexString(HelloWorld)
	set[*key] = true
	b1, err := json.Marshal(set)
	r.NoError(err)
	t.Log(string(b1))
	set2 := make(HashSet)
	err = json.Unmarshal(b1, &set2)
	r.NoError(err)
	r.True(reflect.DeepEqual(set, set2))
}
