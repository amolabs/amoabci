package tx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	tm "github.com/tendermint/tendermint/libs/common"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

func TestParseSetup(t *testing.T) {
	b := tm.HexBytes([]byte("aaaa"))
	id, _ := json.Marshal(b)
	payload := []byte(`{"storage":` + string(id) + `,"url": "http://need_to_check_url_format","registration_fee":"1000000000000000000","hosting_fee":"1000000000000000000"}`)

	expected := SetupParam{
		Storage:         b,
		Url:             "http://need_to_check_url_format",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(1),
	}
	txParam, err := parseSetupParam(payload)
	assert.NoError(t, err)
	assert.Equal(t, expected, txParam)
}

func TestParseClose(t *testing.T) {
	b := tm.HexBytes([]byte("aaaa"))
	id, _ := json.Marshal(b)
	payload := []byte(`{"storage":` + string(id) + `}`)

	expected := CloseParam{
		Storage: b,
	}
	txParam, err := parseCloseParam(payload)
	assert.NoError(t, err)
	assert.Equal(t, expected, txParam)
}

func TestTxSetup(t *testing.T) {
	s := store.NewStore(
		tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NotNil(t, s)

	// initial setup
	param := SetupParam{
		Storage:         []byte("aaaa"),
		Url:             "http://need_to_check_url_format",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(1),
	}
	payload, _ := json.Marshal(param)
	//
	tx := makeTestTx("setup", "provider", payload)
	assert.NotNil(t, tx)
	_, ok := tx.(*TxSetup)
	assert.True(t, ok)
	rc, _ := tx.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	// check store
	sto := s.GetStorage([]byte("aaaa"), false)
	assert.NotNil(t, sto)
	assert.Equal(t, &types.Storage{
		Owner:           makeAccAddr("provider"),
		Url:             "http://need_to_check_url_format",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(1),
		Active:          true,
	}, sto)

	// close
	param2 := CloseParam{
		Storage: []byte("aaaa"),
	}
	payload, _ = json.Marshal(param2)
	//
	tx = makeTestTx("close", "provider", payload)
	assert.NotNil(t, tx)
	_, ok = tx.(*TxClose)
	assert.True(t, ok)
	rc, _ = tx.Check()
	assert.Equal(t, code.TxCodeOK, rc)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	// check whether closed
	sto = s.GetStorage([]byte("aaaa"), false)
	assert.NotNil(t, sto)
	assert.Equal(t, &types.Storage{
		Owner:           makeAccAddr("provider"),
		Url:             "http://need_to_check_url_format",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(1),
		Active:          false,
	}, sto)
	// following-up setup
	payload, _ = json.Marshal(param)
	//
	tx = makeTestTx("setup", "provider", payload)
	rc, _, _ = tx.Execute(s)
	assert.Equal(t, code.TxCodeOK, rc)
	sto = s.GetStorage([]byte("aaaa"), false)
	assert.NotNil(t, sto)
	assert.Equal(t, &types.Storage{
		Owner:           makeAccAddr("provider"),
		Url:             "http://need_to_check_url_format",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(1),
		Active:          true,
	}, sto)
}
