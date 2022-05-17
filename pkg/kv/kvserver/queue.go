package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

const (
	purgatoryReportInterval = 10 * time.Minute

	defaultProcessTimeout = 1 * time.Minute

	defaultQueueMaxSize = 10000
)

var queueGuaranteedProcessingTimeBudget = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.queue.process.guaranteed_time_budget",
	"the guaranteed duration before which the processing of a queue may "+
		"time out",
	defaultProcessTimeout,
)

func init() {
	queueGuaranteedProcessingTimeBudget.SetVisibility(settings.Reserved)
}

func defaultProcessTimeoutFunc(cs *cluster.Settings, _ replicaInQueue) time.Duration {
	__antithesis_instrumentation__.Notify(112211)
	return queueGuaranteedProcessingTimeBudget.Get(&cs.SV)
}

func makeRateLimitedTimeoutFunc(rateSetting *settings.ByteSizeSetting) queueProcessTimeoutFunc {
	__antithesis_instrumentation__.Notify(112212)
	return func(cs *cluster.Settings, r replicaInQueue) time.Duration {
		__antithesis_instrumentation__.Notify(112213)
		minimumTimeout := queueGuaranteedProcessingTimeBudget.Get(&cs.SV)

		repl, ok := r.(interface{ GetMVCCStats() enginepb.MVCCStats })
		if !ok {
			__antithesis_instrumentation__.Notify(112216)
			return minimumTimeout
		} else {
			__antithesis_instrumentation__.Notify(112217)
		}
		__antithesis_instrumentation__.Notify(112214)
		snapshotRate := rateSetting.Get(&cs.SV)
		stats := repl.GetMVCCStats()
		totalBytes := stats.KeyBytes + stats.ValBytes + stats.IntentBytes + stats.SysBytes
		estimatedDuration := time.Duration(totalBytes/snapshotRate) * time.Second
		timeout := estimatedDuration * permittedRangeScanSlowdown
		if timeout < minimumTimeout {
			__antithesis_instrumentation__.Notify(112218)
			timeout = minimumTimeout
		} else {
			__antithesis_instrumentation__.Notify(112219)
		}
		__antithesis_instrumentation__.Notify(112215)
		return timeout
	}
}

const permittedRangeScanSlowdown = 10

type purgatoryError interface {
	error
	purgatoryErrorMarker()
}

type processCallback func(error)

type replicaItem struct {
	rangeID   roachpb.RangeID
	replicaID roachpb.ReplicaID
	seq       int

	priority float64
	index    int

	processing bool
	requeue    bool
	callbacks  []processCallback
}

func (i *replicaItem) setProcessing() {
	__antithesis_instrumentation__.Notify(112220)
	i.priority = 0
	if i.index >= 0 {
		__antithesis_instrumentation__.Notify(112222)
		log.Fatalf(context.Background(),
			"r%d marked as processing but appears in prioQ", i.rangeID,
		)
	} else {
		__antithesis_instrumentation__.Notify(112223)
	}
	__antithesis_instrumentation__.Notify(112221)
	i.processing = true
}

func (i *replicaItem) registerCallback(cb processCallback) {
	__antithesis_instrumentation__.Notify(112224)
	i.callbacks = append(i.callbacks, cb)
}

type priorityQueue struct {
	seqGen int
	sl     []*replicaItem
}

func (pq priorityQueue) Len() int { __antithesis_instrumentation__.Notify(112225); return len(pq.sl) }

func (pq priorityQueue) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(112226)
	a, b := pq.sl[i], pq.sl[j]
	if a.priority == b.priority {
		__antithesis_instrumentation__.Notify(112228)

		return a.seq < b.seq
	} else {
		__antithesis_instrumentation__.Notify(112229)
	}
	__antithesis_instrumentation__.Notify(112227)

	return a.priority > b.priority
}

func (pq priorityQueue) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(112230)
	pq.sl[i], pq.sl[j] = pq.sl[j], pq.sl[i]
	pq.sl[i].index, pq.sl[j].index = i, j
}

func (pq *priorityQueue) Push(x interface{}) {
	__antithesis_instrumentation__.Notify(112231)
	n := len(pq.sl)
	item := x.(*replicaItem)
	item.index = n
	pq.seqGen++
	item.seq = pq.seqGen
	pq.sl = append(pq.sl, item)
}

func (pq *priorityQueue) Pop() interface{} {
	__antithesis_instrumentation__.Notify(112232)
	old := pq.sl
	n := len(old)
	item := old[n-1]
	item.index = -1
	old[n-1] = nil
	pq.sl = old[0 : n-1]
	return item
}

func (pq *priorityQueue) update(item *replicaItem, priority float64) {
	__antithesis_instrumentation__.Notify(112233)
	item.priority = priority
	if len(pq.sl) <= item.index || func() bool {
		__antithesis_instrumentation__.Notify(112235)
		return pq.sl[item.index] != item == true
	}() == true {
		__antithesis_instrumentation__.Notify(112236)
		log.Fatalf(context.Background(), "updating item in heap that's not contained in it: %v", item)
	} else {
		__antithesis_instrumentation__.Notify(112237)
	}
	__antithesis_instrumentation__.Notify(112234)
	heap.Fix(pq, item.index)
}

var (
	errQueueDisabled = errors.New("queue disabled")
	errQueueStopped  = errors.New("queue stopped")
)

func isExpectedQueueError(err error) bool {
	__antithesis_instrumentation__.Notify(112238)
	return err == nil || func() bool {
		__antithesis_instrumentation__.Notify(112239)
		return errors.Is(err, errQueueDisabled) == true
	}() == true
}

