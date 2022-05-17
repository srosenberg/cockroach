package colexecutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type deselectorOp struct {
	colexecop.OneInputHelper
	colexecop.NonExplainable
	unlimitedAllocator *colmem.Allocator
	inputTypes         []*types.T

	output coldata.Batch
}

var _ colexecop.Operator = &deselectorOp{}

func NewDeselectorOp(
	unlimitedAllocator *colmem.Allocator, input colexecop.Operator, typs []*types.T,
) colexecop.Operator {
	__antithesis_instrumentation__.Notify(431728)
	return &deselectorOp{
		OneInputHelper:     colexecop.MakeOneInputHelper(input),
		unlimitedAllocator: unlimitedAllocator,
		inputTypes:         typs,
	}
}

func (p *deselectorOp) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431729)
	batch := p.Input.Next()
	if batch.Selection() == nil || func() bool {
		__antithesis_instrumentation__.Notify(431732)
		return batch.Length() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(431733)
		return batch
	} else {
		__antithesis_instrumentation__.Notify(431734)
	}
	__antithesis_instrumentation__.Notify(431730)

	const maxBatchMemSize = math.MaxInt64
	p.output, _ = p.unlimitedAllocator.ResetMaybeReallocate(
		p.inputTypes, p.output, batch.Length(), maxBatchMemSize,
	)
	sel := batch.Selection()
	p.unlimitedAllocator.PerformOperation(p.output.ColVecs(), func() {
		__antithesis_instrumentation__.Notify(431735)
		for i := range p.inputTypes {
			__antithesis_instrumentation__.Notify(431736)
			toCol := p.output.ColVec(i)
			fromCol := batch.ColVec(i)
			toCol.Copy(
				coldata.SliceArgs{
					Src:       fromCol,
					Sel:       sel,
					SrcEndIdx: batch.Length(),
				},
			)
		}
	})
	__antithesis_instrumentation__.Notify(431731)
	p.output.SetLength(batch.Length())
	return p.output
}
