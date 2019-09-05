package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/db"
	tm "github.com/tendermint/tendermint/types"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/types"
)

const (
	// division by 2 is for safeguarding. tendermint code is not so safe.
	MaxTotalVotingPower = tm.MaxTotalVotingPower / 2
)

var (
	prefixBalance  = []byte("balance:")
	prefixStake    = []byte("stake:")
	prefixDelegate = []byte("delegate:")
	prefixParcel   = []byte("parcel:")
	prefixRequest  = []byte("request:")
	prefixUsage    = []byte("usage:")
)

type Store struct {
	stateDB db.DB // DB for blockchain state: see protocol.md
	indexDB db.DB
	// search index for delegators:
	// XXX: a delegatee can have multiple delegators
	// key: delegatee address || delegator address
	// value: nil
	indexDelegator db.DB
	// search index for validator:
	// key: validator address
	// value: holder address
	indexValidator db.DB
	// ordered cache of effective stakes:
	// key: effective stake (32 bytes) || stake holder address
	// value: nil
	indexEffStake db.DB
}

func NewStore(stateDB db.DB, indexDB db.DB) *Store {
	return &Store{
		stateDB:        stateDB,
		indexDB:        indexDB,
		indexDelegator: db.NewPrefixDB(indexDB, []byte("delegator")),
		indexValidator: db.NewPrefixDB(indexDB, []byte("validator")),
		indexEffStake:  db.NewPrefixDB(indexDB, []byte("effstake")),
	}
}

// Balance store
func getBalanceKey(addr tm.Address) []byte {
	return append(prefixBalance, addr.Bytes()...)
}

func (s Store) Purge() error {
	var itr db.Iterator

	// stateDB
	itr = s.stateDB.Iterator([]byte{}, []byte(nil))

	// TODO: cannot guarantee in multi-thread environment
	// need some sync mechanism
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		// XXX: not sure if this will confuse the iterator
		s.stateDB.Delete(k)
	}

	// TODO: need some method like s.stateDB.Size() to check if the DB has been
	// really emptied.
	itr.Close()

	// indexDB
	itr = s.indexDB.Iterator([]byte{}, []byte(nil))

	// TODO: cannot guarantee in multi-thread environment
	// need some sync mechanism
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		// XXX: not sure if this will confuse the iterator
		s.indexDB.Delete(k)
	}

	// TODO: need some method like s.stateDB.Size() to check if the DB has been
	// really emptied.
	itr.Close()

	return nil
}

func (s Store) SetBalance(addr tm.Address, balance *types.Currency) error {
	b, err := json.Marshal(balance)
	if err != nil {
		return err
	}
	s.stateDB.Set(getBalanceKey(addr), b)
	return nil
}

func (s Store) SetBalanceUint64(addr tm.Address, balance uint64) error {
	b, err := json.Marshal(new(types.Currency).Set(balance))
	if err != nil {
		return err
	}
	s.stateDB.Set(getBalanceKey(addr), b)
	return nil
}

func (s Store) GetBalance(addr tm.Address) *types.Currency {
	c := types.Currency{}
	balance := s.stateDB.Get(getBalanceKey(addr))
	if len(balance) == 0 {
		return &c
	}
	err := json.Unmarshal(balance, &c)
	if err != nil {
		return &c
	}
	return &c
}

// Stake store
func getStakeKey(holder []byte) []byte {
	return append(prefixStake, holder...)
}

func (s Store) SetStake(holder crypto.Address, stake *types.Stake) error {
	b, err := json.Marshal(stake)
	if err != nil {
		return code.TxErrBadParam
	}
	prevHolder := s.indexValidator.Get(stake.Validator.Address())
	if prevHolder != nil && !bytes.Equal(prevHolder, holder) {
		return code.TxErrBadValidator
	}

	if stake.Amount.Sign() == 0 {
		// check if there is a delegate appointed to this stake
		ds := s.GetDelegatesByDelegatee(holder)
		if len(ds) > 0 {
			return code.TxErrDelegateExists
		}

		// check if this is the last stake
		ts := s.GetTopStakes(2)
		if len(ts) == 1 { // requested 2 but got 1
			return code.TxErrLastValidator
		}
	}

	// clean up
	es := s.GetEffStake(holder)
	if es != nil {
		before := makeEffStakeKey(s.GetEffStake(holder).Amount, holder)
		if s.indexEffStake.Has(before) {
			s.indexEffStake.Delete(before)
		}
	}

	// update
	if stake.Amount.Sign() == 0 {
		s.stateDB.Delete(getStakeKey(holder))
		s.indexValidator.Delete(stake.Validator.Address())
	} else {
		s.stateDB.Set(getStakeKey(holder), b)
		s.indexValidator.Set(stake.Validator.Address(), holder)
		after := makeEffStakeKey(s.GetEffStake(holder).Amount, holder)
		s.indexEffStake.Set(after, nil)
	}

	return nil
}

