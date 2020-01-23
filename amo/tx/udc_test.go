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
	payload := []byte(`{"id":65342,"operators":["99FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5"],"desc":"mycoin","amount":"1000000"}`)

	var operator cmn.HexBytes
	err := json.Unmarshal(
		[]byte(`"99FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5"`),
		&operator,
	)
	assert.NoError(t, err)

	expected := IssueParam{
		ID:        65342,
		Operators: []crypto.Address{operator},
		Desc:      "mycoin",
		Amount:    *new(types.Currency).Set(1000000),
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
		ID:        123,
		Operators: []crypto.Address{makeAccAddr("oper1")},
		Desc:      "mycoin",
		Amount:    *new(types.Currency).Set(1000000),
	}
	payload, _ := json.Marshal(param)

	// initial issuing
	tx := makeTestTx("issue", "issuer", payload)
	assert.NotNil(t, tx)
	_, ok := tx.(*TxIssue)
	assert.True(t, ok)
	rc, _ := tx.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	// check validator permission
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)
	// make enough stake and try again
	newStake := types.Stake{}
	newStake.Amount = *new(types.Currency).Set(2000)
	copy(newStake.Validator[:], cmn.RandBytes(32))
	s.SetUnlockedStake(makeAccAddr("issuer"), &newStake)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	udc := s.GetUDC(123, false)
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
	udc = s.GetUDC(123, false)
	assert.NotNil(t, udc)
	assert.Equal(t, *new(types.Currency).Set(2000000), udc.Total)

	// change fields other than total
	param = IssueParam{
		ID:        123,
		Operators: []crypto.Address{makeAccAddr("oper2")},
		Desc:      "my own coin",
		Amount:    *new(types.Currency).Set(0),
	}
	payload, _ = json.Marshal(param)
	tx = makeTestTx("issue", "oper1", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	// check
	udc = s.GetUDC(123, false)
	assert.NotNil(t, udc)
	assert.Equal(t, []crypto.Address{makeAccAddr("oper2")}, udc.Operators)
	assert.Equal(t, "my own coin", udc.Desc)
	assert.Equal(t, *new(types.Currency).Set(2000000), udc.Total)
}

func TestTxUDCBalance(t *testing.T) {
	s := store.NewStore(
		tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NotNil(t, s)

	amo0 := *new(types.Currency)
	amoM := *new(types.Currency).Set(1000000)
	amoK := *new(types.Currency).Set(1000)
	issuer := makeAccAddr("issuer")

	// make necessary stake
	newStake := types.Stake{}
	newStake.Amount = *new(types.Currency).Set(2000)
	copy(newStake.Validator[:], cmn.RandBytes(32))
	s.SetUnlockedStake(makeAccAddr("issuer"), &newStake)

	// issue
	param := IssueParam{
		ID:        123,
		Operators: []crypto.Address{makeAccAddr("oper1")},
		Desc:      "mycoin",
		Amount:    amoM,
	}
	payload, _ := json.Marshal(param)
	tx := makeTestTx("issue", "issuer", payload)
	tx.Execute(s)
	// check
	udc := s.GetUDC(123, false)
	assert.NotNil(t, udc)
	assert.Equal(t, amoM, udc.Total)
	bal := s.GetUDCBalance(123, issuer, false)
	assert.Equal(t, &amoM, bal)

	// issue more
	param = IssueParam{
		ID:        123,
		Operators: nil,
		Desc:      "mycoin",
		Amount:    amoK,
	}
	payload, _ = json.Marshal(param)
	tx = makeTestTx("issue", "issuer", payload)
	tx.Execute(s)
	// check
	tmp := types.Currency{}
	tmp.Add(&amoM)
	tmp.Add(&amoK)
	bal = s.GetUDCBalance(123, issuer, false)
	assert.Equal(t, &tmp, bal)

	// non-UDC balance
	bal = s.GetUDCBalance(0, issuer, false)
	assert.Equal(t, &amo0, bal)

	// parser test for optional tx field
	b := []byte(`{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"}`)
	parsed, err := parseTransferParam(b)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), parsed.UDC)
	b = []byte(`{"udc": 123,"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"}`)
	parsed, err = parseTransferParam(b)
	assert.NoError(t, err)
	assert.Equal(t, uint32(123), parsed.UDC)
	// transfer
	acc1 := makeAccAddr("acc1")
	payload, _ = json.Marshal(TransferParam{
		UDC:    123,
		To:     acc1,
		Amount: amoK,
	})
	tx = makeTestTx("transfer", "issuer", payload)
	rc, _ := tx.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	// check
	bal = s.GetUDCBalance(123, issuer, false)
	assert.Equal(t, &amoM, bal)
	bal = s.GetUDCBalance(123, acc1, false)
	assert.Equal(t, &amoK, bal)
	// not enough  balance
	payload, _ = json.Marshal(TransferParam{
		UDC:    123,
		To:     acc1,
		Amount: amoK,
	})
	tx = makeTestTx("transfer", "acc2", payload)
	rc, _ = tx.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
	// transfer remaining
	payload, _ = json.Marshal(TransferParam{
		UDC:    123,
		To:     acc1,
		Amount: amoM,
	})
	tx = makeTestTx("transfer", "issuer", payload)
	tx.Execute(s)
	// check
	bal = s.GetUDCBalance(123, issuer, false)
	assert.Equal(t, &amo0, bal)
	bal = s.GetUDCBalance(123, acc1, false)
	assert.Equal(t, &tmp, bal)
}
