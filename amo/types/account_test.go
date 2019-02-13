package types

import (
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccountBinary(t *testing.T) {
	r := require.New(t)
	acc := Account{
		Balance:        5000,
		PurchasedFiles: make(HashSet),
	}
	acc.PurchasedFiles[*NewHashFromHexString(HelloWorld)] = true
	b, err := binary.Serialize(acc)
	r.NoError(err)
	var acc2 Account
	err = binary.Deserialize(b, &acc2)
	r.NoError(err)
	r.Equal(acc, acc2)
}
