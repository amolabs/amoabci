package types

import (
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"reflect"
	"testing"
)

func TestAccountBinary(t *testing.T) {
	acc := Account{
		Balance:        5000,
		PurchasedFiles: make(HashSet),
	}
	acc.PurchasedFiles[*NewHashByHexString(HelloWorld)] = true
	b, _ := binary.Serialize(acc)
	var acc2 Account
	err := binary.Deserialize(b, &acc2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(acc, acc2) {
		t.Log(acc, acc2)
		t.Fail()
	}
}
