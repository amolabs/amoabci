package blockchain

import (
	"math/big"

	"github.com/amolabs/amoabci/amo/types"
)

// r = (weight * stake / total) * base
// TODO: eliminate ambiguity in float computation
func partialAmount(weight float64, stake, total *big.Float, base *types.Currency) *types.Currency {
	var wf, t1f, t2f big.Float
	wf.SetFloat64(weight)
	t1f.Set(stake)
	t1f.Mul(&wf, &t1f)
	t2f.Set(total)
	t1f.Quo(&t1f, &t2f)
	t2f.SetInt(&base.Int)
	t1f.Mul(&t1f, &t2f)
	r := types.Currency{}
	t1f.Int(&r.Int)
	return &r
}
