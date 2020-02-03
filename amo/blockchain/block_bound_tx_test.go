package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockBindingTx(t *testing.T) {
	ok := CheckBlockBindingTx(1, 10, 5)
	assert.False(t, ok)
	ok = CheckBlockBindingTx(1, 10, 20)
	assert.True(t, ok)

	ok = CheckBlockBindingTx(11, 10, 20)
	assert.False(t, ok)
	ok = CheckBlockBindingTx(11, 20, 20)
	assert.True(t, ok)
}
