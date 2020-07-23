package tx

import (
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmrand "github.com/tendermint/tendermint/libs/rand"
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

var (
	alice = newUser(p256.GenPrivKeyFromSecret([]byte("alice")))
	bob   = newUser(p256.GenPrivKeyFromSecret([]byte("bob")))
	carol = newUser(p256.GenPrivKeyFromSecret([]byte("carol")))
	eve   = newUser(p256.GenPrivKeyFromSecret([]byte("eve")))
)

var parcelID = []tmbytes.HexBytes{
	[]byte{0xA, 0xA, 0xA, 0xA},
	[]byte{0xB, 0xB, 0xB, 0xB},
	[]byte{0x1, 0x1, 0x1, 0x1},
}

var custody = []tmbytes.HexBytes{
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
	s, _ := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	s.SetBalanceUint64(alice.addr, 3000)
	s.SetBalanceUint64(bob.addr, 1000)
	s.SetBalanceUint64(eve.addr, 50)
	s.SetParcel(parcelID[0], &types.Parcel{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	s.SetParcel(parcelID[1], &types.Parcel{
		Owner:   bob.addr,
		Custody: custody[1],
	})
	s.SetRequest(bob.addr, parcelID[0], &types.Request{
		Payment: *new(types.Currency).Set(100),
	})
	s.SetRequest(alice.addr, parcelID[1], &types.Request{
		Payment: *new(types.Currency).Set(100),
	})
	s.SetUsage(bob.addr, parcelID[0], &types.Usage{
		Custody: custody[0],
	})
	var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))
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
	bytes := []byte(`{"type":"transfer","sender":"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","payload":{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"},"signature":{"pubkey":"0485FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B185FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B1EF1D55E9B1EF1D","sig_bytes":"FFFFFFFF"}}`)
	var sender, tmp, sigbytes tmbytes.HexBytes
	err := json.Unmarshal(
		[]byte(`"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5"`),
		&sender,
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
		Type:       "transfer",
		Payload:    b,
		Sender:     from.PubKey().Address(),
		LastHeight: "1",
	}
	sb := trnx.getSigningBytes()
	_sb := `{"type":"transfer","sender":"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","fee":"0","last_height":"1","payload":{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"}}`
	assert.Equal(t, _sb, string(sb))
	err := trnx.Sign(from)
	assert.NoError(t, err)
	assert.True(t, trnx.Verify())

	// wrong sender address
	trnx.Sender = to
	err = trnx.Sign(from)
	assert.NoError(t, err)
	assert.False(t, trnx.Verify())
}

func TestValidCancel(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	s.SetParcel(parcelID[0], &types.Parcel{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	s.SetRequest(bob.addr, parcelID[0], &types.Request{
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
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	s.SetParcel(parcelID[0], &types.Parcel{
		Owner:        alice.addr,
		Custody:      custody[0],
		ProxyAccount: carol.addr,
	})

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
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	s.SetParcel(parcelID[0], &types.Parcel{
		Owner:   alice.addr,
		Custody: custody[0],
	})

	s.SetParcel(parcelID[1], &types.Parcel{
		Owner:        alice.addr,
		Custody:      custody[1],
		ProxyAccount: bob.addr,
	})

	// target
	param := DiscardParam{
		parcelID[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("discard", "alice", payload)

	param = DiscardParam{
		parcelID[1],
	}
	payload, _ = json.Marshal(param)
	t2 := makeTestTx("discard", "bob", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _ = t2.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidDiscard(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	s.SetParcel(parcelID[0], &types.Parcel{
		Owner:        alice.addr,
		Custody:      custody[0],
		ProxyAccount: bob.addr,
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

func TestRegister(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// target
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, uint32(123))
	parcelID := append(tmp, []byte("parcel")...)

	// wrong proxy_account
	payload, _ := json.Marshal(RegisterParam{
		Target:       parcelID,
		Custody:      []byte("custody"),
		ProxyAccount: []byte("wrong address"),
		Extra:        []byte(`"any json"`),
	})
	t0 := makeTestTx("register", "seller", payload)
	rc, _ := t0.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)

	// empty proxy_account
	payload, _ = json.Marshal(RegisterParam{
		Target:  parcelID,
		Custody: []byte("custody"),
		Extra:   []byte(`"any json"`),
	})
	t0 = makeTestTx("register", "seller", payload)
	rc, _ = t0.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	payload, _ = json.Marshal(RegisterParam{
		Target:       parcelID,
		Custody:      []byte("custody"),
		ProxyAccount: bob.addr,
		Extra:        []byte(`"any json"`),
	})
	t1 := makeTestTx("register", "seller", payload)
	rc, _ = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	// register before storage setup
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNoStorage, rc)

	// dummy storage setup
	mysto := &types.Storage{Active: false}
	assert.NoError(t, s.SetStorage(uint32(123), mysto))

	// register with inactive storage
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNoStorage, rc)

	// dummy storage setup
	mysto = &types.Storage{
		Owner:           makeAccAddr("provider"),
		Url:             "http://dummy",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(1),
		Active:          true,
	}
	assert.NoError(t, s.SetStorage(uint32(123), mysto))

	// register with active storage but not enough balance for registration fee
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)

	// with some balance, do it again
	s.SetBalance(makeAccAddr("seller"), new(types.Currency).SetAMO(1))
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	// check parcel
	parcel := s.GetParcel(parcelID, false)
	assert.NotNil(t, parcel)
	assert.Equal(t, []byte(`"any json"`), []byte(parcel.Extra.Register))
	// check balances
	bal := s.GetBalance(makeAccAddr("seller"), false)
	assert.Equal(t, types.Zero, bal)
	bal = s.GetBalance(makeAccAddr("provider"), false)
	assert.Equal(t, new(types.Currency).SetAMO(1), bal)

	// update already registered parcel
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestRequest(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// target
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, uint32(123))
	parcelID := append(tmp, []byte("parcel")...)

	payload, _ := json.Marshal(RequestParam{
		Target:  parcelID,
		Payment: *new(types.Currency).SetAMO(1),
		Extra:   []byte(`"any json for req"`),
	})
	t1 := makeTestTx("request", "buyer", payload)
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	// request for non-existent parcel
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeParcelNotFound, rc)

	// request for buyer owned parcel
	s.SetParcel(parcelID, &types.Parcel{
		Owner:        makeAccAddr("buyer"),
		Custody:      []byte("custody"),
		ProxyAccount: makeAccAddr("proxy"),
		Extra:        types.Extra{},
	})
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeSelfTransaction, rc)

	// request for already granted parcel
	s.SetParcel(parcelID, &types.Parcel{
		Owner:        makeAccAddr("seller"),
		Custody:      []byte("custody"),
		ProxyAccount: makeAccAddr("proxy"),
		Extra: types.Extra{
			Register: []byte(`"any json for reg"`),
		},
	})
	s.SetUsage(makeAccAddr("buyer"), parcelID, &types.Usage{
		Custody: []byte("custody"),
		Extra:   types.Extra{},
	})
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeAlreadyGranted, rc)
	// clean-up
	s.DeleteUsage(makeAccAddr("buyer"), parcelID)

	// not enough balance
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)

	// with some balance, do it again
	s.SetBalance(makeAccAddr("buyer"), new(types.Currency).SetAMO(2))
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	bal := s.GetBalance(makeAccAddr("buyer"), false)
	assert.Equal(t, new(types.Currency).SetAMO(1), bal)
	req := s.GetRequest(makeAccAddr("buyer"), parcelID, false)
	assert.NotNil(t, req)
	assert.Equal(t, []byte(`"any json for reg"`), []byte(req.Extra.Register))
	assert.Equal(t, []byte(`"any json for req"`), []byte(req.Extra.Request))

	// request for already requested parcel
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeAlreadyRequested, rc)
	// clean-up
	s.DeleteRequest(makeAccAddr("buyer"), parcelID)

	// dealer fee
	// XXX: dealer fee is optional, so the previous tests without dealer fee
	// are valid also.
	payload2, _ := json.Marshal(RequestParam{
		Target:    parcelID,
		Payment:   *new(types.Currency).SetAMO(25),
		Dealer:    makeAccAddr("dealer"),
		DealerFee: *new(types.Currency).SetAMO(50),
		Extra:     []byte(`"any json for req"`),
	})
	t2 := makeTestTx("request", "buyer", payload2)
	rc, _ = t2.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	// at this point, buyer's balance is 1 AMO. not enough
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
	// do again with more money
	s.SetBalance(makeAccAddr("buyer"), new(types.Currency).SetAMO(75))
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	// check balance
	bal = s.GetBalance(makeAccAddr("buyer"), false)
	assert.Equal(t, types.Zero, bal)
	req = s.GetRequest(makeAccAddr("buyer"), parcelID, false)
	assert.Equal(t, new(types.Currency).SetAMO(25), &req.Payment)
	assert.Equal(t, new(types.Currency).SetAMO(50), &req.DealerFee)
}

func TestGrant(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// target
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, uint32(123))
	parcelID := append(tmp, []byte("parcel")...)

	payload, _ := json.Marshal(GrantParam{
		Target:  parcelID,
		Grantee: makeAccAddr("buyer"),
		Custody: []byte("custody"),
		Extra:   []byte(`"any json for grant"`),
	})
	t1 := makeTestTx("grant", "seller", payload)
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	// grant for non-existent parcel
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeParcelNotFound, rc)

	// grant for non-existent request
	s.SetParcel(parcelID, &types.Parcel{
		Owner:        makeAccAddr("seller"),
		Custody:      []byte("custody"),
		ProxyAccount: makeAccAddr("proxy"),
		Extra: types.Extra{
			Register: []byte(`"any json for reg"`),
		},
	})
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeRequestNotFound, rc)

	// grant for already granted parcel
	s.SetUsage(makeAccAddr("buyer"), parcelID, &types.Usage{})
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeAlreadyGranted, rc)
	// clean-up
	s.DeleteUsage(makeAccAddr("buyer"), parcelID)

	// grant without permission
	t2 := makeTestTx("grant", "bogus", payload)
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)

	s.SetRequest(makeAccAddr("buyer"), parcelID, &types.Request{
		Payment:   *new(types.Currency).SetAMO(1),
		Dealer:    makeAccAddr("dealer"),
		DealerFee: *new(types.Currency).SetAMO(1),
		Extra: types.Extra{
			Register: []byte(`"any json for reg"`),
			Request:  []byte(`"any json for req"`),
		},
	})
	// register before storage setup
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNoStorage, rc)

	// dummy storage setup
	assert.NoError(t, s.SetStorage(uint32(123), &types.Storage{
		Owner:           makeAccAddr("provider"),
		Url:             "http://dummy",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(2),
		Active:          true,
	}))

	// owner's grant
	// t1
	// proxy's grant
	t2 = makeTestTx("grant", "proxy", payload)

	// owner's: not enough balance
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
	// proxy's: not enough balance
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)

	// again with some balance
	s.SetBalance(makeAccAddr("seller"), new(types.Currency).SetAMO(1))
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	bal := s.GetBalance(makeAccAddr("seller"), false)
	assert.Equal(t, types.Zero, bal)
	bal = s.GetBalance(makeAccAddr("dealer"), false)
	assert.Equal(t, new(types.Currency).SetAMO(1), bal)
	// check extras
	usage := s.GetUsage(makeAccAddr("buyer"), parcelID, false)
	assert.Equal(t, []byte(`"any json for reg"`), []byte(usage.Extra.Register))
	assert.Equal(t, []byte(`"any json for req"`), []byte(usage.Extra.Request))
	assert.Equal(t, []byte(`"any json for grant"`), []byte(usage.Extra.Grant))
}

