package colmem

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type Allocator struct {
	ctx     context.Context
	acc     *mon.BoundAccount
	factory coldata.ColumnFactory
}

func selVectorSize(capacity int) int64 {
	__antithesis_instrumentation__.Notify(456712)
	return int64(capacity) * memsize.Int
}

func getVecMemoryFootprint(vec coldata.Vec) int64 {
	__antithesis_instrumentation__.Notify(456713)
	if vec == nil {
		__antithesis_instrumentation__.Notify(456716)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(456717)
	}
	__antithesis_instrumentation__.Notify(456714)
	switch vec.CanonicalTypeFamily() {
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(456718)
		return vec.Bytes().Size()
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(456719)
		return sizeOfDecimals(vec.Decimal(), 0)
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(456720)
		return vec.JSON().Size()
	case typeconv.DatumVecCanonicalTypeFamily:
		__antithesis_instrumentation__.Notify(456721)
		return vec.Datum().Size(0)
	default:
		__antithesis_instrumentation__.Notify(456722)
	}
	__antithesis_instrumentation__.Notify(456715)
	return EstimateBatchSizeBytes([]*types.T{vec.Type()}, vec.Capacity())
}

func getVecsMemoryFootprint(vecs []coldata.Vec) int64 {
	__antithesis_instrumentation__.Notify(456723)
	var size int64
	for _, dest := range vecs {
		__antithesis_instrumentation__.Notify(456725)
		size += getVecMemoryFootprint(dest)
	}
	__antithesis_instrumentation__.Notify(456724)
	return size
}

func GetBatchMemSize(b coldata.Batch) int64 {
	__antithesis_instrumentation__.Notify(456726)
	if b == nil || func() bool {
		__antithesis_instrumentation__.Notify(456728)
		return b == coldata.ZeroBatch == true
	}() == true {
		__antithesis_instrumentation__.Notify(456729)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(456730)
	}
	__antithesis_instrumentation__.Notify(456727)

	usesSel := b.Selection() != nil
	b.SetSelection(true)
	memUsage := selVectorSize(cap(b.Selection())) + getVecsMemoryFootprint(b.ColVecs())
	b.SetSelection(usesSel)
	return memUsage
}

func GetProportionalBatchMemSize(b coldata.Batch, length int64) int64 {
	__antithesis_instrumentation__.Notify(456731)
	if length == 0 {
		__antithesis_instrumentation__.Notify(456735)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(456736)
	}
	__antithesis_instrumentation__.Notify(456732)
	usesSel := b.Selection() != nil
	b.SetSelection(true)
	selCapacity := cap(b.Selection())
	b.SetSelection(usesSel)
	proportionalBatchMemSize := int64(0)
	if selCapacity > 0 {
		__antithesis_instrumentation__.Notify(456737)
		proportionalBatchMemSize = selVectorSize(selCapacity) * length / int64(selCapacity)
	} else {
		__antithesis_instrumentation__.Notify(456738)
	}
	__antithesis_instrumentation__.Notify(456733)
	for _, vec := range b.ColVecs() {
		__antithesis_instrumentation__.Notify(456739)
		switch vec.CanonicalTypeFamily() {
		case types.BytesFamily, types.JsonFamily:
			__antithesis_instrumentation__.Notify(456740)
			proportionalBatchMemSize += coldata.ProportionalSize(vec, length)
		default:
			__antithesis_instrumentation__.Notify(456741)
			proportionalBatchMemSize += getVecMemoryFootprint(vec) * length / int64(vec.Capacity())
		}
	}
	__antithesis_instrumentation__.Notify(456734)
	return proportionalBatchMemSize
}

func NewAllocator(
	ctx context.Context, acc *mon.BoundAccount, factory coldata.ColumnFactory,
) *Allocator {
	__antithesis_instrumentation__.Notify(456742)
	return &Allocator{
		ctx:     ctx,
		acc:     acc,
		factory: factory,
	}
}

