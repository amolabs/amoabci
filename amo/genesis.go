package amo

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type GenAmoAppState struct {
	Balances []GenAccBalance `json:"balances"`
}

type GenAccBalance struct {
	Owner  crypto.Address `json:"owner"`
	Amount types.Currency `json:"amount"`
}

func ParseGenesisStateBytes(data []byte) (*GenAmoAppState, error) {
	genState := GenAmoAppState{}
	err := json.Unmarshal(data, &genState)
	if err != nil {
		return nil, err
	}
	return &genState, nil
}

func FillGenesisState(s *store.Store, genState *GenAmoAppState) error {
	err := s.Purge()
	if err != nil {
		return err
	}

	for _, accBal := range genState.Balances {
		s.SetBalance(accBal.Owner, accBal.Amount)
	}

	return nil
}
