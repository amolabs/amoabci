package p256

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"github.com/tendermint/btcd/btcec"
	tmc "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"io"
	"math/big"
)

type PrivKeyP256 [32]byte
type PubKeyP256 [65]byte

func GenPrivKeyFromSecret(secret []byte) PrivKeyP256 {
	privKey32 := sha256.Sum256(secret)
	return PrivKeyP256(privKey32)
}

func GenPrivKey() PrivKeyP256 {
	return genPrivKey(tmc.CReader())
}

func genPrivKey(rand io.Reader) PrivKeyP256 {
	p256, err := ecdsa.GenerateKey(elliptic.P256(), rand)
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

func (privKey PrivKeyP256) Sign(msg []byte) ([]byte, error) {
	priv, _ := btcec.PrivKeyFromBytes(elliptic.P256(), privKey[:])
	sig, err := priv.Sign(tmc.Sha256(msg))
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

func (privKey PrivKeyP256) PubKey() tmc.PubKey {
	_, pub := btcec.PrivKeyFromBytes(elliptic.P256(), privKey[:])
	pubKey := PubKeyP256{}
	copy(pubKey[:], pub.SerializeUncompressed())
	return pubKey
}

func (privKey PrivKeyP256) Equals(other tmc.PrivKey) bool {
	return bytes.Equal(privKey[:], other.Bytes())
}

func (pubKey PubKeyP256) Address() tmc.Address {
	return tmc.Address(tmhash.SumTruncated(pubKey[:]))
}

func (pubKey PubKeyP256) Bytes() []byte {
	return pubKey[:]
}

func (pubKey PubKeyP256) VerifyBytes(msg []byte, sig []byte) bool {
	var pub = btcec.PublicKey{}
	pub.Curve = elliptic.P256()
	pub.X = new(big.Int).SetBytes(pubKey[1:33])
	pub.Y = new(big.Int).SetBytes(pubKey[33:])
	parsedSig, err := btcec.ParseSignature(sig[:], elliptic.P256())
	if err != nil {
		return false
	}
	return parsedSig.Verify(tmc.Sha256(msg), &pub)
}

func (pubKey PubKeyP256) Equals(other tmc.PubKey) bool {
	return bytes.Equal(pubKey[:], other.Bytes())
}

var _ tmc.PrivKey = PrivKeyP256{}
var _ tmc.PubKey = PubKeyP256{}

// TEST CODE
func genPribKeyP256FromHexString(hs string) PrivKeyP256 {
	b, err := hex.DecodeString(hs)
	if err != nil {
		panic(err)
	}
	privKey := PrivKeyP256{}
	copy(privKey[:], b)
	return privKey
}