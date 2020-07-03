package tx

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
)

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