func TestValidRevoke(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	s.SetParcel(parcelID[0], &types.Parcel{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	s.SetUsage(bob.addr, parcelID[0], &types.Usage{
		Custody: custody[0],
	})
	s.SetParcel(parcelID[1], &types.Parcel{
		Owner:        alice.addr,
		Custody:      custody[1],
		ProxyAccount: carol.addr,
	})
	s.SetUsage(bob.addr, parcelID[1], &types.Usage{
		Custody: custody[1],
	})

	// target
	param := RevokeParam{
		Grantee: bob.addr,
		Target:  parcelID[0],
	}
	payload, _ := json.Marshal(param)
	t1 := makeTestTx("revoke", "alice", payload)

	param = RevokeParam{
		Grantee: bob.addr,
		Target:  parcelID[1],
	}
	payload, _ = json.Marshal(param)
	t2 := makeTestTx("revoke", "carol", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _ = t2.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestNonValidRevoke(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	s.SetParcel(parcelID[0], &types.Parcel{
		Owner:        alice.addr,
		Custody:      custody[0],
		ProxyAccount: carol.addr,
	})
	s.SetUsage(bob.addr, parcelID[0], &types.Usage{
		Custody: custody[0],
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
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
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

	aliceBal := s.GetBalance(makeTestAddress("alice"), false)
	assert.Equal(t, new(types.Currency).Set(0), aliceBal)

	bobBal := s.GetBalance(bob.addr, false)
	assert.Equal(t, new(types.Currency).Set(1230), bobBal)
}

func TestNonValidTransfer(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)

	// target
	payload, _ := json.Marshal(TransferParam{
		To:     []byte("bob"),
		Amount: *new(types.Currency).Set(1230),
	})
	t1 := makeTestTx("transfer", "alice", payload)

	payload, _ = json.Marshal(TransferParam{
		To:     bob.addr,
		Amount: *new(types.Currency).Set(500),
	})
	t2 := makeTestTx("transfer", "bob", payload)

	payload, _ = json.Marshal(TransferParam{
		To:     eve.addr,
		Amount: *new(types.Currency).Set(10),
	})
	t3 := makeTestTx("transfer", "eve", payload)

	payload, _ = json.Marshal(TransferParam{
		To:     alice.addr,
		Amount: *new(types.Currency).Set(0),
	})
	t4 := makeTestTx("transfer", "eve", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeBadParam, rc)
	rc, _ = t2.Check()
	assert.Equal(t, code.TxCodeSelfTransaction, rc)
	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)
	rc, _, _ = t4.Execute(s)
	assert.Equal(t, code.TxCodeInvalidAmount, rc)

	aliceBal := s.GetBalance(makeTestAddress("alice"), false)
	assert.Equal(t, new(types.Currency).Set(0), aliceBal)

	bobBal := s.GetBalance(makeTestAddress("bob"), false)
	assert.Equal(t, new(types.Currency).Set(0), bobBal)

	eveBal := s.GetBalance(makeTestAddress("eve"), false)
	assert.Equal(t, new(types.Currency).Set(0), eveBal)
}

func TestValidStake(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	s.SetBalanceUint64(alice.addr, 3000)
	ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(500)

	validator := tmrand.Bytes(32)

	// target
	payload, _ := json.Marshal(StakeParam{
		Validator: validator,
		Amount:    *new(types.Currency).Set(2000),
	})
	t1 := makeTestTx("stake", "alice", payload)

	payload, _ = json.Marshal(StakeParam{
		Validator: validator,
		Amount:    *new(types.Currency).Set(1000),
	})
	t2 := makeTestTx("stake", "alice", payload)

	// test
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	ConfigAMOApp.LockupPeriod += 1 // manipulate

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	stake := s.GetStake(alice.addr, false)
	assert.Equal(t, *new(types.Currency).Set(3000), stake.Amount)
}

func TestNonValidStake(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	s.SetBalanceUint64(alice.addr, 1000)
	ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(500)

	// target
	payload, _ := json.Marshal(StakeParam{
		Validator: tmrand.Bytes(32),
		Amount:    *new(types.Currency).Set(0),
	})

	t1 := makeTestTx("stake", "alice", payload)

	payload, _ = json.Marshal(StakeParam{
		Validator: tmrand.Bytes(32),
		Amount:    *new(types.Currency).Set(2000),
	})

	t2 := makeTestTx("stake", "eve", payload)
	t3 := makeTestTx("stake", "alice", payload)

	// test
	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodeInvalidAmount, rc)

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)

	// env
	s.SetBalanceUint64(eve.addr, 2000)
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	// env
	s.SetBalanceUint64(alice.addr, 2000)

	// test
	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)

	// env
	payload, _ = json.Marshal(StakeParam{
		Validator: tmrand.Bytes(32),
		Amount:    *new(types.Currency).Set(2345),
	})
	s.SetBalanceUint64(eve.addr, 3000)

	t4 := makeTestTx("stake", "eve", payload)

	// test
	rc, _, _ = t4.Execute(s)
	assert.Equal(t, code.TxCodeImproperStakeAmount, rc)
}

func TestValidWithdraw(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))
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
	assert.Equal(t, new(types.Currency).Set(1000), &s.GetStake(alice.addr, false).Amount)

	// add more stakeholder to test stake deletion
	//var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))
	s.SetUnlockedStake(bob.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})

	// test
	rc, _ = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Nil(t, s.GetStake(alice.addr, false))
	assert.NotNil(t, s.GetStake(bob.addr, false))
}

