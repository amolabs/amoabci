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
	TxStake    = "stake"
	TxWithdraw = "withdraw"
	TxDelegate = "delegate"
	TxRetract  = "retract"
)

const (
	NonceSize = 4
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
	Type      string          `json:"type"`
	Sender    crypto.Address  `json:"sender"`
	Nonce     cmn.HexBytes    `json:"nonce"`
	Payload   json.RawMessage `json:"payload"`
	Signature Signature       `json:"signature"`
}

type MessageToSign struct {
	Type    string          `json:"type"`
	Sender  crypto.Address  `json:"sender"`
	Nonce   cmn.HexBytes    `json:"nonce"`
	Payload json.RawMessage `json:"payload"`
}

func (m Message) GetSigningBytes() []byte {
	mts := MessageToSign{
		Type:    m.Type,
		Sender:  m.Sender,
		Nonce:   m.Nonce,
		Payload: m.Payload,
	}
	b, err := json.Marshal(mts)
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
	sb := m.GetSigningBytes()
	sig, err := privKey.Sign(sb)
	if err != nil {
		return err
	}
	sigJson := Signature{
		PubKey:   p256PubKey,
		SigBytes: sig,
	}
	m.Signature = sigJson
	return nil
}

func (m *Message) Verify() bool {
	if len(m.Signature.SigBytes) != p256.SignatureSize {
		return false
	}
	sb := m.GetSigningBytes()
	return m.Signature.PubKey.VerifyBytes(sb, m.Signature.SigBytes)
}

func (m Message) IsValid() bool {
	if len(m.Nonce) != NonceSize {
		return false
	}
	return true
}

type Operation interface {
	Check(store *store.Store, sender crypto.Address) uint32
	Execute(store *store.Store, sender crypto.Address) uint32
}

func ParseTx(tx []byte) (Message, Operation, bool) {
	var message Message

	err := json.Unmarshal(tx, &message)
	if err != nil {
		panic(err)
	}

	isStake := false
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
	case TxStake:
		payload = new(Stake)
		isStake = true
	case TxWithdraw:
		payload = new(Withdraw)
		isStake = true
	case TxDelegate:
		payload = new(Delegate)
		isStake = true
	case TxRetract:
		payload = new(Retract)
		isStake = true
	default:
		panic(cmn.NewError("Invalid operation type: %v", message.Type))
	}

	err = json.Unmarshal(message.Payload, &payload)
	if err != nil {
		panic(err)
	}

	return message, payload.(Operation), isStake
}
