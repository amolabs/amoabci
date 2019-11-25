package blockchain

// BlockBindingManager: check avaiability of given height
// - GracePeriod: period for which tx can be accepted

// GracePeriod: 10
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
	GracePeriod          uint64
	FromHeight, ToHeight uint64
}

func NewBlockBindingManager(gracePeriod uint64, height int64) BlockBindingManager {
	bbm := BlockBindingManager{
		GracePeriod: gracePeriod,
		FromHeight:  1,
		ToHeight:    uint64(height),
	}

	if bbm.ToHeight != 0 && bbm.ToHeight-bbm.FromHeight >= bbm.GracePeriod {
		bbm.FromHeight = bbm.ToHeight - bbm.GracePeriod + 1
	}

	return bbm
}

func (bbm *BlockBindingManager) Check(height int64) bool {
	heightU := uint64(height)
	if bbm.FromHeight <= heightU && heightU <= bbm.ToHeight {
		return true
	}

	return false
}

func (bbm *BlockBindingManager) Update() {
	bbm.ToHeight += 1

	if bbm.ToHeight-bbm.FromHeight == bbm.GracePeriod {
		bbm.FromHeight += 1
	}
}
