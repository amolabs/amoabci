package tx

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/crypto/p256"
)

func makeTestTxV6(txType string, seed string, payload []byte) Tx {
	privKey := p256.GenPrivKeyFromSecret([]byte(seed))
	addr := privKey.PubKey().Address()
	trans := TxBase{
		Type:    txType,
		Sender:  addr,
		Payload: payload,
	}
	trans.Sign(privKey)
	return classifyTxV6(trans)
}

// legacy
func TestTxClaim(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	entry := s.GetDIDEntry("myid", false)
	assert.Nil(t, entry)

	// first claim
	payload, _ := json.Marshal(ClaimParam{
		Target:   "myid",
		Document: []byte(`{}`),
	})
	t1 := makeTestTx("claim", "sender", payload)
	rc, info := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry("myid", false)
	assert.NotNil(t, entry)
	assert.True(t, bytes.Equal(makeAccAddr("sender"), entry.Owner))
	assert.True(t, bytes.Equal([]byte(`{}`), entry.Document))

	// update claim
	payload, _ = json.Marshal(ClaimParam{
		Target:   "myid",
		Document: []byte(`{"haha": "hoho"}`),
	})
	t2 := makeTestTx("claim", "sender", payload)
	rc, info, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry("myid", false)
	assert.NotNil(t, entry)
	assert.True(t, bytes.Equal(makeAccAddr("sender"), entry.Owner))
	// XXX note that retrieved document is a compact representation
	assert.True(t, bytes.Equal([]byte(`{"haha":"hoho"}`), entry.Document))

	// dsmiss
	payload, _ = json.Marshal(DismissParam{
		Target: "myid",
	})
	t3 := makeTestTx("dismiss", "sender", payload)
	rc, info = t3.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)
	rc, info, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry("myid", false)
	assert.Nil(t, entry)
}

