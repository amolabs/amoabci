package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrencyJSON(t *testing.T) {
	testJson := "100"
	var amount Currency
	err := json.Unmarshal([]byte(testJson), &amount)
	assert.NoError(t, err)
	assert.Equal(t, Currency(100), amount)

	// TODO: test for UnmarshalJSON
}
