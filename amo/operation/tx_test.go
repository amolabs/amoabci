package operation

import (
	"encoding/json"
	"github.com/amolabs/tendermint-amo/crypto/p256"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTx(t *testing.T) {
	from := p256.GenPrivKey()
	to := p256.GenPrivKey().PubKey().Address()
	transfer := Transfer{
		To:     to,
		Amount: 100,
	}
	b, _ := json.Marshal(transfer)
	message := Message{
		Command: TxTransfer,
		Payload: b,
	}
	err := message.Sign(from)
	if err != nil {
		panic(err)
	}
	bMsg, _ := json.Marshal(message)
	msg, op := ParseTx(bMsg)
	assert.Equal(t, message, msg)
	assert.Equal(t, &transfer, op)
	assert.True(t, message.Verify())
}
