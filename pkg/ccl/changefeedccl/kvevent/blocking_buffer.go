package kvevent

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type blockingBuffer struct {
	sv       *settings.Values
	metrics  *Metrics
	qp       allocPool
	signalCh chan struct{}

	mu struct {
		syncutil.Mutex
		closed  bool
		reason  error
		drainCh chan struct{}
		blocked bool
		queue   bufferEntryQueue
	}
}

func NewMemBuffer(
	acc mon.BoundAccount, sv *settings.Values, metrics *Metrics, opts ...quotapool.Option,
) Buffer {
	__antithesis_instrumentation__.Notify(16972)
	const slowAcquisitionThreshold = 5 * time.Second

	opts = append(opts,
		quotapool.OnSlowAcquisition(slowAcquisitionThreshold, logSlowAcquisition(slowAcquisitionThreshold)),
		quotapool.OnWaitFinish(
			func(ctx context.Context, poolName string, r quotapool.Request, start time.Time) {
				__antithesis_instrumentation__.Notify(16974)
				metrics.BufferPushbackNanos.Inc(timeutil.Since(start).Nanoseconds())
			}))
	__antithesis_instrumentation__.Notify(16973)

	b := &blockingBuffer{
		signalCh: make(chan struct{}, 1),
		metrics:  metrics,
		sv:       sv,
	}
	quota := &memQuota{acc: acc, notifyOutOfQuota: b.notifyOutOfQuota}
	b.qp = allocPool{
		AbstractPool: quotapool.New("changefeed", quota, opts...),
		metrics:      metrics,
	}

	return b
}

var _ Buffer = (*blockingBuffer)(nil)

func (b *blockingBuffer) pop() (e *bufferEntry, err error) {
	__antithesis_instrumentation__.Notify(16975)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.mu.closed {
		__antithesis_instrumentation__.Notify(16979)
		return nil, ErrBufferClosed{reason: b.mu.reason}
	} else {
		__antithesis_instrumentation__.Notify(16980)
	}
	__antithesis_instrumentation__.Notify(16976)

	e = b.mu.queue.dequeue()

	if e == nil && func() bool {
		__antithesis_instrumentation__.Notify(16981)
		return b.mu.blocked == true
	}() == true {
		__antithesis_instrumentation__.Notify(16982)

		e = newBufferEntry(Event{flush: true})

		b.mu.blocked = false
	} else {
		__antithesis_instrumentation__.Notify(16983)
	}
	__antithesis_instrumentation__.Notify(16977)

	if b.mu.drainCh != nil && func() bool {
		__antithesis_instrumentation__.Notify(16984)
		return b.mu.queue.empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(16985)
		close(b.mu.drainCh)
		b.mu.drainCh = nil
	} else {
		__antithesis_instrumentation__.Notify(16986)
	}
	__antithesis_instrumentation__.Notify(16978)
	return e, nil
}

func (b *blockingBuffer) notifyOutOfQuota() {
	__antithesis_instrumentation__.Notify(16987)
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.mu.closed {
		__antithesis_instrumentation__.Notify(16989)
		return
	} else {
		__antithesis_instrumentation__.Notify(16990)
	}
	__antithesis_instrumentation__.Notify(16988)
	b.mu.blocked = true

	select {
	case b.signalCh <- struct{}{}:
		__antithesis_instrumentation__.Notify(16991)
	default:
		__antithesis_instrumentation__.Notify(16992)
	}
}

func (b *blockingBuffer) Get(ctx context.Context) (ev Event, err error) {
	__antithesis_instrumentation__.Notify(16993)
	for {
		__antithesis_instrumentation__.Notify(16994)
		got, err := b.pop()
		if err != nil {
			__antithesis_instrumentation__.Notify(16997)
			return Event{}, err
		} else {
			__antithesis_instrumentation__.Notify(16998)
		}
		__antithesis_instrumentation__.Notify(16995)

		if got != nil {
			__antithesis_instrumentation__.Notify(16999)
			b.metrics.BufferEntriesOut.Inc(1)
			e := got.e
			bufferEntryPool.Put(got)
			return e, nil
		} else {
			__antithesis_instrumentation__.Notify(17000)
		}
		__antithesis_instrumentation__.Notify(16996)

		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(17001)
			return Event{}, ctx.Err()
		case <-b.signalCh:
			__antithesis_instrumentation__.Notify(17002)
		}
	}
}