func TestNonValidWithdraw(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})

	// target
	payload, _ := json.Marshal(WithdrawParam{
		Amount: *new(types.Currency).Set(0),
	})
	t1 := makeTestTx("withdraw", "alice", payload)

	payload, _ = json.Marshal(WithdrawParam{
		Amount: *new(types.Currency).Set(2000),
	})
	t2 := makeTestTx("withdraw", "eve", payload)
	t3 := makeTestTx("withdraw", "alice", payload)

	// test
	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodeInvalidAmount, rc)

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeNoStake, rc)

	rc, _ = t3.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeLastValidator, rc)

	// env
	s.SetDelegate(bob.addr, &types.Delegate{
		Delegatee: alice.addr,
		Amount:    *new(types.Currency).Set(500),
	})

	// test
	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeDelegateExists, rc)
}

func TestValidDelegate(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	s.SetBalanceUint64(bob.addr, 1000)
	ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(500)

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
	assert.Equal(t, new(types.Currency).Set(1000), &s.GetDelegate(bob.addr, false).Amount)
}

func TestNonValidDelegate(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	copy(k[:], tmrand.Bytes(32))
	s.SetUnlockedStake(eve.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	s.SetBalanceUint64(alice.addr, 1000)
	s.SetBalanceUint64(bob.addr, 1000)
	ConfigAMOApp.MinStakingUnit = *new(types.Currency).Set(500)

	// test
	payload, _ := json.Marshal(DelegateParam{
		Amount: *new(types.Currency).Set(500),
		To:     eve.addr,
	})
	t1 := makeTestTx("delegate", "eve", payload)
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeSelfTransaction, rc)

	payload, _ = json.Marshal(DelegateParam{
		Amount: *new(types.Currency).Set(0),
		To:     alice.addr,
	})
	t1 = makeTestTx("delegate", "eve", payload)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeInvalidAmount, rc)

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

	payload, _ = json.Marshal(DelegateParam{
		Amount: *new(types.Currency).Set(543),
		To:     bob.addr,
	})
	t1 = makeTestTx("delegate", "alice", payload)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeImproperStakeAmount, rc)
}

