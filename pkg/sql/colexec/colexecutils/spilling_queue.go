package colexecutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
	"github.com/marusama/semaphore"
)

type SpillingQueue struct {
	unlimitedAllocator *colmem.Allocator
	maxMemoryLimit     int64

	typs             []*types.T
	items            []coldata.Batch
	curHeadIdx       int
	curTailIdx       int
	numInMemoryItems int
	numOnDiskItems   int
	closed           bool

	nextInMemBatchCapacity int

	diskQueueCfg                colcontainer.DiskQueueCfg
	diskQueue                   colcontainer.Queue
	diskQueueDeselectionScratch coldata.Batch
	fdSemaphore                 semaphore.Semaphore
	dequeueScratch              coldata.Batch

	lastDequeuedBatchMemUsage int64

	rewindable      bool
	rewindableState struct {
		numItemsDequeued int
	}

	testingKnobs struct {
		numEnqueues int

		maxNumBatchesEnqueuedInMemory int
	}

	testingObservability struct {
		spilled     bool
		memoryUsage int64
	}

	diskAcc *mon.BoundAccount
}

var spillingQueueInitialItemsLen = int64(util.ConstantWithMetamorphicTestRange(
	"spilling-queue-initial-len",
	64, 1, 16,
))

type NewSpillingQueueArgs struct {
	UnlimitedAllocator *colmem.Allocator
	Types              []*types.T
	MemoryLimit        int64
	DiskQueueCfg       colcontainer.DiskQueueCfg
	FDSemaphore        semaphore.Semaphore
	DiskAcc            *mon.BoundAccount
}

func NewSpillingQueue(args *NewSpillingQueueArgs) *SpillingQueue {
	__antithesis_instrumentation__.Notify(431869)
	var items []coldata.Batch
	if args.MemoryLimit > 0 {
		__antithesis_instrumentation__.Notify(431871)
		items = make([]coldata.Batch, spillingQueueInitialItemsLen)
	} else {
		__antithesis_instrumentation__.Notify(431872)
	}
	__antithesis_instrumentation__.Notify(431870)
	return &SpillingQueue{
		unlimitedAllocator: args.UnlimitedAllocator,
		maxMemoryLimit:     args.MemoryLimit,
		typs:               args.Types,
		items:              items,
		diskQueueCfg:       args.DiskQueueCfg,
		fdSemaphore:        args.FDSemaphore,
		diskAcc:            args.DiskAcc,
	}
}

func NewRewindableSpillingQueue(args *NewSpillingQueueArgs) *SpillingQueue {
	__antithesis_instrumentation__.Notify(431873)
	q := NewSpillingQueue(args)
	q.rewindable = true
	return q
}

