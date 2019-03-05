package operation

import (
	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/db"
	dtypes "github.com/amolabs/amoabci/amo/db/types"
	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/p256"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type user struct {
	privKey crypto.PrivKey
	pubKey  crypto.PubKey
	addr    crypto.Address
}

var privKeys = []p256.PrivKeyP256{
	p256.GenPrivKeyFromSecret([]byte("alice")),
	p256.GenPrivKeyFromSecret([]byte("bob")),
	p256.GenPrivKeyFromSecret([]byte("eve")),
}

var alice = user{
	privKey: privKeys[0],
	pubKey:  privKeys[0].PubKey(),
	addr:    privKeys[0].PubKey().Address(),
}

var bob = user{
	privKey: privKeys[1],
	pubKey:  privKeys[1].PubKey(),
	addr:    privKeys[1].PubKey().Address(),
}

var eve = user{
	privKey: privKeys[2],
	pubKey:  privKeys[2].PubKey(),
	addr:    privKeys[2].PubKey().Address(),
}

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

func getTestStore() *db.Store {
	store := db.NewMemStore()
	store.SetBalance(alice.addr, 3000)
	store.SetBalance(bob.addr, 1000)
	store.SetBalance(eve.addr, 50)
	store.SetParcel(parcelID[0], &dtypes.ParcelValue{
		Owner:   alice.addr,
		Custody: custody[0],
	})
	store.SetParcel(parcelID[1], &dtypes.ParcelValue{
		Owner:   bob.addr,
		Custody: custody[1],
	})
	store.SetRequest(bob.addr, parcelID[0], &dtypes.RequestValue{
		Payment: 100,
	})
	store.SetRequest(alice.addr, parcelID[1], &dtypes.RequestValue{
		Payment: 100,
	})
	store.SetUsage(bob.addr, parcelID[0], &dtypes.UsageValue{
		Custody: custody[0],
		Exp:     time.Now().UTC().Add(24 * time.Hour),
	})
	return store
}

func TestValidCancel(t *testing.T) {
	store := getTestStore()
	op := Cancel{
		parcelID[0],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(store, bob.addr))
	resCode, _ := op.Execute(store, bob.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidCancel(t *testing.T) {
	store := getTestStore()
	op := Cancel{
		parcelID[0],
	}
	assert.Equal(t, code.TxCodeTargetNotExists, op.Check(store, eve.addr))
}

func TestValidDiscard(t *testing.T) {
	store := getTestStore()
	op := Discard{
		parcelID[0],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(store, alice.addr))
	resCode, _ := op.Execute(store, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidDiscard(t *testing.T) {
	store := getTestStore()
	NEOp := Discard{
		[]byte{0xFF, 0xFF, 0xFF, 0xEE},
	}
	PDOp := Discard{
		parcelID[0],
	}
	assert.Equal(t, code.TxCodeTargetNotExists, NEOp.Check(store, alice.addr))
	assert.Equal(t, code.TxCodePermissionDenied, PDOp.Check(store, eve.addr))
}

func TestValidGrant(t *testing.T) {
	store := getTestStore()
	op := Grant{
		Target:  parcelID[1],
		Grantee: alice.addr,
		Custody: custody[1],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(store, bob.addr))
	resCode, _ := op.Execute(store, bob.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidGrant(t *testing.T) {
	store := getTestStore()
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
	assert.Equal(t, code.TxCodePermissionDenied, PDop.Check(store, eve.addr))
	assert.Equal(t, code.TxCodeTargetAlreadyExists, AEop.Check(store, alice.addr))
}

func TestValidRegister(t *testing.T) {
	store := getTestStore()
	op := Register{
		Target:  parcelID[2],
		Custody: custody[2],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(store, alice.addr))
	resCode, _ := op.Execute(store, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidRegister(t *testing.T) {
	store := getTestStore()
	op := Register{
		Target:  parcelID[0],
		Custody: custody[0],
	}
	assert.Equal(t, code.TxCodeTargetAlreadyExists, op.Check(store, alice.addr))
}

func TestValidRequest(t *testing.T) {
	store := getTestStore()
	op := Request{
		Target:  parcelID[1],
		Payment: 200,
	}
	assert.Equal(t, code.TxCodeOK, op.Check(store, bob.addr))
	resCode, _ := op.Execute(store, bob.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidRequest(t *testing.T) {
	store := getTestStore()
	TNop := Request{
		Target:  []byte{0x0, 0x0, 0x0, 0x0},
		Payment: 100,
	}
	TAop := Request{
		Target:  parcelID[0],
		Payment: 100,
	}
	STop := Request{
		Target:  parcelID[1],
		Payment: 100,
	}
	assert.Equal(t, code.TxCodeTargetNotExists, TNop.Check(store, eve.addr))
	assert.Equal(t, code.TxCodeTargetAlreadyBought, TAop.Check(store, bob.addr))
	assert.Equal(t, code.TxCodeSelfTransaction, STop.Check(store, bob.addr))
}

func TestValidRevoke(t *testing.T) {
	store := getTestStore()
	op := Revoke{
		Grantee: bob.addr,
		Target:  parcelID[0],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(store, alice.addr))
	resCode, _ := op.Execute(store, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidRevoke(t *testing.T) {
	store := getTestStore()
	PDop := Revoke{
		Grantee: eve.addr,
		Target:  parcelID[0],
	}
	TNop := Revoke{
		Grantee: bob.addr,
		Target:  parcelID[2],
	}
	assert.Equal(t, code.TxCodePermissionDenied, PDop.Check(store, eve.addr))
	assert.Equal(t, code.TxCodeTargetNotExists, TNop.Check(store, alice.addr))
}

func TestValidTransfer(t *testing.T) {
	store := getTestStore()
	op := Transfer{
		To:     bob.addr,
		Amount: 1230,
	}
	assert.Equal(t, code.TxCodeOK, op.Check(store, alice.addr))
	resCode, _ := op.Execute(store, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidTransfer(t *testing.T) {
	store := getTestStore()
	BPop := Transfer{
		To:     []byte("bob"),
		Amount: 1230,
	}
	NEop := Transfer{
		To:     bob.addr,
		Amount: 500,
	}
	STop := Transfer{
		To:     eve.addr,
		Amount: 10,
	}
	assert.Equal(t, code.TxCodeBadParam, BPop.Check(store, alice.addr))
	assert.Equal(t, code.TxCodeNotEnoughBalance, NEop.Check(store, eve.addr))
	assert.Equal(t, code.TxCodeSelfTransaction, STop.Check(store, eve.addr))
}
