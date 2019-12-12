package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/tendermint/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tm "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/types"
)

const (
	// division by 2 is for safeguarding. tendermint code is not so safe.
	MaxTotalVotingPower = tm.MaxTotalVotingPower / 2

	merkleTreeCacheSize = 10000
)

var (
	prefixBalance  = []byte("balance:")
	prefixStake    = []byte("stake:")
	prefixDelegate = []byte("delegate:")
	prefixParcel   = []byte("parcel:")
	prefixRequest  = []byte("request:")
	prefixUsage    = []byte("usage:")

	prefixIndexDelegator = []byte("delegator")
	prefixIndexValidator = []byte("validator")
	prefixIndexEffStake  = []byte("effstake")
)

type Store struct {
	// merkle tree for blockchain state
	merkleTree *iavl.MutableTree

	indexDB tmdb.DB
	// search index for delegators:
	// XXX: a delegatee can have multiple delegators
	// key: delegatee address || delegator address
	// value: nil
	indexDelegator tmdb.DB
	// search index for validator:
	// key: validator address
	// value: holder address
	indexValidator tmdb.DB
	// ordered cache of effective stakes:
	// key: effective stake (32 bytes) || stake holder address
	// value: nil
	indexEffStake tmdb.DB

	// search index for block-first delivered txs
	// key: block height
	// value: hash of txs
	indexBlockTx tmdb.DB
	// search index for tx hash-first delivered txs
	// key: tx hash
	// value: block height
	indexTxBlock tmdb.DB

	incentiveDB tmdb.DB
	// search incentive for block height-first:
	// key: block height || stake holder address
	// value: incentive amount
	incentiveHeight tmdb.DB
	// search incentive for address-first:
	// key: stake holder addres || block height
	// value: incentive amount
	incentiveAddress tmdb.DB

	// lazinessCounter database
	lazinessCounterDB tmdb.DB
}

func NewStore(merkleDB, indexDB, incentiveDB, lazinessCounterDB tmdb.DB) *Store {
	return &Store{
		merkleTree: iavl.NewMutableTree(merkleDB, merkleTreeCacheSize),

		indexDB: indexDB,

		indexDelegator: tmdb.NewPrefixDB(indexDB, prefixIndexDelegator),
		indexValidator: tmdb.NewPrefixDB(indexDB, prefixIndexValidator),
		indexEffStake:  tmdb.NewPrefixDB(indexDB, prefixIndexEffStake),

		indexBlockTx: tmdb.NewPrefixDB(indexDB, prefixIndexBlockTx),
		indexTxBlock: tmdb.NewPrefixDB(indexDB, prefixIndexTxBlock),

		incentiveDB:      incentiveDB,
		incentiveHeight:  tmdb.NewPrefixDB(incentiveDB, prefixIncentiveHeight),
		incentiveAddress: tmdb.NewPrefixDB(incentiveDB, prefixIncentiveAddress),

		lazinessCounterDB: lazinessCounterDB,
	}
}

func (s Store) Purge() error {
	var (
		itr tmdb.Iterator
		err error
	)

	// merkleTree
	s.merkleTree.Rollback()

	// delete all available tree versions
	versions := s.merkleTree.AvailableVersions()
	for i := len(versions) - 1; i >= 0; i-- {
		err = s.merkleTree.DeleteVersion(int64(versions[i]))
		if err != nil {
			return err
		}
	}

	// check if merkle tree is really emptied
	if !s.merkleTree.IsEmpty() {
		return errors.New("couldn't purge merkle tree")
	}

	// indexDB
	itr = s.indexDB.Iterator(nil, nil)
	defer itr.Close()
	// TODO: cannot guarantee in multi-thread environment
	// need some sync mechanism
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		// XXX: not sure if this will confuse the iterator
		s.indexDB.Delete(k)
	}

	// TODO: need some method like s.stateDB.Size() to check if the DB has been
	// really emptied.

	// incentiveDB
	itr = s.incentiveDB.Iterator(nil, nil)
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		s.incentiveDB.Delete(k)
	}

	// lazinessCounterDB
	itr = s.lazinessCounterDB.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		s.lazinessCounterDB.Delete(k)
	}

	return nil
}

