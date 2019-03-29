package amo

import (
	"errors"
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tdb "github.com/tendermint/tendermint/libs/db"
	tm "github.com/tendermint/tendermint/types"

	"github.com/amolabs/amoabci/amo/operation"
	"github.com/amolabs/amoabci/amo/types"
)

const (
	nVal = 100
)

var (
	prefixStaked    = []byte("s:")
	prefixDelegated = []byte("d:")
	prefixRanking   = []byte("r:")
)

type stakeInfo struct {
	Type   string
	Amount types.Currency
	Target ed25519.PubKeyEd25519
}

type ValidatorManager struct {
	info       []stakeInfo
	stakeIndex tdb.DB
	ranking    tdb.DB
}

func NewValidatorManager(db tdb.DB) *ValidatorManager {
	return &ValidatorManager{
		stakeIndex: db,
		ranking:    tdb.NewPrefixDB(db, prefixRanking),
		info: make([]stakeInfo, 0, 128),
	}
}

func addUint64(a, b uint64) (uint64, error) {
	prev := a
	a += b
	if prev > a {
		return 0, errors.New("overflow detected")
	}
	return a, nil
}

func (vm *ValidatorManager) AddStakeInfo(t string, pub ed25519.PubKeyEd25519, op operation.Operation) {
	var amount *types.Currency
	switch t {
	case operation.TxStake:
		stake := op.(*operation.Stake)
		amount = &stake.Amount
	case operation.TxWithdraw:
		withdraw := op.(*operation.Withdraw)
		amount = &withdraw.Amount
	case operation.TxDelegate:
		delegate := op.(*operation.Delegate)
		amount = &delegate.Amount
	case operation.TxRetract:
		retract := op.(*operation.Retract)
		amount = &retract.Amount
	}
	vm.info = append(vm.info, stakeInfo{
		Type:   t,
		Amount: *amount,
		Target: pub,
	})
}

func getStakedKey(pub ed25519.PubKeyEd25519) []byte {
	return append(prefixStaked, pub.Bytes()...)
}

func getDelegatedKey(pub ed25519.PubKeyEd25519) []byte {
	return append(prefixDelegated, pub.Bytes()...)
}

func getRankingKey(c *big.Int, suffix []byte) []byte {
	k := make([]byte, 32)
	b := c.Bytes()
	copy(k[32-len(b):], b)
	return append(k[:], suffix...)
}

func (vm *ValidatorManager) Index() {
	for _, info := range vm.info {
		sKey := getStakedKey(info.Target)
		dKey := getDelegatedKey(info.Target)
		sb := vm.stakeIndex.Get(sKey)
		if sb == nil {
			sb = []byte{0x00}
		}
		deb := vm.stakeIndex.Get(dKey)
		if deb == nil {
			deb = []byte{0x00}
		}
		staked := new(big.Int).SetBytes(sb)
		delegated := new(big.Int).SetBytes(deb)
		res := new(big.Int).SetBytes([]byte{0x00})
		switch info.Type {
		case operation.TxStake:
			res.Add(&info.Amount.Int, staked)
			vm.stakeIndex.Set(sKey, res.Bytes())
			res.Add(res, delegated)
		case operation.TxWithdraw:
			res.Sub(staked, &info.Amount.Int)
			vm.stakeIndex.Set(sKey, res.Bytes())
			res.Add(res, delegated)
		case operation.TxDelegate:
			res.Add(&info.Amount.Int, delegated)
			vm.stakeIndex.Set(dKey, res.Bytes())
			res.Add(res, staked)
		case operation.TxRetract:
			res.Sub(delegated, &info.Amount.Int)
			vm.stakeIndex.Set(dKey, res.Bytes())
			res.Add(res, staked)
		}
		total := new(big.Int).SetBytes([]byte{0x00})
		total.Add(total, staked)
		total.Add(total, delegated)
		vm.ranking.Delete(getRankingKey(total, info.Target[0:4]))
		value := make([]byte, 32)
		copy(value, info.Target[:])
		// for memDB;
		vm.ranking.Set(getRankingKey(res, info.Target[0:4]), value)
	}
	vm.info = make([]stakeInfo, 0, 128)
}

func (vm *ValidatorManager) UpdateValidator() abci.ValidatorUpdates {
	max := new(big.Int).SetUint64(uint64(tm.MaxTotalVotingPower))
	adjFactor := uint64(0)
	total := uint64(0)
	vLen := 0
	v := make([]abci.ValidatorUpdate, 0, nVal)
	p := make([]big.Int, 0, nVal)
	for iter := vm.ranking.ReverseIterator(nil, nil); iter.Valid() && vLen < 100; vLen++ {
		p = append(p, big.Int{})
		p[vLen].SetBytes(iter.Key()[:32])
		v = append(v, abci.ValidatorUpdate{})
		v[vLen].PubKey.Type = "ed25519"
		v[vLen].PubKey.Data = iter.Value()
		iter.Next()
	}
	for i := 0; i < vLen; i++ {
		if p[i].Cmp(max) == 1 || total > uint64(tm.MaxTotalVotingPower) {
			for j := 0; j < vLen; j++ {
				p[j].Rsh(&p[j], 1)
			}
			total >>= 1
			adjFactor++
			i--
			continue
		}
		total += p[i].Uint64()
	}
	for i := 0; i < vLen; i++ {
		v[i].Power = p[i].Int64()
	}
	return v
}
