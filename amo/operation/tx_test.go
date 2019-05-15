package operation

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/amolabs/amoabci/amo/types"
	"github.com/amolabs/amoabci/crypto/p256"
)

func TestParseTx(t *testing.T) {
	from := p256.GenPrivKeyFromSecret([]byte("test1"))
	to := p256.GenPrivKeyFromSecret([]byte("test2")).PubKey().Address()
	transfer := Transfer{
		To:     to,
		Amount: *new(types.Currency).Set(1000),
	}
	b, _ := json.Marshal(transfer)
	message := Message{
		Type:    TxTransfer,
		Payload: b,
		Sender:  from.PubKey().Address(),
		Nonce:   []byte{0x12, 0x34, 0x56, 0x78},
	}
	sb := message.GetSigningBytes()
	_sb := `{"type":"transfer","sender":"85FE85FCE6AB426563E5E0749EBCB95E9B1EF1D5","nonce":"12345678","payload":{"to":"218B954DF74E7267E72541CE99AB9F49C410DB96","amount":"1000"}}`
	assert.Equal(t, _sb, string(sb))
	err := message.Sign(from)
	if err != nil {
		panic(err)
	}
	bMsg, _ := json.Marshal(message)
	msg, op, _ := ParseTx(bMsg)
	assert.Equal(t, message, msg)
	assert.Equal(t, &transfer, op)
	assert.True(t, message.Verify())
}