func TestTxDIDClaim(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	myid := "did:amo:70EAD5B53B11DFE78EC8CF131D7960F097D48D70"
	mydoc := Document{
		Id: myid,
	}
	mydocJson, _ := json.Marshal(mydoc)

	// tx check error
	payload, _ := json.Marshal(DIDClaimParam{
		// invalid AMO DID format
		Target:   "did:amo:Z0EAD5B53B11DFE78EC8CF131D7960F097D48D70",
		Document: Document{},
	})
	t1 := makeTestTxV6("did.claim", "controller", payload)
	rc, info := t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "invalid byte")

	// tx check error (mismatching did)
	payload, _ = json.Marshal(DIDClaimParam{
		Target:   "did:amo:70EAD5B53B11DFE78EC8CF131D7960F097D48D70",
		Document: Document{},
	})
	t1 = makeTestTxV6("did.claim", "controller", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Equal(t, "mismatching did", info)

	// adjust test data
	myid = "did:amo:" + makeTestAddress("subject").String()
	mydoc.Id = myid
	mydocJson, _ = json.Marshal(mydoc)

	// tx check error (check verificationMethod)
	payload, _ = json.Marshal(DIDClaimParam{
		Target:   myid,
		Document: mydoc,
	})
	t1 = makeTestTxV6("did.claim", "controller", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Equal(t, "no verificationMethod", info)

	// adjust test data
	mydoc.VerificationMethod = []VerificationMethod{{
		Id:   "asdf#keys-1",
		Type: "jsonWebKey",
		PublicKeyJwk: PublicKeyJwk{
			Kty: "EC",
			Crv: "P-256",
			X:   "FFFF",
			Y:   "EEEE",
		},
	}}

	// tx check error (no authentication)
	payload, _ = json.Marshal(DIDClaimParam{
		Target:   myid,
		Document: mydoc,
	})
	t1 = makeTestTxV6("did.claim", "controller", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "no authentication")

	// adjust test data
	mydoc.Authentication = []json.RawMessage{[]byte(`"unknown-key"`)}

	// tx check error (check authentication)
	payload, _ = json.Marshal(DIDClaimParam{
		Target:   myid,
		Document: mydoc,
	})
	t1 = makeTestTxV6("did.claim", "controller", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "verificationMethod")
	assert.Contains(t, info, "not in authentication")

	// adjust test data
	mydoc.Authentication = []json.RawMessage{[]byte(`"asdf#keys-1"`)}
	controllerId := "did:amo:" + makeTestAddress("controller").String()
	mydoc.Controller = controllerId
	mydocJson, _ = json.Marshal(mydoc)

	// tx check ok
	payload, _ = json.Marshal(DIDClaimParam{
		Target:   myid,
		Document: mydoc,
	})
	t1 = makeTestTxV6("did.claim", "controller", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	// check nil before execute. This make the next claim tx will be for
	// a previously non-existing document.
	entry := s.GetDIDEntry(myid, false)
	assert.Nil(t, entry)

	// claim execute error
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)
	assert.Equal(t, "permission denied", info)

	// first claim
	payload, _ = json.Marshal(DIDClaimParam{
		Target:   myid,
		Document: mydoc,
	})
	// now tx from the ligitimate subject
	t1 = makeTestTxV6("did.claim", "subject", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry(myid, false)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Owner) // in protocl v6, entry.Owner becomes obsolete
	assert.True(t, bytes.Equal(mydocJson, entry.Document))
	var doc Document
	_ = json.Unmarshal(entry.Document, &doc)
	assert.Equal(t, controllerId, doc.Controller)

	// update claim
	mydoc.Controller = ""
	mydocJson, _ = json.Marshal(mydoc)
	payload, _ = json.Marshal(DIDClaimParam{
		Target:   myid,
		Document: mydoc,
	})
	t2 := makeTestTxV6("did.claim", "controller", payload)
	rc, info, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry(myid, false)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Owner) // in protocl v6, entry.Owner becomes obsolete
	assert.True(t, bytes.Equal(mydocJson, entry.Document))

	// Now that controller property is null, further update from controller
	// will fail.
	payload, _ = json.Marshal(DIDClaimParam{
		Target:   myid,
		Document: mydoc,
	})
	t2 = makeTestTxV6("did.claim", "controller", payload)
	rc, info, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)
	assert.Equal(t, "permission denied", info)

	// tx check error
	payload, _ = json.Marshal(DIDDismissParam{
		// invalid AMO DID format
		Target: "did:amo:Z0EAD5B53B11DFE78EC8CF131D7960F097D48D70",
	})
	t1 = makeTestTxV6("did.dismiss", "controller", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "invalid byte")

	// dsmiss execute error
	payload, _ = json.Marshal(DIDDismissParam{
		Target: myid,
	})
	t3 := makeTestTxV6("did.dismiss", "controller", payload)
	rc, info = t3.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)
	rc, info, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)
	assert.Equal(t, "permission denied", info)

	// dismiss
	payload, _ = json.Marshal(DIDDismissParam{
		Target: myid,
	})
	t3 = makeTestTxV6("did.dismiss", "subject", payload)
	rc, info = t3.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)
	rc, info, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	// claim again with controller info
	mydoc.Controller = controllerId
	payload, _ = json.Marshal(DIDClaimParam{
		Target:   myid,
		Document: mydoc,
	})
	t1 = makeTestTxV6("did.claim", "subject", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	// dismiss from controller
	payload, _ = json.Marshal(DIDDismissParam{
		Target: myid,
	})
	t3 = makeTestTxV6("did.dismiss", "controller", payload)
	rc, info = t3.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)
	rc, info, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry(myid, false)
	assert.Nil(t, entry)
}