func (q *SpillingQueue) Enqueue(ctx context.Context, batch coldata.Batch) {
	__antithesis_instrumentation__.Notify(431874)
	if q.rewindable && func() bool {
		__antithesis_instrumentation__.Notify(431883)
		return q.rewindableState.numItemsDequeued > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(431884)
		colexecerror.InternalError(errors.Errorf("attempted to Enqueue to rewindable SpillingQueue after Dequeue has been called"))
	} else {
		__antithesis_instrumentation__.Notify(431885)
	}
	__antithesis_instrumentation__.Notify(431875)

	n := batch.Length()
	if n == 0 {
		__antithesis_instrumentation__.Notify(431886)
		if q.diskQueue != nil {
			__antithesis_instrumentation__.Notify(431888)
			if err := q.diskQueue.Enqueue(ctx, batch); err != nil {
				__antithesis_instrumentation__.Notify(431889)
				HandleErrorFromDiskQueue(err)
			} else {
				__antithesis_instrumentation__.Notify(431890)
			}
		} else {
			__antithesis_instrumentation__.Notify(431891)
		}
		__antithesis_instrumentation__.Notify(431887)
		return
	} else {
		__antithesis_instrumentation__.Notify(431892)
	}
	__antithesis_instrumentation__.Notify(431876)
	q.testingKnobs.numEnqueues++

	alreadySpilled := q.numOnDiskItems > 0
	memoryLimitReached := q.unlimitedAllocator.Used() > q.maxMemoryLimit || func() bool {
		__antithesis_instrumentation__.Notify(431893)
		return q.maxMemoryLimit <= 0 == true
	}() == true
	maxInMemEnqueuesExceeded := q.testingKnobs.maxNumBatchesEnqueuedInMemory != 0 && func() bool {
		__antithesis_instrumentation__.Notify(431894)
		return q.testingKnobs.numEnqueues > q.testingKnobs.maxNumBatchesEnqueuedInMemory == true
	}() == true
	if alreadySpilled || func() bool {
		__antithesis_instrumentation__.Notify(431895)
		return memoryLimitReached == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(431896)
		return maxInMemEnqueuesExceeded == true
	}() == true {
		__antithesis_instrumentation__.Notify(431897)

		if err := q.maybeSpillToDisk(ctx); err != nil {
			__antithesis_instrumentation__.Notify(431901)
			HandleErrorFromDiskQueue(err)
		} else {
			__antithesis_instrumentation__.Notify(431902)
		}
		__antithesis_instrumentation__.Notify(431898)
		if sel := batch.Selection(); sel != nil {
			__antithesis_instrumentation__.Notify(431903)

			const maxBatchMemSize = math.MaxInt64
			q.diskQueueDeselectionScratch, _ = q.unlimitedAllocator.ResetMaybeReallocate(
				q.typs, q.diskQueueDeselectionScratch, n, maxBatchMemSize,
			)
			q.unlimitedAllocator.PerformOperation(q.diskQueueDeselectionScratch.ColVecs(), func() {
				__antithesis_instrumentation__.Notify(431905)
				for i := range q.typs {
					__antithesis_instrumentation__.Notify(431907)
					q.diskQueueDeselectionScratch.ColVec(i).Copy(
						coldata.SliceArgs{
							Src:       batch.ColVec(i),
							Sel:       sel,
							SrcEndIdx: n,
						},
					)
				}
				__antithesis_instrumentation__.Notify(431906)
				q.diskQueueDeselectionScratch.SetLength(n)
			})
			__antithesis_instrumentation__.Notify(431904)
			batch = q.diskQueueDeselectionScratch
		} else {
			__antithesis_instrumentation__.Notify(431908)
		}
		__antithesis_instrumentation__.Notify(431899)
		if err := q.diskQueue.Enqueue(ctx, batch); err != nil {
			__antithesis_instrumentation__.Notify(431909)
			HandleErrorFromDiskQueue(err)
		} else {
			__antithesis_instrumentation__.Notify(431910)
		}
		__antithesis_instrumentation__.Notify(431900)
		q.numOnDiskItems++
		return
	} else {
		__antithesis_instrumentation__.Notify(431911)
	}
	__antithesis_instrumentation__.Notify(431877)

	if q.numInMemoryItems == len(q.items) {
		__antithesis_instrumentation__.Notify(431912)

		newItems := make([]coldata.Batch, q.numInMemoryItems*2)
		if q.curHeadIdx < q.curTailIdx {
			__antithesis_instrumentation__.Notify(431914)
			copy(newItems, q.items[q.curHeadIdx:q.curTailIdx])
		} else {
			__antithesis_instrumentation__.Notify(431915)
			copy(newItems, q.items[q.curHeadIdx:])
			offset := q.numInMemoryItems - q.curHeadIdx
			copy(newItems[offset:], q.items[:q.curTailIdx])
		}
		__antithesis_instrumentation__.Notify(431913)
		q.curHeadIdx = 0
		q.curTailIdx = q.numInMemoryItems
		q.items = newItems
	} else {
		__antithesis_instrumentation__.Notify(431916)
	}
	__antithesis_instrumentation__.Notify(431878)

	alreadyCopied := 0
	if q.numInMemoryItems > 0 {
		__antithesis_instrumentation__.Notify(431917)

		tailBatchIdx := q.curTailIdx - 1
		if tailBatchIdx < 0 {
			__antithesis_instrumentation__.Notify(431919)
			tailBatchIdx = len(q.items) - 1
		} else {
			__antithesis_instrumentation__.Notify(431920)
		}
		__antithesis_instrumentation__.Notify(431918)
		tailBatch := q.items[tailBatchIdx]
		if l, c := tailBatch.Length(), tailBatch.Capacity(); l < c {
			__antithesis_instrumentation__.Notify(431921)
			alreadyCopied = c - l
			if alreadyCopied > n {
				__antithesis_instrumentation__.Notify(431924)
				alreadyCopied = n
			} else {
				__antithesis_instrumentation__.Notify(431925)
			}
			__antithesis_instrumentation__.Notify(431922)
			q.unlimitedAllocator.PerformOperation(tailBatch.ColVecs(), func() {
				__antithesis_instrumentation__.Notify(431926)
				for i := range q.typs {
					__antithesis_instrumentation__.Notify(431928)
					tailBatch.ColVec(i).Copy(
						coldata.SliceArgs{
							Src:         batch.ColVec(i),
							Sel:         batch.Selection(),
							DestIdx:     l,
							SrcStartIdx: 0,
							SrcEndIdx:   alreadyCopied,
						},
					)
				}
				__antithesis_instrumentation__.Notify(431927)
				tailBatch.SetLength(l + alreadyCopied)
			})
			__antithesis_instrumentation__.Notify(431923)
			if alreadyCopied == n {
				__antithesis_instrumentation__.Notify(431929)

				return
			} else {
				__antithesis_instrumentation__.Notify(431930)
			}
		} else {
			__antithesis_instrumentation__.Notify(431931)
		}
	} else {
		__antithesis_instrumentation__.Notify(431932)
	}
	__antithesis_instrumentation__.Notify(431879)

	var newBatchCapacity int
	if q.nextInMemBatchCapacity == coldata.BatchSize() {
		__antithesis_instrumentation__.Notify(431933)

		newBatchCapacity = coldata.BatchSize()
	} else {
		__antithesis_instrumentation__.Notify(431934)
		newBatchCapacity = n - alreadyCopied
		if q.nextInMemBatchCapacity > newBatchCapacity {
			__antithesis_instrumentation__.Notify(431936)
			newBatchCapacity = q.nextInMemBatchCapacity
		} else {
			__antithesis_instrumentation__.Notify(431937)
		}
		__antithesis_instrumentation__.Notify(431935)
		q.nextInMemBatchCapacity = 2 * newBatchCapacity
		if q.nextInMemBatchCapacity > coldata.BatchSize() {
			__antithesis_instrumentation__.Notify(431938)
			q.nextInMemBatchCapacity = coldata.BatchSize()
		} else {
			__antithesis_instrumentation__.Notify(431939)
		}
	}
	__antithesis_instrumentation__.Notify(431880)

	newBatch, _ := q.unlimitedAllocator.ResetMaybeReallocate(
		q.typs,
		nil,
		newBatchCapacity,

		math.MaxInt64,
	)
	q.unlimitedAllocator.PerformOperation(newBatch.ColVecs(), func() {
		__antithesis_instrumentation__.Notify(431940)
		for i := range q.typs {
			__antithesis_instrumentation__.Notify(431942)
			newBatch.ColVec(i).Copy(
				coldata.SliceArgs{
					Src:         batch.ColVec(i),
					Sel:         batch.Selection(),
					SrcStartIdx: alreadyCopied,
					SrcEndIdx:   n,
				},
			)
		}
		__antithesis_instrumentation__.Notify(431941)
		newBatch.SetLength(n - alreadyCopied)
	})
	__antithesis_instrumentation__.Notify(431881)

	q.items[q.curTailIdx] = newBatch
	q.curTailIdx++
	if q.curTailIdx == len(q.items) {
		__antithesis_instrumentation__.Notify(431943)
		q.curTailIdx = 0
	} else {
		__antithesis_instrumentation__.Notify(431944)
	}
	__antithesis_instrumentation__.Notify(431882)
	q.numInMemoryItems++
}

