package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockBindingManager(t *testing.T) {
	// g: 3, f: 1, t: 0
	bbm := NewBlockBindingManager(0, 3)

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

	ok = bbm.Check(4)
	assert.True(t, ok)

	ok = bbm.Check(5)
	assert.True(t, ok)

	ok = bbm.Check(6)
	assert.False(t, ok)

	// g: 3 -> g: 2
	bbm.Set(2)

	// g: 2, f: 5, t: 6
	bbm.Update()

	ok = bbm.Check(4)
	assert.False(t, ok)

	ok = bbm.Check(5)
	assert.True(t, ok)

	ok = bbm.Check(6)
	assert.True(t, ok)

	ok = bbm.Check(7)
	assert.False(t, ok)

	// g: 2 -> g: 4
	bbm.Set(4)

	// g: 4, f: 5, f: 7
	bbm.Update()

	ok = bbm.Check(5)
	assert.True(t, ok)

	// g: 4, f: 5, f: 8
	bbm.Update()

	ok = bbm.Check(5)
	assert.True(t, ok)

	// g: 4, f: 6, f: 9
	bbm.Update()

	ok = bbm.Check(5)
	assert.False(t, ok)

	ok = bbm.Check(6)
	assert.True(t, ok)
}
