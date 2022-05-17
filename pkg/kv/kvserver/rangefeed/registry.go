package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type Stream interface {
	Context() context.Context

	Send(*roachpb.RangeFeedEvent) error
}

type sharedEvent struct {
	event      *roachpb.RangeFeedEvent
	allocation *SharedBudgetAllocation
}

var sharedEventSyncPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(113822)
		return new(sharedEvent)
	},
}

func getPooledSharedEvent(e sharedEvent) *sharedEvent {
	__antithesis_instrumentation__.Notify(113823)
	ev := sharedEventSyncPool.Get().(*sharedEvent)
	*ev = e
	return ev
}

func putPooledSharedEvent(e *sharedEvent) {
	__antithesis_instrumentation__.Notify(113824)
	*e = sharedEvent{}
	sharedEventSyncPool.Put(e)
}

type registration struct {
	span             roachpb.Span
	catchUpTimestamp hlc.Timestamp
	withDiff         bool
	metrics          *Metrics

	catchUpIterConstructor CatchUpIteratorConstructor

	stream Stream
	errC   chan<- *roachpb.Error

	id   int64
	keys interval.Range
	buf  chan *sharedEvent

	mu struct {
		sync.Locker

		overflowed bool

		caughtUp bool

		outputLoopCancelFn func()
		disconnected       bool

		catchUpIter *CatchUpIterator
	}
}

func newRegistration(
	span roachpb.Span,
	startTS hlc.Timestamp,
	catchUpIterConstructor CatchUpIteratorConstructor,
	withDiff bool,
	bufferSz int,
	metrics *Metrics,
	stream Stream,
	errC chan<- *roachpb.Error,
) registration {
	__antithesis_instrumentation__.Notify(113825)
	r := registration{
		span:                   span,
		catchUpTimestamp:       startTS,
		catchUpIterConstructor: catchUpIterConstructor,
		withDiff:               withDiff,
		metrics:                metrics,
		stream:                 stream,
		errC:                   errC,
		buf:                    make(chan *sharedEvent, bufferSz),
	}
	r.mu.Locker = &syncutil.Mutex{}
	r.mu.caughtUp = true
	return r
}

func (r *registration) publish(
	ctx context.Context, event *roachpb.RangeFeedEvent, allocation *SharedBudgetAllocation,
) {
	__antithesis_instrumentation__.Notify(113826)
	r.validateEvent(event)
	e := getPooledSharedEvent(sharedEvent{event: r.maybeStripEvent(event), allocation: allocation})

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mu.overflowed {
		__antithesis_instrumentation__.Notify(113828)
		return
	} else {
		__antithesis_instrumentation__.Notify(113829)
	}
	__antithesis_instrumentation__.Notify(113827)
	allocation.Use()
	select {
	case r.buf <- e:
		__antithesis_instrumentation__.Notify(113830)
		r.mu.caughtUp = false
	default:
		__antithesis_instrumentation__.Notify(113831)

		r.mu.overflowed = true
		allocation.Release(ctx)
	}
}

