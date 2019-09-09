package tx

import (
	"crypto/elliptic"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/crypto/p256"
)

const (
	NonceSize = 4
)

var (
	c = elliptic.P256()
)

type Signature struct {
	PubKey   p256.PubKeyP256 `json:"pubkey"`
	SigBytes tm.HexBytes     `json:"sig_bytes"`
}

func (s Signature) IsValid() bool {
	return len(s.SigBytes) == p256.SignatureSize &&
		c.IsOnCurve(new(big.Int).SetBytes(s.PubKey[1:33]), new(big.Int).SetBytes(s.PubKey[33:]))
}

type Tx struct {
	Type      string          `json:"type"`
	Sender    crypto.Address  `json:"sender"`
	Nonce     tm.HexBytes     `json:"nonce"`
	Payload   json.RawMessage `json:"payload"` // TODO: change to txparam
	Signature Signature       `json:"signature"`
}

type TxToSign struct {
	Type    string          `json:"type"`
	Sender  crypto.Address  `json:"sender"`
	Nonce   tm.HexBytes     `json:"nonce"`
	Payload json.RawMessage `json:"payload"`
}

func ParseTx(txBytes []byte) (Tx, error) {
	var t Tx

	err := json.Unmarshal(txBytes, &t)
	if err != nil {
		return t, err
	}

	t.Type = strings.ToLower(t.Type)

	return t, nil
}

func (t Tx) GetSigningBytes() []byte {
	tts := TxToSign{
		Type:    t.Type,
		Sender:  t.Sender,
		Nonce:   t.Nonce,
		Payload: t.Payload,
	}
	b, err := json.Marshal(tts)
	if err != nil {
		// XXX: nothing to do here
		return b
	}
	return b
}

func (t *Tx) Sign(privKey crypto.PrivKey) error {
	pubKey := privKey.PubKey()
	p256PubKey, ok := pubKey.(p256.PubKeyP256)
	if !ok {
		return tm.NewError("Fail to convert public key to p256 public key")
	}
	sb := t.GetSigningBytes()
	sig, err := privKey.Sign(sb)
	if err != nil {
		return err
	}
	sigJson := Signature{
		PubKey:   p256PubKey,
		SigBytes: sig,
	}
	t.Signature = sigJson
	return nil
}

func (t *Tx) Verify() bool {
	if len(t.Signature.SigBytes) != p256.SignatureSize {
		return false
	}
	sb := t.GetSigningBytes()
	return t.Signature.PubKey.VerifyBytes(sb, t.Signature.SigBytes)
}

// TODO: not used any more. remove this
func (t Tx) IsValid() bool {
	if len(t.Nonce) != NonceSize {
		return false
	}
	return true
}

type Operation interface {
	Check(store *store.Store, sender crypto.Address) uint32
	Execute(store *store.Store, sender crypto.Address) (uint32, []tm.KVPair)
}