func (b *blockingBuffer) ensureOpened(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17003)
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ensureOpenedLocked(ctx)
}

func (b *blockingBuffer) ensureOpenedLocked(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17004)
	if b.mu.closed {
		__antithesis_instrumentation__.Notify(17006)
		logcrash.ReportOrPanic(ctx, b.sv, "buffer unexpectedly closed")
		return errors.AssertionFailedf("buffer unexpectedly closed")
	} else {
		__antithesis_instrumentation__.Notify(17007)
	}
	__antithesis_instrumentation__.Notify(17005)

	return nil
}

func (b *blockingBuffer) enqueue(ctx context.Context, be *bufferEntry) (err error) {
	__antithesis_instrumentation__.Notify(17008)

	b.mu.Lock()
	defer b.mu.Unlock()
	defer func() {
		__antithesis_instrumentation__.Notify(17012)
		if err != nil {
			__antithesis_instrumentation__.Notify(17013)
			bufferEntryPool.Put(be)
		} else {
			__antithesis_instrumentation__.Notify(17014)
		}
	}()
	__antithesis_instrumentation__.Notify(17009)

	if err = b.ensureOpenedLocked(ctx); err != nil {
		__antithesis_instrumentation__.Notify(17015)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17016)
	}
	__antithesis_instrumentation__.Notify(17010)

	b.metrics.BufferEntriesIn.Inc(1)
	b.mu.blocked = false
	b.mu.queue.enqueue(be)

	select {
	case b.signalCh <- struct{}{}:
		__antithesis_instrumentation__.Notify(17017)
	default:
		__antithesis_instrumentation__.Notify(17018)
	}
	__antithesis_instrumentation__.Notify(17011)
	return nil
}

func (b *blockingBuffer) Add(ctx context.Context, e Event) error {
	__antithesis_instrumentation__.Notify(17019)
	if err := b.ensureOpened(ctx); err != nil {
		__antithesis_instrumentation__.Notify(17024)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17025)
	}
	__antithesis_instrumentation__.Notify(17020)

	if e.alloc.ap != nil {
		__antithesis_instrumentation__.Notify(17026)

		return b.enqueue(ctx, newBufferEntry(e))
	} else {
		__antithesis_instrumentation__.Notify(17027)
	}
	__antithesis_instrumentation__.Notify(17021)

	alloc := int64(changefeedbase.EventMemoryMultiplier.Get(b.sv) * float64(e.approxSize))
	if l := changefeedbase.PerChangefeedMemLimit.Get(b.sv); alloc > l {
		__antithesis_instrumentation__.Notify(17028)
		return errors.Newf("event size %d exceeds per changefeed limit %d", alloc, l)
	} else {
		__antithesis_instrumentation__.Notify(17029)
	}
	__antithesis_instrumentation__.Notify(17022)
	e.alloc = Alloc{
		bytes:   alloc,
		entries: 1,
		ap:      &b.qp,
	}
	e.bufferAddTimestamp = timeutil.Now()
	be := newBufferEntry(e)

	if err := b.qp.Acquire(ctx, be); err != nil {
		__antithesis_instrumentation__.Notify(17030)
		bufferEntryPool.Put(be)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17031)
	}
	__antithesis_instrumentation__.Notify(17023)
	b.metrics.BufferEntriesMemAcquired.Inc(alloc)
	return b.enqueue(ctx, be)
}

