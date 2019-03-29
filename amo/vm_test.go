package amo

import (
	"bytes"
	"encoding/json"

	"testing"

	"github.com/stretchr/testify/assert"
	tm "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/common"
	tdb "github.com/tendermint/tendermint/libs/db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/operation"
	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

type user struct {
	key     p256.PrivKeyP256
	nodeKey common.HexBytes
	p       uint64
}

func getUser() user {
	return user{
		key:     p256.GenPrivKey(),
		nodeKey: common.RandBytes(32),
		p:       common.RandUint64(),
	}
}

func getTestApp() *AMOApplication {
	app := NewAMOApplication(tdb.NewMemDB(), tdb.NewMemDB(), nil)
	return app
}

func generateStake(u user, t *testing.T) []byte {
	op := operation.Stake{
		Amount:    *new(types.Currency).Set(u.p),
		Validator: u.nodeKey,
	}
	b, _ := json.Marshal(op)
	msg := operation.Message{
		Type:   operation.TxStake,
		Params: b,
	}
	err := msg.Sign(u.key)
	assert.NoError(t, err)
	b, _ = json.Marshal(msg)
	return b
}

func TestIndex(t *testing.T) {
	const n = 1000
	app := getTestApp()
	u := make([]user, n)
	x, _ := new(types.Currency).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16)
	for i := 0; i < n; i++ {
		u[i] = getUser()
		app.store.SetBalance(u[i].key.PubKey().Address(), x)
		b := generateStake(u[i], t)
		assert.Equal(t, code.TxCodeOK, app.DeliverTx(b).Code)
	}
	var r user
	for i, j:= 1, 0; i < n; i++ {
		j = i
		r = u[j]
		j--
		for j >= 0 && r.p > u[j].p {
			u[j+1] = u[j]
			u[j] = r
			j--
		}
	}
	res := app.EndBlock(tm.RequestEndBlock{
		Height: 1,
	})
	for i, v := range res.ValidatorUpdates {
		assert.True(t, bytes.Equal(v.PubKey.Data, u[i].nodeKey))
	}
}
