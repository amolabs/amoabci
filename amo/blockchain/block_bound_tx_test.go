package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockBindingTx(t *testing.T) {
	err := checkBlockBindingTx(1, 10, 5)
	assert.Error(t, err)
	err = checkBlockBindingTx(1, 10, 20)
	assert.NoError(t, err)

	err = checkBlockBindingTx(11, 10, 20)
	assert.Error(t, err)
	err = checkBlockBindingTx(11, 20, 20)
	assert.NoError(t, err)

	err = checkBlockBindingTx(10, 30, 20)
	assert.Error(t, err)
	err = checkBlockBindingTx(11, 20, 20)
	assert.NoError(t, err)
}
