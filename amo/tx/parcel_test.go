package tx

import (
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

func TestParseTransfer(t *testing.T) {
	bytes := []byte(`{"type":"transfer","sender":"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","payload":{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","parcel":"00000010EFEF"},"signature":{"pubkey":"0485FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B185FE85FCE6AB426563E5E085FE85FCE6AB426563E5E0749EBCB95E9B1EF1D55E9B1EF1D","sig_bytes":"FFFFFFFF"}}`)
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
	var parcel tmbytes.HexBytes
	err = json.Unmarshal([]byte(`"00000010EFEF"`), &parcel)
	assert.NoError(t, err)

	expected := &TxTransfer{
		TxBase{
			Type:    "transfer",
			Sender:  sender,
			Payload: []byte(`{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","parcel":"00000010EFEF"}`),
			Signature: Signature{
				PubKey:   pubkey,
				SigBytes: sigbytes,
			},
		},
		TransferParam{
			To:     to,
			Parcel: parcel,
		},
	}
	parsedTx, err := ParseTx(bytes)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsedTx)
}

func TestTransfer(t *testing.T) {
	// prepare env
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// prepare test parcel
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, uint32(123)) // storage id
	parcelID := append(tmp, []byte("parcel")...) // in-storage id
	s.SetParcel(parcelID, &types.Parcel{
		Owner:        alice.addr,
		Custody:      []byte("custody"),
	})

	// prepare test tx payload
	payload, _ := json.Marshal(TransferParam{
		To:     bob.addr,
		Parcel: parcelID,
	})

	// wrong ownership
	t1 := makeTestTx("transfer", "carol", payload)
	rc, _ := t1.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t1.Execute(s)
	assert.Equal(t, code.TxCodePermissionDenied, rc)
	// check result: no change in ownership
	parcel := s.GetParcel(parcelID, false)
	assert.NotNil(t, parcel)
	assert.Equal(t, alice.addr, parcel.Owner)

	// right ownership
	t2 := makeTestTx("transfer", "alice", payload)
	rc, _ = t2.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = t2.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	// check result
	parcel = s.GetParcel(parcelID, false)
	assert.NotNil(t, parcel)
	assert.Equal(t, bob.addr, parcel.Owner)
}
