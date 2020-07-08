package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/kv"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type MissRuns struct {
	store              *store.Store
	hibernateThreshold int64
	hibernatePeriod    int64
}

func makeRunKey(val crypto.Address, start int64) []byte {
	buf := make([]byte, crypto.AddressSize+8)
	copy(buf[:crypto.AddressSize], val)
	binary.BigEndian.PutUint64(buf[crypto.AddressSize:], uint64(start))

	return buf[:crypto.AddressSize+8]
}

func NewMissRuns(store *store.Store, db tmdb.DB, threshold, period int64) *MissRuns {
	return &MissRuns{
		store:              store,
		hibernateThreshold: threshold,
		hibernatePeriod:    period,
	}
}

func (m MissRuns) UpdateMissRuns(height int64, vals []crypto.Address) (doValUpdate bool, evs []abci.Event, err error) {
	doValUpdate = false

	runDB := m.store.GetMissRunDB()
	batch := runDB.NewBatch()
	defer batch.Close()

	unfinishedRuns := [][]byte{}
	itr, err := runDB.Iterator(nil, nil)
	if err != nil {
		return
	}
	defer itr.Close()
	// guard for rewind
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		b := k[crypto.AddressSize:]
		runStart := int64(binary.BigEndian.Uint64(b))
		b = itr.Value()
		runLen := int64(binary.BigEndian.Uint64(b))

		if runStart > height {
			batch.Delete(k)
		} else if runStart+runLen > height || runLen == 0 {
			// run cropped, and made to be an unfinished run
			unfinishedRuns = append(unfinishedRuns, k)
		}
	}

	// process given missing validators
	for _, v := range vals {
		hit := false
		for i, r := range unfinishedRuns {
			runVal := r[:crypto.AddressSize]
			b := r[crypto.AddressSize:]
			runStart := int64(binary.BigEndian.Uint64(b))
			if bytes.Equal(runVal, v) {
				// continue unfinished run
				hit = true
				// TODO: do Set() only when there is a change of value
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, uint64(0))
				batch.Set(r, buf)
				//fmt.Println(crypto.Address(runVal), height, runStart)
				if height+1-runStart >= m.hibernateThreshold {
					hib := types.Hibernate{
						Start: height,
						End:   height + m.hibernatePeriod,
					}
					m.store.SetHibernate(runVal, &hib)
					addressJson, _ := json.Marshal(vals[i])
					startJson, _ := json.Marshal(height)
					endJson, _ := json.Marshal(height + m.hibernatePeriod)
					ev := abci.Event{
						Type: "hibernate",
						Attributes: []kv.Pair{
							{Key: []byte("validator"), Value: addressJson},
							{Key: []byte("start"), Value: startJson},
							{Key: []byte("end"), Value: endJson},
						},
					}
					evs = append(evs, ev)
					doValUpdate = true
				}
				l := len(unfinishedRuns)
				unfinishedRuns[i] = unfinishedRuns[l-1]
				unfinishedRuns[l-1] = nil
				unfinishedRuns = unfinishedRuns[:l-1]
				break
			}
		}
		// not in the previous set of missing validators
		// start a new run
		if !hit {
			k := makeRunKey(v, height)
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, uint64(0))
			batch.Set(k, buf)
		}
	}

	// close unfinished runs
	for _, r := range unfinishedRuns {
		b := r[crypto.AddressSize:]
		runStart := int64(binary.BigEndian.Uint64(b))
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(height-runStart))
		batch.Set(r, buf)
	}

	// TODO: remove too old runs

	batch.Write()

	return
}

func (m MissRuns) getLastMissRun(val crypto.Address) (start, length int64) {
	start = 0
	length = 0

	runDB := m.store.GetMissRunDB()
	b := make(crypto.Address, crypto.AddressSize+8)
	copy(b, val)
	end := append(b[:crypto.AddressSize],
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}...)
	itr, err := runDB.ReverseIterator(val, end)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		runVal := k[:crypto.AddressSize]
		if !bytes.Equal(runVal, val) {
			return
		}
		b := k[crypto.AddressSize:]
		start = int64(binary.BigEndian.Uint64(b))
		b = itr.Value()
		length = int64(binary.BigEndian.Uint64(b))
		return
	}

	return
}

func max(a, b int64) int64 {
	if a >= b {
		return a
	} else {
		return b
	}
}

func min(a, b int64) int64 {
	if a <= b {
		return a
	} else {
		return b
	}
}

func (m MissRuns) GetMissStat(rangeStart, rangeEnd int64) map[string]int64 {
	stat := map[string]int64{}
	runDB := m.store.GetMissRunDB()

	itr, err := runDB.Iterator(nil, nil)
	if err != nil {
		fmt.Println(err)
		return stat
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		val := crypto.Address(k[:crypto.AddressSize])
		b := k[crypto.AddressSize:]
		start := int64(binary.BigEndian.Uint64(b))
		if start > rangeEnd {
			continue
		}
		b = itr.Value()
		length := int64(binary.BigEndian.Uint64(b))
		if length == 0 {
			length = rangeEnd - start + 1
		}
		if start+length <= rangeStart {
			continue
		}

		m1 := max(start, rangeStart)
		m2 := min(start+length-1, rangeEnd)
		count := m2 - m1 + 1
		if stat[val.String()] == 0 {
			stat[val.String()] = count
		} else {
			stat[val.String()] += count
		}
	}

	return stat
}
