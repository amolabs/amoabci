package operation

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
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
		Amount:    *new(types.Currency).Set(500),
		Delegator: alice.addr,
	})
	return s
}

func TestValidCancel(t *testing.T) {
	s := getTestStore()
	op := Cancel{
		parcelID[0],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, bob.addr))
	resCode := op.Execute(s, bob.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidCancel(t *testing.T) {
	s := getTestStore()
	op := Cancel{
		parcelID[0],
	}
	assert.Equal(t, code.TxCodeTargetNotExists, op.Check(s, eve.addr))
}

func TestValidDiscard(t *testing.T) {
	s := getTestStore()
	op := Discard{
		parcelID[0],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode := op.Execute(s, alice.addr)
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
	assert.Equal(t, code.TxCodeTargetNotExists, NEOp.Check(s, alice.addr))
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
	resCode := op.Execute(s, bob.addr)
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
	assert.Equal(t, code.TxCodeTargetAlreadyExists, AEop.Check(s, alice.addr))
}

func TestValidRegister(t *testing.T) {
	s := getTestStore()
	op := Register{
		Target:  parcelID[2],
		Custody: custody[2],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode := op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidRegister(t *testing.T) {
	s := getTestStore()
	op := Register{
		Target:  parcelID[0],
		Custody: custody[0],
	}
	assert.Equal(t, code.TxCodeTargetAlreadyExists, op.Check(s, alice.addr))
}

func TestValidRequest(t *testing.T) {
	s := getTestStore()
	op := Request{
		Target:  parcelID[1],
		Payment: *new(types.Currency).Set(200),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode := op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidRequest(t *testing.T) {
	s := getTestStore()
	TNop := Request{
		Target:  []byte{0x0, 0x0, 0x0, 0x0},
		Payment: *new(types.Currency).Set(100),
	}
	TAop := Request{
		Target:  parcelID[0],
		Payment: *new(types.Currency).Set(100),
	}
	STop := Request{
		Target:  parcelID[1],
		Payment: *new(types.Currency).Set(100),
	}
	NBop := Request{
		Target:  parcelID[1],
		Payment: *new(types.Currency).Set(100),
	}
	assert.Equal(t, code.TxCodeTargetNotExists, TNop.Check(s, eve.addr))
	assert.Equal(t, code.TxCodeTargetAlreadyBought, TAop.Check(s, bob.addr))
	assert.Equal(t, code.TxCodeSelfTransaction, STop.Check(s, bob.addr))
	assert.Equal(t, code.TxCodeNotEnoughBalance, NBop.Check(s, eve.addr))
}

func TestValidRevoke(t *testing.T) {
	s := getTestStore()
	op := Revoke{
		Grantee: bob.addr,
		Target:  parcelID[0],
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode := op.Execute(s, alice.addr)
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
	assert.Equal(t, code.TxCodeTargetNotExists, TNop.Check(s, alice.addr))
}

func TestValidTransfer(t *testing.T) {
	s := getTestStore()
	op := Transfer{
		To:     bob.addr,
		Amount: *new(types.Currency).Set(1230),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode := op.Execute(s, alice.addr)
	assert.Equal(t, code.TxCodeOK, resCode)
}

func TestNonValidTransfer(t *testing.T) {
	s := getTestStore()
	BPop := Transfer{
		To:     []byte("bob"),
		Amount: *new(types.Currency).Set(1230),
	}
	NEop := Transfer{
		To:     bob.addr,
		Amount: *new(types.Currency).Set(500),
	}
	STop := Transfer{
		To:     eve.addr,
		Amount: *new(types.Currency).Set(10),
	}
	assert.Equal(t, code.TxCodeBadParam, BPop.Check(s, alice.addr))
	assert.Equal(t, code.TxCodeNotEnoughBalance, NEop.Check(s, eve.addr))
	assert.Equal(t, code.TxCodeSelfTransaction, STop.Check(s, eve.addr))
}

func TestValidStake(t *testing.T) {
	s := getTestStore()
	op := Stake{
		Amount:    *new(types.Currency).Set(2000),
		Validator: cmn.RandBytes(32),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	resCode := op.Execute(s, alice.addr)
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
	s := getTestStore()
	op := Withdraw{
		Amount: *new(types.Currency).Set(1000),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	assert.Equal(t, code.TxCodeOK, op.Execute(s, alice.addr))
	assert.Equal(t, new(types.Currency).Set(1000), &s.GetStake(alice.addr).Amount)

	op = Withdraw{
		Amount: *new(types.Currency).Set(1000),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, alice.addr))
	assert.Equal(t, code.TxCodeOK, op.Execute(s, alice.addr))
	assert.Equal(t, zero, &s.GetStake(alice.addr).Amount)
}

func TestNonValidWithdraw(t *testing.T) {
	s := getTestStore()
	op := Withdraw{
		Amount: *new(types.Currency).Set(2000),
	}
	assert.Equal(t, code.TxCodeNotEnoughBalance, op.Check(s, eve.addr))
}

func TestValidDelegate(t *testing.T) {
	s := getTestStore()
	op := Delegate{
		Amount: *new(types.Currency).Set(500),
		To:     alice.addr,
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, bob.addr))
	resCode := op.Execute(s, bob.addr)
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
	assert.Equal(t, code.TxCodeOK, op.Execute(s, bob.addr))
	assert.Equal(t, new(types.Currency).Set(100), &s.GetDelegate(bob.addr).Amount)

	op = Retract{
		Amount: *new(types.Currency).Set(100),
	}
	assert.Equal(t, code.TxCodeOK, op.Check(s, bob.addr))
	assert.Equal(t, code.TxCodeOK, op.Execute(s, bob.addr))
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
