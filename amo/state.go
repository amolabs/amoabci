package amo

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type State struct {
	ProtocolVersion uint64 `json:"protocol_version"`
	MerkleVersion   int64  `json:"merkle_version"`
	Height          int64  `json:"height"` // current block height
	AppHash         []byte `json:"app_hash"`
	LastHeight      int64  `json:"last_height"` // last completed block height
	LastAppHash     []byte `json:"last_app_hash"`
	CounterDue      int64  `json:"counter_due"`
	NextDraftID     uint32 `json:"next_draft_id"`
}

func (s *State) LoadFrom(f *os.File) error {
	file, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return err
	}

	if len(file) > 0 {
		err = json.Unmarshal(file, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *State) SaveTo(f *os.File) error {
	file, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(f.Name(), file, os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}
