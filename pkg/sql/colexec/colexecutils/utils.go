package colexecutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

func MakeWindowIntoBatch(
	windowedBatch, inputBatch coldata.Batch, startIdx, endIdx int, inputTypes []*types.T,
) {
	__antithesis_instrumentation__.Notify(432056)
	windowStart := startIdx
	windowEnd := endIdx
	if sel := inputBatch.Selection(); sel != nil {
		__antithesis_instrumentation__.Notify(432059)

		windowedBatch.SetSelection(true)
		windowIntoSel := sel[startIdx:endIdx]
		copy(windowedBatch.Selection(), windowIntoSel)

		windowStart = 0

		windowEnd = windowIntoSel[len(windowIntoSel)-1] + 1
	} else {
		__antithesis_instrumentation__.Notify(432060)
		windowedBatch.SetSelection(false)
	}
	__antithesis_instrumentation__.Notify(432057)
	for i := range inputTypes {
		__antithesis_instrumentation__.Notify(432061)
		window := inputBatch.ColVec(i).Window(windowStart, windowEnd)
		windowedBatch.ReplaceCol(window, i)
	}
	__antithesis_instrumentation__.Notify(432058)
	windowedBatch.SetLength(endIdx - startIdx)
}

func NewAppendOnlyBufferedBatch(
	allocator *colmem.Allocator, typs []*types.T, colsToStore []int,
) *AppendOnlyBufferedBatch {
	__antithesis_instrumentation__.Notify(432062)
	if colsToStore == nil {
		__antithesis_instrumentation__.Notify(432064)
		colsToStore = make([]int, len(typs))
		for i := range colsToStore {
			__antithesis_instrumentation__.Notify(432065)
			colsToStore[i] = i
		}
	} else {
		__antithesis_instrumentation__.Notify(432066)
	}
	__antithesis_instrumentation__.Notify(432063)
	batch := allocator.NewMemBatchWithFixedCapacity(typs, 0)
	return &AppendOnlyBufferedBatch{
		batch:       batch,
		allocator:   allocator,
		colVecs:     batch.ColVecs(),
		colsToStore: colsToStore,
	}
}

type AppendOnlyBufferedBatch struct {
	batch coldata.Batch

	allocator   *colmem.Allocator
	length      int
	colVecs     []coldata.Vec
	colsToStore []int

	sel []int
}

var _ coldata.Batch = &AppendOnlyBufferedBatch{}

func (b *AppendOnlyBufferedBatch) Length() int {
	__antithesis_instrumentation__.Notify(432067)
	return b.length
}

func (b *AppendOnlyBufferedBatch) SetLength(n int) {
	__antithesis_instrumentation__.Notify(432068)
	b.length = n
}

func (b *AppendOnlyBufferedBatch) Capacity() int {
	__antithesis_instrumentation__.Notify(432069)

	colexecerror.InternalError(errors.AssertionFailedf("Capacity is prohibited on AppendOnlyBufferedBatch"))

	return 0
}

func (b *AppendOnlyBufferedBatch) Width() int {
	__antithesis_instrumentation__.Notify(432070)
	return b.batch.Width()
}

func (b *AppendOnlyBufferedBatch) ColVec(i int) coldata.Vec {
	__antithesis_instrumentation__.Notify(432071)
	return b.colVecs[i]
}

func (b *AppendOnlyBufferedBatch) ColVecs() []coldata.Vec {
	__antithesis_instrumentation__.Notify(432072)
	return b.colVecs
}

func (b *AppendOnlyBufferedBatch) Selection() []int {
	__antithesis_instrumentation__.Notify(432073)
	if b.batch.Selection() != nil {
		__antithesis_instrumentation__.Notify(432075)
		return b.sel
	} else {
		__antithesis_instrumentation__.Notify(432076)
	}
	__antithesis_instrumentation__.Notify(432074)
	return nil
}

func (b *AppendOnlyBufferedBatch) SetSelection(useSel bool) {
	__antithesis_instrumentation__.Notify(432077)
	b.batch.SetSelection(useSel)
	if useSel {
		__antithesis_instrumentation__.Notify(432078)

		if cap(b.sel) < b.length {
			__antithesis_instrumentation__.Notify(432079)
			b.sel = make([]int, b.length)
		} else {
			__antithesis_instrumentation__.Notify(432080)
			b.sel = b.sel[:b.length]
		}
	} else {
		__antithesis_instrumentation__.Notify(432081)
	}
}

func (b *AppendOnlyBufferedBatch) AppendCol(coldata.Vec) {
	__antithesis_instrumentation__.Notify(432082)
	colexecerror.InternalError(errors.AssertionFailedf("AppendCol is prohibited on AppendOnlyBufferedBatch"))
}

func (b *AppendOnlyBufferedBatch) ReplaceCol(coldata.Vec, int) {
	__antithesis_instrumentation__.Notify(432083)
	colexecerror.InternalError(errors.AssertionFailedf("ReplaceCol is prohibited on AppendOnlyBufferedBatch"))
}

