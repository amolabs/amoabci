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
	s.indexBlockTx.SetSync(hb, txsJSON)

	batch := s.indexTxBlock.NewBatch()
	defer batch.Close()

	// update indexTxBlock
	for _, tx := range txs {
		batch.Set(tx, hb)
	}

	err := batch.WriteSync()
	if err != nil {
		s.logger.Error("Store", "AddTxIndexer", err.Error())
	}
}

func (s Store) TxIndexerGetHash(height int64) [][]byte {
	txs := [][]byte{}

	hb := make([]byte, 8)
	binary.BigEndian.PutUint64(hb, uint64(height))
	value, err := s.indexBlockTx.Get(hb)
	if err != nil {
		s.logger.Error("Store", "TxIndexerGetHash", err.Error())
		return nil
	}
	if value == nil {
		return nil
	}

	err = json.Unmarshal(value, &txs)
	if err != nil {
		return nil
	}

	return txs
}

func (s Store) TxIndexerGetHeight(txHash []byte) int64 {
	height := int64(0)

	exist, err := s.indexTxBlock.Has(txHash)
	if err != nil {
		s.logger.Error("Store", "TxIndexerGetHeight", err.Error())
		return int64(0)
	}
	if !exist {
		return int64(0)
	}

	value, err := s.indexTxBlock.Get(txHash)
	if err != nil {
		s.logger.Error("Store", "TxIndexerGetHeight", err.Error())
		return int64(0)
	}
	height = int64(binary.BigEndian.Uint64(value))

	return height
}

func (s Store) TxIndexerDelete(height int64) {
	hb := make([]byte, 8)
	binary.BigEndian.PutUint64(hb, uint64(height))

	exist, err := s.indexBlockTx.Has(hb)
	if err != nil {
		s.logger.Error("Store", "TxIndexerDelete", err.Error())
		return
	}
	if !exist {
		return
	}

	// get txs from indexBlockTx
	txs := s.TxIndexerGetHash(height)

	// delete indexBlockTx of given height
	err = s.indexBlockTx.DeleteSync(hb)
	if err != nil {
		s.logger.Error("Store", "TxIndexerDelete", err.Error())
		return
	}

	// delete txs of given height
	for _, tx := range txs {
		err = s.indexTxBlock.DeleteSync(tx)
		if err != nil {
			s.logger.Error("Store", "TxIndexerDelete", err.Error())
			return
		}
	}
}

func (s Store) TxIndexerPurge() {
	itr, err := s.indexBlockTx.Iterator(nil, nil)
	if err != nil {
		s.logger.Error("Store", "TxIndexerPurge", err.Error())
		return
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		s.indexBlockTx.DeleteSync(itr.Key())
	}

	itr, err = s.indexTxBlock.Iterator(nil, nil)
	if err != nil {
		s.logger.Error("Store", "TxIndexerPurge", err.Error())
		return
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		s.indexTxBlock.DeleteSync(itr.Key())
	}
}