// MERKLE TREE SCOPE
// set -> working tree node (ONLY)
// get(committed: true)  -> the latest saved tree node
// 	  (committed: false) -> working tree node

// MERKLE TREE WORKFLOW
// set 	: working tree
// save : working tree -> saved tree

// node(key, value) -> working tree

func (s Store) has(key []byte) bool {
	return s.merkleTree.Has(key)
}

func (s Store) set(key, value []byte) bool {
	return s.merkleTree.Set(key, value)
}

// { working tree || saved tree } -> node(key, value)
func (s Store) get(key []byte, committed bool) []byte {
	if !committed {
		_, value := s.merkleTree.Get(key)
		return value
	}

	latestVersion, err := s.getLatestVersion()
	if err != nil {
		return nil
	}

	_, value := s.merkleTree.GetVersioned(key, latestVersion)
	return value
}

// working tree, delete node(key, value)
func (s Store) remove(key []byte) ([]byte, bool) {
	return s.merkleTree.Remove(key)
}

// working tree >> saved tree
func (s Store) Save() ([]byte, int64, error) {
	return s.merkleTree.SaveVersion()
}

// Load the latest versioned tree from disk.
func (s Store) Load() (int64, error) {
	return s.merkleTree.Load()
}

func (s Store) Root() []byte {
	// NOTES
	// Hash() : Hash returns the hash of the latest saved version of the tree,
	// as returned by SaveVersion. If no versions have been saved, Hash returns nil.
	//
	// WorkingHash() : WorkingHash returns the hash of the current working tree.

	return s.merkleTree.WorkingHash()
}

func (s Store) Verify(key []byte) (bool, error) {
	return true, nil
}

func (s Store) getLatestVersion() (int64, error) {
	versions := s.merkleTree.AvailableVersions()
	if len(versions) == 0 {
		return int64(0), errors.New("no available versions exist")
	}

	return int64(versions[len(versions)-1]), nil
}

func (s Store) getImmutableTree(committed bool) (*iavl.ImmutableTree, error) {
	if !committed {
		return s.merkleTree.ImmutableTree, nil
	}

	latestVersion, err := s.getLatestVersion()
	if err != nil {
		return nil, err
	}

	imt, err := s.merkleTree.GetImmutable(latestVersion)
	if err != nil {
		return nil, err
	}

	return imt, nil
}

// Balance store
func getBalanceKey(addr tm.Address) []byte {
	return append(prefixBalance, addr.Bytes()...)
}

func (s Store) SetBalance(addr tm.Address, balance *types.Currency) error {
	zero := new(types.Currency).Set(0)
	balanceKey := getBalanceKey(addr)

	if balance.LessThan(zero) {
		return fmt.Errorf("unavailable amount: %s", balance.String())
	}

	// pre-process for setting zero balance, just remove corresponding key
	if s.has(balanceKey) && balance.Equals(zero) {
		s.remove(balanceKey)
		return nil
	}

	b, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	s.set(balanceKey, b)

	return nil
}

func (s Store) SetBalanceUint64(addr tm.Address, balance uint64) error {
	zero := uint64(0)
	balanceKey := getBalanceKey(addr)

	// pre-process for setting zero balance, just remove corresponding key
	if s.has(balanceKey) && balance == zero {
		s.remove(balanceKey)
		return nil
	}

	b, err := json.Marshal(new(types.Currency).Set(balance))
	if err != nil {
		return err
	}

	s.set(balanceKey, b)

	return nil
}

