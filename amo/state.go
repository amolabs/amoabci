package amo

import (
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type State struct {
	ProtocolVersion uint64 `json:"-"`
	Height          int64  `json:"-"` // current block height
	LastHeight      int64  `json:"-"` // last completed block height
	LastAppHash     []byte `json:"-"`
	NextDraftID     uint32 `json:"-"`
}

func (s *State) LoadFrom(sto *store.Store, cfg types.AMOAppConfig) {
	height := sto.GetMerkleVersion() - int64(1)
	if height < int64(0) {
		height = int64(0)
	}

	hash := sto.Root()

	nextDraftID := sto.GetLastDraftID() + uint32(1)
	protocolVersion := cfg.UpgradeProtocolVersion
	if height < cfg.UpgradeProtocolHeight {
		protocolVersion -= uint64(1)
	}

	s.Height = height
	s.LastHeight = height
	s.LastAppHash = hash
	s.NextDraftID = nextDraftID
	s.ProtocolVersion = protocolVersion
}
