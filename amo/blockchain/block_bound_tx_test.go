package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockBindingManager(t *testing.T) {
	// g: 3, f: 1, t: 0
	bbm := NewBlockBindingManager(3, 0)

	ok := bbm.Check(1)
	assert.False(t, ok)

	// g: 3, f: 1, t: 1
	bbm.Update()

	ok = bbm.Check(1)
	assert.True(t, ok)

	// g: 3, f: 1, t: 2
	bbm.Update()

	ok = bbm.Check(1)
	assert.True(t, ok)

	// g: 3, f: 1, t: 3
	bbm.Update()

	ok = bbm.Check(1)
	assert.True(t, ok)

	// g: 3, f: 2, t: 4
	bbm.Update()

	ok = bbm.Check(1)
	assert.False(t, ok)

	// g: 3, f: 3, t: 5
	bbm.Update()

	ok = bbm.Check(2)
	assert.False(t, ok)

	ok = bbm.Check(3)
	assert.True(t, ok)

	ok = bbm.Check(6)
	assert.False(t, ok)
}
