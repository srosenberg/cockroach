package colexecutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
	"github.com/marusama/semaphore"
)

type SpillingBuffer struct {
	memoryLimit        int64
	diskReservedMem    int64
	unlimitedAllocator *colmem.Allocator
	inputTypes         []*types.T
	storedTypes        []*types.T

	colIdxs []int

	bufferedTuples *AppendOnlyBufferedBatch

	scratch coldata.Batch

	diskQueueCfg colcontainer.DiskQueueCfg
	diskQueue    colcontainer.RewindableQueue
	fdSemaphore  semaphore.Semaphore
	diskAcc      *mon.BoundAccount

	dequeueScratch            coldata.Batch
	lastDequeuedBatchMemUsage int64
	numDequeued               int

	doneAppending bool
	length        int
	closed        bool

	testingKnobs struct {
		maxTuplesStoredInMemory int
	}
}

func NewSpillingBuffer(
	unlimitedAllocator *colmem.Allocator,
	memoryLimit int64,
	diskQueueCfg colcontainer.DiskQueueCfg,
	fdSemaphore semaphore.Semaphore,
	inputTypes []*types.T,
	diskAcc *mon.BoundAccount,
	colIdxs ...int,
) *SpillingBuffer {
	__antithesis_instrumentation__.Notify(431773)
	if colIdxs == nil {
		__antithesis_instrumentation__.Notify(431776)
		colIdxs = make([]int, len(inputTypes))
		for i := range colIdxs {
			__antithesis_instrumentation__.Notify(431777)
			colIdxs[i] = i
		}
	} else {
		__antithesis_instrumentation__.Notify(431778)
	}
	__antithesis_instrumentation__.Notify(431774)
	storedTypes := make([]*types.T, len(colIdxs))
	for i, idx := range colIdxs {
		__antithesis_instrumentation__.Notify(431779)
		storedTypes[i] = inputTypes[idx]
	}
	__antithesis_instrumentation__.Notify(431775)

	diskQueueCfg.SetCacheMode(colcontainer.DiskQueueCacheModeClearAndReuseCache)

	diskReservedMem := colmem.EstimateBatchSizeBytes(storedTypes, coldata.BatchSize()) + int64(diskQueueCfg.BufferSizeBytes)
	return &SpillingBuffer{
		unlimitedAllocator: unlimitedAllocator,
		memoryLimit:        memoryLimit,
		diskReservedMem:    diskReservedMem,
		storedTypes:        storedTypes,
		inputTypes:         inputTypes,
		colIdxs:            colIdxs,
		bufferedTuples:     NewAppendOnlyBufferedBatch(unlimitedAllocator, inputTypes, colIdxs),
		scratch:            unlimitedAllocator.NewMemBatchNoCols(storedTypes, 0),
		diskQueueCfg:       diskQueueCfg,
		fdSemaphore:        fdSemaphore,
		diskAcc:            diskAcc,
	}
}

const numSpillingBufferFDs = 1

