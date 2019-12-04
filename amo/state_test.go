package amo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateFile(t *testing.T) {
	setUpTest(t)
	defer tearDownTest(t)

	state := State{
		MerkleVersion: int64(1),
		Height:        int64(1),
		AppHash:       []byte("appHash"),
		LastHeight:    int64(1),
		LastAppHash:   []byte("lastAppHash"),
	}

	err := state.SaveTo(tmpFile)
	assert.NoError(t, err)

	stateToCompare := State{}

	err = stateToCompare.LoadFrom(tmpFile)
	assert.NoError(t, err)

	assert.Equal(t, state, stateToCompare)
}
