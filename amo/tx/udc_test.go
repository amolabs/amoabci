package tx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmrand "github.com/tendermint/tendermint/libs/rand"
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
	payload := []byte(`{"udc":65342,"operators":["99FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5"],"desc":"mycoin","amount":"1000000"}`)

	var operator tmbytes.HexBytes
	err := json.Unmarshal(
		[]byte(`"99FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5"`),
		&operator,
	)
	assert.NoError(t, err)

	expected := IssueParam{
		UDC:       65342,
		Operators: []crypto.Address{operator},
		Desc:      "mycoin",
		Amount:    *new(types.Currency).Set(1000000),
	}
	txParam, err := parseIssueParam(payload)
	assert.NoError(t, err)
	assert.Equal(t, expected, txParam)
}

func TestTxIssue(t *testing.T) {
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	param := IssueParam{
		UDC:       123,
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
	copy(newStake.Validator[:], tmrand.Bytes(32))
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
		UDC:       123,
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
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	amo0 := *new(types.Currency)
	amoM := *new(types.Currency).Set(1000000)
	amoK := *new(types.Currency).Set(1000)
	issuer := makeAccAddr("issuer")

	// make necessary stake
	newStake := types.Stake{}
	newStake.Amount = *new(types.Currency).Set(2000)
	copy(newStake.Validator[:], tmrand.Bytes(32))
	s.SetUnlockedStake(makeAccAddr("issuer"), &newStake)

	// issue
	param := IssueParam{
		UDC:       123,
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
	bal := s.GetUDCBalance(param.UDC, issuer, false)
	assert.Equal(t, &amoM, bal)

	// issue more
	param = IssueParam{
		UDC:       123,
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
	bal = s.GetUDCBalance(param.UDC, issuer, false)
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

func TestParseLock(t *testing.T) {
	payload := []byte(`{"udc":123,"holder":"99FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","amount":"1000000"}`)

	var holder tmbytes.HexBytes
	err := json.Unmarshal(
		[]byte(`"99FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5"`),
		&holder,
	)
	assert.NoError(t, err)

	expected := LockParam{
		UDC:    uint32(123),
		Holder: holder,
		Amount: *new(types.Currency).Set(1000000),
	}
	txParam, err := parseLockParam(payload)
	assert.NoError(t, err)
	assert.Equal(t, expected, txParam)
}

func TestUDCLock(t *testing.T) {
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	param := LockParam{
		UDC:    uint32(123),
		Holder: makeAccAddr("holder"),
		Amount: *new(types.Currency).SetAMO(10),
	}
	payload, _ := json.Marshal(param)

	// basic check
	tx := makeTestTx("lock", "issuer", payload)
	assert.NotNil(t, tx)
	_, ok := tx.(*TxLock)
	assert.True(t, ok)
	rc, _ := tx.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	// no udc
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeUDCNotFound, rc)

	mycoin := &types.UDC{
		makeAccAddr("issuer"),
		"mycoin for test",
		[]crypto.Address{
			makeAccAddr("op1"),
		},
		*new(types.Currency).SetAMO(100),
	}
	assert.NotNil(t, mycoin)
	assert.NoError(t, s.SetUDC(uint32(123), mycoin))

	// no permission
	tx = makeTestTx("lock", "anyone", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)

	// ok
	tx = makeTestTx("lock", "issuer", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	tx = makeTestTx("lock", "op1", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	// set test balance
	s.SetUDCBalance(uint32(123), makeAccAddr("holder"),
		new(types.Currency).SetAMO(12))

	// transfer too much
	payload, _ = json.Marshal(TransferParam{
		UDC:    uint32(123),
		To:     makeAccAddr("recp"),
		Amount: *new(types.Currency).SetAMO(3),
	})
	tx = makeTestTx("transfer", "holder", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)

	// burn too much
	payload, _ = json.Marshal(BurnParam{
		UDC:    uint32(123),
		Amount: *new(types.Currency).SetAMO(3),
	})
	tx = makeTestTx("burn", "holder", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)

	// transfer ok
	payload, _ = json.Marshal(TransferParam{
		UDC:    uint32(123),
		To:     makeAccAddr("recp"),
		Amount: *new(types.Currency).SetAMO(1),
	})
	tx = makeTestTx("transfer", "holder", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	// burn ok
	payload, _ = json.Marshal(BurnParam{
		UDC:    uint32(123),
		Amount: *new(types.Currency).SetAMO(1),
	})
	tx = makeTestTx("burn", "holder", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	// situation changed. balance reduced
	payload, _ = json.Marshal(TransferParam{
		UDC:    uint32(123),
		To:     makeAccAddr("recp"),
		Amount: *new(types.Currency).SetAMO(1),
	})
	tx = makeTestTx("transfer", "holder", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
	// same for burn
	payload, _ = json.Marshal(BurnParam{
		UDC:    uint32(123),
		Amount: *new(types.Currency).SetAMO(1),
	})
	tx = makeTestTx("burn", "holder", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
}

func TestParseBurn(t *testing.T) {
	payload := []byte(`{"udc":123,"amount":"1000000"}`)

	expected := BurnParam{
		UDC:    123,
		Amount: *new(types.Currency).Set(1000000),
	}
	txParam, err := parseBurnParam(payload)
	assert.NoError(t, err)
	assert.Equal(t, expected, txParam)
}

func TestUDCBurn(t *testing.T) {
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	param := BurnParam{
		UDC:    uint32(123),
		Amount: *new(types.Currency).SetAMO(10),
	}
	payload, _ := json.Marshal(param)

	// basic check
	tx := makeTestTx("burn", "holder", payload)
	assert.NotNil(t, tx)
	_, ok := tx.(*TxBurn)
	assert.True(t, ok)
	rc, _ := tx.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	// no udc
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeUDCNotFound, rc)

	mycoin := &types.UDC{
		makeAccAddr("issuer"),
		"mycoin for test",
		[]crypto.Address{
			makeAccAddr("op1"),
		},
		*new(types.Currency).SetAMO(100),
	}
	assert.NotNil(t, mycoin)
	assert.NoError(t, s.SetUDC(uint32(123), mycoin))

	// set test balance
	s.SetUDCBalance(uint32(123), makeAccAddr("holder"),
		new(types.Currency).SetAMO(11),
	)

	// ok
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	bal := s.GetUDCBalance(uint32(123), makeAccAddr("holder"), false)
	assert.Equal(t, new(types.Currency).SetAMO(1), bal)
	mycoinAfter := s.GetUDC(uint32(123), false)
	assert.Equal(t, *mycoin.Total.Sub(new(types.Currency).SetAMO(10)), mycoinAfter.Total)
}
