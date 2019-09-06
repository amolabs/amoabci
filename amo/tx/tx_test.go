package tx

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

type user struct {
	privKey crypto.PrivKey
	pubKey  crypto.PubKey
	addr    crypto.Address
}

func newUser(privKey p256.PrivKeyP256) user {
	return user{
		privKey: privKey,
		pubKey:  privKey.PubKey(),
		addr:    privKey.PubKey().Address(),
	}
}

var alice = newUser(p256.GenPrivKeyFromSecret([]byte("alice")))
var bob = newUser(p256.GenPrivKeyFromSecret([]byte("bob")))
var eve = newUser(p256.GenPrivKeyFromSecret([]byte("eve")))

var parcelID = []cmn.HexBytes{
	[]byte{0xA, 0xA, 0xA, 0xA},
	[]byte{0xB, 0xB, 0xB, 0XB},
	[]byte{0x1, 0x1, 0x1, 0x1},
}

var custody = []cmn.HexBytes{
	[]byte{0xC, 0xC, 0xC, 0XC},
	[]byte{0xD, 0xD, 0xD, 0XD},
	[]byte{0x2, 0x2, 0x2, 0x2},
}

func makeTestTx(txType string, seed string, payload []byte) Tx {
	privKey := p256.GenPrivKeyFromSecret([]byte(seed))
	addr := privKey.PubKey().Address()
	trans := Tx{
		Type:    txType,
		Sender:  addr,
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
		Payload: payload,
	}
	trans.Sign(privKey)
	return trans
}

func makeTestAddress(seed string) crypto.Address {
	privKey := p256.GenPrivKeyFromSecret([]byte(seed))
	addr := privKey.PubKey().Address()
	return addr
}

func getTestStore() *store.Store {
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())
	s.SetBalanceUint64(alice.addr, 3000)
	s.SetBalanceUint64(bob.addr, 1000)
	s.SetBalanceUint64(eve.addr, 50)
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	s.SetParcel(parcelID[1], &types.ParcelValue{
		Owner:   bob.addr,
		Custody: custody[1],
	})
	s.SetRequest(bob.addr, parcelID[0], &types.RequestValue{
		Payment: *new(types.Currency).Set(100),
	})
	s.SetRequest(alice.addr, parcelID[1], &types.RequestValue{
		Payment: *new(types.Currency).Set(100),
	})
	s.SetUsage(bob.addr, parcelID[0], &types.UsageValue{
		Custody: custody[0],
		Exp:     time.Now().UTC().Add(24 * time.Hour),
	})
	var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	s.SetDelegate(bob.addr, &types.Delegate{
		Delegatee: alice.addr,
		Amount:    *new(types.Currency).Set(500),
	})
	return s
}

func TestParseTx(t *testing.T) {
	bytes := []byte(`{"type":"transfer","sender":"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","nonce":"12345678","payload":{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"},"signature":{"pubkey":"0485FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B185FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B1EF1D55E9B1EF1D","sig_bytes":"FFFFFFFF"}}`)
	var sender, nonce, tmp, sigbytes cmn.HexBytes
	err := json.Unmarshal(
		[]byte(`"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5"`),
		&sender,
	)
	assert.NoError(t, err)
	err = json.Unmarshal(
		[]byte(`"12345678"`),
		&nonce,
	)
	assert.NoError(t, err)
	err = json.Unmarshal(
		[]byte(`"0485FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B185FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B1EF1D55E9B1EF1D"`),
		&tmp,
	)
	assert.NoError(t, err)
	var pubkey p256.PubKeyP256
	copy(pubkey[:], tmp)
	err = json.Unmarshal(
		[]byte(`"FFFFFFFF"`),
		&sigbytes,
	)
	assert.NoError(t, err)

	exptected := Tx{
		Type:    "transfer",
		Sender:  sender,
		Nonce:   nonce,
		Payload: []byte(`{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"}`),
		Signature: Signature{
			PubKey:   pubkey,
			SigBytes: sigbytes,
		},
	}
	parsedTx, _, _, err := ParseTx(bytes)
	assert.NoError(t, err)
	assert.Equal(t, exptected, parsedTx)
}

