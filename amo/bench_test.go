package amo

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	tm "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/tx"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

const benchTest = "bench_test"

func setUpBench(b *testing.B) {
	err := tm.EnsureDir(benchTest, 0700)
	assert.NoError(b, err)

	file, err := ioutil.TempFile("", "state_*.json")
	assert.NoError(b, err)

	tmpFile = file
}

func tearDownBench(b *testing.B) {
	err := os.RemoveAll(benchTest)
	assert.NoError(b, err)

	err = os.Remove(tmpFile.Name())
	assert.NoError(b, err)
}

func BenchmarkCheckTransferTx(b *testing.B) {
	setUpBench(b)
	defer tearDownBench(b)

	sdb, err := store.NewDBProxy("state", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, sdb)
	defer sdb.Close()

	idxdb, err := store.NewDBProxy("index", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, idxdb)
	defer idxdb.Close()

	incdb, err := store.NewDBProxy("incentive", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, incdb)
	defer incdb.Close()

	gcdb, err := store.NewDBProxy("group_counter", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, gcdb)
	defer incdb.Close()

	app := NewAMOApp(tmpFile, sdb, idxdb, incdb, gcdb, nil)

	from := p256.GenPrivKeyFromSecret([]byte("alice"))
	//app.store.SetBalanceUint64(from.PubKey().Address(), 1000000000)

	// test tx
	_tx := tx.TransferParam{
		To:     p256.GenPrivKeyFromSecret([]byte("bob")).PubKey().Address(),
		Amount: *new(types.Currency).Set(10),
	}
	payload, _ := json.Marshal(_tx)
	msg := tx.TxBase{
		Type:    "transfer",
		Payload: payload,
		Sender:  from.PubKey().Address(),
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
	}
	msg.Sign(from)
	raw, _ := json.Marshal(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.CheckTx(abci.RequestCheckTx{Tx: raw})
		//app.DeliverTx(abci.RequestDeliverTx{Tx: raw})
	}
}

func BenchmarkDeliverTransferTx(b *testing.B) {
	setUpBench(b)
	defer tearDownBench(b)

	sdb, err := store.NewDBProxy("state", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, sdb)
	defer sdb.Close()

	idxdb, err := store.NewDBProxy("index", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, idxdb)
	defer idxdb.Close()

	incdb, err := store.NewDBProxy("incentive", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, incdb)
	defer incdb.Close()

	gcdb, err := store.NewDBProxy("group_counter", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, gcdb)
	defer incdb.Close()

	app := NewAMOApp(tmpFile, sdb, idxdb, incdb, gcdb, nil)

	from := p256.GenPrivKeyFromSecret([]byte("alice"))
	app.store.SetBalanceUint64(from.PubKey().Address(), 1000000000)

	// test tx
	_tx := tx.TransferParam{
		To:     p256.GenPrivKeyFromSecret([]byte("bob")).PubKey().Address(),
		Amount: *new(types.Currency).Set(10),
	}
	payload, _ := json.Marshal(_tx)
	msg := tx.TxBase{
		Type:    "transfer",
		Payload: payload,
		Sender:  from.PubKey().Address(),
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
	}
	msg.Sign(from)
	raw, _ := json.Marshal(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//app.CheckTx(abci.RequestCheckTx{Tx: raw})
		app.DeliverTx(abci.RequestDeliverTx{Tx: raw})
	}
}
