package blockchain

import (
	"errors"
)

// CheckBlockBindingTx: check avaiability of given txHeight
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
// gracePeriod: 5
//           ====^ (h:14, f: 10, t:14)
//            ====^ (h:15, f: 11, t:15)

func checkBlockBindingTx(txHeight, blockHeight, gracePeriod int64) error {
	var (
		fromHeight int64 = 0
		toHeight   int64 = blockHeight
	)

	if gracePeriod < blockHeight {
		fromHeight = blockHeight - gracePeriod + 1
	}

	if !(fromHeight <= txHeight && txHeight <= toHeight) {
		return errors.New("failed to bind tx to block")
	}

	return nil
}
