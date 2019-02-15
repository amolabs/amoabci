package types

import (
	"encoding/hex"
	"testing"
)

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

	if msg.Timestamp != 1548399359964 || msg.Command != TxPurchase {
		t.Fail()
	}

	t.Log(hex.EncodeToString(purchase.FileHash[:]))
	if hex.EncodeToString(purchase.FileHash[:]) != HelloWorld {
		t.Fail()
	}
}