func TestTxDIDIssue(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	vcid := "amo:cred:70EAD5B53B11DFE78EC8CF131D7960F097D48D7047381938ABC34812DE849023"
	issuerId := "did:amo:" + makeTestAddress("issuer").String()

	// stateless validity check
	invalidVC := `{"id":"amo:cred:80EAD5B53B11DFE78EC8CF131D7960F097D48D70"}`
	payload, _ := json.Marshal(DIDIssueParam{
		Target:     vcid[:64],
		Credential: []byte(invalidVC),
	})
	t1 := makeTestTxV6("did.issue", "issuer", payload)
	rc, info := t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "invalid target VC id")

	payload, _ = json.Marshal(DIDIssueParam{
		Target:     vcid,
		Credential: []byte(invalidVC),
	})
	t1 = makeTestTxV6("did.issue", "issuer", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "mismatching VC id")

	invalidVC = `{"id":"amo:cred:70EAD5B53B11DFE78EC8CF131D7960F097D48D7047381938ABC34812DE849023","issuer":"iiii"}`
	payload, _ = json.Marshal(DIDIssueParam{
		Target:     vcid,
		Credential: []byte(invalidVC),
	})
	t1 = makeTestTxV6("did.issue", "issuer", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "mismatching VC issuer")

	invalidVC = `{"id":"amo:cred:70EAD5B53B11DFE78EC8CF131D7960F097D48D7047381938ABC34812DE849023","issuer":"` + issuerId + `"}`
	payload, _ = json.Marshal(DIDIssueParam{
		Target:     vcid,
		Credential: []byte(invalidVC),
	})
	t1 = makeTestTxV6("did.issue", "issuer", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "missing property: issued")

	validVC := `{"id":"amo:cred:70EAD5B53B11DFE78EC8CF131D7960F097D48D7047381938ABC34812DE849023","issuer":"` + issuerId + `","issued":"ISO date string"}`
	payload, _ = json.Marshal(DIDIssueParam{
		Target:     vcid,
		Credential: []byte(validVC),
	})
	t1 = makeTestTxV6("did.issue", "issuer", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")

	// ensure empty
	entry := s.GetVCEntry(vcid, false)
	assert.Nil(t, entry)

	// valid issue
	vc := struct {
		Id     string `json:"id"`
		Issuer string `json:"issuer"`
		Issued string `json:"issued"`
		Any1   string `json:"any_1"`
		Any2   int64  `json:"any_2"`
	}{
		Id:     vcid,
		Issuer: issuerId,
		Issued: "ISO date string",
		Any1:   "any1",
		Any2:   123,
	}
	vcJson, _ := json.Marshal(vc)

	payload, _ = json.Marshal(DIDIssueParam{
		Target:     vcid,
		Credential: vcJson,
	})
	t1 = makeTestTxV6("did.issue", "issuer", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")

	entry = s.GetVCEntry(vcid, false)
	assert.NotNil(t, entry)
	assert.True(t, bytes.Equal(vcJson, entry.Credential))

	// update vc (permission denied)
	vc = struct {
		Id     string `json:"id"`
		Issuer string `json:"issuer"`
		Issued string `json:"issued"`
		Any1   string `json:"any_1"`
		Any2   int64  `json:"any_2"`
	}{
		Id:     vcid,
		Issuer: "did:amo:" + makeTestAddress("noissuer").String(),
		Issued: "ISO date string",
		Any1:   "any1",
		Any2:   123,
	}
	vcJson, _ = json.Marshal(vc)

	payload, _ = json.Marshal(DIDIssueParam{
		Target:     vcid,
		Credential: vcJson,
	})
	t1 = makeTestTxV6("did.issue", "noissuer", payload) // permission denied
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)
	assert.Contains(t, info, "permission denied")

	// update vc
	vc = struct {
		Id     string `json:"id"`
		Issuer string `json:"issuer"`
		Issued string `json:"issued"`
		Any1   string `json:"any_1"`
		Any2   int64  `json:"any_2"`
	}{
		Id:     vcid,
		Issuer: issuerId,
		Issued: "ISO date string",
		Any1:   "any1",
		Any2:   123,
	}
	vcJson, _ = json.Marshal(vc)

	payload, _ = json.Marshal(DIDIssueParam{
		Target:     vcid,
		Credential: vcJson,
	})
	t1 = makeTestTxV6("did.issue", "issuer", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")

	entry = s.GetVCEntry(vcid, false)
	assert.NotNil(t, entry)
	assert.True(t, bytes.Equal(vcJson, entry.Credential))

	// revoke (validation)
	payload, _ = json.Marshal(DIDRevokeParam{
		Target: "asdf",
	})
	t1 = makeTestTxV6("did.revoke", "issuer", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "invalid target VC id")

	// revoke vc
	payload, _ = json.Marshal(DIDRevokeParam{
		Target: vcid,
	})
	t1 = makeTestTxV6("did.revoke", "noissuer", payload) // permission denied
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)
	assert.Contains(t, info, "permission denied")

	// revoke vc
	payload, _ = json.Marshal(DIDRevokeParam{
		Target: vcid,
	})
	t1 = makeTestTxV6("did.revoke", "issuer", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")

	// revoke vc (not found)
	payload, _ = json.Marshal(DIDRevokeParam{
		Target: vcid,
	})
	t1 = makeTestTxV6("did.revoke", "issuer", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Contains(t, info, "ok")
	rc, info, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNotFound, rc)
	assert.Contains(t, info, "not found")
}
