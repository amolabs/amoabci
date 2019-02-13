package types

import (
	"encoding/json"
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
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
