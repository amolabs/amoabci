package store

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	tmdb "github.com/tendermint/tm-db"

	"github.com/amolabs/amoabci/amo/types"
)

func TestIncentiveRecord(t *testing.T) {
	incentives := []IncentiveInfo{
		{1, makeAccAddr("acc1"), new(types.Currency).Set(1)},
		{1, makeAccAddr("acc2"), new(types.Currency).Set(2)},
		{2, makeAccAddr("acc3"), new(types.Currency).Set(3)},
		{3, makeAccAddr("acc4"), new(types.Currency).Set(4)},
		{4, makeAccAddr("acc2"), new(types.Currency).Set(3)},
	}

	s, err := NewStore(nil, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)

	// unavailable height
	err = s.AddIncentiveRecord(-1, makeAccAddr("test"), new(types.Currency).Set(1))
	assert.Error(t, err)

	// missing address
	err = s.AddIncentiveRecord(1, nil, new(types.Currency).Set(1))
	assert.Error(t, err)

	// unavailable currency
	err = s.AddIncentiveRecord(-1, makeAccAddr("test"), new(types.Currency))
	assert.Error(t, err)

	// 0 incentive
	err = s.AddIncentiveRecord(1, makeAccAddr("test"), new(types.Currency).Set(0))
	assert.Error(t, err)

	// set incentives
	for _, inc := range incentives {
		err = s.AddIncentiveRecord(inc.BlockHeight, inc.Address, inc.Amount)
		assert.NoError(t, err)
	}

	// GetBlockIncentiveRecords Test
	blockIncentiveRecords := s.GetBlockIncentiveRecords(1)
	assert.Equal(t, 2, len(blockIncentiveRecords))

	sort.Slice(blockIncentiveRecords, func(i, j int) bool {
		return blockIncentiveRecords[i].Amount.LessThan(blockIncentiveRecords[j].Amount)
	})

	assert.Equal(t, incentives[0], blockIncentiveRecords[0])
	assert.Equal(t, incentives[1], blockIncentiveRecords[1])

	// GetAddressIncentiveRecords Test
	addressIncentiveRecords := s.GetAddressIncentiveRecords(makeAccAddr("acc2"))
	assert.Equal(t, 2, len(addressIncentiveRecords))

	sort.Slice(addressIncentiveRecords, func(i, j int) bool {
		return addressIncentiveRecords[i].BlockHeight < addressIncentiveRecords[j].BlockHeight
	})

	assert.Equal(t, incentives[1], addressIncentiveRecords[0])
	assert.Equal(t, incentives[4], addressIncentiveRecords[1])

	// GetIncentive Test
	incentive := s.GetIncentiveRecord(2, makeAccAddr("acc3"))
	assert.Equal(t, incentives[2], incentive)

	// GetIncentive Test
	incentive = s.GetIncentiveRecord(3, makeAccAddr("acc4"))
	assert.Equal(t, incentives[3], incentive)
}

func TestPenaltyRecord(t *testing.T) {
	penalties := []PenaltyInfo{
		{1, makeAccAddr("acc1"), new(types.Currency).Set(1)},
		{1, makeAccAddr("acc2"), new(types.Currency).Set(2)},
		{2, makeAccAddr("acc3"), new(types.Currency).Set(3)},
		{3, makeAccAddr("acc4"), new(types.Currency).Set(4)},
		{4, makeAccAddr("acc2"), new(types.Currency).Set(3)},
	}

	s, err := NewStore(nil, tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB(), tmdb.NewMemDB())
	assert.NoError(t, err)

	// unavailable height
	err = s.AddPenaltyRecord(-1, makeAccAddr("test"), new(types.Currency).Set(1))
	assert.Error(t, err)

	// missing address
	err = s.AddPenaltyRecord(1, nil, new(types.Currency).Set(1))
	assert.Error(t, err)

	// unavailable currency
	err = s.AddPenaltyRecord(-1, makeAccAddr("test"), new(types.Currency))
	assert.Error(t, err)

	// 0 penalty
	err = s.AddPenaltyRecord(1, makeAccAddr("test"), new(types.Currency).Set(0))
	assert.Error(t, err)

	// set penalties
	for _, pen := range penalties {
		err = s.AddPenaltyRecord(pen.BlockHeight, pen.Address, pen.Amount)
		assert.NoError(t, err)
	}

	// GetBlockPenaltyRecords Test
	blockPenaltyRecords := s.GetBlockPenaltyRecords(1)
	assert.Equal(t, 2, len(blockPenaltyRecords))

	sort.Slice(blockPenaltyRecords, func(i, j int) bool {
		return blockPenaltyRecords[i].Amount.LessThan(blockPenaltyRecords[j].Amount)
	})

	assert.Equal(t, penalties[0], blockPenaltyRecords[0])
	assert.Equal(t, penalties[1], blockPenaltyRecords[1])

	// GetAddressPenaltyRecords Test
	addressPenaltyRecords := s.GetAddressPenaltyRecords(makeAccAddr("acc2"))
	assert.Equal(t, 2, len(addressPenaltyRecords))

	sort.Slice(addressPenaltyRecords, func(i, j int) bool {
		return addressPenaltyRecords[i].BlockHeight < addressPenaltyRecords[j].BlockHeight
	})

	assert.Equal(t, penalties[1], addressPenaltyRecords[0])
	assert.Equal(t, penalties[4], addressPenaltyRecords[1])

	// GetPenalty Test
	penalty := s.GetPenaltyRecord(2, makeAccAddr("acc3"))
	assert.Equal(t, penalties[2], penalty)

	// GetPenalty Test
	penalty = s.GetPenaltyRecord(3, makeAccAddr("acc4"))
	assert.Equal(t, penalties[3], penalty)
}