func makeEffStakeKey(amount types.Currency, holder crypto.Address) []byte {
	key := make([]byte, 32+20) // 256-bit integer + 20-byte address
	b := amount.Bytes()
	copy(key[32-len(b):], b)
	copy(key[32:], holder)
	return key
}

func (s Store) GetStake(holder crypto.Address) *types.Stake {
	b := s.stateDB.Get(getStakeKey(holder))
	if len(b) == 0 {
		return nil
	}
	var stake types.Stake
	err := json.Unmarshal(b, &stake)
	if err != nil {
		return nil
	}
	return &stake
}

func (s Store) GetStakeByValidator(addr crypto.Address) *types.Stake {
	holder := s.GetHolderByValidator(addr)
	if holder == nil {
		return nil
	}
	return s.GetStake(holder)
}

func (s Store) GetHolderByValidator(addr crypto.Address) []byte {
	return s.indexValidator.Get(addr)
}

// Delegate store
func getDelegateKey(holder []byte) []byte {
	return append(prefixDelegate, holder...)
}

func (s Store) SetDelegate(holder crypto.Address, value *types.Delegate) error {
	b, err := json.Marshal(value)
	if err != nil {
		return code.TxErrBadParam
	}
	// before state update
	es := s.GetEffStake(value.Delegatee)
	if es == nil {
		return code.TxErrNoStake
	}

	// make effStakeKey to find its corresponding value
	before := makeEffStakeKey(es.Amount, value.Delegatee)
	if s.indexEffStake.Has(before) {
		s.indexEffStake.Delete(before)
	}

	// upadate
	if value.Amount.Sign() == 0 {
		s.stateDB.Delete(getDelegateKey(holder))
		s.indexDelegator.Delete(append(value.Delegatee, holder...))
	} else {
		s.stateDB.Set(getDelegateKey(holder), b)
		s.indexDelegator.Set(append(value.Delegatee, holder...), nil)
		after := makeEffStakeKey(s.GetEffStake(value.Delegatee).Amount, value.Delegatee)
		s.indexEffStake.Set(after, nil)
	}

	return nil
}

func (s Store) GetDelegate(holder crypto.Address) *types.Delegate {
	b := s.stateDB.Get(getDelegateKey(holder))
	if len(b) == 0 {
		return nil
	}
	var delegate types.Delegate
	err := json.Unmarshal(b, &delegate)
	if err != nil {
		return nil
	}
	return &delegate
}

func (s Store) GetDelegateEx(holder crypto.Address) *types.DelegateEx {
	delegate := s.GetDelegate(holder)
	if delegate == nil {
		return nil
	}
	return &types.DelegateEx{holder, delegate}
}

func (s Store) GetDelegatesByDelegatee(delegatee crypto.Address) []*types.DelegateEx {
	var itr db.Iterator = s.indexDelegator.Iterator(delegatee, nil)
	defer itr.Close()

	var delegates []*types.DelegateEx
	for ; itr.Valid() && bytes.HasPrefix(itr.Key(), delegatee); itr.Next() {
		delegator := itr.Key()[len(delegatee):]
		delegates = append(delegates, s.GetDelegateEx(delegator))
	}
	return delegates
}

func (s Store) GetEffStake(delegatee crypto.Address) *types.Stake {
	stake := s.GetStake(delegatee)
	if stake == nil {
		return nil
	}
	for _, d := range s.GetDelegatesByDelegatee(delegatee) {
		stake.Amount.Add(&d.Amount)
	}
	return stake
}

