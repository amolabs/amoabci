package amo

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type State struct {
	ProtocolVersion uint64 `json:"protocol_version"`
	Height          int64  `json:"-"` // current block height
	LastHeight      int64  `json:"-"` // last completed block height
	LastAppHash     []byte `json:"-"`
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
