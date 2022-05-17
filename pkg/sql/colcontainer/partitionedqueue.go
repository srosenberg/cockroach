package colcontainer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
	"github.com/marusama/semaphore"
)

type PartitionedQueue interface {
	Enqueue(ctx context.Context, partitionIdx int, batch coldata.Batch) error

	Dequeue(ctx context.Context, partitionIdx int, batch coldata.Batch) error

	CloseAllOpenWriteFileDescriptors(ctx context.Context) error

	CloseAllOpenReadFileDescriptors() error

	CloseInactiveReadPartitions(ctx context.Context) error

	Close(ctx context.Context) error
}

type partitionState int

const (
	partitionStateWriting partitionState = iota

	partitionStateClosedForWriting

	partitionStateReading

	partitionStateClosedForReading

	partitionStatePermanentlyClosed
)

type partition struct {
	Queue
	state partitionState
}

type PartitionerStrategy int

const (
	PartitionerStrategyDefault PartitionerStrategy = iota

	PartitionerStrategyCloseOnNewPartition
)

type PartitionedDiskQueue struct {
	typs     []*types.T
	strategy PartitionerStrategy
	cfg      DiskQueueCfg

	partitionIdxToIndex map[int]int
	partitions          []partition

	lastEnqueuedPartitionIdx int

	numOpenFDs  int
	fdSemaphore semaphore.Semaphore
	diskAcc     *mon.BoundAccount
}

var _ PartitionedQueue = &PartitionedDiskQueue{}

func NewPartitionedDiskQueue(
	typs []*types.T,
	cfg DiskQueueCfg,
	fdSemaphore semaphore.Semaphore,
	partitionerStrategy PartitionerStrategy,
	diskAcc *mon.BoundAccount,
) *PartitionedDiskQueue {
	__antithesis_instrumentation__.Notify(272884)
	return &PartitionedDiskQueue{
		typs:                     typs,
		strategy:                 partitionerStrategy,
		cfg:                      cfg,
		partitionIdxToIndex:      make(map[int]int),
		partitions:               make([]partition, 0),
		lastEnqueuedPartitionIdx: -1,
		fdSemaphore:              fdSemaphore,
		diskAcc:                  diskAcc,
	}
}

type closeWritePartitionArgument int

const (
	retainFD closeWritePartitionArgument = iota
	releaseFD
)

func (p *PartitionedDiskQueue) closeWritePartition(
	ctx context.Context, idx int, releaseFDOption closeWritePartitionArgument,
) error {
	__antithesis_instrumentation__.Notify(272885)
	if p.partitions[idx].state != partitionStateWriting {
		__antithesis_instrumentation__.Notify(272889)
		colexecerror.InternalError(errors.AssertionFailedf("illegal state change from %d to partitionStateClosedForWriting, only partitionStateWriting allowed", p.partitions[idx].state))
	} else {
		__antithesis_instrumentation__.Notify(272890)
	}
	__antithesis_instrumentation__.Notify(272886)
	if err := p.partitions[idx].Enqueue(ctx, coldata.ZeroBatch); err != nil {
		__antithesis_instrumentation__.Notify(272891)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272892)
	}
	__antithesis_instrumentation__.Notify(272887)
	if releaseFDOption == releaseFD && func() bool {
		__antithesis_instrumentation__.Notify(272893)
		return p.fdSemaphore != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(272894)
		p.fdSemaphore.Release(1)
		p.numOpenFDs--
	} else {
		__antithesis_instrumentation__.Notify(272895)
	}
	__antithesis_instrumentation__.Notify(272888)
	p.partitions[idx].state = partitionStateClosedForWriting
	return nil
}

func (p *PartitionedDiskQueue) closeReadPartition(idx int) error {
	__antithesis_instrumentation__.Notify(272896)
	if p.partitions[idx].state != partitionStateReading {
		__antithesis_instrumentation__.Notify(272900)
		colexecerror.InternalError(errors.AssertionFailedf("illegal state change from %d to partitionStateClosedForReading, only partitionStateReading allowed", p.partitions[idx].state))
	} else {
		__antithesis_instrumentation__.Notify(272901)
	}
	__antithesis_instrumentation__.Notify(272897)
	if err := p.partitions[idx].CloseRead(); err != nil {
		__antithesis_instrumentation__.Notify(272902)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272903)
	}
	__antithesis_instrumentation__.Notify(272898)
	if p.fdSemaphore != nil {
		__antithesis_instrumentation__.Notify(272904)
		p.fdSemaphore.Release(1)
		p.numOpenFDs--
	} else {
		__antithesis_instrumentation__.Notify(272905)
	}
	__antithesis_instrumentation__.Notify(272899)
	p.partitions[idx].state = partitionStateClosedForReading
	return nil
}

