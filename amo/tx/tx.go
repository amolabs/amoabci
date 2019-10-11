package tx

import (
	"crypto/elliptic"
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

const (
	NonceSize           = 4
	defaultLockupPeriod = uint64(1000000)
)

var (
	c = elliptic.P256()
)

var (
	// config values from the app
	ConfigLockupPeriod = defaultLockupPeriod
)

type Signature struct {
	PubKey   p256.PubKeyP256 `json:"pubkey"`
	SigBytes tm.HexBytes     `json:"sig_bytes"`
}

type Tx interface {
	// accessors
	GetType() string
	GetSender() crypto.Address
	GetFee() types.Currency
	getNonce() tm.HexBytes
	getPayload() json.RawMessage
	getSignature() Signature
	getSigningBytes() []byte

	// ops
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
	Fee       types.Currency  `json:"fee"`
	Signature Signature       `json:"signature"`
}

type TxToSign struct {
	Type      string          `json:"type"`
	Sender    crypto.Address  `json:"sender"`
	Nonce     tm.HexBytes     `json:"nonce"`
	Payload   json.RawMessage `json:"payload"`
	Fee       types.Currency  `json:"fee"`
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
	case "stake":
		param, _ := parseStakeParam(base.Payload)
		t = &TxStake{
			TxBase: base,
			Param:  param,
		}
	case "withdraw":
		param, _ := parseWithdrawParam(base.Payload)
		t = &TxWithdraw{
			TxBase: base,
			Param:  param,
		}
	case "delegate":
		param, _ := parseDelegateParam(base.Payload)
		t = &TxDelegate{
			TxBase: base,
			Param:  param,
		}
	case "retract":
		param, _ := parseRetractParam(base.Payload)
		t = &TxRetract{
			TxBase: base,
			Param:  param,
		}
	case "register":
		param, _ := parseRegisterParam(base.Payload)
		t = &TxRegister{
			TxBase: base,
			Param:  param,
		}
	case "discard":
		param, _ := parseDiscardParam(base.Payload)
		t = &TxDiscard{
			TxBase: base,
			Param:  param,
		}
	case "request":
		param, _ := parseRequestParam(base.Payload)
		t = &TxRequest{
			TxBase: base,
			Param:  param,
		}
	case "cancel":
		param, _ := parseCancelParam(base.Payload)
		t = &TxCancel{
			TxBase: base,
			Param:  param,
		}
	case "grant":
		param, _ := parseGrantParam(base.Payload)
		t = &TxGrant{
			TxBase: base,
			Param:  param,
		}
	case "revoke":
		param, _ := parseRevokeParam(base.Payload)
		t = &TxRevoke{
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

func (t *TxBase) GetFee() types.Currency {
	return t.Fee
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

// ops

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
	rc := code.TxCodeUnknown
	info := "unknown transaction type"

	return rc, info
}

func (t *TxBase) Execute(store *store.Store) (uint32, string, []tm.KVPair) {
	rc := code.TxCodeUnknown
	info := "unknown transaction type"
	tags := []tm.KVPair(nil)

	return rc, info, tags
}
