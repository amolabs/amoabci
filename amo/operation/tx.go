package operation

import (
	"crypto/elliptic"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/crypto/p256"
)

const (
	TxTransfer = "transfer"
	TxRegister = "register"
	TxRequest  = "request"
	TxCancel   = "cancel"
	TxGrant    = "grant"
	TxRevoke   = "revoke"
	TxDiscard  = "discard"
)

const (
	nonceSize = 4
)

var (
	c = elliptic.P256()
)

type Signature struct {
	PubKey   p256.PubKeyP256 `json:"pubkey"`
	SigBytes cmn.HexBytes    `json:"sig_bytes"`
}

func (s Signature) IsValid() bool {
	return len(s.SigBytes) == p256.SignatureSize &&
		c.IsOnCurve(new(big.Int).SetBytes(s.PubKey[1:33]), new(big.Int).SetBytes(s.PubKey[33:]))
}

type Message struct {
	Type   string          `json:"type"`
	Sender crypto.Address  `json:"sender"`
	Nonce  cmn.HexBytes    `json:"nonce"`
	Params json.RawMessage `json:"param"`
	Sig    Signature `json:"signature"`
}

func (m Message) GetSigningBytes() []byte {
	m.Sig.SigBytes = nil
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func (m *Message) Sign(privKey crypto.PrivKey) error {
	pubKey := privKey.PubKey()
	p256PubKey, ok := pubKey.(p256.PubKeyP256)
	if !ok {
		return cmn.NewError("Fail to convert public key to p256 public key")
	}
	m.Nonce = cmn.RandBytes(nonceSize)
	m.Sender = pubKey.Address()
	sigJson := Signature {
		PubKey: p256PubKey,
	}
	m.Sig = sigJson
	sb := m.GetSigningBytes()
	sig, err := privKey.Sign(sb)
	if err != nil {
		return err
	}
	sigJson.SigBytes = make([]byte, p256.SignatureSize)
	sigLen := copy(sigJson.SigBytes, sig)
	if sigLen != p256.SignatureSize {
		return cmn.NewError("Fail to sign")
	}
	m.Sig = sigJson
	return nil
}

func (m *Message) Verify() bool {
	if len(m.Sig.SigBytes) != p256.SignatureSize {
		return false
	}
	sb := m.GetSigningBytes()
	return m.Sig.PubKey.VerifyBytes(sb, m.Sig.SigBytes)
}

func (m Message) IsValid() bool {
	if len(m.Nonce) != nonceSize {
		return false
	}
	return true
}

type Operation interface {
	Check(store *store.Store, sender crypto.Address) uint32
	Execute(store *store.Store, sender crypto.Address) (uint32, []cmn.KVPair)
}

func ParseTx(tx []byte) (Message, Operation) {
	var message Message

	err := json.Unmarshal(tx, &message)
	if err != nil {
		panic(err)
	}

	message.Type = strings.ToLower(message.Type)
	var payload interface{}
	switch message.Type {
	case TxTransfer:
		payload = new(Transfer)
	case TxRegister:
		payload = new(Register)
	case TxRequest:
		payload = new(Request)
	case TxCancel:
		payload = new(Cancel)
	case TxGrant:
		payload = new(Grant)
	case TxRevoke:
		payload = new(Revoke)
	case TxDiscard:
		payload = new(Discard)
	default:
		panic(cmn.NewError("Invalid operation type: %v", message.Type))
	}

	err = json.Unmarshal(message.Params, &payload)
	if err != nil {
		panic(err)
	}

	return message, payload.(Operation)
}
