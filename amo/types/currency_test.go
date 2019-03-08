package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrencyJSON(t *testing.T) {
	amount := Currency(100)
	d, err := json.Marshal(amount)
	assert.NoError(t, err)
	assert.Equal(t, []byte("\"100\""), d)

	err = json.Unmarshal([]byte("1f"), &amount)
	assert.Error(t, err)
	err = json.Unmarshal([]byte("11"), &amount)
	assert.Error(t, err)
	err = json.Unmarshal([]byte(""), &amount)
	assert.Error(t, err)
	err = json.Unmarshal([]byte("\""), &amount)
	assert.Error(t, err)
	err = json.Unmarshal([]byte("\"1f\""), &amount)
	assert.Error(t, err)

	err = json.Unmarshal([]byte("\"\""), &amount)
	assert.NoError(t, err)
	assert.Equal(t, Currency(0), amount)

	err = json.Unmarshal([]byte("\"100\""), &amount)
	assert.NoError(t, err)
	assert.Equal(t, Currency(100), amount)
}
