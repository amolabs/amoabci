package amo

import (
	"encoding/json"

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

	s.ProtocolVersion = sto.GetProtocolVersion(false)
	if s.ProtocolVersion == 0 {
		// NOTE: Since we cannot determine the protocol version from the saved
		// state, consult the config structure to infer the version. Possible
		// candidate versions are 0x3 and 0x4.
		b := sto.GetAppConfig()
		if b == nil || len(b) == 0 {
			// No config yet stored in the state DB.
			// Assume AMOGenesisProtocolVersion.
			s.ProtocolVersion = 0x3
			return
		}
		var configV4 struct {
			LazinessWindow *int64 `json:"laziness_window"`
		}
		err := json.Unmarshal(b, &configV4)
		if err != nil || configV4.LazinessWindow == nil {
			s.ProtocolVersion = 0x3
			return
		}
		s.ProtocolVersion = 0x4
		return
	}
}