func (p *PartitionedDiskQueue) acquireNewFD(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(272906)
	if p.fdSemaphore == nil {
		__antithesis_instrumentation__.Notify(272909)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(272910)
	}
	__antithesis_instrumentation__.Notify(272907)
	if err := p.fdSemaphore.Acquire(ctx, 1); err != nil {
		__antithesis_instrumentation__.Notify(272911)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272912)
	}
	__antithesis_instrumentation__.Notify(272908)
	p.numOpenFDs++
	return nil
}

func (p *PartitionedDiskQueue) Enqueue(
	ctx context.Context, partitionIdx int, batch coldata.Batch,
) error {
	__antithesis_instrumentation__.Notify(272913)
	idx, ok := p.partitionIdxToIndex[partitionIdx]
	if !ok {
		__antithesis_instrumentation__.Notify(272916)
		needToAcquireFD := true
		if p.strategy == PartitionerStrategyCloseOnNewPartition && func() bool {
			__antithesis_instrumentation__.Notify(272920)
			return p.lastEnqueuedPartitionIdx != -1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(272921)
			idxToClose, found := p.partitionIdxToIndex[p.lastEnqueuedPartitionIdx]
			if !found {
				__antithesis_instrumentation__.Notify(272923)

				return errors.New("PartitionerStrategyCloseOnNewPartition unable to find last Enqueued partition")
			} else {
				__antithesis_instrumentation__.Notify(272924)
			}
			__antithesis_instrumentation__.Notify(272922)
			if p.partitions[idxToClose].state == partitionStateWriting {
				__antithesis_instrumentation__.Notify(272925)

				if err := p.closeWritePartition(ctx, idxToClose, retainFD); err != nil {
					__antithesis_instrumentation__.Notify(272927)
					return err
				} else {
					__antithesis_instrumentation__.Notify(272928)
				}
				__antithesis_instrumentation__.Notify(272926)
				needToAcquireFD = false
			} else {
				__antithesis_instrumentation__.Notify(272929)

				needToAcquireFD = true
			}
		} else {
			__antithesis_instrumentation__.Notify(272930)
		}
		__antithesis_instrumentation__.Notify(272917)

		if needToAcquireFD {
			__antithesis_instrumentation__.Notify(272931)

			if err := p.acquireNewFD(ctx); err != nil {
				__antithesis_instrumentation__.Notify(272932)
				return err
			} else {
				__antithesis_instrumentation__.Notify(272933)
			}
		} else {
			__antithesis_instrumentation__.Notify(272934)
		}
		__antithesis_instrumentation__.Notify(272918)

		q, err := NewDiskQueue(ctx, p.typs, p.cfg, p.diskAcc)
		if err != nil {
			__antithesis_instrumentation__.Notify(272935)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272936)
		}
		__antithesis_instrumentation__.Notify(272919)
		idx = len(p.partitions)
		p.partitions = append(p.partitions, partition{Queue: q})
		p.partitionIdxToIndex[partitionIdx] = idx
	} else {
		__antithesis_instrumentation__.Notify(272937)
	}
	__antithesis_instrumentation__.Notify(272914)
	if state := p.partitions[idx].state; state != partitionStateWriting {
		__antithesis_instrumentation__.Notify(272938)
		if state == partitionStatePermanentlyClosed {
			__antithesis_instrumentation__.Notify(272940)
			return errors.Errorf("partition at index %d permanently closed, cannot Enqueue", partitionIdx)
		} else {
			__antithesis_instrumentation__.Notify(272941)
		}
		__antithesis_instrumentation__.Notify(272939)
		return errors.New("Enqueue illegally called after Dequeue or CloseAllOpenWriteFileDescriptors")
	} else {
		__antithesis_instrumentation__.Notify(272942)
	}
	__antithesis_instrumentation__.Notify(272915)
	p.lastEnqueuedPartitionIdx = partitionIdx
	return p.partitions[idx].Enqueue(ctx, batch)
}