func (b *blockingBuffer) tryDrain() chan struct{} {
	__antithesis_instrumentation__.Notify(17032)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.mu.queue.empty() {
		__antithesis_instrumentation__.Notify(17034)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(17035)
	}
	__antithesis_instrumentation__.Notify(17033)

	b.mu.drainCh = make(chan struct{})
	return b.mu.drainCh
}

func (b *blockingBuffer) Drain(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17036)
	if drained := b.tryDrain(); drained != nil {
		__antithesis_instrumentation__.Notify(17038)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(17039)
			return ctx.Err()
		case <-drained:
			__antithesis_instrumentation__.Notify(17040)
			return nil
		}
	} else {
		__antithesis_instrumentation__.Notify(17041)
	}
	__antithesis_instrumentation__.Notify(17037)

	return nil
}

func (b *blockingBuffer) CloseWithReason(ctx context.Context, reason error) error {
	__antithesis_instrumentation__.Notify(17042)

	b.qp.Close("blocking buffer closing")

	b.qp.Update(func(r quotapool.Resource) (shouldNotify bool) {
		__antithesis_instrumentation__.Notify(17046)
		quota := r.(*memQuota)
		quota.closed = true
		quota.acc.Close(ctx)
		return false
	})
	__antithesis_instrumentation__.Notify(17043)

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.mu.closed {
		__antithesis_instrumentation__.Notify(17047)
		logcrash.ReportOrPanic(ctx, b.sv, "close called multiple times")
		return errors.AssertionFailedf("close called multiple times")
	} else {
		__antithesis_instrumentation__.Notify(17048)
	}
	__antithesis_instrumentation__.Notify(17044)

	b.mu.closed = true
	b.mu.reason = reason
	close(b.signalCh)

	for be := b.mu.queue.dequeue(); be != nil; be = b.mu.queue.dequeue() {
		__antithesis_instrumentation__.Notify(17049)
		bufferEntryPool.Put(be)
	}
	__antithesis_instrumentation__.Notify(17045)

	return nil
}

type memQuota struct {
	closed bool

	allocated int64

	canAllocateBelow int64

	notifyOutOfQuota func()

	acc mon.BoundAccount
}

var _ quotapool.Resource = (*memQuota)(nil)

type bufferEntry struct {
	e    Event
	next *bufferEntry
}

var bufferEntryPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(17050)
		return new(bufferEntry)
	},
}

func newBufferEntry(e Event) *bufferEntry {
	__antithesis_instrumentation__.Notify(17051)
	be := bufferEntryPool.Get().(*bufferEntry)
	be.e = e
	be.next = nil
	return be
}

var _ quotapool.Request = (*bufferEntry)(nil)

func (be *bufferEntry) Acquire(
	ctx context.Context, resource quotapool.Resource,
) (fulfilled bool, tryAgainAfter time.Duration) {
	__antithesis_instrumentation__.Notify(17052)
	quota := resource.(*memQuota)
	fulfilled, tryAgainAfter = be.acquireQuota(ctx, quota)

	if !fulfilled {
		__antithesis_instrumentation__.Notify(17054)
		quota.notifyOutOfQuota()
	} else {
		__antithesis_instrumentation__.Notify(17055)
	}
	__antithesis_instrumentation__.Notify(17053)

	return fulfilled, tryAgainAfter
}

func (be *bufferEntry) acquireQuota(
	ctx context.Context, quota *memQuota,
) (fulfilled bool, tryAgainAfter time.Duration) {
	__antithesis_instrumentation__.Notify(17056)
	if quota.canAllocateBelow > 0 {
		__antithesis_instrumentation__.Notify(17059)
		if quota.allocated > quota.canAllocateBelow {
			__antithesis_instrumentation__.Notify(17061)
			return false, 0
		} else {
			__antithesis_instrumentation__.Notify(17062)
		}
		__antithesis_instrumentation__.Notify(17060)
		quota.canAllocateBelow = 0
	} else {
		__antithesis_instrumentation__.Notify(17063)
	}
	__antithesis_instrumentation__.Notify(17057)

	if err := quota.acc.Grow(ctx, be.e.alloc.bytes); err != nil {
		__antithesis_instrumentation__.Notify(17064)
		if quota.allocated == 0 {
			__antithesis_instrumentation__.Notify(17066)

			return false, time.Second
		} else {
			__antithesis_instrumentation__.Notify(17067)
		}
		__antithesis_instrumentation__.Notify(17065)

		quota.canAllocateBelow = quota.allocated/2 + 1
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(17068)
	}
	__antithesis_instrumentation__.Notify(17058)

	quota.allocated += be.e.alloc.bytes
	quota.canAllocateBelow = 0
	return true, 0
}

