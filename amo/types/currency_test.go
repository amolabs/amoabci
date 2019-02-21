package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAddCurrency(t *testing.T) {
	r := require.New(t)
	a := Currency(100)
	b := Currency(100)
	a.Add(&b)
	r.Equal(uint64(a), uint64(200))
}