func TestValidRetract(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))
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
	assert.Equal(t, new(types.Currency).Set(100), &s.GetDelegate(bob.addr, false).Amount)

	// test
	payload, _ = json.Marshal(RetractParam{
		Amount: *new(types.Currency).Set(100),
	})
	t1 = makeTestTx("retract", "bob", payload)

	rc, _ = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	assert.Nil(t, s.GetDelegate(bob.addr, false))

	assert.Equal(t, new(types.Currency).Set(2000), &s.GetStake(alice.addr, false).Amount)
}

func TestNonValidRetract(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))
	s.SetUnlockedStake(alice.addr, &types.Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: k,
	})
	s.SetDelegate(bob.addr, &types.Delegate{
		Delegatee: alice.addr,
		Amount:    *new(types.Currency).Set(500),
	})

	payload, _ := json.Marshal(RetractParam{
		Amount: *new(types.Currency).Set(0),
	})

	t1 := makeTestTx("retract", "bob", payload)

	payload, _ = json.Marshal(RetractParam{
		Amount: *new(types.Currency).Set(500),
	})

	t2 := makeTestTx("retract", "eve", payload)
	t3 := makeTestTx("retract", "bob", payload)

	rc, _, _ := t1.Execute(s)
	assert.Equal(t, code.TxCodeInvalidAmount, rc)

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeDelegateNotFound, rc)

	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeDelegateNotFound, rc)
}

