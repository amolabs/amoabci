package tx

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

func TestParseSetup(t *testing.T) {
	id := uint32(1)
	payload := []byte(fmt.Sprintf(`{"storage": %d,"url": "http://need_to_check_url_format","registration_fee":"1000000000000000000","hosting_fee":"1000000000000000000"}`, id))

	expected := SetupParam{
		Storage:         id,
		Url:             "http://need_to_check_url_format",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(1),
	}
	txParam, err := parseSetupParam(payload)
	assert.NoError(t, err)
	assert.Equal(t, expected, txParam)
}

func TestParseClose(t *testing.T) {
	id := uint32(1)
	payload := []byte(fmt.Sprintf(`{"storage": %d}`, id))

	expected := CloseParam{
		Storage: 1,
	}
	txParam, err := parseCloseParam(payload)
	assert.NoError(t, err)
	assert.Equal(t, expected, txParam)
}

func TestTxSetup(t *testing.T) {
	s, err := store.NewStore(nil, 1, tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	storageID := uint32(1)

	// initial setup
	param := SetupParam{
		Storage:         storageID,
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
	sto := s.GetStorage(storageID, false)
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
		Storage: storageID,
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
	sto = s.GetStorage(storageID, false)
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
	sto = s.GetStorage(storageID, false)
	assert.NotNil(t, sto)
	assert.Equal(t, &types.Storage{
		Owner:           makeAccAddr("provider"),
		Url:             "http://need_to_check_url_format",
		RegistrationFee: *new(types.Currency).SetAMO(1),
		HostingFee:      *new(types.Currency).SetAMO(1),
		Active:          true,
	}, sto)
}
