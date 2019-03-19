package operation

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

func TestParseTx(t *testing.T) {
	from := p256.GenPrivKey()
	to := p256.GenPrivKey().PubKey().Address()
	transfer := Transfer{
		To:     to,
		Amount: *new(types.Currency).Set(1000),
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
