package amo

import (
	"github.com/amolabs/amoabci/amo/tx"
)

var _ AMOProtocol = (*AMOProtocolV6)(nil)

type AMOProtocolV6 struct {
	AMOProtocolV5
}

func (proto *AMOProtocolV6) Version() uint64 {
	return 0x6
}

func (proto *AMOProtocolV6) ParseTx(txBytes []byte) (tx.Tx, error) {
	return tx.ParseTxV6(txBytes)
}
