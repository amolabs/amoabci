package blockchain

// BlockBindingManager: check avaiability of given height
// - gracePeriod: period for which tx can be accepted

// gracePeriod: 10
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

type BlockBindingManager struct {
	gracePeriod          uint64
	fromHeight, toHeight uint64
}

func NewBlockBindingManager(height int64, gracePeriod uint64) BlockBindingManager {
	bbm := BlockBindingManager{
		gracePeriod: gracePeriod,
		fromHeight:  1,
		toHeight:    uint64(height),
	}

	if bbm.toHeight != 0 && bbm.toHeight-bbm.fromHeight >= bbm.gracePeriod {
		bbm.fromHeight = bbm.toHeight - bbm.gracePeriod + 1
	}

	return bbm
}

// Update() is called at BeginBlock()
func (bbm *BlockBindingManager) Update() {
	bbm.toHeight += 1

	if bbm.toHeight-bbm.fromHeight == bbm.gracePeriod {
		bbm.fromHeight += 1
	}
}

// Check() is called at CheckTx()
func (bbm *BlockBindingManager) Check(height int64) bool {
	heightU := uint64(height)

	if bbm.fromHeight <= heightU && heightU <= bbm.toHeight {
		return true
	}

	return false
}
