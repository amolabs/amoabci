package amo

var _ AMOProtocol = (*AMOProtocolV4)(nil)

type AMOProtocolV4 struct {
	version int64
}

func (proto *AMOProtocolV4) Version() uint64 {
	return 0x4
}
