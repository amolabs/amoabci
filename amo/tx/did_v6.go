package tx

import (
	//"encoding/binary"
	//"encoding/json"
	"encoding/hex"
	"encoding/json"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

//// did.claim

type DIDClaimParam struct {
	Target   string   `json:"target"`
	Document Document `json:"document"`
}

type Document struct {
	Context            string               `json:"@context,omitempty"`
	Id                 string               `json:"id"`
	Controller         string               `json:"controller,omitempty"`
	VerificationMethod []VerificationMethod `json:"verificationMethod"`
	Authentication     string               `json:"authentication"`
	AssertionMethod    string               `json:"assertionMethod,omitempty"`
}

type VerificationMethod struct {
	Id           string       `json:"id"`
	Type         string       `json:"type"`
	Controller   string       `json:"controller,omitempty"`
	PublicKeyJwk PublicKeyJwk `json:"publicKeyJwk"`
}

type PublicKeyJwk struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

func parseDIDClaimParam(raw []byte) (DIDClaimParam, error) {
	var param DIDClaimParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxDIDClaim struct {
	TxBase
	Param DIDClaimParam `json:"-"`
}

var _ Tx = &TxDIDClaim{}

func (t *TxDIDClaim) Check() (uint32, string) {
	param, err := parseDIDClaimParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	// stateless validity check
	ss := strings.Split(param.Target, ":")
	if len(ss) != 3 || ss[0] != "did" || ss[1] != "amo" || len(ss[2]) != 40 {
		return code.TxCodeBadParam, "invalid target did"
	}
	_, err = hex.DecodeString(ss[2])
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}
	if param.Target != param.Document.Id {
		return code.TxCodeBadParam, "mismatching did"
	}

	if len(param.Document.VerificationMethod) == 0 {
		return code.TxCodeBadParam, "no verificationMethod"
	}
	if len(param.Document.Authentication) == 0 {
		return code.TxCodeBadParam, "no authentication"
	}
	if param.Document.Authentication != param.Document.VerificationMethod[0].Id {
		return code.TxCodeBadParam, "unknown verificationMethod for authentication"
	}

	return code.TxCodeOK, "ok"
}

func (t *TxDIDClaim) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseDIDClaimParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	newDoc := txParam.Document
	entry := store.GetDIDEntry(txParam.Target, false)
	senderDID := "did:amo:" + t.GetSender().String()

	if entry != nil {
		var doc Document
		err = json.Unmarshal(entry.Document, &doc)
		if err != nil {
			return code.TxCodeUnknown, "failed to unmarshal document", nil
		}
		if senderDID != doc.Id && senderDID != doc.Controller {
			return code.TxCodePermissionDenied, "permission denied", nil
		}
	} else {
		if senderDID != newDoc.Id {
			return code.TxCodePermissionDenied, "permission denied", nil
		}
	}

	b, err := json.Marshal(newDoc)
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}
	entry = &types.DIDEntry{
		Document: b,
	}
	store.SetDIDEntry(txParam.Target, entry)

	return code.TxCodeOK, "ok", []abci.Event{}
}

//// did.dismiss

type DIDDismissParam struct {
	Target string `json:"target"`
}

// NOTE: this is essentially the same as parseDismissParam()
func parseDIDDismissParam(raw []byte) (DIDDismissParam, error) {
	var param DIDDismissParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxDIDDismiss struct {
	TxBase
	Param DIDDismissParam `json:"-"`
}

var _ Tx = &TxDIDDismiss{}

func (t *TxDIDDismiss) Check() (uint32, string) {
	param, err := parseDIDDismissParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	// stateless validity check
	ss := strings.Split(param.Target, ":")
	if len(ss) != 3 || ss[0] != "did" || ss[1] != "amo" || len(ss[2]) != 40 {
		return code.TxCodeBadParam, "invalid target did"
	}
	_, err = hex.DecodeString(ss[2])
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxDIDDismiss) Execute(store *store.Store) (uint32, string, []abci.Event) {
	txParam, err := parseDIDClaimParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error(), nil
	}

	entry := store.GetDIDEntry(txParam.Target, false)
	senderDID := "did:amo:" + t.GetSender().String()

	if entry != nil {
		var doc Document
		err = json.Unmarshal(entry.Document, &doc)
		if err != nil {
			return code.TxCodeUnknown, "failed to unmarshal document", nil
		}
		if senderDID != doc.Id && senderDID != doc.Controller {
			return code.TxCodePermissionDenied, "permission denied", nil
		}
	} else {
		return code.TxCodeNotFound, "not found", nil
	}

	store.DeleteDIDEntry(txParam.Target)

	return code.TxCodeOK, "ok", []abci.Event{}
}
