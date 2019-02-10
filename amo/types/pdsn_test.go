package types

import (
	"testing"
)

func TestRequestMetaData(t *testing.T) {
	var metaData PDSNMetaData
	err := RequestMetaData(*NewHashByHexString(HelloWorld), &metaData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(metaData)
}
