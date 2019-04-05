package amo

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type GenAmoAppState struct {
	Balances []GenAccBalance `json:"balances"`
	Stakes   []GenAccStake   `json:"stakes"`
}

type GenAccBalance struct {
	Owner  crypto.Address `json:"owner"`
	Amount types.Currency `json:"amount"`
}

type GenAccStake struct {
	Holder    crypto.Address `json:"holder"`
	Amount    types.Currency `json:"amount"`
	Validator []byte         `json:"validator"`
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
		s.SetBalance(accBal.Owner, &accBal.Amount)
	}

	for _, accStake := range genState.Stakes {
		var val25519 ed25519.PubKeyEd25519
		copy(val25519[:], accStake.Validator)
		s.SetStake(accStake.Holder, &types.Stake{
			Amount:    accStake.Amount,
			Validator: val25519,
		})
	}

	return nil
}