func shouldQueueAgain(now, last hlc.Timestamp, minInterval time.Duration) (bool, float64) {
	__antithesis_instrumentation__.Notify(112240)
	if minInterval == 0 || func() bool {
		__antithesis_instrumentation__.Notify(112243)
		return last.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(112244)
		return true, 0
	} else {
		__antithesis_instrumentation__.Notify(112245)
	}
	__antithesis_instrumentation__.Notify(112241)
	if diff := now.GoTime().Sub(last.GoTime()); diff >= minInterval {
		__antithesis_instrumentation__.Notify(112246)
		priority := float64(1)

		if !last.IsEmpty() {
			__antithesis_instrumentation__.Notify(112248)
			priority = float64(diff.Nanoseconds()) / float64(minInterval.Nanoseconds())
		} else {
			__antithesis_instrumentation__.Notify(112249)
		}
		__antithesis_instrumentation__.Notify(112247)
		return true, priority
	} else {
		__antithesis_instrumentation__.Notify(112250)
	}
	__antithesis_instrumentation__.Notify(112242)
	return false, 0
}

type replicaInQueue interface {
	AnnotateCtx(context.Context) context.Context
	ReplicaID() roachpb.ReplicaID
	StoreID() roachpb.StoreID
	GetRangeID() roachpb.RangeID
	IsInitialized() bool
	IsDestroyed() (DestroyReason, error)
	Desc() *roachpb.RangeDescriptor
	maybeInitializeRaftGroup(context.Context)
	redirectOnOrAcquireLease(context.Context) (kvserverpb.LeaseStatus, *roachpb.Error)
	LeaseStatusAt(context.Context, hlc.ClockTimestamp) kvserverpb.LeaseStatus
}

type queueImpl interface {
	shouldQueue(context.Context, hlc.ClockTimestamp, *Replica, spanconfig.StoreReader) (shouldQueue bool, priority float64)

	process(context.Context, *Replica, spanconfig.StoreReader) (processed bool, err error)

	timer(time.Duration) time.Duration

	purgatoryChan() <-chan time.Time
}

type queueProcessTimeoutFunc func(*cluster.Settings, replicaInQueue) time.Duration

type queueConfig struct {
	maxSize int

	maxConcurrency       int
	addOrMaybeAddSemSize int

	needsLease bool

	needsRaftInitialized bool

	needsSystemConfig bool

	acceptsUnsplitRanges bool

	processDestroyedReplicas bool

	processTimeoutFunc queueProcessTimeoutFunc

	successes *metric.Counter

	failures *metric.Counter

	pending *metric.Gauge

	processingNanos *metric.Counter

	purgatory *metric.Gauge
}

type baseQueue struct {
	log.AmbientContext

	name       string
	getReplica func(roachpb.RangeID) (replicaInQueue, error)

	impl  queueImpl
	store *Store
	queueConfig
	incoming         chan struct{}
	processSem       chan struct{}
	addOrMaybeAddSem *quotapool.IntPool
	addLogN          log.EveryN
	processDur       int64
	mu               struct {
		syncutil.Mutex
		replicas  map[roachpb.RangeID]*replicaItem
		priorityQ priorityQueue
		purgatory map[roachpb.RangeID]purgatoryError
		stopped   bool

		disabled bool
	}
}

func newBaseQueue(name string, impl queueImpl, store *Store, cfg queueConfig) *baseQueue {
	__antithesis_instrumentation__.Notify(112251)

	if cfg.processTimeoutFunc == nil {
		__antithesis_instrumentation__.Notify(112257)
		cfg.processTimeoutFunc = defaultProcessTimeoutFunc
	} else {
		__antithesis_instrumentation__.Notify(112258)
	}
	__antithesis_instrumentation__.Notify(112252)
	if cfg.maxConcurrency == 0 {
		__antithesis_instrumentation__.Notify(112259)
		cfg.maxConcurrency = 1
	} else {
		__antithesis_instrumentation__.Notify(112260)
	}
	__antithesis_instrumentation__.Notify(112253)

	if cfg.addOrMaybeAddSemSize == 0 {
		__antithesis_instrumentation__.Notify(112261)
		cfg.addOrMaybeAddSemSize = 20
	} else {
		__antithesis_instrumentation__.Notify(112262)
	}
	__antithesis_instrumentation__.Notify(112254)

	ambient := store.cfg.AmbientCtx
	ambient.AddLogTag(name, nil)

	if !cfg.acceptsUnsplitRanges && func() bool {
		__antithesis_instrumentation__.Notify(112263)
		return !cfg.needsSystemConfig == true
	}() == true {
		__antithesis_instrumentation__.Notify(112264)
		log.Fatalf(ambient.AnnotateCtx(context.Background()),
			"misconfigured queue: acceptsUnsplitRanges=false requires needsSystemConfig=true; got %+v", cfg)
	} else {
		__antithesis_instrumentation__.Notify(112265)
	}
	__antithesis_instrumentation__.Notify(112255)

	bq := baseQueue{
		AmbientContext:   ambient,
		name:             name,
		impl:             impl,
		store:            store,
		queueConfig:      cfg,
		incoming:         make(chan struct{}, 1),
		processSem:       make(chan struct{}, cfg.maxConcurrency),
		addOrMaybeAddSem: quotapool.NewIntPool("queue-add", uint64(cfg.addOrMaybeAddSemSize)),
		addLogN:          log.Every(5 * time.Second),
		getReplica: func(id roachpb.RangeID) (replicaInQueue, error) {
			__antithesis_instrumentation__.Notify(112266)
			repl, err := store.GetReplica(id)
			if repl == nil || func() bool {
				__antithesis_instrumentation__.Notify(112268)
				return err != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(112269)

				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(112270)
			}
			__antithesis_instrumentation__.Notify(112267)
			return repl, err
		},
	}
	__antithesis_instrumentation__.Notify(112256)
	bq.mu.replicas = map[roachpb.RangeID]*replicaItem{}

	return &bq
}

func (bq *baseQueue) Name() string {
	__antithesis_instrumentation__.Notify(112271)
	return bq.name
}

func (bq *baseQueue) NeedsLease() bool {
	__antithesis_instrumentation__.Notify(112272)
	return bq.needsLease
}

func (bq *baseQueue) Length() int {
	__antithesis_instrumentation__.Notify(112273)
	bq.mu.Lock()
	defer bq.mu.Unlock()
	return bq.mu.priorityQ.Len()
}

func (bq *baseQueue) PurgatoryLength() int {
	__antithesis_instrumentation__.Notify(112274)

	defer bq.lockProcessing()()

	bq.mu.Lock()
	defer bq.mu.Unlock()
	return len(bq.mu.purgatory)
}