func (s Store) GetBalance(addr tm.Address, committed bool) *types.Currency {
	c := types.Currency{}
	balance := s.get(getBalanceKey(addr), committed)
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
func makeStakeKey(holder []byte) []byte {
	return append(prefixStake, holder...)
}

func makeLockedStakeKey(holder []byte, height int64) []byte {
	hb := make([]byte, 8)
	binary.BigEndian.PutUint64(hb, uint64(height))
	dbKey := append(prefixStake, holder...)
	dbKey = append(dbKey, hb...)
	return dbKey
}

func splitLockedStakeKey(key []byte) (crypto.Address, int64) {
	if len(key) != len(prefixStake)+crypto.AddressSize+8 {
		return nil, 0
	}
	h := binary.BigEndian.Uint64(key[len(prefixStake)+crypto.AddressSize:])
	return key[len(prefixStake) : len(prefixStake)+crypto.AddressSize], int64(h)
}

func (s Store) checkValidatorMatch(holder crypto.Address, stake *types.Stake) error {
	// TODO: use s.GetHolderByValidator(stake.Validator)
	prevHolder := s.indexValidator.Get(stake.Validator.Address())
	if prevHolder != nil && !bytes.Equal(prevHolder, holder) {
		return code.TxErrPermissionDenied
	}
	prevStake := s.GetStake(holder, false)
	if prevStake != nil &&
		!bytes.Equal(prevStake.Validator[:], stake.Validator[:]) {
		return code.TxErrBadValidator
	}
	return nil
}

func (s Store) checkStakeDeletion(holder crypto.Address, stake *types.Stake, height int64) error {
	if stake.Amount.Sign() == 0 {
		whole := s.GetStake(holder, false)
		if whole == nil {
			// something wrong. but harmless for now.
			return nil
		}
		var target *types.Stake
		if height == 0 {
			target = s.GetUnlockedStake(holder, false)
		} else if height > 0 {
			target = s.GetLockedStake(holder, height, false)
		} else { // height must not be negative
			return code.TxErrUnknown
		}
		whole.Amount.Sub(&target.Amount)
		if whole.Amount.Sign() == 0 {
			// whole stake for this holder goes to zero. need to check this is
			// allowed.

			// check if there is a delegate appointed to this stake
			ds := s.GetDelegatesByDelegatee(holder, false)
			if len(ds) > 0 {
				return code.TxErrDelegateExists
			}

			// check if this is the last stake
			ts := s.GetTopStakes(2, false)
			if len(ts) == 1 {
				// requested 2 but got 1. it means this is the last validator.
				return code.TxErrLastValidator
			}
		}
	}

	return nil
}

func (s Store) SetUnlockedStake(holder crypto.Address, stake *types.Stake) error {
	b, err := json.Marshal(stake)
	if err != nil {
		return code.TxErrBadParam
	}

	// condition checks
	err = s.checkValidatorMatch(holder, stake)
	if err != nil {
		return err
	}
	err = s.checkStakeDeletion(holder, stake, 0)
	if err != nil {
		return err
	}

	// clean up
	es := s.GetEffStake(holder, false)
	if es != nil {
		before := makeEffStakeKey(s.GetEffStake(holder, false).Amount, holder)
		if s.indexEffStake.Has(before) {
			s.indexEffStake.Delete(before)
		}
	}
	// update
	if stake.Amount.Sign() == 0 {
		s.remove(makeStakeKey(holder))
		s.indexValidator.Delete(stake.Validator.Address())
	} else {
		s.set(makeStakeKey(holder), b)
		s.indexValidator.SetSync(stake.Validator.Address(), holder)
		after := makeEffStakeKey(s.GetEffStake(holder, false).Amount, holder)
		s.indexEffStake.SetSync(after, nil)
	}

	return nil
}

// SetLockedStake stores a stake locked at *height*. The stake's height is
// decremented each time when LoosenLockedStakes is called.
func (s Store) SetLockedStake(holder crypto.Address, stake *types.Stake, height int64) error {
	b, err := json.Marshal(stake)
	if err != nil {
		return code.TxErrBadParam
	}

	// condition checks
	err = s.checkValidatorMatch(holder, stake)
	if err != nil {
		return err
	}
	err = s.checkStakeDeletion(holder, stake, height)
	if err != nil {
		return err
	}

	// clean up
	es := s.GetEffStake(holder, false)
	if es != nil {
		before := makeEffStakeKey(s.GetEffStake(holder, false).Amount, holder)
		if s.indexEffStake.Has(before) {
			s.indexEffStake.Delete(before)
		}
	}

	// update
	if stake.Amount.Sign() == 0 {
		s.remove(makeLockedStakeKey(holder, height))
		s.indexValidator.Delete(stake.Validator.Address())
	} else {
		s.set(makeLockedStakeKey(holder, height), b)
		s.indexValidator.SetSync(stake.Validator.Address(), holder)
		after := makeEffStakeKey(s.GetEffStake(holder, false).Amount, holder)
		s.indexEffStake.SetSync(after, nil)
	}

	return nil
}

func (s Store) SlashStakes(holder crypto.Address, amount types.Currency, committed bool) {
	zeroAmount := new(types.Currency).Set(0)

	total := s.GetStake(holder, committed)
	if total == nil {
		return
	}

	unlockedStake := s.GetUnlockedStake(holder, committed)
	lockedStakes, heights := s.GetLockedStakesWithHeight(holder, committed)

	// slash until amount gets 0
	// slash unlockedStake first, then lockedStakes

	if unlockedStake != nil {
		switch amount.Cmp(&unlockedStake.Amount.Int) {
		case -1: // amount < unlockedStake.Amount
			unlockedStake.Amount.Sub(&amount)
			amount.Set(0)
		case 0: // amount == unlockedStake.Amount
			unlockedStake.Amount.Set(0)
			amount.Set(0)
		case 1: // amount > unlockedStake.Amount
			unlockedStake.Amount.Set(0)
			amount.Sub(&unlockedStake.Amount)
		}

		s.SetUnlockedStake(holder, unlockedStake)

		// check end of slash
		if amount.Equals(zeroAmount) {
			return
		}
	}

	// from height(close to unlocked) to EOA
	for i := len(lockedStakes) - 1; i >= 0; i-- {
		lockedStake := lockedStakes[i]
		height := heights[i]

		switch amount.Cmp(&lockedStake.Amount.Int) {
		case -1: // amount < lockedStake.Amount
			lockedStake.Amount.Sub(&amount)
			amount.Set(0)
		case 0: // amount == lockedStake.Amount
			lockedStake.Amount.Set(0)
			amount.Set(0)
		case 1: // amount > lockedStake.Amount
			lockedStake.Amount.Set(0)
			amount.Sub(&lockedStake.Amount)
		}

		s.SetLockedStake(holder, lockedStake, height)

		if amount.Equals(zeroAmount) {
			break
		}
	}
}

func (s Store) UnlockStakes(holder crypto.Address, height int64, committed bool) {
	start := makeLockedStakeKey(holder, 0)
	end := makeLockedStakeKey(holder, height)

	unlocked := s.GetUnlockedStake(holder, committed)

	imt, err := s.getImmutableTree(committed)
	if err != nil {
		return
	}

	imt.IterateRangeInclusive(start, end, true, func(key []byte, value []byte, version int64) bool {
		stake := new(types.Stake)
		err := json.Unmarshal(value, stake)
		if err != nil {
			// We cannot recover from this error.
			// Since this function returns nothing, just skip this stake.
			return false // same as 'continue'
		}
		s.remove(key)
		if unlocked == nil {
			unlocked = stake
		} else {
			unlocked.Amount.Add(&stake.Amount)
		}
		return false
	})
	s.SetUnlockedStake(holder, unlocked)
}

func (s Store) LoosenLockedStakes(committed bool) {
	imt, err := s.getImmutableTree(committed)
	if err != nil {
		return
	}

	imt.IterateRangeInclusive(prefixStake, nil, true, func(key []byte, value []byte, version int64) bool {
		if !bytes.HasPrefix(key, prefixStake) {
			return false
		}

		if len(key) == len(prefixStake)+crypto.AddressSize {
			// unlocked stake
			return false // continue
		}

		holder, height := splitLockedStakeKey(key)
		if holder == nil || height <= 0 {
			// db corruption detected. but we can do nothing here. just skip.
			return false // continue
		}

		stake := new(types.Stake)
		err := json.Unmarshal(value, stake)
		if err != nil {
			// We cannot recover from this error.
			// Since this function returns nothing, just skip this stake.
			return false // continue
		}

		s.remove(key)
		height--
		if height == 0 {
			unlocked := s.GetUnlockedStake(holder, committed)
			if unlocked == nil {
				unlocked = stake
			} else {
				unlocked.Amount.Add(&stake.Amount)
			}
			err := s.SetUnlockedStake(holder, unlocked)
			if err != nil {
				return false // continue
			}
		} else {
			err := s.SetLockedStake(holder, stake, height)
			if err != nil {
				return false // continue
			}
		}
		return false
	})
}

func makeEffStakeKey(amount types.Currency, holder crypto.Address) []byte {
	key := make([]byte, 32+20) // 256-bit integer + 20-byte address
	b := amount.Bytes()
	copy(key[32-len(b):], b)
	copy(key[32:], holder)
	return key
}

func (s Store) GetStake(holder crypto.Address, committed bool) *types.Stake {
	stake := s.GetUnlockedStake(holder, committed)

	stakes := s.GetLockedStakes(holder, committed)
	for _, item := range stakes {
		if stake == nil {
			stake = item
		} else {
			// check db integrity
			if !bytes.Equal(stake.Validator[:], item.Validator[:]) {
				return nil
			}
			stake.Amount.Add(&item.Amount)
		}
	}

	return stake
}

func (s Store) GetUnlockedStake(holder crypto.Address, committed bool) *types.Stake {
	b := s.get(makeStakeKey(holder), committed)
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

func (s Store) GetLockedStake(holder crypto.Address, height int64, committed bool) *types.Stake {
	b := s.get(makeLockedStakeKey(holder, height), committed)
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

func (s Store) GetLockedStakes(holder crypto.Address, committed bool) []*types.Stake {
	holderKey := makeStakeKey(holder)
	start := makeLockedStakeKey(holder, 0)

	var stakes []*types.Stake
	// XXX: This routine may be used to get all free and locked stakes for a
	// holder. But, let's differentiate getUnlockedStake() and
	// GetLockedStakes() for now.

	imt, err := s.getImmutableTree(committed)
	if err != nil {
		return nil
	}

	imt.IterateRangeInclusive(start, nil, false, func(key []byte, value []byte, version int64) bool {
		if !bytes.HasPrefix(key, holderKey) {
			return false
		}

		stake := new(types.Stake)
		err := json.Unmarshal(value, stake)
		if err != nil {
			// We cannot recover from this error
			return false
		}

		stakes = append(stakes, stake)

		return false
	})

	return stakes
}

func (s Store) GetLockedStakesWithHeight(holder crypto.Address, committed bool) ([]*types.Stake, []int64) {
	holderKey := makeStakeKey(holder)
	start := makeLockedStakeKey(holder, 0)

	var (
		stakes  []*types.Stake
		heights []int64
	)

	imt, err := s.getImmutableTree(committed)
	if err != nil {
		return nil, nil
	}

	imt.IterateRangeInclusive(start, nil, false, func(key []byte, value []byte, version int64) bool {
		if !bytes.HasPrefix(key, holderKey) {
			return false
		}

		stake := new(types.Stake)
		err := json.Unmarshal(value, stake)
		if err != nil {
			// We cannot recover from this error
			return false
		}

		height := binary.BigEndian.Uint64(key[len(prefixStake)+crypto.AddressSize:])

		stakes = append(stakes, stake)
		heights = append(heights, int64(height))

		return false
	})

	return stakes, heights
}
func (s Store) GetStakeByValidator(addr crypto.Address, committed bool) *types.Stake {
	holder := s.GetHolderByValidator(addr, committed)
	if holder == nil {
		return nil
	}
	return s.GetStake(holder, committed)
}

func (s Store) GetHolderByValidator(addr crypto.Address, committed bool) []byte {
	return s.indexValidator.Get(addr)
}

// Delegate store
func getDelegateKey(holder []byte) []byte {
	return append(prefixDelegate, holder...)
}

// Update data on stateDB, indexDelegator, indexEffStake
func (s Store) SetDelegate(holder crypto.Address, delegate *types.Delegate) error {
	b, err := json.Marshal(delegate)
	if err != nil {
		return code.TxErrBadParam
	}
	// before state update
	es := s.GetEffStake(delegate.Delegatee, false)
	if es == nil {
		return code.TxErrNoStake
	}

	// make effStakeKey to find its corresponding value
	before := makeEffStakeKey(es.Amount, delegate.Delegatee)
	if s.indexEffStake.Has(before) {
		s.indexEffStake.Delete(before)
	}

	// upadate
	if delegate.Amount.Sign() == 0 {
		s.remove(getDelegateKey(holder))
		s.indexDelegator.Delete(append(delegate.Delegatee, holder...))
	} else {
		s.set(getDelegateKey(holder), b)
		s.indexDelegator.SetSync(append(delegate.Delegatee, holder...), nil)
	}

	after := makeEffStakeKey(
		s.GetEffStake(delegate.Delegatee, false).Amount,
		delegate.Delegatee,
	)

	s.indexEffStake.SetSync(after, nil)

	return nil
}

func (s Store) GetDelegate(holder crypto.Address, committed bool) *types.Delegate {
	b := s.get(getDelegateKey(holder), committed)
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

func (s Store) GetDelegateEx(holder crypto.Address, committed bool) *types.DelegateEx {
	delegate := s.GetDelegate(holder, committed)
	if delegate == nil {
		return nil
	}
	return &types.DelegateEx{holder, delegate}
}

func (s Store) GetDelegatesByDelegatee(delegatee crypto.Address, committed bool) []*types.DelegateEx {
	var itr tmdb.Iterator = s.indexDelegator.Iterator(delegatee, nil)
	defer itr.Close()

	var delegates []*types.DelegateEx
	for ; itr.Valid() && bytes.HasPrefix(itr.Key(), delegatee); itr.Next() {
		delegator := itr.Key()[len(delegatee):]
		delegateEx := s.GetDelegateEx(delegator, committed)
		if delegateEx == nil {
			continue
		}
		delegates = append(delegates, delegateEx)
	}
	return delegates
}

func (s Store) GetEffStake(delegatee crypto.Address, committed bool) *types.Stake {
	stake := s.GetStake(delegatee, committed)
	if stake == nil {
		return nil
	}
	for _, d := range s.GetDelegatesByDelegatee(delegatee, committed) {
		stake.Amount.Add(&d.Amount)
	}
	return stake
}

func (s Store) GetTopStakes(max uint64, committed bool) []*types.Stake {
	var stakes []*types.Stake
	var itr tmdb.Iterator = s.indexEffStake.ReverseIterator(nil, nil)
	var cnt uint64 = 0
	for ; itr.Valid(); itr.Next() {
		if cnt >= max {
			break
		}
		key := itr.Key()
		var amount types.Currency
		amount.SetBytes(key[:32])
		holder := key[32:]
		stake := s.GetStake(holder, committed)
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
	s.set(getParcelKey(parcelID), b)
	return nil
}

func (s Store) GetParcel(parcelID []byte, committed bool) *types.ParcelValue {
	b := s.get(getParcelKey(parcelID), committed)
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
	s.remove(getParcelKey(parcelID))
}

// Request store
func getRequestKey(buyer crypto.Address, parcelID []byte) (buyerParcelKey, parcelBuyerKey []byte) {
	buyerParcelKey = append(prefixRequest, append(append(buyer, ':'), parcelID...)...)
	parcelBuyerKey = append(prefixRequest, append(append(parcelID, ':'), buyer...)...)
	return
}

func splitParcelBuyerKey(prefix, key []byte) (parcelID []byte, buyer crypto.Address) {
	// prefix + parcelID + buyer
	parcelID = key[len(prefix) : len(key)-crypto.AddressSize-1]
	buyer = key[len(key)-crypto.AddressSize:]
	return
}

func (s Store) SetRequest(buyer crypto.Address, parcelID []byte, value *types.RequestValue) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	buyerParcelKey, parcelBuyerKey := getRequestKey(buyer, parcelID)

	// parcelBuyerKey has only nil as value to use it as index
	s.set(buyerParcelKey, b)
	s.set(parcelBuyerKey, []byte{0})

	return nil
}

func (s Store) GetRequest(buyer crypto.Address, parcelID []byte, committed bool) *types.RequestValue {
	buyerParcelKey, _ := getRequestKey(buyer, parcelID)

	b := s.get(buyerParcelKey, committed)
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

func (s Store) GetRequests(parcelID []byte, committed bool) []*types.RequestValueEx {
	prefixRequestKey := append(prefixRequest, append(parcelID, ':')...)
	requests := []*types.RequestValueEx{}

	imt, err := s.getImmutableTree(committed)
	if err != nil {
		return nil
	}

	imt.IterateRangeInclusive(prefixRequestKey, nil, true,
		func(key []byte, value []byte, version int64) bool {
			if !bytes.HasPrefix(key, prefixRequestKey) {
				return false
			}

			// TODO: Is this really the best ?
			parcelID, buyer := splitParcelBuyerKey(prefixRequest, key)
			requestValue := s.GetRequest(buyer, parcelID, committed)
			if requestValue == nil {
				return false
			}
			request := types.RequestValueEx{
				RequestValue: requestValue,
				Buyer:        buyer,
			}

			requests = append(requests, &request)

			return false
		},
	)

	return requests
}

func (s Store) DeleteRequest(buyer crypto.Address, parcelID []byte) {
	buyerParcelKey, parcelBuyerKey := getRequestKey(buyer, parcelID)

	s.remove(buyerParcelKey)
	s.remove(parcelBuyerKey)
}

// Usage store
func getUsageKey(buyer crypto.Address, parcelID []byte) (buyerParcelKey, parcelBuyerKey []byte) {
	buyerParcelKey = append(prefixUsage, append(append(buyer, ':'), parcelID...)...)
	parcelBuyerKey = append(prefixUsage, append(append(parcelID, ':'), buyer...)...)
	return
}

func (s Store) SetUsage(buyer crypto.Address, parcelID []byte, value *types.UsageValue) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	buyerParcelKey, parcelBuyerKey := getUsageKey(buyer, parcelID)

	// parcelBuyerKey has only nil as value to use it as index
	s.set(buyerParcelKey, b)
	s.set(parcelBuyerKey, []byte{0})

	return nil
}

func (s Store) GetUsage(buyer crypto.Address, parcelID []byte, committed bool) *types.UsageValue {
	buyerParcelKey, _ := getUsageKey(buyer, parcelID)
	b := s.get(buyerParcelKey, committed)
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

func (s Store) GetUsages(parcelID []byte, committed bool) []*types.UsageValueEx {
	prefixUsageKey := append(prefixUsage, append(parcelID, ':')...)
	usages := []*types.UsageValueEx{}

	imt, err := s.getImmutableTree(committed)
	if err != nil {
		return nil
	}

	imt.IterateRangeInclusive(prefixUsageKey, nil, true,
		func(key []byte, value []byte, version int64) bool {
			if !bytes.HasPrefix(key, prefixUsageKey) {
				return false
			}

			// TODO: Is this really the best ?
			parcelID, buyer := splitParcelBuyerKey(prefixUsage, key)
			usage := types.UsageValueEx{
				UsageValue: s.GetUsage(buyer, parcelID, committed),
				Buyer:      buyer,
			}

			usages = append(usages, &usage)

			return false
		},
	)

	return usages
}

func (s Store) DeleteUsage(buyer crypto.Address, parcelID []byte) {
	buyerParcelKey, parcelBuyerKey := getUsageKey(buyer, parcelID)

	s.remove(buyerParcelKey)
	s.remove(parcelBuyerKey)
}

func (s Store) GetValidators(max uint64, committed bool) abci.ValidatorUpdates {
	var vals abci.ValidatorUpdates
	stakes := s.GetTopStakes(max, committed)
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
