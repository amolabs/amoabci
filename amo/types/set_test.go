package types

import (
	"encoding/json"
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"reflect"
	"testing"
)

func TestAddressSetType(t *testing.T) {
	set := make(AddressSet)
	key := testAddr
	set[key] = true

	b1, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b1))
	set2 := make(AddressSet)
	err = json.Unmarshal(b1, &set2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(set2)
}

func TestHashSetType(t *testing.T) {
	set := make(HashSet)
	key := NewHashFromHexString(HelloWorld)
	set[*key] = true

	b1, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b1))
	set2 := make(HashSet)
	err = json.Unmarshal(b1, &set2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(set2)
}

func TestAddressSetBinary(t *testing.T) {
	set := make(AddressSet)
	set[testAddr] = true
	set[testAddr2] = true
	b, _ := binary.Serialize(set)
	t.Log(b)
	var set2 AddressSet
	_ = binary.Deserialize(b, &set2)
	t.Log(set2)
	if !reflect.DeepEqual(set, set2) {
		t.Fail()
	}
}