func (bq *baseQueue) SetDisabled(disabled bool) {
	__antithesis_instrumentation__.Notify(112275)
	bq.mu.Lock()
	bq.mu.disabled = disabled
	bq.mu.Unlock()
}

func (bq *baseQueue) lockProcessing() func() {
	__antithesis_instrumentation__.Notify(112276)
	semCount := cap(bq.processSem)

	for i := 0; i < semCount; i++ {
		__antithesis_instrumentation__.Notify(112278)
		bq.processSem <- struct{}{}
	}
	__antithesis_instrumentation__.Notify(112277)

	return func() {
		__antithesis_instrumentation__.Notify(112279)

		for i := 0; i < semCount; i++ {
			__antithesis_instrumentation__.Notify(112280)
			<-bq.processSem
		}
	}
}

func (bq *baseQueue) Start(stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(112281)
	bq.processLoop(stopper)
}

type baseQueueHelper struct {
	bq *baseQueue
}

func (h baseQueueHelper) MaybeAdd(
	ctx context.Context, repl replicaInQueue, now hlc.ClockTimestamp,
) {
	__antithesis_instrumentation__.Notify(112282)
	h.bq.maybeAdd(ctx, repl, now)
}

func (h baseQueueHelper) Add(ctx context.Context, repl replicaInQueue, prio float64) {
	__antithesis_instrumentation__.Notify(112283)
	_, err := h.bq.addInternal(ctx, repl.Desc(), repl.ReplicaID(), prio)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(112284)
		return log.V(1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(112285)
		log.Infof(ctx, "during Add: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(112286)
	}
}

type queueHelper interface {
	MaybeAdd(ctx context.Context, repl replicaInQueue, now hlc.ClockTimestamp)
	Add(ctx context.Context, repl replicaInQueue, prio float64)
}

func (bq *baseQueue) Async(
	ctx context.Context, opName string, wait bool, fn func(ctx context.Context, h queueHelper),
) {
	__antithesis_instrumentation__.Notify(112287)
	if log.V(3) {
		__antithesis_instrumentation__.Notify(112289)
		log.InfofDepth(ctx, 2, "%s", redact.Safe(opName))
	} else {
		__antithesis_instrumentation__.Notify(112290)
	}
	__antithesis_instrumentation__.Notify(112288)
	opName += " (" + bq.name + ")"
	bgCtx := bq.AnnotateCtx(context.Background())
	if err := bq.store.stopper.RunAsyncTaskEx(bgCtx,
		stop.TaskOpts{
			TaskName:   opName,
			Sem:        bq.addOrMaybeAddSem,
			WaitForSem: wait,
		},
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(112291)
			fn(ctx, baseQueueHelper{bq})
		}); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(112292)
		return bq.addLogN.ShouldLog() == true
	}() == true {
		__antithesis_instrumentation__.Notify(112293)
		log.Infof(ctx, "rate limited in %s: %s", redact.Safe(opName), err)
	} else {
		__antithesis_instrumentation__.Notify(112294)
	}
}

func (bq *baseQueue) MaybeAddAsync(
	ctx context.Context, repl replicaInQueue, now hlc.ClockTimestamp,
) {
	__antithesis_instrumentation__.Notify(112295)
	bq.Async(ctx, "MaybeAdd", false, func(ctx context.Context, h queueHelper) {
		__antithesis_instrumentation__.Notify(112296)
		h.MaybeAdd(ctx, repl, now)
	})
}

func (bq *baseQueue) AddAsync(ctx context.Context, repl replicaInQueue, prio float64) {
	__antithesis_instrumentation__.Notify(112297)
	bq.Async(ctx, "Add", false, func(ctx context.Context, h queueHelper) {
		__antithesis_instrumentation__.Notify(112298)
		h.Add(ctx, repl, prio)
	})
}