func (b *AppendOnlyBufferedBatch) Reset([]*types.T, int, coldata.ColumnFactory) {
	__antithesis_instrumentation__.Notify(432084)
	colexecerror.InternalError(errors.AssertionFailedf("Reset is prohibited on AppendOnlyBufferedBatch"))
}

func (b *AppendOnlyBufferedBatch) ResetInternalBatch() int64 {
	__antithesis_instrumentation__.Notify(432085)
	b.SetLength(0)
	return b.batch.ResetInternalBatch()
}

func (b *AppendOnlyBufferedBatch) String() string {
	__antithesis_instrumentation__.Notify(432086)

	b.batch.SetLength(b.length)
	return b.batch.String()
}

func (b *AppendOnlyBufferedBatch) AppendTuples(batch coldata.Batch, startIdx, endIdx int) {
	__antithesis_instrumentation__.Notify(432087)
	b.allocator.PerformAppend(b, func() {
		__antithesis_instrumentation__.Notify(432088)
		for _, colIdx := range b.colsToStore {
			__antithesis_instrumentation__.Notify(432090)
			b.colVecs[colIdx].Append(
				coldata.SliceArgs{
					Src:         batch.ColVec(colIdx),
					Sel:         batch.Selection(),
					DestIdx:     b.length,
					SrcStartIdx: startIdx,
					SrcEndIdx:   endIdx,
				},
			)
		}
		__antithesis_instrumentation__.Notify(432089)
		b.length += endIdx - startIdx
	})
}

func MaybeAllocateUint64Array(array []uint64, length int) []uint64 {
	__antithesis_instrumentation__.Notify(432091)
	if cap(array) < length {
		__antithesis_instrumentation__.Notify(432094)
		return make([]uint64, length)
	} else {
		__antithesis_instrumentation__.Notify(432095)
	}
	__antithesis_instrumentation__.Notify(432092)
	array = array[:length]
	for n := 0; n < length; n += copy(array[n:], ZeroUint64Column) {
		__antithesis_instrumentation__.Notify(432096)
	}
	__antithesis_instrumentation__.Notify(432093)
	return array
}

func MaybeAllocateBoolArray(array []bool, length int) []bool {
	__antithesis_instrumentation__.Notify(432097)
	if cap(array) < length {
		__antithesis_instrumentation__.Notify(432100)
		return make([]bool, length)
	} else {
		__antithesis_instrumentation__.Notify(432101)
	}
	__antithesis_instrumentation__.Notify(432098)
	array = array[:length]
	for n := 0; n < length; n += copy(array[n:], ZeroBoolColumn) {
		__antithesis_instrumentation__.Notify(432102)
	}
	__antithesis_instrumentation__.Notify(432099)
	return array
}

func MaybeAllocateLimitedUint64Array(array []uint64, length int) []uint64 {
	__antithesis_instrumentation__.Notify(432103)
	if cap(array) < length {
		__antithesis_instrumentation__.Notify(432105)
		return make([]uint64, length)
	} else {
		__antithesis_instrumentation__.Notify(432106)
	}
	__antithesis_instrumentation__.Notify(432104)
	array = array[:length]
	copy(array, ZeroUint64Column)
	return array
}

func MaybeAllocateLimitedBoolArray(array []bool, length int) []bool {
	__antithesis_instrumentation__.Notify(432107)
	if cap(array) < length {
		__antithesis_instrumentation__.Notify(432109)
		return make([]bool, length)
	} else {
		__antithesis_instrumentation__.Notify(432110)
	}
	__antithesis_instrumentation__.Notify(432108)
	array = array[:length]
	copy(array, ZeroBoolColumn)
	return array
}

var (
	ZeroBoolColumn = make([]bool, coldata.MaxBatchSize)

	ZeroUint64Column = make([]uint64, coldata.MaxBatchSize)
)

func HandleErrorFromDiskQueue(err error) {
	__antithesis_instrumentation__.Notify(432111)
	if sqlerrors.IsDiskFullError(err) {
		__antithesis_instrumentation__.Notify(432112)

		colexecerror.ExpectedError(err)
	} else {
		__antithesis_instrumentation__.Notify(432113)
		colexecerror.InternalError(err)
	}
}

func EnsureSelectionVectorLength(old []int, length int) []int {
	__antithesis_instrumentation__.Notify(432114)
	if cap(old) >= length {
		__antithesis_instrumentation__.Notify(432116)
		return old[:length]
	} else {
		__antithesis_instrumentation__.Notify(432117)
	}
	__antithesis_instrumentation__.Notify(432115)
	return make([]int, length)
}

func UpdateBatchState(batch coldata.Batch, length int, usesSel bool, sel []int) {
	__antithesis_instrumentation__.Notify(432118)
	batch.SetSelection(usesSel)
	if usesSel {
		__antithesis_instrumentation__.Notify(432120)
		copy(batch.Selection()[:length], sel[:length])
	} else {
		__antithesis_instrumentation__.Notify(432121)
	}
	__antithesis_instrumentation__.Notify(432119)
	batch.SetLength(length)
}

var DefaultSelectionVector []int

func init() {
	DefaultSelectionVector = make([]int, coldata.MaxBatchSize)
	for i := range DefaultSelectionVector {
		DefaultSelectionVector[i] = i
	}
}
