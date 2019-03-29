package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrencyJSON(t *testing.T) {
	amount := new(Currency).Set(100)
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
	assert.Equal(t, new(Currency).Set(0), amount)

	err = json.Unmarshal([]byte("\"100\""), &amount)
	assert.NoError(t, err)
	assert.Equal(t, new(Currency).Set(100), amount)

	err = json.Unmarshal([]byte("\"12QQ\""), &amount)
	assert.Error(t, err)
}

func TestCurrencyAdd(t *testing.T) {
	x := new(Currency).Set(1000)
	y := new(Currency).Set(2000)
	z := new(Currency).Set(3000)

	// x += 2000
	x.Add(y)
	assert.Equal(t, z, x)
	assert.True(t, x.Equals(z))
	// x = 2000
	x.Set(1000)
	z.Sub(y)
	assert.Equal(t, x, z)
	// x < z
	assert.True(t, x.LessThan(y))
	// x > z
	assert.False(t, x.GreaterThan(y))
}

func TestCurrencyMax(t *testing.T) {
	_, err := new(Currency).SetString("7" + maxCurrencyHex[1:] , 16)
	assert.NoError(t, err)
	_, err = new(Currency).SetString(maxCurrencyHex + "FF" , 16)
	assert.Error(t, err)
}