func (bq *baseQueue) maybeAdd(ctx context.Context, repl replicaInQueue, now hlc.ClockTimestamp) {
	__antithesis_instrumentation__.Notify(112299)
	ctx = repl.AnnotateCtx(ctx)

	var confReader spanconfig.StoreReader
	if bq.needsSystemConfig {
		__antithesis_instrumentation__.Notify(112307)
		var err error
		confReader, err = bq.store.GetConfReader(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(112308)
			if errors.Is(err, errSysCfgUnavailable) && func() bool {
				__antithesis_instrumentation__.Notify(112310)
				return log.V(1) == true
			}() == true {
				__antithesis_instrumentation__.Notify(112311)
				log.Warningf(ctx, "unable to retrieve system config, skipping: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(112312)
			}
			__antithesis_instrumentation__.Notify(112309)
			return
		} else {
			__antithesis_instrumentation__.Notify(112313)
		}
	} else {
		__antithesis_instrumentation__.Notify(112314)
	}
	__antithesis_instrumentation__.Notify(112300)

	bq.mu.Lock()
	stopped := bq.mu.stopped || func() bool {
		__antithesis_instrumentation__.Notify(112315)
		return bq.mu.disabled == true
	}() == true
	bq.mu.Unlock()

	if stopped {
		__antithesis_instrumentation__.Notify(112316)
		return
	} else {
		__antithesis_instrumentation__.Notify(112317)
	}
	__antithesis_instrumentation__.Notify(112301)

	if !repl.IsInitialized() {
		__antithesis_instrumentation__.Notify(112318)
		return
	} else {
		__antithesis_instrumentation__.Notify(112319)
	}
	__antithesis_instrumentation__.Notify(112302)

	if bq.needsRaftInitialized {
		__antithesis_instrumentation__.Notify(112320)
		repl.maybeInitializeRaftGroup(ctx)
	} else {
		__antithesis_instrumentation__.Notify(112321)
	}
	__antithesis_instrumentation__.Notify(112303)

	if !bq.acceptsUnsplitRanges && func() bool {
		__antithesis_instrumentation__.Notify(112322)
		return confReader.NeedsSplit(ctx, repl.Desc().StartKey, repl.Desc().EndKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(112323)

		if log.V(1) {
			__antithesis_instrumentation__.Notify(112325)
			log.Infof(ctx, "split needed; not adding")
		} else {
			__antithesis_instrumentation__.Notify(112326)
		}
		__antithesis_instrumentation__.Notify(112324)
		return
	} else {
		__antithesis_instrumentation__.Notify(112327)
	}
	__antithesis_instrumentation__.Notify(112304)

	if bq.needsLease {
		__antithesis_instrumentation__.Notify(112328)

		st := repl.LeaseStatusAt(ctx, now)
		if st.IsValid() && func() bool {
			__antithesis_instrumentation__.Notify(112329)
			return !st.OwnedBy(repl.StoreID()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(112330)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(112332)
				log.Infof(ctx, "needs lease; not adding: %v", st.Lease)
			} else {
				__antithesis_instrumentation__.Notify(112333)
			}
			__antithesis_instrumentation__.Notify(112331)
			return
		} else {
			__antithesis_instrumentation__.Notify(112334)
		}
	} else {
		__antithesis_instrumentation__.Notify(112335)
	}
	__antithesis_instrumentation__.Notify(112305)

	realRepl, _ := repl.(*Replica)
	should, priority := bq.impl.shouldQueue(ctx, now, realRepl, confReader)
	if !should {
		__antithesis_instrumentation__.Notify(112336)
		return
	} else {
		__antithesis_instrumentation__.Notify(112337)
	}
	__antithesis_instrumentation__.Notify(112306)
	if _, err := bq.addInternal(ctx, repl.Desc(), repl.ReplicaID(), priority); !isExpectedQueueError(err) {
		__antithesis_instrumentation__.Notify(112338)
		log.Errorf(ctx, "unable to add: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(112339)
	}
}

func (bq *baseQueue) addInternal(
	ctx context.Context, desc *roachpb.RangeDescriptor, replicaID roachpb.ReplicaID, priority float64,
) (bool, error) {
	__antithesis_instrumentation__.Notify(112340)

	if !desc.IsInitialized() {
		__antithesis_instrumentation__.Notify(112349)

		return false, errors.New("replica not initialized")
	} else {
		__antithesis_instrumentation__.Notify(112350)
	}
	__antithesis_instrumentation__.Notify(112341)

	bq.mu.Lock()
	defer bq.mu.Unlock()

	if bq.mu.stopped {
		__antithesis_instrumentation__.Notify(112351)
		return false, errQueueStopped
	} else {
		__antithesis_instrumentation__.Notify(112352)
	}
	__antithesis_instrumentation__.Notify(112342)

	if bq.mu.disabled {
		__antithesis_instrumentation__.Notify(112353)
		if log.V(3) {
			__antithesis_instrumentation__.Notify(112355)
			log.Infof(ctx, "queue disabled")
		} else {
			__antithesis_instrumentation__.Notify(112356)
		}
		__antithesis_instrumentation__.Notify(112354)
		return false, errQueueDisabled
	} else {
		__antithesis_instrumentation__.Notify(112357)
	}
	__antithesis_instrumentation__.Notify(112343)

	if _, ok := bq.mu.purgatory[desc.RangeID]; ok {
		__antithesis_instrumentation__.Notify(112358)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(112359)
	}
	__antithesis_instrumentation__.Notify(112344)

	item, ok := bq.mu.replicas[desc.RangeID]
	if ok {
		__antithesis_instrumentation__.Notify(112360)

		if item.processing {
			__antithesis_instrumentation__.Notify(112363)
			wasRequeued := item.requeue
			item.requeue = true
			return !wasRequeued, nil
		} else {
			__antithesis_instrumentation__.Notify(112364)
		}
		__antithesis_instrumentation__.Notify(112361)

		if priority > item.priority {
			__antithesis_instrumentation__.Notify(112365)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(112367)
				log.Infof(ctx, "updating priority: %0.3f -> %0.3f", item.priority, priority)
			} else {
				__antithesis_instrumentation__.Notify(112368)
			}
			__antithesis_instrumentation__.Notify(112366)
			bq.mu.priorityQ.update(item, priority)
		} else {
			__antithesis_instrumentation__.Notify(112369)
		}
		__antithesis_instrumentation__.Notify(112362)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(112370)
	}
	__antithesis_instrumentation__.Notify(112345)

	if log.V(3) {
		__antithesis_instrumentation__.Notify(112371)
		log.Infof(ctx, "adding: priority=%0.3f", priority)
	} else {
		__antithesis_instrumentation__.Notify(112372)
	}
	__antithesis_instrumentation__.Notify(112346)
	item = &replicaItem{rangeID: desc.RangeID, replicaID: replicaID, priority: priority}
	bq.addLocked(item)

	if pqLen := bq.mu.priorityQ.Len(); pqLen > bq.maxSize {
		__antithesis_instrumentation__.Notify(112373)
		bq.removeLocked(bq.mu.priorityQ.sl[pqLen-1])
	} else {
		__antithesis_instrumentation__.Notify(112374)
	}
	__antithesis_instrumentation__.Notify(112347)

	select {
	case bq.incoming <- struct{}{}:
		__antithesis_instrumentation__.Notify(112375)
	default:
		__antithesis_instrumentation__.Notify(112376)

	}
	__antithesis_instrumentation__.Notify(112348)
	return true, nil
}

func (bq *baseQueue) MaybeAddCallback(rangeID roachpb.RangeID, cb processCallback) bool {
	__antithesis_instrumentation__.Notify(112377)
	bq.mu.Lock()
	defer bq.mu.Unlock()

	if purgatoryErr, ok := bq.mu.purgatory[rangeID]; ok {
		__antithesis_instrumentation__.Notify(112380)
		cb(purgatoryErr)
		return true
	} else {
		__antithesis_instrumentation__.Notify(112381)
	}
	__antithesis_instrumentation__.Notify(112378)
	if item, ok := bq.mu.replicas[rangeID]; ok {
		__antithesis_instrumentation__.Notify(112382)
		item.registerCallback(cb)
		return true
	} else {
		__antithesis_instrumentation__.Notify(112383)
	}
	__antithesis_instrumentation__.Notify(112379)
	return false
}

func (bq *baseQueue) MaybeRemove(rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(112384)
	bq.mu.Lock()
	defer bq.mu.Unlock()

	if bq.mu.stopped {
		__antithesis_instrumentation__.Notify(112386)
		return
	} else {
		__antithesis_instrumentation__.Notify(112387)
	}
	__antithesis_instrumentation__.Notify(112385)

	if item, ok := bq.mu.replicas[rangeID]; ok {
		__antithesis_instrumentation__.Notify(112388)
		ctx := bq.AnnotateCtx(context.TODO())
		if log.V(3) {
			__antithesis_instrumentation__.Notify(112390)
			log.Infof(ctx, "%s: removing", item.rangeID)
		} else {
			__antithesis_instrumentation__.Notify(112391)
		}
		__antithesis_instrumentation__.Notify(112389)
		bq.removeLocked(item)
	} else {
		__antithesis_instrumentation__.Notify(112392)
	}
}

func (bq *baseQueue) processLoop(stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(112393)
	ctx := bq.AnnotateCtx(context.Background())
	done := func() {
		__antithesis_instrumentation__.Notify(112395)
		bq.mu.Lock()
		bq.mu.stopped = true
		bq.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(112394)
	if err := stopper.RunAsyncTaskEx(ctx,
		stop.TaskOpts{TaskName: "queue-loop", SpanOpt: stop.SterileRootSpan},
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(112396)
			defer done()

			var nextTime <-chan time.Time

			immediately := make(chan time.Time)
			close(immediately)

			for {
				__antithesis_instrumentation__.Notify(112397)
				select {

				case <-stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(112398)
					return

				case <-bq.incoming:
					__antithesis_instrumentation__.Notify(112399)
					if nextTime == nil {
						__antithesis_instrumentation__.Notify(112402)

						nextTime = immediately

						bq.impl.timer(0)
					} else {
						__antithesis_instrumentation__.Notify(112403)
					}

				case <-nextTime:
					__antithesis_instrumentation__.Notify(112400)

					bq.processSem <- struct{}{}

					repl := bq.pop()
					if repl != nil {
						__antithesis_instrumentation__.Notify(112404)
						annotatedCtx := repl.AnnotateCtx(ctx)
						if stopper.RunAsyncTaskEx(annotatedCtx, stop.TaskOpts{
							TaskName: bq.processOpName() + " [outer]",
						},
							func(ctx context.Context) {
								__antithesis_instrumentation__.Notify(112405)

								defer func() { __antithesis_instrumentation__.Notify(112407); <-bq.processSem }()
								__antithesis_instrumentation__.Notify(112406)

								start := timeutil.Now()
								err := bq.processReplica(ctx, repl)

								duration := timeutil.Since(start)
								bq.recordProcessDuration(ctx, duration)

								bq.finishProcessingReplica(ctx, stopper, repl, err)
							}) != nil {
							__antithesis_instrumentation__.Notify(112408)

							<-bq.processSem
							return
						} else {
							__antithesis_instrumentation__.Notify(112409)
						}
					} else {
						__antithesis_instrumentation__.Notify(112410)

						<-bq.processSem
					}
					__antithesis_instrumentation__.Notify(112401)

					if bq.Length() == 0 {
						__antithesis_instrumentation__.Notify(112411)
						nextTime = nil
					} else {
						__antithesis_instrumentation__.Notify(112412)

						lastDur := bq.lastProcessDuration()
						switch t := bq.impl.timer(lastDur); t {
						case 0:
							__antithesis_instrumentation__.Notify(112413)
							nextTime = immediately
						default:
							__antithesis_instrumentation__.Notify(112414)
							nextTime = time.After(t)
						}
					}
				}
			}
		}); err != nil {
		__antithesis_instrumentation__.Notify(112415)
		done()
	} else {
		__antithesis_instrumentation__.Notify(112416)
	}
}

func (bq *baseQueue) lastProcessDuration() time.Duration {
	__antithesis_instrumentation__.Notify(112417)
	return time.Duration(atomic.LoadInt64(&bq.processDur))
}

func (bq *baseQueue) recordProcessDuration(ctx context.Context, dur time.Duration) {
	__antithesis_instrumentation__.Notify(112418)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(112420)
		log.Infof(ctx, "done %s", dur)
	} else {
		__antithesis_instrumentation__.Notify(112421)
	}
	__antithesis_instrumentation__.Notify(112419)
	bq.processingNanos.Inc(dur.Nanoseconds())
	atomic.StoreInt64(&bq.processDur, int64(dur))
}

func (bq *baseQueue) processReplica(ctx context.Context, repl replicaInQueue) error {
	__antithesis_instrumentation__.Notify(112422)

	var confReader spanconfig.StoreReader
	if bq.needsSystemConfig {
		__antithesis_instrumentation__.Notify(112425)
		var err error
		confReader, err = bq.store.GetConfReader(ctx)
		if errors.Is(err, errSysCfgUnavailable) {
			__antithesis_instrumentation__.Notify(112427)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(112429)
				log.Warningf(ctx, "unable to retrieve conf reader, skipping: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(112430)
			}
			__antithesis_instrumentation__.Notify(112428)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(112431)
		}
		__antithesis_instrumentation__.Notify(112426)
		if err != nil {
			__antithesis_instrumentation__.Notify(112432)
			return err
		} else {
			__antithesis_instrumentation__.Notify(112433)
		}
	} else {
		__antithesis_instrumentation__.Notify(112434)
	}
	__antithesis_instrumentation__.Notify(112423)

	if !bq.acceptsUnsplitRanges && func() bool {
		__antithesis_instrumentation__.Notify(112435)
		return confReader.NeedsSplit(ctx, repl.Desc().StartKey, repl.Desc().EndKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(112436)

		log.VEventf(ctx, 3, "split needed; skipping")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(112437)
	}
	__antithesis_instrumentation__.Notify(112424)

	ctx, span := tracing.EnsureChildSpan(ctx, bq.Tracer, bq.processOpName())
	defer span.Finish()
	return contextutil.RunWithTimeout(ctx, fmt.Sprintf("%s queue process replica %d", bq.name, repl.GetRangeID()),
		bq.processTimeoutFunc(bq.store.ClusterSettings(), repl), func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(112438)
			log.VEventf(ctx, 1, "processing replica")

			if !repl.IsInitialized() {
				__antithesis_instrumentation__.Notify(112444)

				return errors.New("cannot process uninitialized replica")
			} else {
				__antithesis_instrumentation__.Notify(112445)
			}
			__antithesis_instrumentation__.Notify(112439)

			if reason, err := repl.IsDestroyed(); err != nil {
				__antithesis_instrumentation__.Notify(112446)
				if !bq.queueConfig.processDestroyedReplicas || func() bool {
					__antithesis_instrumentation__.Notify(112447)
					return reason == destroyReasonRemoved == true
				}() == true {
					__antithesis_instrumentation__.Notify(112448)
					log.VEventf(ctx, 3, "replica destroyed (%s); skipping", err)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(112449)
				}
			} else {
				__antithesis_instrumentation__.Notify(112450)
			}
			__antithesis_instrumentation__.Notify(112440)

			if bq.needsLease {
				__antithesis_instrumentation__.Notify(112451)
				if _, pErr := repl.redirectOnOrAcquireLease(ctx); pErr != nil {
					__antithesis_instrumentation__.Notify(112452)
					switch v := pErr.GetDetail().(type) {
					case *roachpb.NotLeaseHolderError, *roachpb.RangeNotFoundError:
						__antithesis_instrumentation__.Notify(112453)
						log.VEventf(ctx, 3, "%s; skipping", v)
						return nil
					default:
						__antithesis_instrumentation__.Notify(112454)
						log.VErrEventf(ctx, 2, "could not obtain lease: %s", pErr)
						return errors.Wrapf(pErr.GoError(), "%s: could not obtain lease", repl)
					}
				} else {
					__antithesis_instrumentation__.Notify(112455)
				}
			} else {
				__antithesis_instrumentation__.Notify(112456)
			}
			__antithesis_instrumentation__.Notify(112441)

			log.VEventf(ctx, 3, "processing...")

			realRepl, _ := repl.(*Replica)
			processed, err := bq.impl.process(ctx, realRepl, confReader)
			if err != nil {
				__antithesis_instrumentation__.Notify(112457)
				return err
			} else {
				__antithesis_instrumentation__.Notify(112458)
			}
			__antithesis_instrumentation__.Notify(112442)
			if processed {
				__antithesis_instrumentation__.Notify(112459)
				log.VEventf(ctx, 3, "processing... done")
				bq.successes.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(112460)
			}
			__antithesis_instrumentation__.Notify(112443)
			return nil
		})
}

type benignError struct {
	cause error
}

func (be *benignError) Error() string {
	__antithesis_instrumentation__.Notify(112461)
	return be.cause.Error()
}
func (be *benignError) Cause() error { __antithesis_instrumentation__.Notify(112462); return be.cause }

func isBenign(err error) bool {
	__antithesis_instrumentation__.Notify(112463)
	return errors.HasType(err, (*benignError)(nil))
}

func isPurgatoryError(err error) (purgatoryError, bool) {
	__antithesis_instrumentation__.Notify(112464)
	var purgErr purgatoryError
	return purgErr, errors.As(err, &purgErr)
}

func (bq *baseQueue) assertInvariants() {
	__antithesis_instrumentation__.Notify(112465)
	bq.mu.Lock()
	defer bq.mu.Unlock()

	ctx := bq.AnnotateCtx(context.Background())
	for _, item := range bq.mu.priorityQ.sl {
		__antithesis_instrumentation__.Notify(112469)
		if item.processing {
			__antithesis_instrumentation__.Notify(112472)
			log.Fatalf(ctx, "processing item found in prioQ: %v", item)
		} else {
			__antithesis_instrumentation__.Notify(112473)
		}
		__antithesis_instrumentation__.Notify(112470)
		if _, inReplicas := bq.mu.replicas[item.rangeID]; !inReplicas {
			__antithesis_instrumentation__.Notify(112474)
			log.Fatalf(ctx, "item found in prioQ but not in mu.replicas: %v", item)
		} else {
			__antithesis_instrumentation__.Notify(112475)
		}
		__antithesis_instrumentation__.Notify(112471)
		if _, inPurg := bq.mu.purgatory[item.rangeID]; inPurg {
			__antithesis_instrumentation__.Notify(112476)
			log.Fatalf(ctx, "item found in prioQ and purgatory: %v", item)
		} else {
			__antithesis_instrumentation__.Notify(112477)
		}
	}
	__antithesis_instrumentation__.Notify(112466)
	for rangeID := range bq.mu.purgatory {
		__antithesis_instrumentation__.Notify(112478)
		item, inReplicas := bq.mu.replicas[rangeID]
		if !inReplicas {
			__antithesis_instrumentation__.Notify(112480)
			log.Fatalf(ctx, "item found in purg but not in mu.replicas: %v", item)
		} else {
			__antithesis_instrumentation__.Notify(112481)
		}
		__antithesis_instrumentation__.Notify(112479)
		if item.processing {
			__antithesis_instrumentation__.Notify(112482)
			log.Fatalf(ctx, "processing item found in purgatory: %v", item)
		} else {
			__antithesis_instrumentation__.Notify(112483)
		}

	}
	__antithesis_instrumentation__.Notify(112467)

	var nNotProcessing int
	for _, item := range bq.mu.replicas {
		__antithesis_instrumentation__.Notify(112484)
		if !item.processing {
			__antithesis_instrumentation__.Notify(112485)
			nNotProcessing++
		} else {
			__antithesis_instrumentation__.Notify(112486)
		}
	}
	__antithesis_instrumentation__.Notify(112468)
	if nNotProcessing != len(bq.mu.purgatory)+len(bq.mu.priorityQ.sl) {
		__antithesis_instrumentation__.Notify(112487)
		log.Fatalf(ctx, "have %d non-processing replicas in mu.replicas, "+
			"but %d in purgatory and %d in prioQ; the latter two should add up"+
			"to the former", nNotProcessing, len(bq.mu.purgatory), len(bq.mu.priorityQ.sl))
	} else {
		__antithesis_instrumentation__.Notify(112488)
	}
}

func (bq *baseQueue) finishProcessingReplica(
	ctx context.Context, stopper *stop.Stopper, repl replicaInQueue, err error,
) {
	__antithesis_instrumentation__.Notify(112489)
	bq.mu.Lock()

	item := bq.mu.replicas[repl.GetRangeID()]
	processing := item.processing
	callbacks := item.callbacks
	requeue := item.requeue
	item.callbacks = nil
	bq.removeFromReplicaSetLocked(repl.GetRangeID())
	item = nil
	bq.mu.Unlock()

	if !processing {
		__antithesis_instrumentation__.Notify(112493)
		log.Fatalf(ctx, "%s: attempt to remove non-processing replica %v", bq.name, repl)
	} else {
		__antithesis_instrumentation__.Notify(112494)
	}
	__antithesis_instrumentation__.Notify(112490)

	for _, cb := range callbacks {
		__antithesis_instrumentation__.Notify(112495)
		cb(err)
	}
	__antithesis_instrumentation__.Notify(112491)

	if err != nil {
		__antithesis_instrumentation__.Notify(112496)
		benign := isBenign(err)

		bq.failures.Inc(1)

		if purgErr, ok := isPurgatoryError(err); ok {
			__antithesis_instrumentation__.Notify(112498)
			bq.mu.Lock()
			bq.addToPurgatoryLocked(ctx, stopper, repl, purgErr)
			bq.mu.Unlock()
			return
		} else {
			__antithesis_instrumentation__.Notify(112499)
		}
		__antithesis_instrumentation__.Notify(112497)

		if !benign {
			__antithesis_instrumentation__.Notify(112500)
			log.Errorf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(112501)
		}
	} else {
		__antithesis_instrumentation__.Notify(112502)
	}
	__antithesis_instrumentation__.Notify(112492)

	if requeue {
		__antithesis_instrumentation__.Notify(112503)
		bq.maybeAdd(ctx, repl, bq.store.Clock().NowAsClockTimestamp())
	} else {
		__antithesis_instrumentation__.Notify(112504)
	}
}

func (bq *baseQueue) addToPurgatoryLocked(
	ctx context.Context, stopper *stop.Stopper, repl replicaInQueue, purgErr purgatoryError,
) {
	__antithesis_instrumentation__.Notify(112505)
	bq.mu.AssertHeld()

	if bq.impl.purgatoryChan() == nil {
		__antithesis_instrumentation__.Notify(112511)
		log.Errorf(ctx, "queue does not support purgatory errors, but saw %v", purgErr)
		return
	} else {
		__antithesis_instrumentation__.Notify(112512)
	}
	__antithesis_instrumentation__.Notify(112506)

	if log.V(1) {
		__antithesis_instrumentation__.Notify(112513)
		log.Infof(ctx, "purgatory: %v", purgErr)
	} else {
		__antithesis_instrumentation__.Notify(112514)
	}
	__antithesis_instrumentation__.Notify(112507)

	if _, found := bq.mu.replicas[repl.GetRangeID()]; found {
		__antithesis_instrumentation__.Notify(112515)

		return
	} else {
		__antithesis_instrumentation__.Notify(112516)
	}
	__antithesis_instrumentation__.Notify(112508)

	item := &replicaItem{rangeID: repl.GetRangeID(), replicaID: repl.ReplicaID(), index: -1}
	bq.mu.replicas[repl.GetRangeID()] = item

	defer func() {
		__antithesis_instrumentation__.Notify(112517)
		bq.purgatory.Update(int64(len(bq.mu.purgatory)))
	}()
	__antithesis_instrumentation__.Notify(112509)

	if bq.mu.purgatory != nil {
		__antithesis_instrumentation__.Notify(112518)
		bq.mu.purgatory[repl.GetRangeID()] = purgErr
		return
	} else {
		__antithesis_instrumentation__.Notify(112519)
	}
	__antithesis_instrumentation__.Notify(112510)

	bq.mu.purgatory = map[roachpb.RangeID]purgatoryError{
		repl.GetRangeID(): purgErr,
	}

	workerCtx := bq.AnnotateCtx(context.Background())
	_ = stopper.RunAsyncTaskEx(workerCtx, stop.TaskOpts{TaskName: bq.name + ".purgatory", SpanOpt: stop.SterileRootSpan}, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(112520)
		ticker := time.NewTicker(purgatoryReportInterval)
		for {
			__antithesis_instrumentation__.Notify(112521)
			select {
			case <-bq.impl.purgatoryChan():
				__antithesis_instrumentation__.Notify(112522)
				func() {
					__antithesis_instrumentation__.Notify(112528)

					bq.processSem <- struct{}{}
					defer func() { __antithesis_instrumentation__.Notify(112531); <-bq.processSem }()
					__antithesis_instrumentation__.Notify(112529)

					bq.mu.Lock()
					ranges := make([]*replicaItem, 0, len(bq.mu.purgatory))
					for rangeID := range bq.mu.purgatory {
						__antithesis_instrumentation__.Notify(112532)
						item := bq.mu.replicas[rangeID]
						if item == nil {
							__antithesis_instrumentation__.Notify(112534)
							log.Fatalf(ctx, "r%d is in purgatory but not in replicas", rangeID)
						} else {
							__antithesis_instrumentation__.Notify(112535)
						}
						__antithesis_instrumentation__.Notify(112533)
						item.setProcessing()
						ranges = append(ranges, item)
						bq.removeFromPurgatoryLocked(item)
					}
					__antithesis_instrumentation__.Notify(112530)
					bq.mu.Unlock()

					for _, item := range ranges {
						__antithesis_instrumentation__.Notify(112536)
						repl, err := bq.getReplica(item.rangeID)
						if err != nil || func() bool {
							__antithesis_instrumentation__.Notify(112538)
							return item.replicaID != repl.ReplicaID() == true
						}() == true {
							__antithesis_instrumentation__.Notify(112539)
							continue
						} else {
							__antithesis_instrumentation__.Notify(112540)
						}
						__antithesis_instrumentation__.Notify(112537)
						annotatedCtx := repl.AnnotateCtx(ctx)
						if stopper.RunTask(
							annotatedCtx, bq.processOpName(), func(ctx context.Context) {
								__antithesis_instrumentation__.Notify(112541)
								err := bq.processReplica(ctx, repl)
								bq.finishProcessingReplica(ctx, stopper, repl, err)
							}) != nil {
							__antithesis_instrumentation__.Notify(112542)
							return
						} else {
							__antithesis_instrumentation__.Notify(112543)
						}
					}
				}()
				__antithesis_instrumentation__.Notify(112523)

				bq.mu.Lock()
				if len(bq.mu.purgatory) == 0 {
					__antithesis_instrumentation__.Notify(112544)
					log.Infof(ctx, "purgatory is now empty")
					bq.mu.purgatory = nil
					bq.mu.Unlock()
					return
				} else {
					__antithesis_instrumentation__.Notify(112545)
				}
				__antithesis_instrumentation__.Notify(112524)
				bq.mu.Unlock()
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(112525)

				bq.mu.Lock()
				errMap := map[string]int{}
				for _, err := range bq.mu.purgatory {
					__antithesis_instrumentation__.Notify(112546)
					errMap[err.Error()]++
				}
				__antithesis_instrumentation__.Notify(112526)
				bq.mu.Unlock()
				for errStr, count := range errMap {
					__antithesis_instrumentation__.Notify(112547)
					log.Errorf(ctx, "%d replicas failing with %q", count, errStr)
				}
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(112527)
				return
			}
		}
	})
}

