package store

import (
	"encoding/binary"
	"encoding/json"
)

// indexBlockTx
// key: block height
// value: hash of txs

// indexTxBlock
// key: tx hash
// value: block height

var (
	prefixIndexBlockTx = []byte("blocktx")
	prefixIndexTxBlock = []byte("txblock")
)

func (s Store) AddTxIndexer(height int64, txs [][]byte) {
	hb := make([]byte, 8)
	binary.BigEndian.PutUint64(hb, uint64(height))
	txsJSON, _ := json.Marshal(txs)

	// update indexBlockTx
	s.indexBlockTx.Set(hb, txsJSON)

	s.indexTxBlock.NewBatch().Write()
	// update indexTxBlock
	for _, tx := range txs {
		s.indexTxBlock.Set(tx, hb)
	}
}

func (s Store) GetTxIndexerHeight(height int64) [][]byte {
	txs := [][]byte{}

	hb := make([]byte, 8)
	binary.BigEndian.PutUint64(hb, uint64(height))
	value := s.indexBlockTx.Get(hb)
	if value == nil {
		return nil
	}

	err := json.Unmarshal(value, &txs)
	if err != nil {
		return nil
	}

	return txs
}

func (s Store) GetTxIndexerHash(tx []byte) int64 {
	height := int64(0)

	if !s.indexTxBlock.Has(tx) {
		return height
	}

	value := s.indexTxBlock.Get(tx)
	height = int64(binary.BigEndian.Uint64(value))

	return height
}

func (s Store) DeleteTxIndexer(height int64) {
	hb := make([]byte, 8)
	binary.BigEndian.PutUint64(hb, uint64(height))

	if !s.indexBlockTx.Has(hb) {
		return
	}

	// get txs from indexBlockTx
	txs := s.GetTxIndexerHeight(height)

	// delete indexBlockTx of given height
	s.indexBlockTx.Delete(hb)

	// delete txs of given height
	for _, tx := range txs {
		s.indexTxBlock.Delete(tx)
	}
}

func (s Store) PurgeTxIndexer() {
	itr := s.indexBlockTx.Iterator(nil, nil)
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		s.indexBlockTx.Delete(itr.Key())
	}

	itr = s.indexTxBlock.Iterator(nil, nil)
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		s.indexTxBlock.Delete(itr.Key())
	}
}
