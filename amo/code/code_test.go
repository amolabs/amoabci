package code

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCode(t *testing.T) {
	assert.Equal(t, uint32(0), TxCodeOK)
	assert.Equal(t, nil, GetError(TxCodeOK))
	assert.Equal(t, errors.New("Unknown"), GetError(TxCodeUnknown))
	assert.NotEqual(t, errors.New("VoteNotOpen"), GetError(TxCodeAlreadyVoted))
	assert.Equal(t, uint32(0), QueryCodeOK)
	assert.Equal(t, nil, GetError(QueryCodeOK))
	assert.Equal(t, errors.New("BadPath"), GetError(QueryCodeBadPath))
	assert.NotEqual(t, errors.New("BadKey"), GetError(QueryCodeNoKey))
	assert.Panics(t, func() {
		TxNonExistingCode := uint32(5555555)
		GetError(TxNonExistingCode)
	})
}