func (bq *baseQueue) pop() replicaInQueue {
	__antithesis_instrumentation__.Notify(112548)
	bq.mu.Lock()
	for {
		__antithesis_instrumentation__.Notify(112549)
		if bq.mu.priorityQ.Len() == 0 {
			__antithesis_instrumentation__.Notify(112553)
			bq.mu.Unlock()
			return nil
		} else {
			__antithesis_instrumentation__.Notify(112554)
		}
		__antithesis_instrumentation__.Notify(112550)
		item := heap.Pop(&bq.mu.priorityQ).(*replicaItem)
		if item.processing {
			__antithesis_instrumentation__.Notify(112555)
			log.Fatalf(bq.AnnotateCtx(context.Background()), "%s pulled processing item from heap: %v", bq.name, item)
		} else {
			__antithesis_instrumentation__.Notify(112556)
		}
		__antithesis_instrumentation__.Notify(112551)
		item.setProcessing()
		bq.pending.Update(int64(bq.mu.priorityQ.Len()))
		bq.mu.Unlock()

		repl, _ := bq.getReplica(item.rangeID)
		if repl != nil && func() bool {
			__antithesis_instrumentation__.Notify(112557)
			return item.replicaID == repl.ReplicaID() == true
		}() == true {
			__antithesis_instrumentation__.Notify(112558)
			return repl
		} else {
			__antithesis_instrumentation__.Notify(112559)
		}
		__antithesis_instrumentation__.Notify(112552)

		bq.mu.Lock()
		bq.removeFromReplicaSetLocked(item.rangeID)
	}
}

