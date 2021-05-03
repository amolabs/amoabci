package amo

import (
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type State struct {
	ProtocolVersion uint64 `json:"protocol_version"`
	Height          int64  `json:"-"` // current block height
	LastHeight      int64  `json:"-"` // last completed block height
	LastAppHash     []byte `json:"-"`
	NextDraftID     uint32 `json:"-"`
}

func (s *State) InferFrom(sto *store.Store, cfg types.AMOAppConfig) {
	height := sto.GetMerkleVersion() - int64(1)
	if height < int64(0) {
		height = int64(0)
	}

	hash := sto.Root()

	nextDraftID := sto.GetLastDraftID() + uint32(1)

	s.Height = height
	s.LastHeight = height
	s.LastAppHash = hash
	s.NextDraftID = nextDraftID

	s.ProtocolVersion = sto.GetProtocolVersion(false)
	if s.ProtocolVersion == 0 {
		// NOTE: This can be done since we are writing a SW in a retrospective
		// manner. That is, we already observed a state DB which holds data
		// produced via protocol version greater than 3.
		if cfg.UpgradeProtocolHeight > s.Height {
			s.ProtocolVersion = cfg.UpgradeProtocolVersion - 1
		} else if s.Height > 0 && cfg.UpgradeProtocolHeight <= s.Height {
			s.ProtocolVersion = cfg.UpgradeProtocolVersion
		} else {
			s.ProtocolVersion = AMOGenesisProtocolVersion
		}
	}
}
