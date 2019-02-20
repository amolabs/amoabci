package p256

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"github.com/tendermint/go-amino"
	tmc "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"io"
	"math/big"
)

type PrivKeyP256 [32]byte
type PubKeyP256 [65]byte

var (
	c = elliptic.P256()
	h = tmc.Sha256
)

const (
	PrivKeyAminoName = "amo/PrivKeyP256"
	PubKeyAminoName = "amo/PubKeyP256"
)

func RegisterAmino(cdc *amino.Codec) {
	cdc.RegisterConcrete(PrivKeyP256{}, PrivKeyAminoName, nil)
	cdc.RegisterConcrete(PubKeyP256{}, PubKeyAminoName, nil)
}

func GenPrivKeyFromSecret(secret []byte) PrivKeyP256 {
	privKey32 := h(secret)
	priv := PrivKeyP256{}
	copy(priv[:], privKey32)
	return priv
}

func GenPrivKey() PrivKeyP256 {
	return genPrivKey(tmc.CReader())
}

func genPrivKey(rand io.Reader) PrivKeyP256 {
	p256, err := ecdsa.GenerateKey(c, rand)
	if err != nil {
		panic(err)
	}
	var privKey PrivKeyP256
	copy(privKey[:], p256.D.Bytes())
	return privKey
}

func (privKey PrivKeyP256) Bytes() []byte {
	return privKey[:]
}

func (privKey PrivKeyP256) ToECDSA() *ecdsa.PrivateKey {
	X, Y := c.ScalarBaseMult(privKey[:])
	return &ecdsa.PrivateKey{
		D: new(big.Int).SetBytes(privKey[:]),
		PublicKey: ecdsa.PublicKey{
			Curve: c,
			X: X,
			Y: Y,
		},
	}
}

func (privKey PrivKeyP256) Sign(msg []byte) ([]byte, error) {
	priv :=  privKey.ToECDSA()
	r, s, err := ecdsa.Sign(tmc.CReader(), priv, h(msg))
	if err != nil {
		return nil, err
	}
	rb := r.Bytes()
	sb := s.Bytes()
	sig := make([]byte, 0, len(rb)+len(sb))
	sig = append(sig, rb...)
	sig = append(sig, sb...)
	// concat r, s
	return sig, nil
}

func (privKey PrivKeyP256) PubKey() tmc.PubKey {
	priv := privKey.ToECDSA()
	pubKey := PubKeyP256{0x04}
	copy(pubKey[1:], priv.X.Bytes())
	copy(pubKey[33:], priv.Y.Bytes())
	return pubKey
}

func (privKey PrivKeyP256) Equals(other tmc.PrivKey) bool {
	return bytes.Equal(privKey[:], other.Bytes())
}

func (privKey *PrivKeyP256) SetBytes(buf []byte) {
	copy(privKey[:], buf)
}

func (privKey PrivKeyP256) String() string {
	return hex.EncodeToString(privKey[:])
}

func (pubKey PubKeyP256) Address() tmc.Address {
	return tmc.Address(tmhash.SumTruncated(pubKey[:]))
}

func (pubKey PubKeyP256) Bytes() []byte {
	return pubKey[:]
}

func (pubKey PubKeyP256) ToECDSA() *ecdsa.PublicKey {
	return &ecdsa.PublicKey{
		Curve: c,
		X: new(big.Int).SetBytes(pubKey[1:33]),
		Y: new(big.Int).SetBytes(pubKey[33:]),
	}
}

func (pubKey PubKeyP256) VerifyBytes(msg []byte, sig []byte) bool {
	if len(sig) != 64 {
		return false
	}
	return ecdsa.Verify(pubKey.ToECDSA(), h(msg), new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:]))
}

func (pubKey PubKeyP256) Equals(other tmc.PubKey) bool {
	return bytes.Equal(pubKey[:], other.Bytes())
}

func (pubKey PubKeyP256) String() string {
	return hex.EncodeToString(pubKey[:])
}

var _ tmc.PrivKey = PrivKeyP256{}
var _ tmc.PubKey = PubKeyP256{}
