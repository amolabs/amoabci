package amo

import (
	"github.com/amolabs/amoabci/amo/tx"
)

var _ AMOProtocol = (*AMOProtocolV4)(nil)

type AMOProtocolV4 struct {
	version int64
}

func (proto *AMOProtocolV4) Version() uint64 {
	return 0x4
}

func (proto *AMOProtocolV4) ParseTx(txBytes []byte) (tx.Tx, error) {
	return tx.ParseTx(txBytes)
}