func (b *SpillingBuffer) AppendTuples(
	ctx context.Context, batch coldata.Batch, startIdx, endIdx int,
) {
	__antithesis_instrumentation__.Notify(431780)
	var err error
	if b.doneAppending {
		__antithesis_instrumentation__.Notify(431788)
		colexecerror.InternalError(
			errors.AssertionFailedf("attempted to append to SpillingBuffer after calling GetVecWithTuple"))
	} else {
		__antithesis_instrumentation__.Notify(431789)
	}
	__antithesis_instrumentation__.Notify(431781)
	if startIdx >= endIdx || func() bool {
		__antithesis_instrumentation__.Notify(431790)
		return startIdx < 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(431791)
		return endIdx > batch.Length() == true
	}() == true {
		__antithesis_instrumentation__.Notify(431792)
		colexecerror.InternalError(
			errors.AssertionFailedf("invalid indexes into source batch: %d, %d", startIdx, endIdx))
	} else {
		__antithesis_instrumentation__.Notify(431793)
	}
	__antithesis_instrumentation__.Notify(431782)
	if batch.Selection() != nil {
		__antithesis_instrumentation__.Notify(431794)
		colexecerror.InternalError(
			errors.AssertionFailedf("attempted to append batch with selection to SpillingBuffer"))
	} else {
		__antithesis_instrumentation__.Notify(431795)
	}
	__antithesis_instrumentation__.Notify(431783)
	b.length += endIdx - startIdx
	memLimitReached := b.unlimitedAllocator.Used()+b.diskReservedMem > b.memoryLimit
	maxInMemTuplesLimitReached := b.testingKnobs.maxTuplesStoredInMemory > 0 && func() bool {
		__antithesis_instrumentation__.Notify(431796)
		return b.bufferedTuples.Length() >= b.testingKnobs.maxTuplesStoredInMemory == true
	}() == true
	if !memLimitReached && func() bool {
		__antithesis_instrumentation__.Notify(431797)
		return b.diskQueue == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(431798)
		return !maxInMemTuplesLimitReached == true
	}() == true {
		__antithesis_instrumentation__.Notify(431799)
		b.bufferedTuples.AppendTuples(batch, startIdx, endIdx)
		return
	} else {
		__antithesis_instrumentation__.Notify(431800)
	}
	__antithesis_instrumentation__.Notify(431784)

	if b.diskQueue == nil {
		__antithesis_instrumentation__.Notify(431801)
		if b.fdSemaphore != nil {
			__antithesis_instrumentation__.Notify(431804)
			if err = b.fdSemaphore.Acquire(ctx, numSpillingBufferFDs); err != nil {
				__antithesis_instrumentation__.Notify(431805)
				colexecerror.InternalError(err)
			} else {
				__antithesis_instrumentation__.Notify(431806)
			}
		} else {
			__antithesis_instrumentation__.Notify(431807)
		}
		__antithesis_instrumentation__.Notify(431802)
		if b.diskQueue, err = colcontainer.NewRewindableDiskQueue(
			ctx, b.storedTypes, b.diskQueueCfg, b.diskAcc); err != nil {
			__antithesis_instrumentation__.Notify(431808)
			colexecerror.InternalError(err)
		} else {
			__antithesis_instrumentation__.Notify(431809)
		}
		__antithesis_instrumentation__.Notify(431803)
		log.VEvent(ctx, 1, "spilled to disk")
	} else {
		__antithesis_instrumentation__.Notify(431810)
	}
	__antithesis_instrumentation__.Notify(431785)

	for i, idx := range b.colIdxs {
		__antithesis_instrumentation__.Notify(431811)
		window := batch.ColVec(idx).Window(startIdx, endIdx)
		b.scratch.ReplaceCol(window, i)
	}
	__antithesis_instrumentation__.Notify(431786)
	b.scratch.SetLength(endIdx - startIdx)
	if err = b.diskQueue.Enqueue(ctx, b.scratch); err != nil {
		__antithesis_instrumentation__.Notify(431812)
		HandleErrorFromDiskQueue(err)
	} else {
		__antithesis_instrumentation__.Notify(431813)
	}
	__antithesis_instrumentation__.Notify(431787)

	for i := range b.scratch.ColVecs() {
		__antithesis_instrumentation__.Notify(431814)
		b.scratch.ColVecs()[i] = nil
	}
}