func (r *registration) validateEvent(event *roachpb.RangeFeedEvent) {
	__antithesis_instrumentation__.Notify(113832)
	switch t := event.GetValue().(type) {
	case *roachpb.RangeFeedValue:
		__antithesis_instrumentation__.Notify(113833)
		if t.Key == nil {
			__antithesis_instrumentation__.Notify(113841)
			panic(fmt.Sprintf("unexpected empty RangeFeedValue.Key: %v", t))
		} else {
			__antithesis_instrumentation__.Notify(113842)
		}
		__antithesis_instrumentation__.Notify(113834)
		if t.Value.RawBytes == nil {
			__antithesis_instrumentation__.Notify(113843)
			panic(fmt.Sprintf("unexpected empty RangeFeedValue.Value.RawBytes: %v", t))
		} else {
			__antithesis_instrumentation__.Notify(113844)
		}
		__antithesis_instrumentation__.Notify(113835)
		if t.Value.Timestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(113845)
			panic(fmt.Sprintf("unexpected empty RangeFeedValue.Value.Timestamp: %v", t))
		} else {
			__antithesis_instrumentation__.Notify(113846)
		}
	case *roachpb.RangeFeedCheckpoint:
		__antithesis_instrumentation__.Notify(113836)
		if t.Span.Key == nil {
			__antithesis_instrumentation__.Notify(113847)
			panic(fmt.Sprintf("unexpected empty RangeFeedCheckpoint.Span.Key: %v", t))
		} else {
			__antithesis_instrumentation__.Notify(113848)
		}
	case *roachpb.RangeFeedSSTable:
		__antithesis_instrumentation__.Notify(113837)
		if len(t.Data) == 0 {
			__antithesis_instrumentation__.Notify(113849)
			panic(fmt.Sprintf("unexpected empty RangeFeedSSTable.Data: %v", t))
		} else {
			__antithesis_instrumentation__.Notify(113850)
		}
		__antithesis_instrumentation__.Notify(113838)
		if len(t.Span.Key) == 0 {
			__antithesis_instrumentation__.Notify(113851)
			panic(fmt.Sprintf("unexpected empty RangeFeedSSTable.Span: %v", t))
		} else {
			__antithesis_instrumentation__.Notify(113852)
		}
		__antithesis_instrumentation__.Notify(113839)
		if t.WriteTS.IsEmpty() {
			__antithesis_instrumentation__.Notify(113853)
			panic(fmt.Sprintf("unexpected empty RangeFeedSSTable.Timestamp: %v", t))
		} else {
			__antithesis_instrumentation__.Notify(113854)
		}
	default:
		__antithesis_instrumentation__.Notify(113840)
		panic(fmt.Sprintf("unexpected RangeFeedEvent variant: %v", t))
	}
}

func (r *registration) maybeStripEvent(event *roachpb.RangeFeedEvent) *roachpb.RangeFeedEvent {
	__antithesis_instrumentation__.Notify(113855)
	ret := event
	copyOnWrite := func() interface{} {
		__antithesis_instrumentation__.Notify(113858)
		if ret == event {
			__antithesis_instrumentation__.Notify(113860)
			ret = event.ShallowCopy()
		} else {
			__antithesis_instrumentation__.Notify(113861)
		}
		__antithesis_instrumentation__.Notify(113859)
		return ret.GetValue()
	}
	__antithesis_instrumentation__.Notify(113856)

	switch t := ret.GetValue().(type) {
	case *roachpb.RangeFeedValue:
		__antithesis_instrumentation__.Notify(113862)
		if t.PrevValue.IsPresent() && func() bool {
			__antithesis_instrumentation__.Notify(113866)
			return !r.withDiff == true
		}() == true {
			__antithesis_instrumentation__.Notify(113867)

			t = copyOnWrite().(*roachpb.RangeFeedValue)
			t.PrevValue = roachpb.Value{}
		} else {
			__antithesis_instrumentation__.Notify(113868)
		}
	case *roachpb.RangeFeedCheckpoint:
		__antithesis_instrumentation__.Notify(113863)
		if !t.Span.EqualValue(r.span) {
			__antithesis_instrumentation__.Notify(113869)

			if !t.Span.Contains(r.span) {
				__antithesis_instrumentation__.Notify(113871)
				panic(fmt.Sprintf("registration span %v larger than checkpoint span %v", r.span, t.Span))
			} else {
				__antithesis_instrumentation__.Notify(113872)
			}
			__antithesis_instrumentation__.Notify(113870)
			t = copyOnWrite().(*roachpb.RangeFeedCheckpoint)
			t.Span = r.span
		} else {
			__antithesis_instrumentation__.Notify(113873)
		}
	case *roachpb.RangeFeedSSTable:
		__antithesis_instrumentation__.Notify(113864)

	default:
		__antithesis_instrumentation__.Notify(113865)
		panic(fmt.Sprintf("unexpected RangeFeedEvent variant: %v", t))
	}
	__antithesis_instrumentation__.Notify(113857)
	return ret
}

