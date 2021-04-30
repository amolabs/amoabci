package amo

var _ AMOProtocol = (*AMOProtocolV5)(nil)

type AMOProtocolV5 struct {
	AMOProtocolV4
}

func (proto *AMOProtocolV5) Version() uint64 {
	return 0x5
}
