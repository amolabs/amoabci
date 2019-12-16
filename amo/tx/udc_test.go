package tx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

func makeAccAddr(seed string) crypto.Address {
	return p256.GenPrivKeyFromSecret([]byte(seed)).PubKey().Address()
}

func TestParseIssue(t *testing.T) {
	payload := []byte(`{"id":"ff3e","operators":["99FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5"],"desc":"mycoin","total":"1000000"}`)

	var operator cmn.HexBytes
	err := json.Unmarshal(
		[]byte(`"99FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5"`),
		&operator,
	)
	assert.NoError(t, err)

	expected := IssueParam{
		Id:        []byte{0xff, 0x3e},
		Operators: []crypto.Address{operator},
		Desc:      "mycoin",
		Total:     *new(types.Currency).Set(1000000),
	}
	txParam, err := parseIssueParam(payload)
	assert.NoError(t, err)
	assert.Equal(t, expected, txParam)
}

func TestTxIssue(t *testing.T) {
	s := store.NewStore(
		tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NotNil(t, s)

	param := IssueParam{
		Id:        []byte("mycoin"),
		Operators: []crypto.Address{makeAccAddr("oper1")},
		Desc:      "mycoin",
		Total:     *new(types.Currency).Set(1000000),
	}
	payload, _ := json.Marshal(param)

	// initial issuing
	tx := makeTestTx("issue", "issuer", payload)
	assert.NotNil(t, tx)
	rc, _ := tx.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	// TODO: test validator permission
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	udc := s.GetUDC([]byte("mycoin"), false)
	assert.NotNil(t, udc)
	assert.Equal(t, *new(types.Currency).Set(1000000), udc.Total)

	// additional issuing (fail)
	tx = makeTestTx("issue", "bogus", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)

	// additional issuing
	tx = makeTestTx("issue", "issuer", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	// check
	udc = s.GetUDC([]byte("mycoin"), false)
	assert.NotNil(t, udc)
	assert.Equal(t, *new(types.Currency).Set(2000000), udc.Total)

	// change fields other than total
	param = IssueParam{
		Id:        []byte("mycoin"),
		Operators: []crypto.Address{makeAccAddr("oper2")},
		Desc:      "my own coin",
		Total:     *new(types.Currency).Set(0),
	}
	payload, _ = json.Marshal(param)
	tx = makeTestTx("issue", "oper1", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	// check
	udc = s.GetUDC([]byte("mycoin"), false)
	assert.NotNil(t, udc)
	assert.Equal(t, []crypto.Address{makeAccAddr("oper2")}, udc.Operators)
	assert.Equal(t, "my own coin", udc.Desc)
	assert.Equal(t, *new(types.Currency).Set(2000000), udc.Total)
}

func TestTxUDCBalance(t *testing.T) {
	s := store.NewStore(
		tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NotNil(t, s)

	amoM := *new(types.Currency).Set(1000000)
	amoK := *new(types.Currency).Set(1000)

	// issue
	param := IssueParam{
		Id:        []byte("mycoin"),
		Operators: []crypto.Address{makeAccAddr("oper1")},
		Desc:      "mycoin",
		Total:     amoM,
	}
	payload, _ := json.Marshal(param)
	tx := makeTestTx("issue", "issuer", payload)
	tx.Execute(s)
	// check
	udc := s.GetUDC([]byte("mycoin"), false)
	assert.NotNil(t, udc)
	assert.Equal(t, amoM, udc.Total)
	bal := s.GetUDCBalance(udc.Id, makeAccAddr("issuer"), false)
	assert.Equal(t, &amoM, bal)

	// issue more
	param = IssueParam{
		Id:        []byte("mycoin"),
		Operators: nil,
		Desc:      "mycoin",
		Total:     amoK,
	}
	payload, _ = json.Marshal(param)
	tx = makeTestTx("issue", "issuer", payload)
	tx.Execute(s)
	// check
	tmp := types.Currency{}
	tmp.Add(&amoM)
	tmp.Add(&amoK)
	bal = s.GetUDCBalance(udc.Id, makeAccAddr("issuer"), false)
	assert.Equal(t, &tmp, bal)
}
