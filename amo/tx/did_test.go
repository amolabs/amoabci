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
	rc, _, _ = t1.Execute(s)
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
	rc, _, _ = t2.Execute(s)
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
	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry("myid", false)
	assert.Nil(t, entry)
}

func TestTxClaimV6(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	entry := s.GetDIDEntry("myid", false)
	assert.Nil(t, entry)

	// tx check error
	payload, _ := json.Marshal(ClaimParam{
		// invalid AMO DID format
		Target: "did:amo:Z0EAD5B53B11DFE78EC8CF131D7960F097D48D70",
		// invalid DID document (no "id")
		Document: []byte(`{}`),
	})
	t1 := makeTestTxV6("claim", "sender", payload)
	rc, info := t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Contains(t, info, "invalid byte")

	payload, _ = json.Marshal(ClaimParam{
		// valid AMO DID format
		Target: "did:amo:70EAD5B53B11DFE78EC8CF131D7960F097D48D70",
		// invalid DID document (no "id")
		Document: []byte(`{}`),
	})
	t1 = makeTestTxV6("claim", "sender", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	assert.Equal(t, "mismatching did", info)

	// first claim
	myid := "did:amo:70EAD5B53B11DFE78EC8CF131D7960F097D48D70"
	mydoc := Document{
		Id: myid,
	}
	mydocJson, _ := json.Marshal(mydoc)
	payload, _ = json.Marshal(ClaimParamV6{
		Target:   myid,
		Document: mydoc,
	})
	t1 = makeTestTxV6("claim", "sender", payload)
	rc, info = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry(myid, false)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Owner) // in protocl v6, entry.Owner becomes obsolete
	assert.True(t, bytes.Equal(mydocJson, entry.Document))

	// update claim
	mydoc.Controller = "did:amo:0687D766FF0563B86BFF078B7F560AFC070C81AD"
	mydocJson, _ = json.Marshal(mydoc)
	payload, _ = json.Marshal(ClaimParamV6{
		Target:   myid,
		Document: mydoc,
	})
	t2 := makeTestTxV6("claim", "sender", payload)
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry(myid, false)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Owner) // in protocl v6, entry.Owner becomes obsolete
	// XXX note that retrieved document is a compact representation
	assert.True(t, bytes.Equal(mydocJson, entry.Document))

	// dsmiss
	payload, _ = json.Marshal(DismissParam{
		Target: myid,
	})
	t3 := makeTestTxV6("dismiss", "sender", payload)
	rc, info = t3.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)
	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, "ok", info)

	entry = s.GetDIDEntry(myid, false)
	assert.Nil(t, entry)
}