func (a *Allocator) NewMemBatchWithFixedCapacity(typs []*types.T, capacity int) coldata.Batch {
	__antithesis_instrumentation__.Notify(456743)
	estimatedMemoryUsage := selVectorSize(capacity) + EstimateBatchSizeBytes(typs, capacity)
	if err := a.acc.Grow(a.ctx, estimatedMemoryUsage); err != nil {
		__antithesis_instrumentation__.Notify(456745)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(456746)
	}
	__antithesis_instrumentation__.Notify(456744)
	return coldata.NewMemBatchWithCapacity(typs, capacity, a.factory)
}

func (a *Allocator) NewMemBatchWithMaxCapacity(typs []*types.T) coldata.Batch {
	__antithesis_instrumentation__.Notify(456747)
	return a.NewMemBatchWithFixedCapacity(typs, coldata.BatchSize())
}

func (a *Allocator) NewMemBatchNoCols(typs []*types.T, capacity int) coldata.Batch {
	__antithesis_instrumentation__.Notify(456748)
	estimatedMemoryUsage := selVectorSize(capacity)
	if err := a.acc.Grow(a.ctx, estimatedMemoryUsage); err != nil {
		__antithesis_instrumentation__.Notify(456750)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(456751)
	}
	__antithesis_instrumentation__.Notify(456749)
	return coldata.NewMemBatchNoCols(typs, capacity)
}

func (a *Allocator) ResetMaybeReallocate(
	typs []*types.T, oldBatch coldata.Batch, minDesiredCapacity int, maxBatchMemSize int64,
) (newBatch coldata.Batch, reallocated bool) {
	__antithesis_instrumentation__.Notify(456752)
	if minDesiredCapacity < 0 {
		__antithesis_instrumentation__.Notify(456755)
		colexecerror.InternalError(errors.AssertionFailedf("invalid minDesiredCapacity %d", minDesiredCapacity))
	} else {
		__antithesis_instrumentation__.Notify(456756)
		if minDesiredCapacity == 0 {
			__antithesis_instrumentation__.Notify(456757)
			minDesiredCapacity = 1
		} else {
			__antithesis_instrumentation__.Notify(456758)
			if minDesiredCapacity > coldata.BatchSize() {
				__antithesis_instrumentation__.Notify(456759)
				minDesiredCapacity = coldata.BatchSize()
			} else {
				__antithesis_instrumentation__.Notify(456760)
			}
		}
	}
	__antithesis_instrumentation__.Notify(456753)
	reallocated = true
	if oldBatch == nil {
		__antithesis_instrumentation__.Notify(456761)
		newBatch = a.NewMemBatchWithFixedCapacity(typs, minDesiredCapacity)
	} else {
		__antithesis_instrumentation__.Notify(456762)

		useOldBatch := oldBatch.Capacity() == coldata.BatchSize()

		var oldBatchMemSize int64
		if !useOldBatch {
			__antithesis_instrumentation__.Notify(456764)

			oldBatchMemSize = GetBatchMemSize(oldBatch)
			useOldBatch = oldBatchMemSize >= maxBatchMemSize
		} else {
			__antithesis_instrumentation__.Notify(456765)
		}
		__antithesis_instrumentation__.Notify(456763)
		if useOldBatch {
			__antithesis_instrumentation__.Notify(456766)
			reallocated = false
			a.ReleaseMemory(oldBatch.ResetInternalBatch())
			newBatch = oldBatch
		} else {
			__antithesis_instrumentation__.Notify(456767)
			a.ReleaseMemory(oldBatchMemSize)
			newCapacity := oldBatch.Capacity() * 2
			if newCapacity < minDesiredCapacity {
				__antithesis_instrumentation__.Notify(456770)
				newCapacity = minDesiredCapacity
			} else {
				__antithesis_instrumentation__.Notify(456771)
			}
			__antithesis_instrumentation__.Notify(456768)
			if newCapacity > coldata.BatchSize() {
				__antithesis_instrumentation__.Notify(456772)
				newCapacity = coldata.BatchSize()
			} else {
				__antithesis_instrumentation__.Notify(456773)
			}
			__antithesis_instrumentation__.Notify(456769)
			newBatch = a.NewMemBatchWithFixedCapacity(typs, newCapacity)
		}
	}
	__antithesis_instrumentation__.Notify(456754)
	return newBatch, reallocated
}

