package blockchain

import (
	"crypto/sha256"
	"errors"

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

	indexRange int64
	fromHeight int64

	txBucket TxBucket
}

func NewReplayPreventer(
	store *store.Store,
	blockHeight int64,
	indexRange int64,
) ReplayPreventer {

	fromHeight := int64(1)
	if blockHeight != 0 && blockHeight-fromHeight >= indexRange {
		fromHeight = blockHeight - indexRange + 1
	}

	return ReplayPreventer{
		store:      store,
		indexRange: indexRange,
		fromHeight: fromHeight,
		txBucket:   make(TxBucket),
	}
}

// Update() is called at BeginBlock()
func (rp *ReplayPreventer) Update(blockHeight, indexRange int64) {
	// indexRange gets shrinked
	if rp.indexRange > indexRange {
		// flush orphan data
		for i := rp.fromHeight; i <= blockHeight-indexRange; i++ {
			rp.store.TxIndexerDelete(int64(i))
		}
	}

	rp.indexRange = indexRange

	if blockHeight > indexRange {
		rp.fromHeight = blockHeight - indexRange + 1
	}
}

// Check() is called at CheckTx()
func (rp *ReplayPreventer) Check(tx []byte, txHeight, blockHeight int64) (TxHash, error) {
	// check if given tx is block bound
	err := checkBlockBindingTx(txHeight, blockHeight, rp.indexRange)
	if err != nil {
		return TxHash{}, err
	}

	txHash := sha256.Sum256(tx)

	// check if given tx already exists in txBucket
	if _, exist := rp.txBucket[txHash]; exist {
		return TxHash{}, errors.New("already processed tx")
	}

	// check if given tx already exists in txIndexer
	if rp.store.TxIndexerGetHeight(txHash[:]) > 0 {
		return TxHash{}, errors.New("already processed tx")
	}

	return txHash, nil
}

// Append() is called at DeliverTx()
func (rp *ReplayPreventer) Append(tx []byte, txHeight, blockHeight int64) error {
	txHash, err := rp.Check(tx, txHeight, blockHeight)
	if err != nil {
		return err
	}

	rp.txBucket[txHash] = true

	return nil
}

// Index() is called at EndBlock()
func (rp *ReplayPreventer) Index(blockHeight int64) {
	txs := make([][]byte, 0, len(rp.txBucket))
	for key, _ := range rp.txBucket {
		tx := key
		txs = append(txs, tx[:])
	}

	rp.store.AddTxIndexer(blockHeight, txs)

	if blockHeight-rp.fromHeight+1 == rp.indexRange {
		rp.store.TxIndexerDelete(rp.fromHeight)
	}

	// clear txBucket
	for key, _ := range rp.txBucket {
		delete(rp.txBucket, key)
	}
}
