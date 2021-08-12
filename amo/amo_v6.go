package amo

var _ AMOProtocol = (*AMOProtocolV6)(nil)

type AMOProtocolV6 struct {
	AMOProtocolV5
}

func (proto *AMOProtocolV6) Version() uint64 {
	return 0x6
}

