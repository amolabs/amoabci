package amo

import (
	"github.com/amolabs/amoabci/amo/tx"
)

var _ AMOProtocol = (*AMOProtocolV5)(nil)

type AMOProtocolV5 struct {
	AMOProtocolV4
}

func (proto *AMOProtocolV5) Version() uint64 {
	return 0x5
}

func (proto *AMOProtocolV5) ParseTx(txBytes []byte) (tx.Tx, error) {
	return tx.ParseTxV5(txBytes)
}
