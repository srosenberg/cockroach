package contention

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/contention/contentionutils"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const (
	eventBatchSize   = 64
	eventChannelSize = 24
)

type eventWriter interface {
	addEvent(contentionpb.ExtendedContentionEvent)
}

type eventReader interface {
	ForEachEvent(func(*contentionpb.ExtendedContentionEvent) error) error
}

type timeSource func() time.Time

type eventBatch [eventBatchSize]contentionpb.ExtendedContentionEvent

func (b *eventBatch) len() int {
	__antithesis_instrumentation__.Notify(459047)
	for i := 0; i < eventBatchSize; i++ {
		__antithesis_instrumentation__.Notify(459049)
		if !b[i].Valid() {
			__antithesis_instrumentation__.Notify(459050)
			return i
		} else {
			__antithesis_instrumentation__.Notify(459051)
		}
	}
	__antithesis_instrumentation__.Notify(459048)
	return eventBatchSize
}

var eventBatchPool = &sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(459052)
		return &eventBatch{}
	},
}

type eventStore struct {
	st *cluster.Settings

	guard struct {
		*contentionutils.ConcurrentBufferGuard

		buffer *eventBatch
	}

	eventBatchChan chan *eventBatch
	closeCh        chan struct{}

	resolver resolverQueue

	mu struct {
		syncutil.RWMutex

		store *cache.UnorderedCache
	}

	atomic struct {
		storageSize int64
	}

	timeSrc timeSource
}

var (
	_ eventWriter = &eventStore{}
	_ eventReader = &eventStore{}
)

func newEventStore(
	st *cluster.Settings, endpoint ResolverEndpoint, timeSrc timeSource, metrics *Metrics,
) *eventStore {
	__antithesis_instrumentation__.Notify(459053)
	s := &eventStore{
		st:             st,
		resolver:       newResolver(endpoint, metrics, eventBatchSize),
		eventBatchChan: make(chan *eventBatch, eventChannelSize),
		closeCh:        make(chan struct{}),
		timeSrc:        timeSrc,
	}

	s.mu.store = cache.NewUnorderedCache(cache.Config{
		Policy: cache.CacheFIFO,
		ShouldEvict: func(_ int, _, _ interface{}) bool {
			__antithesis_instrumentation__.Notify(459056)
			capacity := StoreCapacity.Get(&st.SV)
			size := atomic.LoadInt64(&s.atomic.storageSize)
			return size > capacity
		},
		OnEvictedEntry: func(entry *cache.Entry) {
			__antithesis_instrumentation__.Notify(459057)
			event := entry.Value.(contentionpb.ExtendedContentionEvent)
			entrySize := int64(entryBytes(&event))
			atomic.AddInt64(&s.atomic.storageSize, -entrySize)
		},
	})
	__antithesis_instrumentation__.Notify(459054)

	s.guard.buffer = eventBatchPool.Get().(*eventBatch)
	s.guard.ConcurrentBufferGuard = contentionutils.NewConcurrentBufferGuard(
		func() int64 {
			__antithesis_instrumentation__.Notify(459058)
			return eventBatchSize
		},
		func(_ int64) {
			__antithesis_instrumentation__.Notify(459059)
			select {
			case s.eventBatchChan <- s.guard.buffer:
				__antithesis_instrumentation__.Notify(459061)
			case <-s.closeCh:
				__antithesis_instrumentation__.Notify(459062)
			}
			__antithesis_instrumentation__.Notify(459060)
			s.guard.buffer = eventBatchPool.Get().(*eventBatch)
		},
	)
	__antithesis_instrumentation__.Notify(459055)

	return s
}

func (s *eventStore) start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(459063)
	s.startEventIntake(ctx, stopper)
	s.startResolver(ctx, stopper)
}

func (s *eventStore) startEventIntake(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(459064)
	handleInsert := func(batch []contentionpb.ExtendedContentionEvent) {
		__antithesis_instrumentation__.Notify(459067)
		s.resolver.enqueue(batch)
		s.upsertBatch(batch)
	}
	__antithesis_instrumentation__.Notify(459065)

	consumeBatch := func(batch *eventBatch) {
		__antithesis_instrumentation__.Notify(459068)
		batchLen := batch.len()
		handleInsert(batch[:batchLen])
		*batch = eventBatch{}
		eventBatchPool.Put(batch)
	}
	__antithesis_instrumentation__.Notify(459066)

	if err := stopper.RunAsyncTask(ctx, "contention-event-intake", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(459069)
		for {
			__antithesis_instrumentation__.Notify(459070)
			select {
			case batch := <-s.eventBatchChan:
				__antithesis_instrumentation__.Notify(459071)
				consumeBatch(batch)
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(459072)
				close(s.closeCh)
				return
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(459073)
		close(s.closeCh)
	} else {
		__antithesis_instrumentation__.Notify(459074)
	}
}

func (s *eventStore) startResolver(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(459075)
	_ = stopper.RunAsyncTask(ctx, "contention-event-resolver", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(459076)

		var resolutionIntervalChanged = make(chan struct{}, 1)
		TxnIDResolutionInterval.SetOnChange(&s.st.SV, func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(459078)
			resolutionIntervalChanged <- struct{}{}
		})
		__antithesis_instrumentation__.Notify(459077)

		initialDelay := s.resolutionIntervalWithJitter()
		timer := timeutil.NewTimer()
		timer.Reset(initialDelay)

		for {
			__antithesis_instrumentation__.Notify(459079)
			waitInterval := s.resolutionIntervalWithJitter()
			timer.Reset(waitInterval)

			select {
			case <-timer.C:
				__antithesis_instrumentation__.Notify(459080)
				if err := s.flushAndResolve(ctx); err != nil {
					__antithesis_instrumentation__.Notify(459084)
					if log.V(1) {
						__antithesis_instrumentation__.Notify(459085)
						log.Warningf(ctx, "unexpected error encountered when performing "+
							"txn id resolution %s", err)
					} else {
						__antithesis_instrumentation__.Notify(459086)
					}
				} else {
					__antithesis_instrumentation__.Notify(459087)
				}
				__antithesis_instrumentation__.Notify(459081)
				timer.Read = true
			case <-resolutionIntervalChanged:
				__antithesis_instrumentation__.Notify(459082)
				continue
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(459083)
				return
			}
		}
	})
}

