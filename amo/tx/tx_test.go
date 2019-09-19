package tx

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmdb "github.com/tendermint/tm-db"

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
	[]byte{0xB, 0xB, 0xB, 0xB},
	[]byte{0x1, 0x1, 0x1, 0x1},
}

var custody = []cmn.HexBytes{
	[]byte{0xC, 0xC, 0xC, 0xC},
	[]byte{0xD, 0xD, 0xD, 0xD},
	[]byte{0x2, 0x2, 0x2, 0x2},
}

func makeTestTx(txType string, seed string, payload []byte) Tx {
	privKey := p256.GenPrivKeyFromSecret([]byte(seed))
	addr := privKey.PubKey().Address()
	trans := TxBase{
		Type:    txType,
		Sender:  addr,
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
		Payload: payload,
	}
	trans.Sign(privKey)
	return classifyTx(trans)
}

func makeTestAddress(seed string) crypto.Address {
	privKey := p256.GenPrivKeyFromSecret([]byte(seed))
	addr := privKey.PubKey().Address()
	return addr
}

func getTestStore() *store.Store {
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
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
	s.SetUnlockedStake(alice.addr, &types.Stake{
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

	var to crypto.Address
	err = json.Unmarshal(
		[]byte(`"218B954DF74E7267E72541CE99AB9F49C410DB96"`),
		&to,
	)
	assert.NoError(t, err)

	expected := &TxTransfer{
		TxBase{
			Type:    "transfer",
			Sender:  sender,
			Nonce:   nonce,
			Payload: []byte(`{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"}`),
			Signature: Signature{
				PubKey:   pubkey,
				SigBytes: sigbytes,
			},
		},
		TransferParam{
			To:     to,
			Amount: *new(types.Currency).Set(1000),
		},
	}
	parsedTx, err := ParseTx(bytes)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsedTx)
}

func TestTxSignature(t *testing.T) {
	from := p256.GenPrivKeyFromSecret([]byte("test1"))
	to := p256.GenPrivKeyFromSecret([]byte("test2")).PubKey().Address()
	transfer := TransferParam{
		To:     to,
		Amount: *new(types.Currency).Set(1000),
	}
	b, _ := json.Marshal(transfer)
	trnx := &TxBase{
		Type:    "transfer",
		Payload: b,
		Sender:  from.PubKey().Address(),
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
	}
	sb := trnx.getSigningBytes()
	_sb := `{"type":"transfer","sender":"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","nonce":"12345678","payload":{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"}}`
	assert.Equal(t, _sb, string(sb))
	err := trnx.Sign(from)
	if err != nil {
		panic(err)
	}
	assert.True(t, trnx.Verify())
}

func TestValidCancel(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
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
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidCancel(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())

	// target
	param := CancelParam{
		parcelID[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("cancel", "eve", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeRequestNotFound, rc)
}

func TestValidDiscard(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})

	// target
	param := DiscardParam{
		parcelID[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("discard", "alice", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidDiscard(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})

	// target
	param := DiscardParam{
		[]byte{0xFF, 0xFF, 0xFF, 0xEE},
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("discard", "alice", payload)

	param = DiscardParam{
		parcelID[0],
	}
	payload, _ = json.Marshal(param)
	t2 := makeTestTx("discard", "eve", payload)

	// test
	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodeParcelNotFound, rc)

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)
}

func TestValidGrant(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetParcel(parcelID[1], &types.ParcelValue{
		Owner:   bob.addr,
		Custody: custody[1],
	})
	s.SetRequest(alice.addr, parcelID[1], &types.RequestValue{
		Payment: *new(types.Currency).Set(100),
	})

	// target
	param := GrantParam{
		Target:  parcelID[1],
		Grantee: alice.addr,
		Custody: custody[1],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("grant", "bob", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidGrant(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	s.SetUsage(bob.addr, parcelID[0], &types.UsageValue{
		Custody: custody[0],
		Exp:     time.Now().UTC().Add(24 * time.Hour),
	})

	// target
	param := GrantParam{
		Target:  parcelID[0],
		Grantee: eve.addr,
		Custody: custody[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("grant", "eve", payload)

	param = GrantParam{
		Target:  parcelID[0],
		Grantee: bob.addr,
		Custody: custody[0],
	}
	payload, _ = json.Marshal(param)
	t2 := makeTestTx("grant", "alice", payload)

	// test
	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeAlreadyGranted, rc)
}

func TestValidRegister(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())

	// target
	param := RegisterParam{
		Target:  parcelID[2],
		Custody: custody[2],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("register", "alice", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidRegister(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
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
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeAlreadyRegistered, rc)
}

func TestValidRequest(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
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
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
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
	rc, _, _ := t1.Execute(s)
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
	rc, _, _ = t2.Execute(s)
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
	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeSelfTransaction, rc)

	// target
	param = RequestParam{
		Target:  parcelID[1],
		Payment: *new(types.Currency).Set(100),
	}
	payload, _ = json.Marshal(param)
	t4 := makeTestTx("request", "eve", payload)

	// test
	rc, _, _ = t4.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
}

func TestValidRevoke(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	s.SetUsage(bob.addr, parcelID[0], &types.UsageValue{
		Custody: custody[0],
		Exp:     time.Now().UTC().Add(24 * time.Hour),
	})

	// target
	param := RevokeParam{
		Grantee: bob.addr,
		Target:  parcelID[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("revoke", "alice", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidRevoke(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetParcel(parcelID[0], &types.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	s.SetUsage(bob.addr, parcelID[0], &types.UsageValue{
		Custody: custody[0],
		Exp:     time.Now().UTC().Add(24 * time.Hour),
	})

	// target
	param := RevokeParam{
		Grantee: eve.addr,
		Target:  parcelID[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("revoke", "eve", payload)

	param = RevokeParam{
		Grantee: bob.addr,
		Target:  parcelID[2],
	}
	payload, _ = json.Marshal(param)
	t2 := makeTestTx("revoke", "alice", payload)

	// test
	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeParcelNotFound, rc)
}

func TestValidTransfer(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetBalanceUint64(makeTestAddress("alice"), 1230)

	// target
	param := TransferParam{
		To:     bob.addr,
		Amount: *new(types.Currency).Set(1230),
	}
	payload, _ := json.Marshal(param)
	trans := makeTestTx("transfer", "alice", payload)

	// test
	rc, _ := trans.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = trans.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidTransfer(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())

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
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	rc, _ = t2.Check()
	assert.Equal(t, code.TxCodeSelfTransaction, rc)
	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
}

func TestValidStake(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetBalanceUint64(alice.addr, 3000)

	// target
	param := StakeParam{
		Validator: cmn.RandBytes(32),
		Amount:    *new(types.Currency).Set(2000),
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("stake", "alice", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidStake(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetBalanceUint64(alice.addr, 3000)

	// target
	param := StakeParam{
		Validator: cmn.RandBytes(32),
		Amount:    *new(types.Currency).Set(2000),
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("stake", "eve", payload)

	t2 := makeTestTx("stake", "alice", payload)

	// test
	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)

	// env
	s.SetBalanceUint64(eve.addr, 2000)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	// test
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)
}

func TestValidWithdraw(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})

	// target
	param := WithdrawParam{
		Amount: *new(types.Currency).Set(1000),
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("withdraw", "alice", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, new(types.Currency).Set(1000), &s.GetStake(alice.addr).Amount)

	// add more stakeholder to test stake deletion
	//var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetUnlockedStake(bob.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})

	// test
	rc, _ = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Nil(t, s.GetStake(alice.addr))
	assert.NotNil(t, s.GetStake(bob.addr))
}

func TestNonValidWithdraw(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})

	// target
	param := WithdrawParam{
		Amount: *new(types.Currency).Set(2000),
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("withdraw", "eve", payload)
	t2 := makeTestTx("withdraw", "alice", payload)

	// test
	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodeNoStake, rc)

	// test
	rc, _ = t2.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeLastValidator, rc)

	// env
	s.SetDelegate(bob.addr, &types.Delegate{
		Delegatee: alice.addr,
		Amount:    *new(types.Currency).Set(500),
	})

	// test
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeDelegateExists, rc)
}

func TestValidDelegate(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	s.SetBalanceUint64(bob.addr, 1000)

	// target
	param := DelegateParam{
		Amount: *new(types.Currency).Set(1000),
		To:     alice.addr,
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("delegate", "bob", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, new(types.Currency).Set(1000), &s.GetDelegate(bob.addr).Amount)
}

func TestNonValidDelegate(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	copy(k[:], cmn.RandBytes(32))
	s.SetUnlockedStake(eve.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	s.SetBalanceUint64(alice.addr, 1000)
	s.SetBalanceUint64(bob.addr, 1000)

	// test
	payload, _ := json.Marshal(DelegateParam{
		Amount: *new(types.Currency).Set(500),
		To:     eve.addr,
	})
	t1 := makeTestTx("delegate", "eve", payload)
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeSelfTransaction, rc)

	payload, _ = json.Marshal(DelegateParam{
		Amount: *new(types.Currency).Set(500),
		To:     alice.addr,
	})
	t1 = makeTestTx("delegate", "eve", payload)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)

	t1 = makeTestTx("delegate", "bob", payload)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	payload, _ = json.Marshal(DelegateParam{
		Amount: *new(types.Currency).Set(500),
		To:     eve.addr,
	})
	t1 = makeTestTx("delegate", "bob", payload)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeMultipleDelegates, rc)

	payload, _ = json.Marshal(DelegateParam{
		Amount: *new(types.Currency).Set(500),
		To:     bob.addr,
	})
	t1 = makeTestTx("delegate", "alice", payload)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNoStake, rc)
}

func TestValidRetract(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	s.SetDelegate(bob.addr, &types.Delegate{
		Delegatee: alice.addr,
		Amount:    *new(types.Currency).Set(500),
	})

	// test
	payload, _ := json.Marshal(RetractParam{
		Amount: *new(types.Currency).Set(400),
	})
	t1 := makeTestTx("retract", "bob", payload)

	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Equal(t, new(types.Currency).Set(100), &s.GetDelegate(bob.addr).Amount)

	// test
	payload, _ = json.Marshal(RetractParam{
		Amount: *new(types.Currency).Set(100),
	})
	t1 = makeTestTx("retract", "bob", payload)

	rc, _ = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Nil(t, s.GetDelegate(bob.addr))

	assert.Equal(t, new(types.Currency).Set(2000), &s.GetStake(alice.addr).Amount)
}

func TestNonValidRetract(t *testing.T) {
	// env
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	var k ed25519.PubKeyEd25519
	copy(k[:], cmn.RandBytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	s.SetDelegate(bob.addr, &types.Delegate{
		Delegatee: alice.addr,
		Amount:    *new(types.Currency).Set(500),
	})

	payload, _ := json.Marshal(RetractParam{
		Amount: *new(types.Currency).Set(500),
	})
	t1 := makeTestTx("retract", "eve", payload)
	t2 := makeTestTx("retract", "bob", payload)

	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodeDelegateNotFound, rc)

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeDelegateNotFound, rc)
}

func TestStakeLockup(t *testing.T) {
	s := store.NewStore(tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetBalanceUint64(alice.addr, 3000)

	// setup lock-up period config
	ConfigLockupPeriod = 2

	// deposit stake
	stakeParam := StakeParam{
		Validator: cmn.RandBytes(32),
		Amount:    *new(types.Currency).Set(2000),
	}
	payload, _ := json.Marshal(stakeParam)
	t1 := makeTestTx("stake", "alice", payload)
	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	// withdraw stake
	withdrawParam := WithdrawParam{
		Amount: *new(types.Currency).Set(1000),
	}
	payload, _ = json.Marshal(withdrawParam)
	t2 := makeTestTx("withdraw", "alice", payload)

	// test
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeStakeLocked, rc)

	// stake is locked at height 2. loosen 2 times.
	s.LoosenLockedStakes()
	s.LoosenLockedStakes()

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	stake := s.GetStake(makeTestAddress("alice"))
	var val ed25519.PubKeyEd25519
	copy(val[:], stakeParam.Validator)
	assert.Equal(t, &types.Stake{
		Validator: val,
		Amount:    *new(types.Currency).Set(1000),
	}, stake)

	// TODO: test last validator error later
	//rc, _, _ = t2.Execute(s)
	//assert.Equal(t, code.TxCodeOK, rc)

	//stake = s.GetStake(makeTestAddress("alice"))
	//assert.Nil(t, stake)
}