func (s Store) GetTopStakes(max uint64) []*types.Stake {
	var stakes []*types.Stake
	var itr db.Iterator = s.indexEffStake.ReverseIterator(nil, nil)
	var cnt uint64 = 0
	for ; itr.Valid(); itr.Next() {
		if cnt >= max {
			break
		}
		key := itr.Key()
		var amount types.Currency
		amount.SetBytes(key[:32])
		holder := key[32:]
		stake := s.GetStake(holder)
		stake.Amount = amount
		stakes = append(stakes, stake)
		cnt++
		// TODO: assert GetEffStake() gives the same result
	}
	return stakes
}

// Parcel store
func getParcelKey(parcelID []byte) []byte {
	return append(prefixParcel, parcelID...)
}

func (s Store) SetParcel(parcelID []byte, value *types.ParcelValue) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.stateDB.Set(getParcelKey(parcelID), b)
	return nil
}

func (s Store) GetParcel(parcelID []byte) *types.ParcelValue {
	b := s.stateDB.Get(getParcelKey(parcelID))
	if len(b) == 0 {
		return nil
	}
	var parcel types.ParcelValue
	err := json.Unmarshal(b, &parcel)
	if err != nil {
		return nil
	}
	return &parcel
}

func (s Store) DeleteParcel(parcelID []byte) {
	s.stateDB.DeleteSync(getParcelKey(parcelID))
}

// Request store
func getRequestKey(buyer crypto.Address, parcelID []byte) []byte {
	return append(prefixRequest, append(append(buyer, ':'), parcelID...)...)
}

func (s Store) SetRequest(buyer crypto.Address, parcelID []byte, value *types.RequestValue) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.stateDB.Set(getRequestKey(buyer, parcelID), b)
	return nil
}

func (s Store) GetRequest(buyer crypto.Address, parcelID []byte) *types.RequestValue {
	b := s.stateDB.Get(getRequestKey(buyer, parcelID))
	if len(b) == 0 {
		return nil
	}
	var request types.RequestValue
	err := json.Unmarshal(b, &request)
	if err != nil {
		return nil
	}
	return &request
}

func (s Store) DeleteRequest(buyer crypto.Address, parcelID []byte) {
	s.stateDB.DeleteSync(getRequestKey(buyer, parcelID))
}

// Usage store
func getUsageKey(buyer crypto.Address, parcelID []byte) []byte {
	return append(prefixUsage, append(append(buyer, ':'), parcelID...)...)
}

func (s Store) SetUsage(buyer crypto.Address, parcelID []byte, value *types.UsageValue) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.stateDB.Set(getUsageKey(buyer, parcelID), b)
	return nil
}

func (s Store) GetUsage(buyer crypto.Address, parcelID []byte) *types.UsageValue {
	b := s.stateDB.Get(getUsageKey(buyer, parcelID))
	if len(b) == 0 {
		return nil
	}
	var usage types.UsageValue
	err := json.Unmarshal(b, &usage)
	if err != nil {
		return nil
	}
	return &usage
}

func (s Store) DeleteUsage(buyer crypto.Address, parcelID []byte) {
	s.stateDB.DeleteSync(getUsageKey(buyer, parcelID))
}

func (s Store) GetValidators(max uint64) abci.ValidatorUpdates {
	var vals abci.ValidatorUpdates
	stakes := s.GetTopStakes(max)
	adjFactor := calcAdjustFactor(stakes)
	for _, stake := range stakes {
		key := abci.PubKey{ // TODO
			Type: "ed25519",
			Data: stake.Validator[:],
		}
		var power big.Int
		power.Rsh(&stake.Amount.Int, adjFactor)
		val := abci.ValidatorUpdate{
			PubKey: key,
			Power:  power.Int64(),
		}
		if val.Power > 0 {
			vals = append(vals, val)
		}
	}
	return vals
}

func calcAdjustFactor(stakes []*types.Stake) uint {
	var vp big.Int
	max := MaxTotalVotingPower
	var vps int64 = 0
	var shifts uint = 0
	for _, stake := range stakes {
		vp.Rsh(&stake.Amount.Int, shifts)
		for !vp.IsInt64() {
			vp.Rsh(&vp, 1)
			shifts++
			vps >>= 1
		}
		vpi := vp.Int64()
		tmp := vps + vpi
		for tmp < vps || tmp > max {
			vps >>= 1
			vpi >>= 1
			shifts++
			tmp = vps + vpi
		}
		vps = tmp
	}
	return shifts
}