func (p *PartitionedDiskQueue) Dequeue(
	ctx context.Context, partitionIdx int, batch coldata.Batch,
) error {
	__antithesis_instrumentation__.Notify(272943)
	idx, ok := p.partitionIdxToIndex[partitionIdx]
	if !ok {
		__antithesis_instrumentation__.Notify(272949)
		batch.SetLength(0)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(272950)
	}
	__antithesis_instrumentation__.Notify(272944)
	switch state := p.partitions[idx].state; state {
	case partitionStateWriting:
		__antithesis_instrumentation__.Notify(272951)

		if err := p.closeWritePartition(ctx, idx, retainFD); err != nil {
			__antithesis_instrumentation__.Notify(272958)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272959)
		}
		__antithesis_instrumentation__.Notify(272952)
		p.partitions[idx].state = partitionStateReading
	case partitionStateClosedForWriting, partitionStateClosedForReading:
		__antithesis_instrumentation__.Notify(272953)

		if err := p.acquireNewFD(ctx); err != nil {
			__antithesis_instrumentation__.Notify(272960)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272961)
		}
		__antithesis_instrumentation__.Notify(272954)
		p.partitions[idx].state = partitionStateReading
	case partitionStateReading:
		__antithesis_instrumentation__.Notify(272955)

	case partitionStatePermanentlyClosed:
		__antithesis_instrumentation__.Notify(272956)
		return errors.Errorf("partition at index %d permanently closed, cannot Dequeue", partitionIdx)
	default:
		__antithesis_instrumentation__.Notify(272957)
		colexecerror.InternalError(errors.AssertionFailedf("unhandled state %d", state))
	}
	__antithesis_instrumentation__.Notify(272945)
	notEmpty, err := p.partitions[idx].Dequeue(ctx, batch)
	if err != nil {
		__antithesis_instrumentation__.Notify(272962)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272963)
	}
	__antithesis_instrumentation__.Notify(272946)
	if batch.Length() == 0 {
		__antithesis_instrumentation__.Notify(272964)

		if err := p.closeReadPartition(idx); err != nil {
			__antithesis_instrumentation__.Notify(272965)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272966)
		}
	} else {
		__antithesis_instrumentation__.Notify(272967)
	}
	__antithesis_instrumentation__.Notify(272947)
	if !notEmpty {
		__antithesis_instrumentation__.Notify(272968)

		colexecerror.InternalError(
			errors.AssertionFailedf("DiskQueue unexpectedly returned that more data will be added"))
	} else {
		__antithesis_instrumentation__.Notify(272969)
	}
	__antithesis_instrumentation__.Notify(272948)
	return nil
}

func (p *PartitionedDiskQueue) CloseAllOpenWriteFileDescriptors(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(272970)
	for i, q := range p.partitions {
		__antithesis_instrumentation__.Notify(272972)
		if q.state != partitionStateWriting {
			__antithesis_instrumentation__.Notify(272974)
			continue
		} else {
			__antithesis_instrumentation__.Notify(272975)
		}
		__antithesis_instrumentation__.Notify(272973)

		if err := p.closeWritePartition(ctx, i, releaseFD); err != nil {
			__antithesis_instrumentation__.Notify(272976)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272977)
		}
	}
	__antithesis_instrumentation__.Notify(272971)
	return nil
}

func (p *PartitionedDiskQueue) CloseAllOpenReadFileDescriptors() error {
	__antithesis_instrumentation__.Notify(272978)
	for i, q := range p.partitions {
		__antithesis_instrumentation__.Notify(272980)
		if q.state != partitionStateReading {
			__antithesis_instrumentation__.Notify(272982)
			continue
		} else {
			__antithesis_instrumentation__.Notify(272983)
		}
		__antithesis_instrumentation__.Notify(272981)

		if err := p.closeReadPartition(i); err != nil {
			__antithesis_instrumentation__.Notify(272984)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272985)
		}
	}
	__antithesis_instrumentation__.Notify(272979)
	return nil
}

func (p *PartitionedDiskQueue) CloseInactiveReadPartitions(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(272986)
	var lastErr error
	for i, q := range p.partitions {
		__antithesis_instrumentation__.Notify(272988)
		if q.state != partitionStateClosedForReading {
			__antithesis_instrumentation__.Notify(272990)
			continue
		} else {
			__antithesis_instrumentation__.Notify(272991)
		}
		__antithesis_instrumentation__.Notify(272989)
		lastErr = q.Close(ctx)
		p.partitions[i].state = partitionStatePermanentlyClosed
	}
	__antithesis_instrumentation__.Notify(272987)
	return lastErr
}

func (p *PartitionedDiskQueue) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(272992)
	var lastErr error
	for i, q := range p.partitions {
		__antithesis_instrumentation__.Notify(272995)
		if q.state == partitionStatePermanentlyClosed {
			__antithesis_instrumentation__.Notify(272997)

			continue
		} else {
			__antithesis_instrumentation__.Notify(272998)
		}
		__antithesis_instrumentation__.Notify(272996)
		lastErr = q.Close(ctx)
		p.partitions[i].state = partitionStatePermanentlyClosed
	}
	__antithesis_instrumentation__.Notify(272993)
	if p.numOpenFDs != 0 {
		__antithesis_instrumentation__.Notify(272999)

		p.fdSemaphore.Release(p.numOpenFDs)
		p.numOpenFDs = 0
	} else {
		__antithesis_instrumentation__.Notify(273000)
	}
	__antithesis_instrumentation__.Notify(272994)
	return lastErr
}