func (bq *baseQueue) addLocked(item *replicaItem) {
	__antithesis_instrumentation__.Notify(112560)
	heap.Push(&bq.mu.priorityQ, item)
	bq.pending.Update(int64(bq.mu.priorityQ.Len()))
	bq.mu.replicas[item.rangeID] = item
}

func (bq *baseQueue) removeLocked(item *replicaItem) {
	__antithesis_instrumentation__.Notify(112561)
	if item.processing {
		__antithesis_instrumentation__.Notify(112562)

		item.requeue = false
	} else {
		__antithesis_instrumentation__.Notify(112563)
		if _, inPurg := bq.mu.purgatory[item.rangeID]; inPurg {
			__antithesis_instrumentation__.Notify(112565)
			bq.removeFromPurgatoryLocked(item)
		} else {
			__antithesis_instrumentation__.Notify(112566)
			if item.index >= 0 {
				__antithesis_instrumentation__.Notify(112567)
				bq.removeFromQueueLocked(item)
			} else {
				__antithesis_instrumentation__.Notify(112568)
				log.Fatalf(bq.AnnotateCtx(context.Background()),
					"item for r%d is only in replicas map, but is not processing",
					item.rangeID,
				)
			}
		}
		__antithesis_instrumentation__.Notify(112564)
		bq.removeFromReplicaSetLocked(item.rangeID)
	}
}