func TestStakeLockup(t *testing.T) {
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	s.SetBalanceUint64(alice.addr, 3000)

	// setup lock-up period config
	ConfigAMOApp.LockupPeriod = 2

	// deposit stake
	stakeParam := StakeParam{
		Validator: tmrand.Bytes(32),
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
	s.LoosenLockedStakes(false)
	s.LoosenLockedStakes(false)

	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	stake := s.GetStake(makeTestAddress("alice"), false)
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

func TestPropose(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)
	ConfigAMOApp = types.AMOAppConfig{
		MaxValidators:         uint64(100),
		WeightValidator:       float64(2),
		WeightDelegator:       float64(1),
		MinStakingUnit:        *new(types.Currency).Set(100),
		BlkReward:             *new(types.Currency).Set(1000),
		TxReward:              *new(types.Currency).Set(1000),
		PenaltyRatioM:         float64(0.1),
		PenaltyRatioL:         float64(0.1),
		LazinessCounterWindow: int64(10000),
		LazinessThreshold:     float64(0.9),
		BlockBindingWindow:    int64(10000),
		LockupPeriod:          int64(10000),
		DraftOpenCount:        int64(10000),
		DraftCloseCount:       int64(10000),
		DraftApplyCount:       int64(10000),
		DraftDeposit:          *new(types.Currency).Set(1000),
		DraftQuorumRate:       float64(0.1),
		DraftPassRate:         float64(0.7),
		DraftRefundRate:       float64(0.2),
	}

	// target
	payload, _ := json.Marshal(ProposeParam{
		DraftID: uint32(2),
		Config:  []byte(`{"min_staking_unit": "0"}`),
		Desc:    "any json",
	})
	t1 := makeTestTx("propose", "proposer", payload)
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	// propose before proposer acquires permission
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)

	// proposer stake
	var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))

	assert.NoError(t, s.SetUnlockedStake(makeAccAddr("proposer"), &types.Stake{
		Validator: k,
		Amount:    *new(types.Currency).Set(10000000),
	}))

	// propose with improper draft id
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeImproperDraftID, rc)

	// imitate next draft id for test
	StateNextDraftID = 2

	// propose without having enough balance in proposer's account
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNotEnoughBalance, rc)

	// give some balance
	s.SetBalance(makeAccAddr("proposer"), new(types.Currency).Set(1000))

	// propose draft with improper draft config
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeImproperDraftConfig, rc)

	// modify draft config to make it proper
	payload, _ = json.Marshal(ProposeParam{
		DraftID: uint32(2),
		Config:  []byte(`{"min_staking_unit": "100"}`),
		Desc:    "any json",
	})
	t1 = makeTestTx("propose", "proposer", payload)
	rc, _ = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	// propose draft with proper draft config
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	// check balance
	bal := s.GetBalance(makeAccAddr("proposer"), false)
	assert.Equal(t, types.Zero, bal)

	// propose same draft by proposerDup
	copy(k[:], tmrand.Bytes(32))
	s.SetBalance(makeAccAddr("proposerDup"), new(types.Currency).Set(1000))
	assert.NoError(t, s.SetUnlockedStake(makeAccAddr("proposerDup"), &types.Stake{
		Validator: k,
		Amount:    *new(types.Currency).Set(10000000),
	}))
	t1 = makeTestTx("propose", "proposerDup", payload)
	rc, _ = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeProposedDraft, rc)

	// propose other draft while there exists a draft in process
	StateNextDraftID = 3
	payload, _ = json.Marshal(ProposeParam{
		DraftID: uint32(3),
		Config:  []byte(`{"tx_reward": "0"}`),
		Desc:    "i don't want other vals to earn tx rewards",
	})
	t1 = makeTestTx("propose", "proposerDup", payload)
	rc, _ = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeAnotherDraftInProcess, rc)

	// imitate next draft id for test
	StateNextDraftID = 4

	// propose a draft having config left empty on purpose
	s.SetBalance(makeAccAddr("proposer"), new(types.Currency).Set(1000))
	payload, _ = json.Marshal(ProposeParam{
		DraftID: uint32(4),
		Config:  []byte(``),
		Desc:    "empty config is used to give an opinion",
	})
	t1 = makeTestTx("propose", "proposer", payload)
	rc, _ = t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
}