func (a *Allocator) NewMemColumn(t *types.T, capacity int) coldata.Vec {
	__antithesis_instrumentation__.Notify(456774)
	estimatedMemoryUsage := EstimateBatchSizeBytes([]*types.T{t}, capacity)
	if err := a.acc.Grow(a.ctx, estimatedMemoryUsage); err != nil {
		__antithesis_instrumentation__.Notify(456776)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(456777)
	}
	__antithesis_instrumentation__.Notify(456775)
	return coldata.NewMemColumn(t, capacity, a.factory)
}

func (a *Allocator) MaybeAppendColumn(b coldata.Batch, t *types.T, colIdx int) {
	__antithesis_instrumentation__.Notify(456778)
	if b.Length() == 0 {
		__antithesis_instrumentation__.Notify(456783)
		colexecerror.InternalError(errors.AssertionFailedf("trying to add a column to zero length batch"))
	} else {
		__antithesis_instrumentation__.Notify(456784)
	}
	__antithesis_instrumentation__.Notify(456779)
	width := b.Width()
	desiredCapacity := b.Capacity()
	if desiredCapacity == 0 {
		__antithesis_instrumentation__.Notify(456785)

		desiredCapacity = b.Length()
	} else {
		__antithesis_instrumentation__.Notify(456786)
	}
	__antithesis_instrumentation__.Notify(456780)
	if colIdx < width {
		__antithesis_instrumentation__.Notify(456787)
		presentVec := b.ColVec(colIdx)
		presentType := presentVec.Type()
		if presentType.Family() == types.UnknownFamily {
			__antithesis_instrumentation__.Notify(456790)

			return
		} else {
			__antithesis_instrumentation__.Notify(456791)
		}
		__antithesis_instrumentation__.Notify(456788)
		if presentType.Identical(t) {
			__antithesis_instrumentation__.Notify(456792)

			if presentVec.Capacity() < desiredCapacity {
				__antithesis_instrumentation__.Notify(456795)

				oldMemUsage := getVecMemoryFootprint(presentVec)
				newEstimatedMemoryUsage := EstimateBatchSizeBytes([]*types.T{t}, desiredCapacity)
				if err := a.acc.Grow(a.ctx, newEstimatedMemoryUsage-oldMemUsage); err != nil {
					__antithesis_instrumentation__.Notify(456797)
					colexecerror.InternalError(err)
				} else {
					__antithesis_instrumentation__.Notify(456798)
				}
				__antithesis_instrumentation__.Notify(456796)
				b.ReplaceCol(a.NewMemColumn(t, desiredCapacity), colIdx)
				return
			} else {
				__antithesis_instrumentation__.Notify(456799)
			}
			__antithesis_instrumentation__.Notify(456793)
			if presentVec.CanonicalTypeFamily() == typeconv.DatumVecCanonicalTypeFamily {
				__antithesis_instrumentation__.Notify(456800)
				a.ReleaseMemory(presentVec.Datum().Reset())
			} else {
				__antithesis_instrumentation__.Notify(456801)
				coldata.ResetIfBytesLike(presentVec)
			}
			__antithesis_instrumentation__.Notify(456794)
			return
		} else {
			__antithesis_instrumentation__.Notify(456802)
		}
		__antithesis_instrumentation__.Notify(456789)

		colexecerror.InternalError(errors.AssertionFailedf(
			"trying to add a column of %s type at index %d but %s vector already present",
			t, colIdx, presentType,
		))
	} else {
		__antithesis_instrumentation__.Notify(456803)
		if colIdx > width {
			__antithesis_instrumentation__.Notify(456804)

			colexecerror.InternalError(errors.AssertionFailedf(
				"trying to add a column of %s type at index %d but batch has width %d",
				t, colIdx, width,
			))
		} else {
			__antithesis_instrumentation__.Notify(456805)
		}
	}
	__antithesis_instrumentation__.Notify(456781)
	estimatedMemoryUsage := EstimateBatchSizeBytes([]*types.T{t}, desiredCapacity)
	if err := a.acc.Grow(a.ctx, estimatedMemoryUsage); err != nil {
		__antithesis_instrumentation__.Notify(456806)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(456807)
	}
	__antithesis_instrumentation__.Notify(456782)
	b.AppendCol(a.NewMemColumn(t, desiredCapacity))
}

