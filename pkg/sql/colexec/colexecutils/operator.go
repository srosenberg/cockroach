package colexecutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type zeroOperator struct {
	colexecop.OneInputHelper
	colexecop.NonExplainable
}

var _ colexecop.Operator = &zeroOperator{}

func NewZeroOp(input colexecop.Operator) colexecop.Operator {
	__antithesis_instrumentation__.Notify(431737)
	return &zeroOperator{OneInputHelper: colexecop.MakeOneInputHelper(input)}
}

func (s *zeroOperator) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431738)
	return coldata.ZeroBatch
}

type fixedNumTuplesNoInputOp struct {
	colexecop.ZeroInputNode
	colexecop.NonExplainable
	batch          coldata.Batch
	numTuplesLeft  int
	opToInitialize colexecop.Operator
}

var _ colexecop.Operator = &fixedNumTuplesNoInputOp{}

func NewFixedNumTuplesNoInputOp(
	allocator *colmem.Allocator, numTuples int, opToInitialize colexecop.Operator,
) colexecop.Operator {
	__antithesis_instrumentation__.Notify(431739)
	capacity := numTuples
	if capacity > coldata.BatchSize() {
		__antithesis_instrumentation__.Notify(431741)
		capacity = coldata.BatchSize()
	} else {
		__antithesis_instrumentation__.Notify(431742)
	}
	__antithesis_instrumentation__.Notify(431740)
	return &fixedNumTuplesNoInputOp{
		batch:          allocator.NewMemBatchWithFixedCapacity(nil, capacity),
		numTuplesLeft:  numTuples,
		opToInitialize: opToInitialize,
	}
}

func (s *fixedNumTuplesNoInputOp) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431743)
	if s.opToInitialize != nil {
		__antithesis_instrumentation__.Notify(431744)
		s.opToInitialize.Init(ctx)
	} else {
		__antithesis_instrumentation__.Notify(431745)
	}
}

func (s *fixedNumTuplesNoInputOp) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431746)
	if s.numTuplesLeft == 0 {
		__antithesis_instrumentation__.Notify(431749)
		return coldata.ZeroBatch
	} else {
		__antithesis_instrumentation__.Notify(431750)
	}
	__antithesis_instrumentation__.Notify(431747)
	s.batch.ResetInternalBatch()
	length := s.numTuplesLeft
	if length > coldata.BatchSize() {
		__antithesis_instrumentation__.Notify(431751)
		length = coldata.BatchSize()
	} else {
		__antithesis_instrumentation__.Notify(431752)
	}
	__antithesis_instrumentation__.Notify(431748)
	s.numTuplesLeft -= length
	s.batch.SetLength(length)
	return s.batch
}

type vectorTypeEnforcer struct {
	colexecop.OneInputInitCloserHelper
	colexecop.NonExplainable

	allocator *colmem.Allocator
	typ       *types.T
	idx       int
}

var _ colexecop.ResettableOperator = &vectorTypeEnforcer{}

func NewVectorTypeEnforcer(
	allocator *colmem.Allocator, input colexecop.Operator, typ *types.T, idx int,
) colexecop.Operator {
	__antithesis_instrumentation__.Notify(431753)
	return &vectorTypeEnforcer{
		OneInputInitCloserHelper: colexecop.MakeOneInputInitCloserHelper(input),
		allocator:                allocator,
		typ:                      typ,
		idx:                      idx,
	}
}

func (e *vectorTypeEnforcer) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431754)
	b := e.Input.Next()
	if b.Length() == 0 {
		__antithesis_instrumentation__.Notify(431756)
		return b
	} else {
		__antithesis_instrumentation__.Notify(431757)
	}
	__antithesis_instrumentation__.Notify(431755)
	e.allocator.MaybeAppendColumn(b, e.typ, e.idx)
	return b
}

func (e *vectorTypeEnforcer) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431758)
	if r, ok := e.Input.(colexecop.Resetter); ok {
		__antithesis_instrumentation__.Notify(431759)
		r.Reset(ctx)
	} else {
		__antithesis_instrumentation__.Notify(431760)
	}
}

type BatchSchemaSubsetEnforcer struct {
	colexecop.OneInputInitCloserHelper
	colexecop.NonExplainable

	allocator                    *colmem.Allocator
	typs                         []*types.T
	subsetStartIdx, subsetEndIdx int
}

var _ colexecop.Operator = &BatchSchemaSubsetEnforcer{}

func NewBatchSchemaSubsetEnforcer(
	allocator *colmem.Allocator,
	input colexecop.Operator,
	typs []*types.T,
	subsetStartIdx, subsetEndIdx int,
) *BatchSchemaSubsetEnforcer {
	__antithesis_instrumentation__.Notify(431761)
	return &BatchSchemaSubsetEnforcer{
		OneInputInitCloserHelper: colexecop.MakeOneInputInitCloserHelper(input),
		allocator:                allocator,
		typs:                     typs,
		subsetStartIdx:           subsetStartIdx,
		subsetEndIdx:             subsetEndIdx,
	}
}

func (e *BatchSchemaSubsetEnforcer) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431762)
	if e.subsetStartIdx >= e.subsetEndIdx {
		__antithesis_instrumentation__.Notify(431764)
		colexecerror.InternalError(errors.AssertionFailedf("unexpectedly subsetStartIdx is not less than subsetEndIdx"))
	} else {
		__antithesis_instrumentation__.Notify(431765)
	}
	__antithesis_instrumentation__.Notify(431763)
	e.OneInputInitCloserHelper.Init(ctx)
}

func (e *BatchSchemaSubsetEnforcer) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431766)
	b := e.Input.Next()
	if b.Length() == 0 {
		__antithesis_instrumentation__.Notify(431769)
		return b
	} else {
		__antithesis_instrumentation__.Notify(431770)
	}
	__antithesis_instrumentation__.Notify(431767)
	for i := e.subsetStartIdx; i < e.subsetEndIdx; i++ {
		__antithesis_instrumentation__.Notify(431771)
		e.allocator.MaybeAppendColumn(b, e.typs[i], i)
	}
	__antithesis_instrumentation__.Notify(431768)
	return b
}

func (e *BatchSchemaSubsetEnforcer) SetTypes(typs []*types.T) {
	__antithesis_instrumentation__.Notify(431772)
	e.typs = typs
	e.subsetEndIdx = len(typs)
}
