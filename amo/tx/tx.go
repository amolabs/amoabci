package tx

import (
	"crypto/elliptic"
	"encoding/json"
	"math/big"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
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

type Tx interface {
	GetType() string
	GetSender() crypto.Address
	getNonce() tm.HexBytes
	getPayload() json.RawMessage
	getSignature() Signature
	getSigningBytes() []byte

	Sign(privKey crypto.PrivKey) error
	Verify() bool
	Check() (uint32, string)
	Execute(store *store.Store) (uint32, string, []tm.KVPair)
}

var _ Tx = &TxBase{}

type TxBase struct {
	Type      string          `json:"type"`
	Sender    crypto.Address  `json:"sender"`
	Nonce     tm.HexBytes     `json:"nonce"`
	Payload   json.RawMessage `json:"payload"` // TODO: change to txparam
	Signature Signature       `json:"signature"`
}

type TxToSign struct {
	Type      string          `json:"type"`
	Sender    crypto.Address  `json:"sender"`
	Nonce     tm.HexBytes     `json:"nonce"`
	Payload   json.RawMessage `json:"payload"`
	Signature Signature       `json:"-"`
}

func classifyTx(base TxBase) Tx {
	var t Tx
	switch base.Type {
	case "transfer":
		param, _ := parseTransferParam(base.Payload)
		t = &TxTransfer{
			TxBase: base,
			Param:  param,
		}
	default:
		t = &base
	}
	return t
}

func ParseTx(txBytes []byte) (Tx, error) {
	var base TxBase

	err := json.Unmarshal(txBytes, &base)
	if err != nil {
		return nil, err
	}

	return classifyTx(base), nil
}

// accessors

func (t *TxBase) GetType() string {
	return t.Type
}

func (t *TxBase) GetSender() crypto.Address {
	return t.Sender
}

func (t *TxBase) getNonce() tm.HexBytes {
	return t.Nonce
}

func (t *TxBase) getPayload() json.RawMessage {
	return t.Payload
}

func (t *TxBase) getSignature() Signature {
	return t.Signature
}

func (t *TxBase) getSigningBytes() []byte {
	var tts TxToSign = TxToSign(*t)
	b, _ := json.Marshal(tts)
	/* XXX: nothing to do here
	if err != nil {
		return b
	}
	*/
	return b
}

func (t *TxBase) Sign(privKey crypto.PrivKey) error {
	pubKey := privKey.PubKey()
	p256PubKey, ok := pubKey.(p256.PubKeyP256)
	if !ok {
		return tm.NewError("Fail to convert public key to p256 public key")
	}
	sb := t.getSigningBytes()
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

func (t *TxBase) Verify() bool {
	if len(t.Signature.SigBytes) != p256.SignatureSize {
		return false
	}
	sb := t.getSigningBytes()
	return t.Signature.PubKey.VerifyBytes(sb, t.Signature.SigBytes)
}

func (t *TxBase) Check() (uint32, string) {
	var rc uint32
	var info string

	switch t.Type {
	case "stake":
		rc, info = CheckStake(t)
	case "withdraw":
		rc, info = CheckWithdraw(t)
	case "delegate":
		rc, info = CheckDelegate(t)
	case "retract":
		rc, info = CheckRetract(t)
	case "register":
		rc, info = CheckRegister(t)
	case "request":
		rc, info = CheckRequest(t)
	case "cancel":
		rc, info = CheckCancel(t)
	case "grant":
		rc, info = CheckGrant(t)
	case "revoke":
		rc, info = CheckRevoke(t)
	case "discard":
		rc, info = CheckDiscard(t)
	default:
		rc = code.TxCodeUnknown
		info = "unknown transaction type"
	}
	return rc, info
}

func (t *TxBase) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	var rc uint32
	var info string
	var tags []tm.KVPair
	switch t.Type {
	case "stake":
		rc, info, tags = ExecuteStake(t, store)
	case "withdraw":
		rc, info, tags = ExecuteWithdraw(t, store)
	case "delegate":
		rc, info, tags = ExecuteDelegate(t, store)
	case "retract":
		rc, info, tags = ExecuteRetract(t, store)
	case "register":
		rc, info, tags = ExecuteRegister(t, store)
	case "request":
		rc, info, tags = ExecuteRequest(t, store)
	case "cancel":
		rc, info, tags = ExecuteCancel(t, store)
	case "grant":
		rc, info, tags = ExecuteGrant(t, store)
	case "revoke":
		rc, info, tags = ExecuteRevoke(t, store)
	case "discard":
		rc, info, tags = ExecuteDiscard(t, store)
	default:
		rc = code.TxCodeUnknown
		info = "unknown transaction type"
		tags = nil
	}

	return rc, info, tags
}