func (s *eventStore) addEvent(e contentionpb.ExtendedContentionEvent) {
	__antithesis_instrumentation__.Notify(459088)

	if TxnIDResolutionInterval.Get(&s.st.SV) == 0 {
		__antithesis_instrumentation__.Notify(459091)
		return
	} else {
		__antithesis_instrumentation__.Notify(459092)
	}
	__antithesis_instrumentation__.Notify(459089)

	if threshold := DurationThreshold.Get(&s.st.SV); threshold > 0 {
		__antithesis_instrumentation__.Notify(459093)
		if e.BlockingEvent.Duration < threshold {
			__antithesis_instrumentation__.Notify(459094)
			return
		} else {
			__antithesis_instrumentation__.Notify(459095)
		}
	} else {
		__antithesis_instrumentation__.Notify(459096)
	}
	__antithesis_instrumentation__.Notify(459090)

	s.guard.AtomicWrite(func(writerIdx int64) {
		__antithesis_instrumentation__.Notify(459097)
		e.CollectionTs = s.timeSrc()
		s.guard.buffer[writerIdx] = e
	})
}

func (s *eventStore) ForEachEvent(
	op func(event *contentionpb.ExtendedContentionEvent) error,
) error {
	__antithesis_instrumentation__.Notify(459098)

	s.mu.RLock()
	keys := make([]uint64, 0, s.mu.store.Len())
	s.mu.store.Do(func(entry *cache.Entry) {
		__antithesis_instrumentation__.Notify(459101)
		keys = append(keys, entry.Key.(uint64))
	})
	__antithesis_instrumentation__.Notify(459099)
	s.mu.RUnlock()

	for i := range keys {
		__antithesis_instrumentation__.Notify(459102)
		event, ok := s.getEventByEventHash(keys[i])
		if !ok {
			__antithesis_instrumentation__.Notify(459104)

			continue
		} else {
			__antithesis_instrumentation__.Notify(459105)
		}
		__antithesis_instrumentation__.Notify(459103)
		if err := op(&event); err != nil {
			__antithesis_instrumentation__.Notify(459106)
			return err
		} else {
			__antithesis_instrumentation__.Notify(459107)
		}
	}
	__antithesis_instrumentation__.Notify(459100)

	return nil
}

func (s *eventStore) getEventByEventHash(
	hash uint64,
) (_ contentionpb.ExtendedContentionEvent, ok bool) {
	__antithesis_instrumentation__.Notify(459108)
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.mu.store.Get(hash)
	return event.(contentionpb.ExtendedContentionEvent), ok
}

func (s *eventStore) flushAndResolve(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(459109)

	s.guard.ForceSync()

	result, err := s.resolver.dequeue(ctx)

	s.upsertBatch(result)

	return err
}

func (s *eventStore) upsertBatch(events []contentionpb.ExtendedContentionEvent) {
	__antithesis_instrumentation__.Notify(459110)
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range events {
		__antithesis_instrumentation__.Notify(459111)
		blockingTxnID := events[i].BlockingEvent.TxnMeta.ID
		_, ok := s.mu.store.Get(blockingTxnID)
		if !ok {
			__antithesis_instrumentation__.Notify(459113)
			atomic.AddInt64(&s.atomic.storageSize, int64(entryBytes(&events[i])))
		} else {
			__antithesis_instrumentation__.Notify(459114)
		}
		__antithesis_instrumentation__.Notify(459112)
		s.mu.store.Add(events[i].Hash(), events[i])
	}
}

func (s *eventStore) resolutionIntervalWithJitter() time.Duration {
	__antithesis_instrumentation__.Notify(459115)
	baseInterval := TxnIDResolutionInterval.Get(&s.st.SV)

	frac := 1 + (2*rand.Float64()-1)*0.15
	jitteredInterval := time.Duration(frac * float64(baseInterval.Nanoseconds()))
	return jitteredInterval
}

func entryBytes(event *contentionpb.ExtendedContentionEvent) int {
	__antithesis_instrumentation__.Notify(459116)

	return event.Size() + 8
}
