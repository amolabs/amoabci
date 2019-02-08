package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

const HelloWorld = "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

func TestTest(t *testing.T) {
	var x = map[string]int{
		"x": 1,
	}
	t.Log(x["x"])
	t.Log(x["y"])
}

func TestParseTx(t *testing.T) {
	// {
	//  "type": "purchase",
	//  "timestamp": 1548399359964,
	//  "payload": {
	//    "from": "bbbbb",
	//    "file_hash": "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	//  }
	// }
	MSG, _ := hex.DecodeString("7b2274797065223a227075726368617365222c2274696d657374616d70223a3135343833393" +
		"93335393936342c227061796c6f6164223a7b2266726f6d223a226262626262222c2266696c655f68617368223a2262393464" +
		"32376239393334643365303861353265353264376461376461626661633438346566653337613533383065653930383866376" +
		"1636532656663646539227d7d")

	msg, payload := ParseTx(MSG)
	purchase := payload.(*Purchase)

	if msg.Timestamp != 1548399359964 || msg.Type != "purchase" {
		t.Fail()
	}

	t.Log(hex.EncodeToString(purchase.FileHash[:]))
	if hex.EncodeToString(purchase.FileHash[:]) != HelloWorld {
		t.Fail()
	}
}

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

func TestHashSetType(t *testing.T) {
	set := make(HashSet)
	key := NewHashByHexString(HelloWorld)
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

func TestAddressSetType(t *testing.T) {
	set := make(AddressSet)
	key := Address("aaaaa")
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