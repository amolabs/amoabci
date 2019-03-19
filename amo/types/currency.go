package types

import (
	"errors"
	"math/big"

	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	maxCurrencyHex    = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
	currencyLen       = 256 / 8
)

// Currency uses big endian for compatibility to big.Int
type Currency struct {
	big.Int
}

var (
	maxCurrency big.Int
)

func init() {
	maxCurrency.SetString(maxCurrencyHex, 16)
}

func isExceed(i *big.Int) bool {
	return maxCurrency.Cmp(i) != 1
}

func (c *Currency) Set(x uint64) *Currency {
	c.SetUint64(x)
	return c
}

func (c *Currency) SetString(x string, base int) (*Currency, error) {
	i, ok := c.Int.SetString(x, base)
	if !ok {
		return nil, cmn.NewError("Fail to convert hex string(%v)", x)
	}
	if isExceed(i) {
		return nil, cmn.NewError("Currency supports up to 32 bytes;%v", x)
	}
	*c = Currency{
		Int: *i,
	}
	return c, nil
}

func (c Currency) String() string {
	return c.Text(10)
}

func (c Currency) MarshalJSON() ([]byte, error) {
	return []byte("\"" + c.String() + "\""), nil
}

func (c *Currency) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return errors.New(
			"Currency should be represented as double-quoted string.")
	}
	*c = Currency{}
	s = s[1 : len(s)-1]
	if len(s) > currencyLen*2 {
		return cmn.NewError("Currency supports up to 32 bytes(%v)", s)
	}
	if len(s) == 0 {
		return nil
	}
	_, err := c.SetString(s, 10)
	return err
}

func (c *Currency) Add(a *Currency) *Currency {
	if isExceed(c.Int.Add(&c.Int, &a.Int)) {
		panic(cmn.NewError("Cannot add"))
	}
	return c
}

func (c *Currency) Sub(a *Currency) *Currency {
	if c.Int.Sub(&c.Int, &a.Int).Sign() == -1 {
		panic(cmn.NewError("Cannot subtract"))
	}
	return c
}

func (c Currency) Equals(a *Currency) bool {
	return c.Cmp(&a.Int) == 0
}

func (c Currency) GreaterThan(a *Currency) bool {
	return c.Cmp(&a.Int) == 1
}

func (c Currency) LessThan(a *Currency) bool {
	return c.Cmp(&a.Int) == -1
}