func (be *bufferEntry) ShouldWait() bool {
	__antithesis_instrumentation__.Notify(17069)
	return true
}

type bufferEntryQueue struct {
	head, tail *bufferEntry
}

func (l *bufferEntryQueue) enqueue(be *bufferEntry) {
	__antithesis_instrumentation__.Notify(17070)
	if l.tail == nil {
		__antithesis_instrumentation__.Notify(17071)
		l.head, l.tail = be, be
	} else {
		__antithesis_instrumentation__.Notify(17072)
		l.tail.next = be
		l.tail = be
	}
}

func (l *bufferEntryQueue) empty() bool {
	__antithesis_instrumentation__.Notify(17073)
	return l.head == nil
}

func (l *bufferEntryQueue) dequeue() *bufferEntry {
	__antithesis_instrumentation__.Notify(17074)
	if l.head == nil {
		__antithesis_instrumentation__.Notify(17077)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(17078)
	}
	__antithesis_instrumentation__.Notify(17075)
	ret := l.head
	if l.head = l.head.next; l.head == nil {
		__antithesis_instrumentation__.Notify(17079)
		l.tail = nil
	} else {
		__antithesis_instrumentation__.Notify(17080)
	}
	__antithesis_instrumentation__.Notify(17076)
	return ret
}

type allocPool struct {
	*quotapool.AbstractPool
	metrics *Metrics
}

func (ap allocPool) Release(ctx context.Context, bytes, entries int64) {
	__antithesis_instrumentation__.Notify(17081)
	ap.AbstractPool.Update(func(r quotapool.Resource) (shouldNotify bool) {
		__antithesis_instrumentation__.Notify(17082)
		quota := r.(*memQuota)
		if quota.closed {
			__antithesis_instrumentation__.Notify(17084)
			return false
		} else {
			__antithesis_instrumentation__.Notify(17085)
		}
		__antithesis_instrumentation__.Notify(17083)
		quota.acc.Shrink(ctx, bytes)
		quota.allocated -= bytes
		ap.metrics.BufferEntriesMemReleased.Inc(bytes)
		ap.metrics.BufferEntriesReleased.Inc(entries)
		return true
	})
}

func logSlowAcquisition(slowAcquisitionThreshold time.Duration) quotapool.SlowAcquisitionFunc {
	__antithesis_instrumentation__.Notify(17086)
	logSlowAcquire := log.Every(slowAcquisitionThreshold)

	return func(ctx context.Context, poolName string, r quotapool.Request, start time.Time) func() {
		__antithesis_instrumentation__.Notify(17087)
		shouldLog := logSlowAcquire.ShouldLog()
		if shouldLog {
			__antithesis_instrumentation__.Notify(17089)
			log.Warningf(ctx, "have been waiting %s attempting to acquire changefeed quota",
				timeutil.Since(start))
		} else {
			__antithesis_instrumentation__.Notify(17090)
		}
		__antithesis_instrumentation__.Notify(17088)

		return func() {
			__antithesis_instrumentation__.Notify(17091)
			if shouldLog {
				__antithesis_instrumentation__.Notify(17092)
				log.Infof(ctx, "acquired changefeed quota after %s", timeutil.Since(start))
			} else {
				__antithesis_instrumentation__.Notify(17093)
			}
		}
	}
}