func (q *SpillingQueue) Dequeue(ctx context.Context) (coldata.Batch, error) {
	__antithesis_instrumentation__.Notify(431945)
	if q.Empty() {
		__antithesis_instrumentation__.Notify(431950)
		if (!q.rewindable || func() bool {
			__antithesis_instrumentation__.Notify(431952)
			return q.numOnDiskItems != 0 == true
		}() == true) && func() bool {
			__antithesis_instrumentation__.Notify(431953)
			return q.lastDequeuedBatchMemUsage != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(431954)

			q.unlimitedAllocator.ReleaseMemory(q.lastDequeuedBatchMemUsage)
			q.lastDequeuedBatchMemUsage = 0
		} else {
			__antithesis_instrumentation__.Notify(431955)
		}
		__antithesis_instrumentation__.Notify(431951)
		return coldata.ZeroBatch, nil
	} else {
		__antithesis_instrumentation__.Notify(431956)
	}
	__antithesis_instrumentation__.Notify(431946)

	if (q.rewindable && func() bool {
		__antithesis_instrumentation__.Notify(431957)
		return q.numInMemoryItems <= q.rewindableState.numItemsDequeued == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(431958)
		return (!q.rewindable && func() bool {
			__antithesis_instrumentation__.Notify(431959)
			return q.numInMemoryItems == 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(431960)

		if !q.rewindable && func() bool {
			__antithesis_instrumentation__.Notify(431966)
			return q.curHeadIdx != q.curTailIdx == true
		}() == true {
			__antithesis_instrumentation__.Notify(431967)
			colexecerror.InternalError(errors.AssertionFailedf("assertion failed in SpillingQueue: curHeadIdx != curTailIdx, %d != %d", q.curHeadIdx, q.curTailIdx))
		} else {
			__antithesis_instrumentation__.Notify(431968)
		}
		__antithesis_instrumentation__.Notify(431961)

		if q.dequeueScratch == nil {
			__antithesis_instrumentation__.Notify(431969)

			q.dequeueScratch = q.unlimitedAllocator.NewMemBatchWithFixedCapacity(q.typs, coldata.BatchSize())
			q.unlimitedAllocator.ReleaseMemory(colmem.GetBatchMemSize(q.dequeueScratch))
		} else {
			__antithesis_instrumentation__.Notify(431970)
		}
		__antithesis_instrumentation__.Notify(431962)
		ok, err := q.diskQueue.Dequeue(ctx, q.dequeueScratch)
		if err != nil {
			__antithesis_instrumentation__.Notify(431971)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(431972)
		}
		__antithesis_instrumentation__.Notify(431963)
		if !ok {
			__antithesis_instrumentation__.Notify(431973)

			colexecerror.InternalError(errors.AssertionFailedf("disk queue was not empty but failed to Dequeue element in SpillingQueue"))
		} else {
			__antithesis_instrumentation__.Notify(431974)
		}
		__antithesis_instrumentation__.Notify(431964)

		q.unlimitedAllocator.ReleaseMemory(q.lastDequeuedBatchMemUsage)
		q.lastDequeuedBatchMemUsage = colmem.GetBatchMemSize(q.dequeueScratch)
		q.unlimitedAllocator.AdjustMemoryUsage(q.lastDequeuedBatchMemUsage)
		if q.rewindable {
			__antithesis_instrumentation__.Notify(431975)
			q.rewindableState.numItemsDequeued++
		} else {
			__antithesis_instrumentation__.Notify(431976)
			q.numOnDiskItems--
		}
		__antithesis_instrumentation__.Notify(431965)
		return q.dequeueScratch, nil
	} else {
		__antithesis_instrumentation__.Notify(431977)
	}
	__antithesis_instrumentation__.Notify(431947)

	res := q.items[q.curHeadIdx]
	if q.rewindable {
		__antithesis_instrumentation__.Notify(431978)

		q.rewindableState.numItemsDequeued++
	} else {
		__antithesis_instrumentation__.Notify(431979)

		q.items[q.curHeadIdx] = nil

		q.unlimitedAllocator.ReleaseMemory(q.lastDequeuedBatchMemUsage)
		q.lastDequeuedBatchMemUsage = colmem.GetBatchMemSize(res)
		q.numInMemoryItems--
	}
	__antithesis_instrumentation__.Notify(431948)
	q.curHeadIdx++
	if q.curHeadIdx == len(q.items) {
		__antithesis_instrumentation__.Notify(431980)
		q.curHeadIdx = 0
	} else {
		__antithesis_instrumentation__.Notify(431981)
	}
	__antithesis_instrumentation__.Notify(431949)
	return res, nil
}

func (q *SpillingQueue) numFDsOpenAtAnyGivenTime() int {
	__antithesis_instrumentation__.Notify(431982)
	if q.diskQueueCfg.CacheMode != colcontainer.DiskQueueCacheModeIntertwinedCalls {
		__antithesis_instrumentation__.Notify(431984)

		return 1
	} else {
		__antithesis_instrumentation__.Notify(431985)
	}
	__antithesis_instrumentation__.Notify(431983)

	return 2
}

func (q *SpillingQueue) maybeSpillToDisk(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(431986)
	if q.diskQueue != nil {
		__antithesis_instrumentation__.Notify(431993)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(431994)
	}
	__antithesis_instrumentation__.Notify(431987)
	var err error

	if q.fdSemaphore != nil {
		__antithesis_instrumentation__.Notify(431995)
		if err = q.fdSemaphore.Acquire(ctx, q.numFDsOpenAtAnyGivenTime()); err != nil {
			__antithesis_instrumentation__.Notify(431996)
			return err
		} else {
			__antithesis_instrumentation__.Notify(431997)
		}
	} else {
		__antithesis_instrumentation__.Notify(431998)
	}
	__antithesis_instrumentation__.Notify(431988)
	log.VEvent(ctx, 1, "spilled to disk")
	var diskQueue colcontainer.Queue
	if q.rewindable {
		__antithesis_instrumentation__.Notify(431999)
		diskQueue, err = colcontainer.NewRewindableDiskQueue(ctx, q.typs, q.diskQueueCfg, q.diskAcc)
	} else {
		__antithesis_instrumentation__.Notify(432000)
		diskQueue, err = colcontainer.NewDiskQueue(ctx, q.typs, q.diskQueueCfg, q.diskAcc)
	}
	__antithesis_instrumentation__.Notify(431989)
	if err != nil {
		__antithesis_instrumentation__.Notify(432001)
		return err
	} else {
		__antithesis_instrumentation__.Notify(432002)
	}
	__antithesis_instrumentation__.Notify(431990)

	q.diskQueue = diskQueue

	q.maxMemoryLimit -= int64(q.diskQueueCfg.BufferSizeBytes)

	var queueTailToMove []coldata.Batch
	for q.numInMemoryItems > 0 && func() bool {
		__antithesis_instrumentation__.Notify(432003)
		return q.unlimitedAllocator.Used() > q.maxMemoryLimit == true
	}() == true {
		__antithesis_instrumentation__.Notify(432004)
		tailBatchIdx := q.curTailIdx - 1
		if tailBatchIdx < 0 {
			__antithesis_instrumentation__.Notify(432006)
			tailBatchIdx = len(q.items) - 1
		} else {
			__antithesis_instrumentation__.Notify(432007)
		}
		__antithesis_instrumentation__.Notify(432005)
		tailBatch := q.items[tailBatchIdx]
		queueTailToMove = append(queueTailToMove, tailBatch)
		q.items[tailBatchIdx] = nil

		q.unlimitedAllocator.ReleaseMemory(colmem.GetBatchMemSize(tailBatch))
		q.numInMemoryItems--
		q.curTailIdx--
		if q.curTailIdx < 0 {
			__antithesis_instrumentation__.Notify(432008)
			q.curTailIdx = len(q.items) - 1
		} else {
			__antithesis_instrumentation__.Notify(432009)
		}
	}
	__antithesis_instrumentation__.Notify(431991)
	for i := len(queueTailToMove) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(432010)

		if err := q.diskQueue.Enqueue(ctx, queueTailToMove[i]); err != nil {
			__antithesis_instrumentation__.Notify(432012)
			return err
		} else {
			__antithesis_instrumentation__.Notify(432013)
		}
		__antithesis_instrumentation__.Notify(432011)
		q.numOnDiskItems++
	}
	__antithesis_instrumentation__.Notify(431992)
	return nil
}

func (q *SpillingQueue) Empty() bool {
	__antithesis_instrumentation__.Notify(432014)
	if q.rewindable {
		__antithesis_instrumentation__.Notify(432016)
		return q.numInMemoryItems+q.numOnDiskItems == q.rewindableState.numItemsDequeued
	} else {
		__antithesis_instrumentation__.Notify(432017)
	}
	__antithesis_instrumentation__.Notify(432015)
	return q.numInMemoryItems == 0 && func() bool {
		__antithesis_instrumentation__.Notify(432018)
		return q.numOnDiskItems == 0 == true
	}() == true
}

func (q *SpillingQueue) Spilled() bool {
	__antithesis_instrumentation__.Notify(432019)
	if q.closed {
		__antithesis_instrumentation__.Notify(432021)
		return q.testingObservability.spilled
	} else {
		__antithesis_instrumentation__.Notify(432022)
	}
	__antithesis_instrumentation__.Notify(432020)
	return q.diskQueue != nil
}

func (q *SpillingQueue) MemoryUsage() int64 {
	__antithesis_instrumentation__.Notify(432023)
	if q.closed {
		__antithesis_instrumentation__.Notify(432025)
		return q.testingObservability.memoryUsage
	} else {
		__antithesis_instrumentation__.Notify(432026)
	}
	__antithesis_instrumentation__.Notify(432024)
	return q.unlimitedAllocator.Used()
}

func (q *SpillingQueue) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(432027)
	if q == nil || func() bool {
		__antithesis_instrumentation__.Notify(432031)
		return q.closed == true
	}() == true {
		__antithesis_instrumentation__.Notify(432032)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(432033)
	}
	__antithesis_instrumentation__.Notify(432028)
	q.closed = true
	q.testingObservability.spilled = q.diskQueue != nil
	q.testingObservability.memoryUsage = q.unlimitedAllocator.Used()
	q.unlimitedAllocator.ReleaseMemory(q.unlimitedAllocator.Used())

	for i := range q.items {
		__antithesis_instrumentation__.Notify(432034)
		q.items[i] = nil
	}
	__antithesis_instrumentation__.Notify(432029)
	q.diskQueueDeselectionScratch = nil
	q.dequeueScratch = nil
	if q.diskQueue != nil {
		__antithesis_instrumentation__.Notify(432035)
		if err := q.diskQueue.Close(ctx); err != nil {
			__antithesis_instrumentation__.Notify(432038)
			return err
		} else {
			__antithesis_instrumentation__.Notify(432039)
		}
		__antithesis_instrumentation__.Notify(432036)
		q.diskQueue = nil
		if q.fdSemaphore != nil {
			__antithesis_instrumentation__.Notify(432040)
			q.fdSemaphore.Release(q.numFDsOpenAtAnyGivenTime())
		} else {
			__antithesis_instrumentation__.Notify(432041)
		}
		__antithesis_instrumentation__.Notify(432037)
		q.maxMemoryLimit += int64(q.diskQueueCfg.BufferSizeBytes)
	} else {
		__antithesis_instrumentation__.Notify(432042)
	}
	__antithesis_instrumentation__.Notify(432030)
	return nil
}

func (q *SpillingQueue) Rewind() error {
	__antithesis_instrumentation__.Notify(432043)
	if !q.rewindable {
		__antithesis_instrumentation__.Notify(432046)
		return errors.Newf("unexpectedly Rewind() called when spilling queue is not rewindable")
	} else {
		__antithesis_instrumentation__.Notify(432047)
	}
	__antithesis_instrumentation__.Notify(432044)
	if q.diskQueue != nil {
		__antithesis_instrumentation__.Notify(432048)
		if err := q.diskQueue.(colcontainer.RewindableQueue).Rewind(); err != nil {
			__antithesis_instrumentation__.Notify(432049)
			return err
		} else {
			__antithesis_instrumentation__.Notify(432050)
		}
	} else {
		__antithesis_instrumentation__.Notify(432051)
	}
	__antithesis_instrumentation__.Notify(432045)
	q.curHeadIdx = 0
	q.lastDequeuedBatchMemUsage = 0
	q.rewindableState.numItemsDequeued = 0
	return nil
}

func (q *SpillingQueue) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(432052)
	if err := q.Close(ctx); err != nil {
		__antithesis_instrumentation__.Notify(432054)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(432055)
	}
	__antithesis_instrumentation__.Notify(432053)
	q.closed = false
	q.numInMemoryItems = 0
	q.numOnDiskItems = 0
	q.curHeadIdx = 0
	q.curTailIdx = 0
	q.nextInMemBatchCapacity = 0
	q.lastDequeuedBatchMemUsage = 0
	q.rewindableState.numItemsDequeued = 0
	q.testingKnobs.numEnqueues = 0
}