func TestTxSignature(t *testing.T) {
	from := p256.GenPrivKeyFromSecret([]byte("test1"))
	to := p256.GenPrivKeyFromSecret([]byte("test2")).PubKey().Address()
	transfer := TransferParam{
		To:     to,
		Amount: *new(types.Currency).Set(1000),
	}
	b, _ := json.Marshal(transfer)
	message := Tx{
		Type:    "transfer",
		Payload: b,
		Sender:  from.PubKey().Address(),
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
	}
	sb := message.GetSigningBytes()
	_sb := `{"type":"transfer","sender":"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","nonce":"12345678","payload":{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"}}`
	assert.Equal(t, _sb, string(sb))
	err := message.Sign(from)
	if err != nil {
		panic(err)
	}
	assert.True(t, message.Verify())
}

func TestValidCancel(t *testing.T) {
	// env
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	s.SetRequest(bob.addr, parcelID[0], &types.RequestValue{
		Payment: *new(types.Currency).Set(100),
	})

	// target
	param := CancelParam{
		parcelID[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("cancel", "bob", payload)

	// test
	rc, _ := CheckCancel(t1)
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = ExecuteCancel(t1, s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidCancel(t *testing.T) {
	// env
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())

	// target
	param := CancelParam{
		parcelID[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("cancel", "eve", payload)

	// test
	rc, _ := CheckCancel(t1)
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = ExecuteCancel(t1, s)
	assert.Equal(t, code.TxCodeRequestNotFound, rc)
}

func TestValidDiscard(t *testing.T) {
	s := getTestStore()
	op := Discard{
		parcelID[0],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode, _ := op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidDiscard(t *testing.T) {
	s := getTestStore()
	NEOp := Discard{
		[]byte{0xFF, 0xFF, 0xFF, 0xEE},
	}
	PDOp := Discard{
		parcelID[0],
	}
	assert.Equal(t, code.TxCodeParcelNotFound, NEOp.Check(s, alice.addr))
	assert.Equal(t, code.TxCodePermissionDenied, PDOp.Check(s, eve.addr))
}

func TestValidGrant(t *testing.T) {
	s := getTestStore()
	op := Grant{
		Target:  parcelID[1],
		Grantee: alice.addr,
		Custody: custody[1],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, bob.addr))
	resCode, _ := op.Execute(s, bob.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidGrant(t *testing.T) {
	s := getTestStore()
	PDop := Grant{
		Target:  parcelID[0],
		Grantee: eve.addr,
		Custody: custody[0],
	}
	AEop := Grant{
		Target:  parcelID[0],
		Grantee: bob.addr,
		Custody: custody[0],
	}
	assert.Equal(t, code.TxCodePermissionDenied, PDop.Check(s, eve.addr))
	assert.Equal(t, code.TxCodeAlreadyGranted, AEop.Check(s, alice.addr))
}

func TestValidRegister(t *testing.T) {
	// env
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())

	// target
	param := RegisterParam{
		Target:  parcelID[2],
		Custody: custody[2],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("register", "alice", payload)

	// test
	rc, _ := CheckRegister(t1)
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = ExecuteRegister(t1, s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidRegister(t *testing.T) {
	// env
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})

	// target
	param := RegisterParam{
		Target:  parcelID[0],
		Custody: custody[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("register", "alice", payload)

	// test
	rc, _ := CheckRegister(t1)
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = ExecuteRegister(t1, s)
	assert.Equal(t, code.TxCodeAlreadyRegistered, rc)
}

func TestValidRequest(t *testing.T) {
	// env
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())
	s.SetBalanceUint64(alice.addr, 200)
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   bob.addr,
		Custody: custody[0],
	})

	// target
	param := RequestParam{
		Target:  parcelID[0],
		Payment: *new(types.Currency).Set(200),
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("request", "alice", payload)

	// test
	rc, _ := CheckRequest(t1)
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = ExecuteRequest(t1, s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidRequest(t *testing.T) {
	// env
	s := getTestStore()

	// target
	param := RequestParam{
		Target:  []byte{0x0, 0x0, 0x0, 0x0},
		Payment: *new(types.Currency).Set(100),
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("request", "eve", payload)

	// test
	rc, _, _ := ExecuteRequest(t1, s)
	assert.Equal(t, code.TxCodeParcelNotFound, rc)

	// env
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	s.SetUsage(bob.addr, parcelID[0], &types.UsageValue{
		Custody: custody[0],
		Exp:     time.Now().UTC().Add(24 * time.Hour),
	})

	// target
	param = RequestParam{
		Target:  parcelID[0],
		Payment: *new(types.Currency).Set(100),
	}
	payload, _ = json.Marshal(param)
	t2 := makeTestTx("request", "bob", payload)

	// test
	rc, _, _ = ExecuteRequest(t2, s)
	assert.Equal(t, code.TxCodeAlreadyGranted, rc)

	// env
	s.SetParcel(parcelID[1], &types.ParcelValue{
		Owner:   bob.addr,
		Custody: custody[1],
	})

	// target
	param = RequestParam{
		Target:  parcelID[1],
		Payment: *new(types.Currency).Set(100),
	}
	payload, _ = json.Marshal(param)
	t3 := makeTestTx("request", "bob", payload)

	// test
	rc, _, _ = ExecuteRequest(t3, s)
	assert.Equal(t, code.TxCodeSelfTransaction, rc)

	// target
	param = RequestParam{
		Target:  parcelID[1],
		Payment: *new(types.Currency).Set(100),
	}
	payload, _ = json.Marshal(param)
	t4 := makeTestTx("request", "eve", payload)

	// test
	rc, _, _ = ExecuteRequest(t4, s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
}

func TestValidRevoke(t *testing.T) {
	s := getTestStore()
	op := Revoke{
		Grantee: bob.addr,
		Target:  parcelID[0],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode, _ := op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidRevoke(t *testing.T) {
	s := getTestStore()
	PDop := Revoke{
		Grantee: eve.addr,
		Target:  parcelID[0],
	}
	TNop := Revoke{
		Grantee: bob.addr,
		Target:  parcelID[2],
	}
	assert.Equal(t, code.TxCodePermissionDenied, PDop.Check(s, eve.addr))
	assert.Equal(t, code.TxCodeParcelNotFound, TNop.Check(s, alice.addr))
}

func TestValidTransfer(t *testing.T) {
	// env
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())
	s.SetBalanceUint64(makeTestAddress("alice"), 1230)

	// target
	param := TransferParam{
		To:     bob.addr,
		Amount: *new(types.Currency).Set(1230),
	}
	payload, _ := json.Marshal(param)
	trans := makeTestTx("transfer", "alice", payload)

	// test
	rc, _ := CheckTransfer(trans)
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = ExecuteTransfer(trans, s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidTransfer(t *testing.T) {
	// env
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())

	// target
	param := TransferParam{
		To:     []byte("bob"),
		Amount: *new(types.Currency).Set(1230),
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("transfer", "alice", payload)

	param = TransferParam{
		To:     bob.addr,
		Amount: *new(types.Currency).Set(500),
	}
	payload, _ = json.Marshal(param)
	t2 := makeTestTx("transfer", "bob", payload)

	param = TransferParam{
		To:     eve.addr,
		Amount: *new(types.Currency).Set(10),
	}
	payload, _ = json.Marshal(param)
	t3 := makeTestTx("transfer", "eve", payload)

	// test
	rc, _ := CheckTransfer(t1)
	assert.Equal(t, code.TxCodeBadParam, rc)
	rc, _ = CheckTransfer(t2)
	assert.Equal(t, code.TxCodeSelfTransaction, rc)
	rc, _, _ = ExecuteTransfer(t3, s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
}

func TestValidStake(t *testing.T) {
	s := getTestStore()
	op := Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: cmn.RandBytes(32),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode, _ := op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidStake(t *testing.T) {
	s := getTestStore()
	NEop := Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: cmn.RandBytes(32),
	}
	BVop := Stake{
		Amount:    *new(types.Currency).Set(500),
		Validator: cmn.RandBytes(33),
	}
	assert.Equal(t, code.TxCodeNotEnoughBalance, NEop.Check(s, eve.addr))
	assert.Equal(t, code.TxCodeBadValidator, BVop.Check(s, alice.addr))
}

func TestValidWithdraw(t *testing.T) {
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())
	var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})

	op := Withdraw{
		Amount: *new(types.Currency).Set(1000),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode, _ := op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
	assert.Equal(t, new(types.Currency).Set(1000), &s.GetStake(alice.addr).Amount)

	// add more stakeholder to test stake deletion
	//var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetStake(bob.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})

	op = Withdraw{
		Amount: *new(types.Currency).Set(1000),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode, _ = op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
	assert.Nil(t, s.GetStake(alice.addr))
}

func TestNonValidWithdraw(t *testing.T) {
	// env
	s := store.NewStore(db.NewMemDB(), db.NewMemDB())
	var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})

	// test
	op := Withdraw{
		Amount: *new(types.Currency).Set(2000),
	}
	assert.Equal(t, code.TxCodeNotEnoughBalance, op.Check(s, eve.addr))

	// test
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode, _ := op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeLastValidator, resCode)

	// prepare
	s.SetDelegate(bob.addr, &types.Delegate{
		Delegatee: alice.addr,
		Amount:    *new(types.Currency).Set(500),
	})

	// test
	resCode, _ = op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeDelegateExists, resCode)
}

func TestValidDelegate(t *testing.T) {
	s := getTestStore()
	op := Delegate{
		Amount: *new(types.Currency).Set(500),
		To:     alice.addr,
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, bob.addr))
	resCode, _ := op.Execute(s, bob.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
	assert.Equal(t, new(types.Currency).Set(1000), &s.GetDelegate(bob.addr).Amount)
}

func TestNonValidDelegate(t *testing.T) {
	s := getTestStore()
	STop := Delegate{
		Amount: *new(types.Currency).Set(500),
		To:     eve.addr,
	}
	NEop := Delegate{
		Amount: *new(types.Currency).Set(500),
		To:     alice.addr,
	}
	ADop := Delegate{
		Amount: *new(types.Currency).Set(500),
		To:     eve.addr,
	}
	NSop := Delegate{
		Amount: *new(types.Currency).Set(500),
		To:     eve.addr,
	}
	assert.Equal(t, code.TxCodeSelfTransaction, STop.Check(s, eve.addr))
	assert.Equal(t, code.TxCodeNotEnoughBalance, NEop.Check(s, eve.addr))
	assert.Equal(t, code.TxCodeMultipleDelegates, ADop.Check(s, bob.addr))
	assert.Equal(t, code.TxCodeNoStake, NSop.Check(s, alice.addr))
}

func TestValidRetract(t *testing.T) {
	s := getTestStore()
	op := Retract{
		Amount: *new(types.Currency).Set(400),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, bob.addr))
	resCode, _ := op.Execute(s, bob.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
	assert.Equal(t, new(types.Currency).Set(100), &s.GetDelegate(bob.addr).Amount)

	op = Retract{
		Amount: *new(types.Currency).Set(100),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, bob.addr))
	resCode, _ = op.Execute(s, bob.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
	assert.Equal(t, zero, &s.GetDelegate(bob.addr).Amount)
}

func TestNonValidRetract(t *testing.T) {
	s := getTestStore()
	op := Retract{
		Amount: *new(types.Currency).Set(500),
	}
	NEop := Retract{
		Amount: *new(types.Currency).Set(1000),
	}
	assert.Equal(t, code.TxCodeDelegationNotExists, op.Check(s, eve.addr))
	assert.Equal(t, code.TxCodeOK, op.Check(s, bob.addr))
	assert.Equal(t, code.TxCodeNotEnoughBalance, NEop.Check(s, bob.addr))
}
