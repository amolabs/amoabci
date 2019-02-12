package types

import (
	"testing"
)

func TestRequestMetaData(t *testing.T) {
	var metaData PDSNMetaData
	err := RequestMetaData(*NewHashFromHexString(HelloWorld), &metaData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(metaData)
}