func (bq *baseQueue) removeFromPurgatoryLocked(item *replicaItem) {
	__antithesis_instrumentation__.Notify(112569)
	delete(bq.mu.purgatory, item.rangeID)
	bq.purgatory.Update(int64(len(bq.mu.purgatory)))
}

func (bq *baseQueue) removeFromQueueLocked(item *replicaItem) {
	__antithesis_instrumentation__.Notify(112570)
	heap.Remove(&bq.mu.priorityQ, item.index)
	bq.pending.Update(int64(bq.mu.priorityQ.Len()))
}

func (bq *baseQueue) removeFromReplicaSetLocked(rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(112571)
	if _, found := bq.mu.replicas[rangeID]; !found {
		__antithesis_instrumentation__.Notify(112573)
		log.Fatalf(bq.AnnotateCtx(context.Background()),
			"attempted to remove r%d from queue, but it isn't in it",
			rangeID,
		)
	} else {
		__antithesis_instrumentation__.Notify(112574)
	}
	__antithesis_instrumentation__.Notify(112572)
	delete(bq.mu.replicas, rangeID)
}

func (bq *baseQueue) DrainQueue(stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(112575)

	defer bq.lockProcessing()()

	ctx := bq.AnnotateCtx(context.Background())
	for repl := bq.pop(); repl != nil; repl = bq.pop() {
		__antithesis_instrumentation__.Notify(112576)
		annotatedCtx := repl.AnnotateCtx(ctx)
		err := bq.processReplica(annotatedCtx, repl)
		bq.finishProcessingReplica(annotatedCtx, stopper, repl, err)
	}
}

func (bq *baseQueue) processOpName() string {
	__antithesis_instrumentation__.Notify(112577)
	return fmt.Sprintf("queue.%s: processing replica", bq.name)
}