func (a *Allocator) PerformOperation(destVecs []coldata.Vec, operation func()) {
	__antithesis_instrumentation__.Notify(456808)
	before := getVecsMemoryFootprint(destVecs)

	operation()
	after := getVecsMemoryFootprint(destVecs)

	a.AdjustMemoryUsage(after - before)
}

func (a *Allocator) PerformAppend(batch coldata.Batch, operation func()) {
	__antithesis_instrumentation__.Notify(456809)
	prevLength := batch.Length()
	var before int64
	for _, dest := range batch.ColVecs() {
		__antithesis_instrumentation__.Notify(456812)
		switch dest.CanonicalTypeFamily() {
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(456813)

			before += sizeOfDecimals(dest.Decimal(), prevLength)
		case typeconv.DatumVecCanonicalTypeFamily:
			__antithesis_instrumentation__.Notify(456814)
			before += dest.Datum().Size(prevLength)
		default:
			__antithesis_instrumentation__.Notify(456815)
			before += getVecMemoryFootprint(dest)
		}
	}
	__antithesis_instrumentation__.Notify(456810)
	operation()
	var after int64
	for _, dest := range batch.ColVecs() {
		__antithesis_instrumentation__.Notify(456816)
		switch dest.CanonicalTypeFamily() {
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(456817)
			after += sizeOfDecimals(dest.Decimal(), prevLength)
		case typeconv.DatumVecCanonicalTypeFamily:
			__antithesis_instrumentation__.Notify(456818)
			after += dest.Datum().Size(prevLength)
		default:
			__antithesis_instrumentation__.Notify(456819)
			after += getVecMemoryFootprint(dest)
		}
	}
	__antithesis_instrumentation__.Notify(456811)
	a.AdjustMemoryUsage(after - before)
}

func (a *Allocator) Used() int64 {
	__antithesis_instrumentation__.Notify(456820)
	return a.acc.Used()
}

func (a *Allocator) AdjustMemoryUsage(delta int64) {
	__antithesis_instrumentation__.Notify(456821)
	if delta > 0 {
		__antithesis_instrumentation__.Notify(456822)
		if err := a.acc.Grow(a.ctx, delta); err != nil {
			__antithesis_instrumentation__.Notify(456823)
			colexecerror.InternalError(err)
		} else {
			__antithesis_instrumentation__.Notify(456824)
		}
	} else {
		__antithesis_instrumentation__.Notify(456825)
		if delta < 0 {
			__antithesis_instrumentation__.Notify(456826)
			a.ReleaseMemory(-delta)
		} else {
			__antithesis_instrumentation__.Notify(456827)
		}
	}
}

func (a *Allocator) ReleaseMemory(size int64) {
	__antithesis_instrumentation__.Notify(456828)
	if size < 0 {
		__antithesis_instrumentation__.Notify(456831)
		colexecerror.InternalError(errors.AssertionFailedf("unexpectedly negative size in ReleaseMemory: %d", size))
	} else {
		__antithesis_instrumentation__.Notify(456832)
	}
	__antithesis_instrumentation__.Notify(456829)
	if size > a.acc.Used() {
		__antithesis_instrumentation__.Notify(456833)
		size = a.acc.Used()
	} else {
		__antithesis_instrumentation__.Notify(456834)
	}
	__antithesis_instrumentation__.Notify(456830)
	a.acc.Shrink(a.ctx, size)
}