func (b *SpillingBuffer) GetVecWithTuple(
	ctx context.Context, colIdx, idx int,
) (_ coldata.Vec, rowIdx int, length int) {
	__antithesis_instrumentation__.Notify(431815)
	var err error
	if idx < 0 || func() bool {
		__antithesis_instrumentation__.Notify(431821)
		return idx >= b.Length() == true
	}() == true {
		__antithesis_instrumentation__.Notify(431822)
		colexecerror.InternalError(
			errors.AssertionFailedf("index out of range for spilling buffer: %d", idx))
	} else {
		__antithesis_instrumentation__.Notify(431823)
	}
	__antithesis_instrumentation__.Notify(431816)
	if idx < b.bufferedTuples.Length() {
		__antithesis_instrumentation__.Notify(431824)

		return b.bufferedTuples.ColVec(b.colIdxs[colIdx]), idx, b.bufferedTuples.Length()
	} else {
		__antithesis_instrumentation__.Notify(431825)
	}
	__antithesis_instrumentation__.Notify(431817)

	idx -= b.bufferedTuples.Length()
	if idx < b.numDequeued {
		__antithesis_instrumentation__.Notify(431826)

		if err = b.diskQueue.Rewind(); err != nil {
			__antithesis_instrumentation__.Notify(431828)
			colexecerror.InternalError(err)
		} else {
			__antithesis_instrumentation__.Notify(431829)
		}
		__antithesis_instrumentation__.Notify(431827)
		b.numDequeued = 0
		if b.dequeueScratch != nil {
			__antithesis_instrumentation__.Notify(431830)
			b.dequeueScratch.SetLength(0)
		} else {
			__antithesis_instrumentation__.Notify(431831)
		}
	} else {
		__antithesis_instrumentation__.Notify(431832)
	}
	__antithesis_instrumentation__.Notify(431818)
	if b.dequeueScratch == nil {
		__antithesis_instrumentation__.Notify(431833)

		b.dequeueScratch = b.unlimitedAllocator.NewMemBatchWithFixedCapacity(
			b.storedTypes, coldata.BatchSize())
		b.unlimitedAllocator.ReleaseMemory(colmem.GetBatchMemSize(b.dequeueScratch))
	} else {
		__antithesis_instrumentation__.Notify(431834)
	}
	__antithesis_instrumentation__.Notify(431819)
	if !b.doneAppending {
		__antithesis_instrumentation__.Notify(431835)
		b.doneAppending = true

		if err = b.diskQueue.Enqueue(ctx, coldata.ZeroBatch); err != nil {
			__antithesis_instrumentation__.Notify(431836)
			HandleErrorFromDiskQueue(err)
		} else {
			__antithesis_instrumentation__.Notify(431837)
		}
	} else {
		__antithesis_instrumentation__.Notify(431838)
	}
	__antithesis_instrumentation__.Notify(431820)

	for {
		__antithesis_instrumentation__.Notify(431839)
		if idx-b.numDequeued < b.dequeueScratch.Length() {
			__antithesis_instrumentation__.Notify(431842)

			rowIdx = idx - b.numDequeued

			b.unlimitedAllocator.ReleaseMemory(b.lastDequeuedBatchMemUsage)
			b.lastDequeuedBatchMemUsage = colmem.GetBatchMemSize(b.dequeueScratch)
			b.unlimitedAllocator.AdjustMemoryUsage(b.lastDequeuedBatchMemUsage)
			return b.dequeueScratch.ColVec(colIdx), rowIdx, b.dequeueScratch.Length()
		} else {
			__antithesis_instrumentation__.Notify(431843)
		}
		__antithesis_instrumentation__.Notify(431840)

		var ok bool
		b.numDequeued += b.dequeueScratch.Length()
		if ok, err = b.diskQueue.Dequeue(ctx, b.dequeueScratch); err != nil {
			__antithesis_instrumentation__.Notify(431844)
			colexecerror.InternalError(err)
		} else {
			__antithesis_instrumentation__.Notify(431845)
		}
		__antithesis_instrumentation__.Notify(431841)
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(431846)
			return b.dequeueScratch.Length() == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(431847)
			colexecerror.InternalError(
				errors.AssertionFailedf("index out of range for SpillingBuffer"))
		} else {
			__antithesis_instrumentation__.Notify(431848)
		}
	}
}

func (b *SpillingBuffer) Length() int {
	__antithesis_instrumentation__.Notify(431849)
	return b.length
}

func (b *SpillingBuffer) closeSpillingQueue(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431850)
	if b.diskQueue != nil {
		__antithesis_instrumentation__.Notify(431851)
		if err := b.diskQueue.Close(ctx); err != nil {
			__antithesis_instrumentation__.Notify(431854)
			colexecerror.InternalError(err)
		} else {
			__antithesis_instrumentation__.Notify(431855)
		}
		__antithesis_instrumentation__.Notify(431852)
		if b.fdSemaphore != nil {
			__antithesis_instrumentation__.Notify(431856)
			b.fdSemaphore.Release(numSpillingBufferFDs)
		} else {
			__antithesis_instrumentation__.Notify(431857)
		}
		__antithesis_instrumentation__.Notify(431853)
		b.diskQueue = nil
	} else {
		__antithesis_instrumentation__.Notify(431858)
	}
}

func (b *SpillingBuffer) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431859)
	if b.closed {
		__antithesis_instrumentation__.Notify(431861)
		return
	} else {
		__antithesis_instrumentation__.Notify(431862)
	}
	__antithesis_instrumentation__.Notify(431860)
	b.unlimitedAllocator.ReleaseMemory(b.unlimitedAllocator.Used())
	b.closeSpillingQueue(ctx)

	*b = SpillingBuffer{closed: true}
}

func (b *SpillingBuffer) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431863)
	b.doneAppending = false
	b.numDequeued = 0
	b.length = 0
	if b.diskQueue != nil {
		__antithesis_instrumentation__.Notify(431865)

		b.closeSpillingQueue(ctx)
		b.unlimitedAllocator.ReleaseMemory(b.unlimitedAllocator.Used() - b.lastDequeuedBatchMemUsage)
		b.bufferedTuples = NewAppendOnlyBufferedBatch(b.unlimitedAllocator, b.inputTypes, b.colIdxs)
	} else {
		__antithesis_instrumentation__.Notify(431866)

		b.bufferedTuples.ResetInternalBatch()
	}
	__antithesis_instrumentation__.Notify(431864)
	if b.dequeueScratch != nil {
		__antithesis_instrumentation__.Notify(431867)
		b.dequeueScratch.SetLength(0)
	} else {
		__antithesis_instrumentation__.Notify(431868)
	}
}