func (r *registration) disconnect(pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(113874)
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.mu.disconnected {
		__antithesis_instrumentation__.Notify(113875)
		if r.mu.catchUpIter != nil {
			__antithesis_instrumentation__.Notify(113878)
			r.mu.catchUpIter.Close()
			r.mu.catchUpIter = nil
		} else {
			__antithesis_instrumentation__.Notify(113879)
		}
		__antithesis_instrumentation__.Notify(113876)
		if r.mu.outputLoopCancelFn != nil {
			__antithesis_instrumentation__.Notify(113880)
			r.mu.outputLoopCancelFn()
		} else {
			__antithesis_instrumentation__.Notify(113881)
		}
		__antithesis_instrumentation__.Notify(113877)
		r.mu.disconnected = true
		r.errC <- pErr
	} else {
		__antithesis_instrumentation__.Notify(113882)
	}
}

func (r *registration) outputLoop(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(113883)

	if err := r.maybeRunCatchUpScan(); err != nil {
		__antithesis_instrumentation__.Notify(113885)
		err = errors.Wrap(err, "catch-up scan failed")
		log.Errorf(ctx, "%v", err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(113886)
	}
	__antithesis_instrumentation__.Notify(113884)

	for {
		__antithesis_instrumentation__.Notify(113887)
		overflowed := false
		r.mu.Lock()
		if len(r.buf) == 0 {
			__antithesis_instrumentation__.Notify(113890)
			overflowed = r.mu.overflowed
			r.mu.caughtUp = true
		} else {
			__antithesis_instrumentation__.Notify(113891)
		}
		__antithesis_instrumentation__.Notify(113888)
		r.mu.Unlock()
		if overflowed {
			__antithesis_instrumentation__.Notify(113892)
			return newErrBufferCapacityExceeded().GoError()
		} else {
			__antithesis_instrumentation__.Notify(113893)
		}
		__antithesis_instrumentation__.Notify(113889)

		select {
		case nextEvent := <-r.buf:
			__antithesis_instrumentation__.Notify(113894)
			err := r.stream.Send(nextEvent.event)
			nextEvent.allocation.Release(ctx)
			putPooledSharedEvent(nextEvent)
			if err != nil {
				__antithesis_instrumentation__.Notify(113897)
				return err
			} else {
				__antithesis_instrumentation__.Notify(113898)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(113895)
			return ctx.Err()
		case <-r.stream.Context().Done():
			__antithesis_instrumentation__.Notify(113896)
			return r.stream.Context().Err()
		}
	}
}

func (r *registration) runOutputLoop(ctx context.Context, _forStacks roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(113899)
	r.mu.Lock()
	if r.mu.disconnected {
		__antithesis_instrumentation__.Notify(113901)

		r.mu.Unlock()
		return
	} else {
		__antithesis_instrumentation__.Notify(113902)
	}
	__antithesis_instrumentation__.Notify(113900)
	ctx, r.mu.outputLoopCancelFn = context.WithCancel(ctx)
	r.mu.Unlock()
	err := r.outputLoop(ctx)
	r.disconnect(roachpb.NewError(err))
}

func (r *registration) drainAllocations(ctx context.Context) {
	__antithesis_instrumentation__.Notify(113903)
	for {
		__antithesis_instrumentation__.Notify(113904)
		select {
		case e, ok := <-r.buf:
			__antithesis_instrumentation__.Notify(113905)
			if !ok {
				__antithesis_instrumentation__.Notify(113908)
				return
			} else {
				__antithesis_instrumentation__.Notify(113909)
			}
			__antithesis_instrumentation__.Notify(113906)
			e.allocation.Release(ctx)
			putPooledSharedEvent(e)
		default:
			__antithesis_instrumentation__.Notify(113907)
			return
		}
	}
}

func (r *registration) maybeRunCatchUpScan() error {
	__antithesis_instrumentation__.Notify(113910)
	catchUpIter := r.detachCatchUpIter()
	if catchUpIter == nil {
		__antithesis_instrumentation__.Notify(113913)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(113914)
	}
	__antithesis_instrumentation__.Notify(113911)
	start := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(113915)
		catchUpIter.Close()
		r.metrics.RangeFeedCatchUpScanNanos.Inc(timeutil.Since(start).Nanoseconds())
	}()
	__antithesis_instrumentation__.Notify(113912)

	startKey := storage.MakeMVCCMetadataKey(r.span.Key)
	endKey := storage.MakeMVCCMetadataKey(r.span.EndKey)

	return catchUpIter.CatchUpScan(startKey, endKey, r.catchUpTimestamp, r.withDiff, r.stream.Send)
}

func (r *registration) ID() uintptr {
	__antithesis_instrumentation__.Notify(113916)
	return uintptr(r.id)
}

func (r *registration) Range() interval.Range {
	__antithesis_instrumentation__.Notify(113917)
	return r.keys
}

func (r registration) String() string {
	__antithesis_instrumentation__.Notify(113918)
	return fmt.Sprintf("[%s @ %s+]", r.span, r.catchUpTimestamp)
}

type registry struct {
	tree    interval.Tree
	idAlloc int64
}

func makeRegistry() registry {
	__antithesis_instrumentation__.Notify(113919)
	return registry{
		tree: interval.NewTree(interval.ExclusiveOverlapper),
	}
}

func (reg *registry) Len() int {
	__antithesis_instrumentation__.Notify(113920)
	return reg.tree.Len()
}

func (reg *registry) NewFilter() *Filter {
	__antithesis_instrumentation__.Notify(113921)
	return newFilterFromRegistry(reg)
}

func (reg *registry) Register(r *registration) {
	__antithesis_instrumentation__.Notify(113922)
	r.id = reg.nextID()
	r.keys = r.span.AsRange()
	if err := reg.tree.Insert(r, false); err != nil {
		__antithesis_instrumentation__.Notify(113923)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(113924)
	}
}

func (reg *registry) nextID() int64 {
	__antithesis_instrumentation__.Notify(113925)
	reg.idAlloc++
	return reg.idAlloc
}

func (reg *registry) PublishToOverlapping(
	ctx context.Context,
	span roachpb.Span,
	event *roachpb.RangeFeedEvent,
	allocation *SharedBudgetAllocation,
) {
	__antithesis_instrumentation__.Notify(113926)

	var minTS hlc.Timestamp
	switch t := event.GetValue().(type) {
	case *roachpb.RangeFeedValue:
		__antithesis_instrumentation__.Notify(113928)
		minTS = t.Value.Timestamp
	case *roachpb.RangeFeedSSTable:
		__antithesis_instrumentation__.Notify(113929)
		minTS = t.WriteTS
	case *roachpb.RangeFeedCheckpoint:
		__antithesis_instrumentation__.Notify(113930)

		minTS = hlc.MaxTimestamp
	default:
		__antithesis_instrumentation__.Notify(113931)
		panic(fmt.Sprintf("unexpected RangeFeedEvent variant: %v", t))
	}
	__antithesis_instrumentation__.Notify(113927)

	reg.forOverlappingRegs(span, func(r *registration) (bool, *roachpb.Error) {
		__antithesis_instrumentation__.Notify(113932)

		if r.catchUpTimestamp.Less(minTS) {
			__antithesis_instrumentation__.Notify(113934)
			r.publish(ctx, event, allocation)
		} else {
			__antithesis_instrumentation__.Notify(113935)
		}
		__antithesis_instrumentation__.Notify(113933)
		return false, nil
	})
}

func (reg *registry) Unregister(ctx context.Context, r *registration) {
	__antithesis_instrumentation__.Notify(113936)
	if err := reg.tree.Delete(r, false); err != nil {
		__antithesis_instrumentation__.Notify(113938)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(113939)
	}
	__antithesis_instrumentation__.Notify(113937)
	r.drainAllocations(ctx)
}

func (reg *registry) Disconnect(span roachpb.Span) {
	__antithesis_instrumentation__.Notify(113940)
	reg.DisconnectWithErr(span, nil)
}

func (reg *registry) DisconnectWithErr(span roachpb.Span, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(113941)
	reg.forOverlappingRegs(span, func(_ *registration) (bool, *roachpb.Error) {
		__antithesis_instrumentation__.Notify(113942)
		return true, pErr
	})
}

var all = roachpb.Span{Key: roachpb.KeyMin, EndKey: roachpb.KeyMax}

func (reg *registry) forOverlappingRegs(
	span roachpb.Span, fn func(*registration) (disconnect bool, pErr *roachpb.Error),
) {
	__antithesis_instrumentation__.Notify(113943)
	var toDelete []interval.Interface
	matchFn := func(i interval.Interface) (done bool) {
		__antithesis_instrumentation__.Notify(113946)
		r := i.(*registration)
		dis, pErr := fn(r)
		if dis {
			__antithesis_instrumentation__.Notify(113948)
			r.disconnect(pErr)
			toDelete = append(toDelete, i)
		} else {
			__antithesis_instrumentation__.Notify(113949)
		}
		__antithesis_instrumentation__.Notify(113947)
		return false
	}
	__antithesis_instrumentation__.Notify(113944)
	if span.EqualValue(all) {
		__antithesis_instrumentation__.Notify(113950)
		reg.tree.Do(matchFn)
	} else {
		__antithesis_instrumentation__.Notify(113951)
		reg.tree.DoMatching(matchFn, span.AsRange())
	}
	__antithesis_instrumentation__.Notify(113945)

	if len(toDelete) == reg.tree.Len() {
		__antithesis_instrumentation__.Notify(113952)
		reg.tree.Clear()
	} else {
		__antithesis_instrumentation__.Notify(113953)
		if len(toDelete) == 1 {
			__antithesis_instrumentation__.Notify(113954)
			if err := reg.tree.Delete(toDelete[0], false); err != nil {
				__antithesis_instrumentation__.Notify(113955)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(113956)
			}
		} else {
			__antithesis_instrumentation__.Notify(113957)
			if len(toDelete) > 1 {
				__antithesis_instrumentation__.Notify(113958)
				for _, i := range toDelete {
					__antithesis_instrumentation__.Notify(113960)
					if err := reg.tree.Delete(i, true); err != nil {
						__antithesis_instrumentation__.Notify(113961)
						panic(err)
					} else {
						__antithesis_instrumentation__.Notify(113962)
					}
				}
				__antithesis_instrumentation__.Notify(113959)
				reg.tree.AdjustRanges()
			} else {
				__antithesis_instrumentation__.Notify(113963)
			}
		}
	}
}

func (r *registration) waitForCaughtUp() error {
	__antithesis_instrumentation__.Notify(113964)
	opts := retry.Options{
		InitialBackoff: 5 * time.Millisecond,
		Multiplier:     2,
		MaxBackoff:     10 * time.Second,
		MaxRetries:     50,
	}
	for re := retry.Start(opts); re.Next(); {
		__antithesis_instrumentation__.Notify(113966)
		r.mu.Lock()
		caughtUp := len(r.buf) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(113967)
			return r.mu.caughtUp == true
		}() == true
		r.mu.Unlock()
		if caughtUp {
			__antithesis_instrumentation__.Notify(113968)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(113969)
		}
	}
	__antithesis_instrumentation__.Notify(113965)
	return errors.Errorf("registration %v failed to empty in time", r.Range())
}

