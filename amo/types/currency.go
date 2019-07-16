package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

const (
	maxCurrencyHex = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
	currencyLen    = 256 / 8
	OneAMOUint64   = 1000000000000000000 // in decimal
	//oneAMO         = 0xDE0B6B3A7640000
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

func isTooBig(i *big.Int) bool {
	return maxCurrency.Cmp(i) != 1
}

func (c *Currency) Set(x uint64) *Currency {
	c.SetUint64(x)
	return c
}

func (c *Currency) SetAMO(x float64) *Currency {
	c.SetUint64(OneAMOUint64)
	var f1, f2 big.Float
	f1.SetInt(&c.Int)
	f2.SetFloat64(x)
	f1.Mul(&f1, &f2)
	f1.Int(&c.Int)
	return c
}

func (c *Currency) SetString(x string, base int) (*Currency, error) {
	i, ok := c.Int.SetString(x, base)
	if !ok {
		return nil, fmt.Errorf("Fail to convert hex string(%v)", x)
	}
	if isTooBig(i) {
		return nil, fmt.Errorf("Currency supports up to 32 bytes;%v", x)
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
			"Currency should be represented as double-quoted string(hex:" +
				hex.EncodeToString(data) +
				",str:" +
				s +
				").")
	}
	*c = Currency{}
	s = s[1 : len(s)-1]
	if len(s) == 0 {
		return nil
	}
	_, err := c.SetString(s, 10)
	if c.Int.Cmp(&maxCurrency) == 1 {
		return fmt.Errorf("Currency supports up to 32 bytes(%v)", s)
	}
	return err
}

func (c *Currency) Add(a *Currency) *Currency {
	// XXX Well.. This is a problem.
	if isTooBig(c.Int.Add(&c.Int, &a.Int)) {
		c.Int.Set(&maxCurrency)
	}
	return c
}

func (c *Currency) Sub(a *Currency) *Currency {
	c.Int.Sub(&c.Int, &a.Int)
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
