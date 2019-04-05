package store

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/types"

	atypes "github.com/amolabs/amoabci/amo/types"
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
	// search index for delegators:
	// key: delegator address || holder address
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
		indexDelegator: db.NewPrefixDB(indexDB, []byte("delegator")),
		indexValidator: db.NewPrefixDB(indexDB, []byte("validator")),
		indexEffStake:  db.NewPrefixDB(indexDB, []byte("effstake")),
	}
}

// Balance store
func getBalanceKey(addr types.Address) []byte {
	return append(prefixBalance, addr.Bytes()...)
}

func (s Store) Purge() error {
	var itr db.Iterator = s.stateDB.Iterator([]byte{}, []byte(nil))
	defer itr.Close()

	// TODO: cannot guarantee in multi-thread environment
	// need some sync mechanism
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		// XXX: not sure if this will confuse the iterator
		s.stateDB.Delete(k)
	}

	// TODO: need some method like s.stateDB.Size() to check if the DB has been
	// really emptied.

	return nil
}

func (s Store) SetBalance(addr types.Address, balance *atypes.Currency) {
	b, err := json.Marshal(balance)
	if err != nil {
		panic(err)
	}
	s.stateDB.Set(getBalanceKey(addr), b)
}

func (s Store) SetBalanceUint64(addr types.Address, balance uint64) {
	b, err := json.Marshal(new(atypes.Currency).Set(balance))
	if err != nil {
		panic(err)
	}
	s.stateDB.Set(getBalanceKey(addr), b)
}

func (s Store) GetBalance(addr types.Address) *atypes.Currency {
	c := atypes.Currency{}
	balance := s.stateDB.Get(getBalanceKey(addr))
	if len(balance) == 0 {
		return &c
	}
	err := json.Unmarshal(balance, &c)
	if err != nil {
		panic(err)
	}
	return &c
}

// Stake store
func getStakeKey(holder []byte) []byte {
	return append(prefixStake, holder...)
}

func (s Store) SetStake(holder crypto.Address, stake *atypes.Stake) {
	b, err := json.Marshal(stake)
	if err != nil {
		panic(err)
	}
	// before state update
	es := s.GetEffStake(holder)
	if es != nil {
		before := makeEffStakeKey(s.GetEffStake(holder).Amount, holder)
		if s.indexEffStake.Has(before) {
			s.indexEffStake.Delete(before)
		}
	}
	s.stateDB.Set(getStakeKey(holder), b)
	// after state update
	s.indexValidator.Set(stake.Validator.Address(), holder)
	after := makeEffStakeKey(s.GetEffStake(holder).Amount, holder)
	s.indexEffStake.Set(after, nil)
}

func makeEffStakeKey(amount atypes.Currency, holder crypto.Address) []byte {
	key := make([]byte, 32+20) // 256-bit integer + 20-byte address
	b := amount.Bytes()
	copy(key[32-len(b):], b)
	copy(key[32:], holder)
	return key
}

func (s Store) GetStake(holder crypto.Address) *atypes.Stake {
	b := s.stateDB.Get(getStakeKey(holder))
	if len(b) == 0 {
		return nil
	}
	var stake atypes.Stake
	err := json.Unmarshal(b, &stake)
	if err != nil {
		panic(err)
	}
	return &stake
}

func (s Store) GetStakeByValidator(addr crypto.Address) *atypes.Stake {
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

func (s Store) SetDelegate(holder crypto.Address, value *atypes.Delegate) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	// before state update
	es := s.GetEffStake(value.Delegator)
	if es == nil {
		return errors.New("No stake for a delegator")
	} else {
		before := makeEffStakeKey(es.Amount, value.Delegator)
		if s.indexEffStake.Has(before) {
			s.indexEffStake.Delete(before)
		}
	}
	// state update
	s.stateDB.Set(getDelegateKey(holder), b)
	// after state update
	s.indexDelegator.Set(append(value.Delegator, holder...), nil)
	after := makeEffStakeKey(s.GetEffStake(value.Delegator).Amount, value.Delegator)
	s.indexEffStake.Set(after, nil)
	return nil
}

func (s Store) GetDelegate(holder crypto.Address) *atypes.Delegate {
	b := s.stateDB.Get(getDelegateKey(holder))
	if len(b) == 0 {
		return nil
	}
	var delegate atypes.Delegate
	err := json.Unmarshal(b, &delegate)
	if err != nil {
		panic(err)
	}
	delegate.Holder = holder
	return &delegate
}

func (s Store) GetDelegatesByDelegator(delegator crypto.Address) []*atypes.Delegate {
	var itr db.Iterator = s.indexDelegator.Iterator(delegator, nil)
	defer itr.Close()

	var delegates []*atypes.Delegate
	for ; itr.Valid() && bytes.HasPrefix(itr.Key(), delegator); itr.Next() {
		holder := itr.Key()[len(delegator):]
		delegates = append(delegates, s.GetDelegate(holder))
	}
	return delegates
}

func (s Store) GetEffStake(delegator crypto.Address) *atypes.Stake {
	stake := s.GetStake(delegator)
	if stake == nil {
		return nil
	}
	for _, d := range s.GetDelegatesByDelegator(delegator) {
		stake.Amount.Add(&d.Amount)
	}
	return stake
}

func (s Store) GetTopStakes(max uint64) []*atypes.Stake {
	var stakes []*atypes.Stake
	var itr db.Iterator = s.indexEffStake.ReverseIterator(nil, nil)
	var cnt uint64 = 0
	for ; itr.Valid(); itr.Next() {
		if cnt >= max {
			break
		}
		key := itr.Key()
		var amount atypes.Currency
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

func (s Store) SetParcel(parcelID []byte, value *atypes.ParcelValue) {
	b, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	s.stateDB.Set(getParcelKey(parcelID), b)
}

func (s Store) GetParcel(parcelID []byte) *atypes.ParcelValue {
	b := s.stateDB.Get(getParcelKey(parcelID))
	if len(b) == 0 {
		return nil
	}
	var parcel atypes.ParcelValue
	err := json.Unmarshal(b, &parcel)
	if err != nil {
		panic(err)
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

func (s Store) SetRequest(buyer crypto.Address, parcelID []byte, value *atypes.RequestValue) {
	b, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	s.stateDB.Set(getRequestKey(buyer, parcelID), b)
}

func (s Store) GetRequest(buyer crypto.Address, parcelID []byte) *atypes.RequestValue {
	b := s.stateDB.Get(getRequestKey(buyer, parcelID))
	if len(b) == 0 {
		return nil
	}
	var request atypes.RequestValue
	err := json.Unmarshal(b, &request)
	if err != nil {
		panic(err)
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

func (s Store) SetUsage(buyer crypto.Address, parcelID []byte, value *atypes.UsageValue) {
	b, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	s.stateDB.Set(getUsageKey(buyer, parcelID), b)
}

func (s Store) GetUsage(buyer crypto.Address, parcelID []byte) *atypes.UsageValue {
	b := s.stateDB.Get(getUsageKey(buyer, parcelID))
	if len(b) == 0 {
		return nil
	}
	var usage atypes.UsageValue
	err := json.Unmarshal(b, &usage)
	if err != nil {
		panic(err)
	}
	return &usage
}

func (s Store) DeleteUsage(buyer crypto.Address, parcelID []byte) {
	s.stateDB.DeleteSync(getUsageKey(buyer, parcelID))
}