func (r *registration) maybeConstructCatchUpIter() {
	__antithesis_instrumentation__.Notify(113970)
	if r.catchUpIterConstructor == nil {
		__antithesis_instrumentation__.Notify(113972)
		return
	} else {
		__antithesis_instrumentation__.Notify(113973)
	}
	__antithesis_instrumentation__.Notify(113971)

	catchUpIter := r.catchUpIterConstructor()
	r.catchUpIterConstructor = nil

	r.mu.Lock()
	defer r.mu.Unlock()
	r.mu.catchUpIter = catchUpIter
}

func (r *registration) detachCatchUpIter() *CatchUpIterator {
	__antithesis_instrumentation__.Notify(113974)
	r.mu.Lock()
	defer r.mu.Unlock()
	catchUpIter := r.mu.catchUpIter
	r.mu.catchUpIter = nil
	return catchUpIter
}

func (reg *registry) waitForCaughtUp(span roachpb.Span) error {
	__antithesis_instrumentation__.Notify(113975)
	var outerErr error
	reg.forOverlappingRegs(span, func(r *registration) (bool, *roachpb.Error) {
		__antithesis_instrumentation__.Notify(113977)
		if outerErr == nil {
			__antithesis_instrumentation__.Notify(113979)
			outerErr = r.waitForCaughtUp()
		} else {
			__antithesis_instrumentation__.Notify(113980)
		}
		__antithesis_instrumentation__.Notify(113978)
		return false, nil
	})
	__antithesis_instrumentation__.Notify(113976)
	return outerErr
}