func sizeOfDecimals(decimals coldata.Decimals, startIdx int) int64 {
	__antithesis_instrumentation__.Notify(456835)
	if startIdx >= cap(decimals) {
		__antithesis_instrumentation__.Notify(456840)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(456841)
	}
	__antithesis_instrumentation__.Notify(456836)
	if startIdx >= len(decimals) {
		__antithesis_instrumentation__.Notify(456842)
		return int64(cap(decimals)-startIdx) * memsize.Decimal
	} else {
		__antithesis_instrumentation__.Notify(456843)
	}
	__antithesis_instrumentation__.Notify(456837)
	if startIdx < 0 {
		__antithesis_instrumentation__.Notify(456844)
		startIdx = 0
	} else {
		__antithesis_instrumentation__.Notify(456845)
	}
	__antithesis_instrumentation__.Notify(456838)

	size := int64(cap(decimals)-len(decimals)) * memsize.Decimal
	for i := startIdx; i < decimals.Len(); i++ {
		__antithesis_instrumentation__.Notify(456846)
		size += int64(decimals[i].Size())
	}
	__antithesis_instrumentation__.Notify(456839)
	return size
}

var SizeOfBatchSizeSelVector = int64(coldata.BatchSize()) * memsize.Int

func EstimateBatchSizeBytes(vecTypes []*types.T, batchLength int) int64 {
	__antithesis_instrumentation__.Notify(456847)
	if batchLength == 0 {
		__antithesis_instrumentation__.Notify(456850)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(456851)
	}
	__antithesis_instrumentation__.Notify(456848)

	var acc int64
	numBytesVectors := 0
	for _, t := range vecTypes {
		__antithesis_instrumentation__.Notify(456852)
		switch typeconv.TypeFamilyToCanonicalTypeFamily(t.Family()) {
		case types.BytesFamily, types.JsonFamily:
			__antithesis_instrumentation__.Notify(456853)
			numBytesVectors++
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(456854)

			acc += memsize.Decimal
		case typeconv.DatumVecCanonicalTypeFamily:
			__antithesis_instrumentation__.Notify(456855)

			acc += memsize.DatumOverhead
		case
			types.BoolFamily,
			types.IntFamily,
			types.FloatFamily,
			types.TimestampTZFamily,
			types.IntervalFamily:
			__antithesis_instrumentation__.Notify(456856)

			acc += GetFixedSizeTypeSize(t)
		default:
			__antithesis_instrumentation__.Notify(456857)
			colexecerror.InternalError(errors.AssertionFailedf("unhandled type %s", t))
		}
	}
	__antithesis_instrumentation__.Notify(456849)

	bytesVectorsSize := int64(numBytesVectors) * (coldata.FlatBytesOverhead + int64(batchLength)*coldata.ElementSize)
	return acc*int64(batchLength) + bytesVectorsSize
}

func GetFixedSizeTypeSize(t *types.T) (size int64) {
	__antithesis_instrumentation__.Notify(456858)
	switch typeconv.TypeFamilyToCanonicalTypeFamily(t.Family()) {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(456860)
		size = memsize.Bool
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(456861)
		switch t.Width() {
		case 16:
			__antithesis_instrumentation__.Notify(456866)
			size = memsize.Int16
		case 32:
			__antithesis_instrumentation__.Notify(456867)
			size = memsize.Int32
		default:
			__antithesis_instrumentation__.Notify(456868)
			size = memsize.Int64
		}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(456862)
		size = memsize.Float64
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(456863)

		size = memsize.Time
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(456864)
		size = memsize.Duration
	default:
		__antithesis_instrumentation__.Notify(456865)
		colexecerror.InternalError(errors.AssertionFailedf("unhandled type %s", t))
	}
	__antithesis_instrumentation__.Notify(456859)
	return size
}

type SetAccountingHelper struct {
	Allocator *Allocator

	allFixedLength bool

	bytesLikeVecIdxs util.FastIntSet

	bytesLikeVectors []*coldata.Bytes

	prevBytesLikeTotalSize int64

	decimalVecIdxs util.FastIntSet

	decimalVecs []coldata.Decimals

	decimalSizes []int64

	varLenDatumVecIdxs util.FastIntSet

	varLenDatumVecs []coldata.DatumVec
}

