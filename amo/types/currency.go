package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/big"

	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	maxCurrencyHex    = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
	currencyLen       = 256 / 8
)

// Currency uses big endian for compatibility to big.Int
type Currency [currencyLen]byte

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
	binary.BigEndian.PutUint64(c[currencyLen-8:], x)
	return c
}

func (c *Currency) SetString(x string, base int) (*Currency, error) {
	i, ok := new(big.Int).SetString(x, base)
	if !ok {
		return nil, cmn.NewError("Fail to convert hex string(%v)", x)
	}
	if isExceed(i) {
		return nil, cmn.NewError("Currency supports up to 32 bytes;%v", x)
	}
	b := i.Bytes()
	*c = Currency{}
	for i := 0; i < len(b); i++ {
		c[currencyLen-i-1] = b[len(b)-i-1]
	}
	return c, nil
}

func (c Currency) String() string {
	return new(big.Int).SetBytes(c[:]).Text(10)
}

func (c Currency) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	i := 0
	for ; i < len(c); i++ {
		if c[i] != 0 {
			break
		}
	}
	buf.WriteByte(byte(currencyLen-i))
	buf.Write(c[i:])
	return buf.Bytes(), nil
}

func (c *Currency) Deserialize(data []byte) error {
	for i := 0; i < len(data[1:]); i++ {
		c[currencyLen-data[0]+byte(i)] = data[i+1]
	}
	return nil
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
	x, y := new(big.Int).SetBytes(c[:]), new(big.Int).SetBytes(a[:])
	if isExceed(x.Add(x, y)) {
		panic(cmn.NewError("Cannot add"))
	}
	c.Set(0)
	b := x.Bytes()
	for i := 0; i < len(b); i++ {
		c[currencyLen-i-1] = b[len(b)-i-1]
	}
	return c
}

func (c *Currency) Sub(a *Currency) *Currency {
	x, y := new(big.Int).SetBytes(c[:]), new(big.Int).SetBytes(a[:])
	if x.Sub(x, y).Sign() == -1 {
		panic(cmn.NewError("Cannot subtract"))
	}
	c.Set(0)
	b := x.Bytes()
	for i := 0; i < len(b); i++ {
		c[currencyLen-i-1] = b[len(b)-i-1]
	}
	return c
}

func (c Currency) Equals(a *Currency) bool {
	return bytes.Equal(c[:], a[:])
}

func (c Currency) GreaterThan(a *Currency) bool {
	x, y := new(big.Int).SetBytes(c[:]), new(big.Int).SetBytes(a[:])
	return x.Cmp(y) == 1
}

func (c Currency) LessThan(a *Currency) bool {
	x, y := new(big.Int).SetBytes(c[:]), new(big.Int).SetBytes(a[:])
	return x.Cmp(y) == -1
}
