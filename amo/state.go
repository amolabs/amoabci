package amo

import (
	"github.com/amolabs/amoabci/amo/store"
)

type State struct {
	ProtocolVersion uint64 `json:"protocol_version"`
	Height          int64  `json:"-"` // current block height
	LastHeight      int64  `json:"-"` // last completed block height
	LastAppHash     []byte `json:"-"`
	NextDraftID     uint32 `json:"-"`
}

func (s *State) InferFrom(sto *store.Store) {
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

	s.ProtocolVersion = AMOGenesisProtocolVersion
}