func (h *SetAccountingHelper) Init(allocator *Allocator, typs []*types.T) {
	__antithesis_instrumentation__.Notify(456869)
	h.Allocator = allocator

	for vecIdx, typ := range typs {
		__antithesis_instrumentation__.Notify(456871)
		switch typeconv.TypeFamilyToCanonicalTypeFamily(typ.Family()) {
		case types.BytesFamily, types.JsonFamily:
			__antithesis_instrumentation__.Notify(456872)
			h.bytesLikeVecIdxs.Add(vecIdx)
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(456873)
			h.decimalVecIdxs.Add(vecIdx)
		case typeconv.DatumVecCanonicalTypeFamily:
			__antithesis_instrumentation__.Notify(456874)
			h.varLenDatumVecIdxs.Add(vecIdx)
		default:
			__antithesis_instrumentation__.Notify(456875)
		}
	}
	__antithesis_instrumentation__.Notify(456870)

	h.allFixedLength = h.bytesLikeVecIdxs.Empty() && func() bool {
		__antithesis_instrumentation__.Notify(456876)
		return h.decimalVecIdxs.Empty() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(456877)
		return h.varLenDatumVecIdxs.Empty() == true
	}() == true
	h.bytesLikeVectors = make([]*coldata.Bytes, h.bytesLikeVecIdxs.Len())
	h.decimalVecs = make([]coldata.Decimals, h.decimalVecIdxs.Len())
	h.varLenDatumVecs = make([]coldata.DatumVec, h.varLenDatumVecIdxs.Len())
}

func (h *SetAccountingHelper) getBytesLikeTotalSize() int64 {
	__antithesis_instrumentation__.Notify(456878)
	var bytesLikeTotalSize int64
	for _, b := range h.bytesLikeVectors {
		__antithesis_instrumentation__.Notify(456880)
		bytesLikeTotalSize += b.Size()
	}
	__antithesis_instrumentation__.Notify(456879)
	return bytesLikeTotalSize
}

func (h *SetAccountingHelper) ResetMaybeReallocate(
	typs []*types.T, oldBatch coldata.Batch, minCapacity int, maxBatchMemSize int64,
) (newBatch coldata.Batch, reallocated bool) {
	__antithesis_instrumentation__.Notify(456881)
	newBatch, reallocated = h.Allocator.ResetMaybeReallocate(
		typs, oldBatch, minCapacity, maxBatchMemSize,
	)
	if reallocated && func() bool {
		__antithesis_instrumentation__.Notify(456883)
		return !h.allFixedLength == true
	}() == true {
		__antithesis_instrumentation__.Notify(456884)

		vecs := newBatch.ColVecs()
		if !h.bytesLikeVecIdxs.Empty() {
			__antithesis_instrumentation__.Notify(456887)
			h.bytesLikeVectors = h.bytesLikeVectors[:0]
			for vecIdx, ok := h.bytesLikeVecIdxs.Next(0); ok; vecIdx, ok = h.bytesLikeVecIdxs.Next(vecIdx + 1) {
				__antithesis_instrumentation__.Notify(456889)
				if vecs[vecIdx].CanonicalTypeFamily() == types.BytesFamily {
					__antithesis_instrumentation__.Notify(456890)
					h.bytesLikeVectors = append(h.bytesLikeVectors, vecs[vecIdx].Bytes())
				} else {
					__antithesis_instrumentation__.Notify(456891)
					h.bytesLikeVectors = append(h.bytesLikeVectors, &vecs[vecIdx].JSON().Bytes)
				}
			}
			__antithesis_instrumentation__.Notify(456888)
			h.prevBytesLikeTotalSize = h.getBytesLikeTotalSize()
		} else {
			__antithesis_instrumentation__.Notify(456892)
		}
		__antithesis_instrumentation__.Notify(456885)
		if !h.decimalVecIdxs.Empty() {
			__antithesis_instrumentation__.Notify(456893)
			h.decimalVecs = h.decimalVecs[:0]
			for vecIdx, ok := h.decimalVecIdxs.Next(0); ok; vecIdx, ok = h.decimalVecIdxs.Next(vecIdx + 1) {
				__antithesis_instrumentation__.Notify(456895)
				h.decimalVecs = append(h.decimalVecs, vecs[vecIdx].Decimal())
			}
			__antithesis_instrumentation__.Notify(456894)
			h.decimalSizes = make([]int64, newBatch.Capacity())
			for i := range h.decimalSizes {
				__antithesis_instrumentation__.Notify(456896)

				h.decimalSizes[i] = int64(len(h.decimalVecs)) * memsize.Decimal
			}
		} else {
			__antithesis_instrumentation__.Notify(456897)
		}
		__antithesis_instrumentation__.Notify(456886)
		if !h.varLenDatumVecIdxs.Empty() {
			__antithesis_instrumentation__.Notify(456898)
			h.varLenDatumVecs = h.varLenDatumVecs[:0]
			for vecIdx, ok := h.varLenDatumVecIdxs.Next(0); ok; vecIdx, ok = h.varLenDatumVecIdxs.Next(vecIdx + 1) {
				__antithesis_instrumentation__.Notify(456899)
				h.varLenDatumVecs = append(h.varLenDatumVecs, vecs[vecIdx].Datum())
			}
		} else {
			__antithesis_instrumentation__.Notify(456900)
		}
	} else {
		__antithesis_instrumentation__.Notify(456901)
	}
	__antithesis_instrumentation__.Notify(456882)
	return newBatch, reallocated
}

