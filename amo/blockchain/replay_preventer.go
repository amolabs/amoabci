package blockchain

import (
	"crypto/sha256"

	"github.com/amolabs/amoabci/amo/store"
)

// ReplayPreventer: check whether given tx is already delivered
// - IndexRange: range in which txs are indexed and checked

// IndexRange: 10
// 0    5    10   15   20   25   30
// |----|----|----|----|----|----|
// ^ (h:0, f:1, t:0 - initChain)
//  ^ (h:1, f:1, t:1)
//  =^ (h:2, f:1, t:2)
//  ==^ (h:3, f:1, t:3)
//  ===^ (h:4, f:1, t:4)
//  ...
//  =========^ (h:10, f:1, t:10)
//   =========^ (h:11, f:2, t:11)
//    =========^ (h:12, f:3, t:12)
//     =========^ (h:13, f:4, t:13)
//
//
//   \       /              (NewReplayPreventer)
//    \     /               : txBucket is initialized
//     \___/
//
//   \       /    \ @@@   / (Append)
//    \     /  ->  \@@@@@/  : txBucket gets filled with txs
//     \___/        \@@@/
//
//   \@@@@@@@/     |@@@@  | (Index)
//    \@@@@@/  ->  |@@@@@@| : txs in txBucket are recorded on DISK storage
//     \@@@/   |   |@@@@@@|   and old txs out of gracePeriod gets deleted
//             |   +------+
//             |
//             |  \       /
//             ->  \     /  : txBucket gets emptyed
//                  \___/

type (
	TxHash   [32]byte
	TxBucket map[TxHash]bool
)

type ReplayPreventer struct {
	store *store.Store
	// TODO: maybe caching txs is required ?

	indexRange           uint64
	fromHeight, toHeight uint64

	txBucket TxBucket
}

func NewReplayPreventer(
	store *store.Store,
	indexRange uint64,
	height int64,
) ReplayPreventer {

	rp := ReplayPreventer{
		store:      store,
		indexRange: indexRange,
		fromHeight: 1,
		toHeight:   uint64(height),
		txBucket:   make(TxBucket),
	}

	return rp
}

// Update() is called at BeginBlock()
func (rp *ReplayPreventer) Update() {
	rp.toHeight += 1

	if rp.toHeight-rp.fromHeight == rp.indexRange {
		rp.fromHeight += 1
	}
}

// Check() is called at CheckTx()
func (rp *ReplayPreventer) Check(tx []byte) bool {
	txHash := sha256.Sum256(tx)

	// check if given tx already exists in txBucket
	if _, exist := rp.txBucket[txHash]; exist {
		return false
	}

	// check if given tx already exists in txIndexer
	height := rp.store.TxIndexerGetHeight(txHash[:])
	if height > 0 {
		return false
	}

	return true
}

// Append() is called at DeliverTx()
func (rp *ReplayPreventer) Append(tx []byte) bool {
	if ok := rp.Check(tx); !ok {
		return false
	}

	txHash := sha256.Sum256(tx)
	rp.txBucket[txHash] = true

	return true
}

// Index() is called at EndBlock()
func (rp *ReplayPreventer) Index() {
	txs := make([][]byte, 0, len(rp.txBucket))
	for key, _ := range rp.txBucket {
		tx := key
		txs = append(txs, tx[:])
	}

	rp.store.AddTxIndexer(int64(rp.toHeight), txs)

	if rp.toHeight-rp.fromHeight+1 == rp.indexRange {
		rp.store.TxIndexerDelete(int64(rp.fromHeight))
	}

	// clear txBucket
	for key, _ := range rp.txBucket {
		delete(rp.txBucket, key)
	}
}
