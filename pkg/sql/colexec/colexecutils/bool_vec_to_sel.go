package colexecutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

func BoolOrUnknownToSelOp(
	input colexecop.Operator, typs []*types.T, vecIdx int,
) (colexecop.Operator, error) {
	__antithesis_instrumentation__.Notify(431671)
	switch typs[vecIdx].Family() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(431672)
		return NewBoolVecToSelOp(input, vecIdx), nil
	case types.UnknownFamily:
		__antithesis_instrumentation__.Notify(431673)

		return NewZeroOp(input), nil
	default:
		__antithesis_instrumentation__.Notify(431674)
		return nil, errors.Errorf("unexpectedly %s is neither bool nor unknown", typs[vecIdx])
	}
}

type BoolVecToSelOp struct {
	colexecop.OneInputHelper
	colexecop.NonExplainable

	OutputCol []bool

	ProcessOnlyOneBatch bool
}

var _ colexecop.ResettableOperator = &BoolVecToSelOp{}

func (p *BoolVecToSelOp) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431675)

	for {
		__antithesis_instrumentation__.Notify(431676)
		batch := p.Input.Next()
		n := batch.Length()
		if n == 0 {
			__antithesis_instrumentation__.Notify(431679)
			return batch
		} else {
			__antithesis_instrumentation__.Notify(431680)
		}
		__antithesis_instrumentation__.Notify(431677)
		outputCol := p.OutputCol

		idx := 0
		if sel := batch.Selection(); sel != nil {
			__antithesis_instrumentation__.Notify(431681)
			sel = sel[:n]
			for s := range sel {
				__antithesis_instrumentation__.Notify(431682)
				i := sel[s]
				var inc int

				if outputCol[i] {
					__antithesis_instrumentation__.Notify(431684)
					inc = 1
				} else {
					__antithesis_instrumentation__.Notify(431685)
				}
				__antithesis_instrumentation__.Notify(431683)
				sel[idx] = i
				idx += inc
			}
		} else {
			__antithesis_instrumentation__.Notify(431686)
			batch.SetSelection(true)
			sel := batch.Selection()
			col := outputCol[:n]
			for i := range col {
				__antithesis_instrumentation__.Notify(431687)
				var inc int

				if col[i] {
					__antithesis_instrumentation__.Notify(431689)
					inc = 1
				} else {
					__antithesis_instrumentation__.Notify(431690)
				}
				__antithesis_instrumentation__.Notify(431688)
				sel[idx] = i
				idx += inc
			}
		}
		__antithesis_instrumentation__.Notify(431678)

		if idx > 0 || func() bool {
			__antithesis_instrumentation__.Notify(431691)
			return p.ProcessOnlyOneBatch == true
		}() == true {
			__antithesis_instrumentation__.Notify(431692)
			batch.SetLength(idx)
			return batch
		} else {
			__antithesis_instrumentation__.Notify(431693)
		}
	}
}

func (p *BoolVecToSelOp) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431694)
	if r, ok := p.Input.(colexecop.Resetter); ok {
		__antithesis_instrumentation__.Notify(431695)
		r.Reset(ctx)
	} else {
		__antithesis_instrumentation__.Notify(431696)
	}
}

func NewBoolVecToSelOp(input colexecop.Operator, colIdx int) colexecop.Operator {
	__antithesis_instrumentation__.Notify(431697)
	d := &selBoolOp{OneInputHelper: colexecop.MakeOneInputHelper(input), colIdx: colIdx}
	ret := &BoolVecToSelOp{OneInputHelper: colexecop.MakeOneInputHelper(d)}
	d.boolVecToSelOp = ret
	return ret
}

type selBoolOp struct {
	colexecop.OneInputHelper
	colexecop.NonExplainable
	boolVecToSelOp *BoolVecToSelOp
	colIdx         int
}

func (d *selBoolOp) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431698)
	batch := d.Input.Next()
	n := batch.Length()
	if n == 0 {
		__antithesis_instrumentation__.Notify(431701)
		return batch
	} else {
		__antithesis_instrumentation__.Notify(431702)
	}
	__antithesis_instrumentation__.Notify(431699)
	inputCol := batch.ColVec(d.colIdx)
	d.boolVecToSelOp.OutputCol = inputCol.Bool()
	if inputCol.MaybeHasNulls() {
		__antithesis_instrumentation__.Notify(431703)

		outputCol := d.boolVecToSelOp.OutputCol
		sel := batch.Selection()
		nulls := inputCol.Nulls()
		if sel != nil {
			__antithesis_instrumentation__.Notify(431704)
			sel = sel[:n]
			for _, i := range sel {
				__antithesis_instrumentation__.Notify(431705)
				if nulls.NullAt(i) {
					__antithesis_instrumentation__.Notify(431706)
					outputCol[i] = false
				} else {
					__antithesis_instrumentation__.Notify(431707)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(431708)
			outputCol = outputCol[0:n]
			for i := range outputCol {
				__antithesis_instrumentation__.Notify(431709)
				if nulls.NullAt(i) {
					__antithesis_instrumentation__.Notify(431710)
					outputCol[i] = false
				} else {
					__antithesis_instrumentation__.Notify(431711)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(431712)
	}
	__antithesis_instrumentation__.Notify(431700)
	return batch
}