func (h *SetAccountingHelper) AccountForSet(rowIdx int) {
	__antithesis_instrumentation__.Notify(456902)
	if h.allFixedLength {
		__antithesis_instrumentation__.Notify(456906)

		return
	} else {
		__antithesis_instrumentation__.Notify(456907)
	}
	__antithesis_instrumentation__.Notify(456903)

	if len(h.bytesLikeVectors) > 0 {
		__antithesis_instrumentation__.Notify(456908)
		newBytesLikeTotalSize := h.getBytesLikeTotalSize()
		h.Allocator.AdjustMemoryUsage(newBytesLikeTotalSize - h.prevBytesLikeTotalSize)
		h.prevBytesLikeTotalSize = newBytesLikeTotalSize
	} else {
		__antithesis_instrumentation__.Notify(456909)
	}
	__antithesis_instrumentation__.Notify(456904)

	if !h.decimalVecIdxs.Empty() {
		__antithesis_instrumentation__.Notify(456910)
		var newDecimalSizes int64
		for _, decimalVec := range h.decimalVecs {
			__antithesis_instrumentation__.Notify(456912)
			d := decimalVec.Get(rowIdx)
			newDecimalSizes += int64(d.Size())
		}
		__antithesis_instrumentation__.Notify(456911)
		h.Allocator.AdjustMemoryUsage(newDecimalSizes - h.decimalSizes[rowIdx])
		h.decimalSizes[rowIdx] = newDecimalSizes
	} else {
		__antithesis_instrumentation__.Notify(456913)
	}
	__antithesis_instrumentation__.Notify(456905)

	if !h.varLenDatumVecIdxs.Empty() {
		__antithesis_instrumentation__.Notify(456914)
		var newVarLengthDatumSize int64
		for _, datumVec := range h.varLenDatumVecs {
			__antithesis_instrumentation__.Notify(456916)
			datumSize := datumVec.Get(rowIdx).(tree.Datum).Size()

			newVarLengthDatumSize += int64(datumSize)
		}
		__antithesis_instrumentation__.Notify(456915)
		h.Allocator.AdjustMemoryUsage(newVarLengthDatumSize)
	} else {
		__antithesis_instrumentation__.Notify(456917)
	}
}

func (h *SetAccountingHelper) Release() {
	__antithesis_instrumentation__.Notify(456918)
	*h = SetAccountingHelper{}
}
