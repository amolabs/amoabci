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
	Payload   json.RawMessage `json:"payload"`
	Signature Signature       `json:"signature"`
}

type TxToSign struct {
	Type    string          `json:"type"`
	Sender  crypto.Address  `json:"sender"`
	Nonce   tm.HexBytes     `json:"nonce"`
	Payload json.RawMessage `json:"payload"`
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
		panic(err)
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

// TODO: too clumsy prototype. improve it
func ParseTx(txBytes []byte) (Tx, Operation, bool, error) {
	var t Tx

	err := json.Unmarshal(txBytes, &t)
	if err != nil {
		return t, nil, false, err
	}

	isStake := false
	t.Type = strings.ToLower(t.Type)
	var payload interface{}
	switch t.Type {
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
		return t, nil, false, tm.NewError("Invalid operation type: %v", t.Type)
	}

	err = json.Unmarshal(t.Payload, &payload)
	if err != nil {
		return t, nil, false, err
	}

	return t, payload.(Operation), isStake, nil
}
