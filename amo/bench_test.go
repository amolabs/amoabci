package amo

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/tx"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

const benchTest = "bench_test"

func setUpBench(b *testing.B) {
	err := tmos.EnsureDir(benchTest, 0700)
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

	sdb, err := tmdb.NewGoLevelDB("state", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, sdb)
	defer sdb.Close()

	idxdb, err := tmdb.NewGoLevelDB("index", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, idxdb)
	defer idxdb.Close()

	gcdb, err := tmdb.NewGoLevelDB("group_counter", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, gcdb)
	defer gcdb.Close()

	app := NewAMOApp(tmpFile, 1, sdb, idxdb, nil)

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

	sdb, err := tmdb.NewGoLevelDB("state", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, sdb)
	defer sdb.Close()

	idxdb, err := tmdb.NewGoLevelDB("index", benchTest)
	assert.NoError(b, err)
	assert.NotNil(b, idxdb)
	defer idxdb.Close()

	app := NewAMOApp(tmpFile, 1, sdb, idxdb, nil)
	assert.NoError(b, err)

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
	}
	msg.Sign(from)
	raw, _ := json.Marshal(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//app.CheckTx(abci.RequestCheckTx{Tx: raw})
		app.DeliverTx(abci.RequestDeliverTx{Tx: raw})
	}
}