func TestVote(t *testing.T) {
	// env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// target
	payload, _ := json.Marshal(VoteParam{
		DraftID: uint32(1),
		Approve: true,
	})
	t1 := makeTestTx("vote", "voter1", payload)
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)

	// voter1 vote without permission (without stake)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)

	// voter1 stake
	var k ed25519.PubKeyEd25519
	copy(k[:], tmrand.Bytes(32))
	assert.NoError(t, s.SetUnlockedStake(makeAccAddr("voter1"), &types.Stake{
		Validator: k,
		Amount:    *new(types.Currency).Set(10000000),
	}))

	// voter1 tries to vote for non-existing draft
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeNonExistingDraft, rc)

	// set dummy draft
	cfg := types.AMOAppConfig{
		MaxValidators:         uint64(100),
		WeightValidator:       float64(2),
		WeightDelegator:       float64(1),
		MinStakingUnit:        *new(types.Currency).Set(100),
		BlkReward:             *new(types.Currency).Set(1000),
		TxReward:              *new(types.Currency).Set(1000),
		PenaltyRatioM:         float64(0.1),
		PenaltyRatioL:         float64(0.1),
		LazinessCounterWindow: int64(10000),
		LazinessThreshold:     float64(0.9),
		BlockBindingWindow:    int64(10000),
		LockupPeriod:          int64(10000),
		DraftOpenCount:        int64(10000),
		DraftCloseCount:       int64(10000),
		DraftApplyCount:       int64(10000),
		DraftDeposit:          *new(types.Currency).Set(1000),
		DraftQuorumRate:       float64(0.1),
		DraftPassRate:         float64(0.7),
		DraftRefundRate:       float64(0.2),
	}

	StateNextDraftID = uint32(1)
	draftID := StateNextDraftID
	s.SetDraft(draftID, &types.Draft{
		Proposer: makeAccAddr("proposer"),
		Config:   cfg,
		Desc:     "any desc",

		// imitate beginning of draft vote situation
		OpenCount:  int64(0),
		CloseCount: int64(1000),
		ApplyCount: int64(10000),
		Deposit:    *new(types.Currency).Set(100),

		TallyQuorum:  *new(types.Currency).Set(0),
		TallyApprove: *new(types.Currency).Set(0),
		TallyReject:  *new(types.Currency).Set(0),
	})

	// proposer stake
	copy(k[:], tmrand.Bytes(32))
	assert.NoError(t, s.SetUnlockedStake(makeAccAddr("proposer"), &types.Stake{
		Validator: k,
		Amount:    *new(types.Currency).Set(10000000),
	}))

	// proposer tries to vote on his own draft
	t2 := makeTestTx("vote", "proposer", payload)
	rc, _ = t2.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeSelfTransaction, rc)

	StateNextDraftID = uint32(2)

	// voter1 vote again
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)

	// voter1 tries to already voted draft vote
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodeAlreadyVoted, rc)

	// voter2 stake
	copy(k[:], tmrand.Bytes(32))
	assert.NoError(t, s.SetUnlockedStake(makeAccAddr("voter2"), &types.Stake{
		Validator: k,
		Amount:    *new(types.Currency).Set(10000000),
	}))

	// relpace draft value to imitate ending of draft vote
	s.SetDraft(draftID, &types.Draft{
		Proposer: makeAccAddr("proposer"),
		Config:   cfg,
		Desc:     "any desc",

		// imitate ending of draft vote situation
		OpenCount:  int64(0),
		CloseCount: int64(0),
		ApplyCount: int64(10000),
		Deposit:    *new(types.Currency).Set(100),

		TallyQuorum:  *new(types.Currency).Set(0),
		TallyApprove: *new(types.Currency).Set(0),
		TallyReject:  *new(types.Currency).Set(0),
	})

	// voter2 tries to vote for closed draft vote
	t3 := makeTestTx("vote", "voter2", payload)
	rc, _ = t3.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t3.Execute(s)
	assert.Equal(t, code.TxCodeVoteNotOpen, rc)
}
