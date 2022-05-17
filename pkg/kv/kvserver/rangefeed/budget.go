package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

var useBudgets = envutil.EnvOrDefaultBool("COCKROACH_USE_RANGEFEED_MEM_BUDGETS", true)

var totalSharedFeedBudgetFraction = envutil.EnvOrDefaultFloat64("COCKROACH_RANGEFEED_FEED_MEM_FRACTION",
	0.5)

var maxFeedFraction = envutil.EnvOrDefaultFloat64("COCKROACH_RANGEFEED_TOTAL_MEM_FRACTION", 0.05)

var systemRangeFeedBudget = envutil.EnvOrDefaultInt64("COCKROACH_RANGEFEED_SYSTEM_BUDGET",
	2*64*1024*1024)

var RangefeedBudgetsEnabled = settings.RegisterBoolSetting(
	settings.SystemOnly,
	"kv.rangefeed.memory_budgets.enabled",
	"if set, rangefeed memory budgets are enabled",
	true,
)

var budgetAllocationSyncPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(113463)
		return new(SharedBudgetAllocation)
	},
}

func getPooledBudgetAllocation(ba SharedBudgetAllocation) *SharedBudgetAllocation {
	__antithesis_instrumentation__.Notify(113464)
	b := budgetAllocationSyncPool.Get().(*SharedBudgetAllocation)
	*b = ba
	return b
}

func putPooledBudgetAllocation(ba *SharedBudgetAllocation) {
	__antithesis_instrumentation__.Notify(113465)
	*ba = SharedBudgetAllocation{}
	budgetAllocationSyncPool.Put(ba)
}

type FeedBudget struct {
	mu struct {
		syncutil.Mutex

		memBudget *mon.BoundAccount

		closed bool
	}

	limit int64

	replenishC chan interface{}

	stopC chan interface{}

	settings *settings.Values

	closed sync.Once
}

func NewFeedBudget(budget *mon.BoundAccount, limit int64, settings *settings.Values) *FeedBudget {
	__antithesis_instrumentation__.Notify(113466)
	if budget == nil {
		__antithesis_instrumentation__.Notify(113469)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(113470)
	}
	__antithesis_instrumentation__.Notify(113467)

	if limit <= 0 {
		__antithesis_instrumentation__.Notify(113471)
		limit = (1 << 63) - 1
	} else {
		__antithesis_instrumentation__.Notify(113472)
	}
	__antithesis_instrumentation__.Notify(113468)
	f := &FeedBudget{
		replenishC: make(chan interface{}, 1),
		stopC:      make(chan interface{}),
		limit:      limit,
		settings:   settings,
	}
	f.mu.memBudget = budget
	return f
}

func (f *FeedBudget) TryGet(ctx context.Context, amount int64) (*SharedBudgetAllocation, error) {
	__antithesis_instrumentation__.Notify(113473)
	if !RangefeedBudgetsEnabled.Get(f.settings) {
		__antithesis_instrumentation__.Notify(113478)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(113479)
	}
	__antithesis_instrumentation__.Notify(113474)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.mu.closed {
		__antithesis_instrumentation__.Notify(113480)
		log.Info(ctx, "trying to get allocation from already closed budget")
		return nil, errors.Errorf("budget unexpectedly closed")
	} else {
		__antithesis_instrumentation__.Notify(113481)
	}
	__antithesis_instrumentation__.Notify(113475)
	var err error
	if f.mu.memBudget.Used()+amount > f.limit {
		__antithesis_instrumentation__.Notify(113482)
		return nil, errors.Wrap(f.mu.memBudget.Monitor().Resource().NewBudgetExceededError(amount,
			f.mu.memBudget.Used(),
			f.limit), "rangefeed budget")
	} else {
		__antithesis_instrumentation__.Notify(113483)
	}
	__antithesis_instrumentation__.Notify(113476)
	if err = f.mu.memBudget.Grow(ctx, amount); err != nil {
		__antithesis_instrumentation__.Notify(113484)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(113485)
	}
	__antithesis_instrumentation__.Notify(113477)
	return getPooledBudgetAllocation(SharedBudgetAllocation{size: amount, refCount: 1, feed: f}), nil
}

func (f *FeedBudget) WaitAndGet(
	ctx context.Context, amount int64,
) (*SharedBudgetAllocation, error) {
	__antithesis_instrumentation__.Notify(113486)
	for {
		__antithesis_instrumentation__.Notify(113487)
		select {
		case <-f.replenishC:
			__antithesis_instrumentation__.Notify(113488)
			alloc, err := f.TryGet(ctx, amount)
			if err == nil {
				__antithesis_instrumentation__.Notify(113491)
				return alloc, nil
			} else {
				__antithesis_instrumentation__.Notify(113492)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(113489)

			return f.TryGet(ctx, amount)
		case <-f.stopC:
			__antithesis_instrumentation__.Notify(113490)

			return nil, nil
		}
	}
}

func (f *FeedBudget) returnAllocation(ctx context.Context, amount int64) {
	__antithesis_instrumentation__.Notify(113493)
	f.mu.Lock()
	if f.mu.closed {
		__antithesis_instrumentation__.Notify(113496)
		f.mu.Unlock()
		return
	} else {
		__antithesis_instrumentation__.Notify(113497)
	}
	__antithesis_instrumentation__.Notify(113494)
	if amount > 0 {
		__antithesis_instrumentation__.Notify(113498)
		f.mu.memBudget.Shrink(ctx, amount)
	} else {
		__antithesis_instrumentation__.Notify(113499)
	}
	__antithesis_instrumentation__.Notify(113495)
	f.mu.Unlock()
	select {
	case f.replenishC <- struct{}{}:
		__antithesis_instrumentation__.Notify(113500)
	default:
		__antithesis_instrumentation__.Notify(113501)
	}
}

func (f *FeedBudget) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(113502)
	if f == nil {
		__antithesis_instrumentation__.Notify(113504)
		return
	} else {
		__antithesis_instrumentation__.Notify(113505)
	}
	__antithesis_instrumentation__.Notify(113503)
	f.closed.Do(func() {
		__antithesis_instrumentation__.Notify(113506)
		f.mu.Lock()
		f.mu.closed = true
		f.mu.memBudget.Close(ctx)
		close(f.stopC)
		f.mu.Unlock()
	})
}

type SharedBudgetAllocation struct {
	refCount int32
	size     int64
	feed     *FeedBudget
}

func (a *SharedBudgetAllocation) Use() {
	__antithesis_instrumentation__.Notify(113507)
	if a != nil {
		__antithesis_instrumentation__.Notify(113508)
		if atomic.AddInt32(&a.refCount, 1) == 1 {
			__antithesis_instrumentation__.Notify(113509)
			panic("unexpected shared memory allocation usage increase after free")
		} else {
			__antithesis_instrumentation__.Notify(113510)
		}
	} else {
		__antithesis_instrumentation__.Notify(113511)
	}
}

func (a *SharedBudgetAllocation) Release(ctx context.Context) {
	__antithesis_instrumentation__.Notify(113512)
	if a != nil && func() bool {
		__antithesis_instrumentation__.Notify(113513)
		return atomic.AddInt32(&a.refCount, -1) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(113514)
		a.feed.returnAllocation(ctx, a.size)
		putPooledBudgetAllocation(a)
	} else {
		__antithesis_instrumentation__.Notify(113515)
	}
}

type BudgetFactory struct {
	limit              int64
	adjustLimit        func(int64) int64
	feedBytesMon       *mon.BytesMonitor
	systemFeedBytesMon *mon.BytesMonitor

	settings *settings.Values

	metrics *FeedBudgetPoolMetrics
}

type BudgetFactoryConfig struct {
	rootMon                 *mon.BytesMonitor
	provisionalFeedLimit    int64
	adjustLimit             func(int64) int64
	totalRangeReedBudget    int64
	histogramWindowInterval time.Duration
	settings                *settings.Values
}

func (b BudgetFactoryConfig) empty() bool {
	__antithesis_instrumentation__.Notify(113516)
	return b.rootMon == nil
}

func CreateBudgetFactoryConfig(
	rootMon *mon.BytesMonitor,
	memoryPoolSize int64,
	histogramWindowInterval time.Duration,
	adjustLimit func(int64) int64,
	settings *settings.Values,
) BudgetFactoryConfig {
	__antithesis_instrumentation__.Notify(113517)
	if rootMon == nil || func() bool {
		__antithesis_instrumentation__.Notify(113519)
		return !useBudgets == true
	}() == true {
		__antithesis_instrumentation__.Notify(113520)
		return BudgetFactoryConfig{}
	} else {
		__antithesis_instrumentation__.Notify(113521)
	}
	__antithesis_instrumentation__.Notify(113518)
	totalRangeReedBudget := int64(float64(memoryPoolSize) * totalSharedFeedBudgetFraction)
	feedSizeLimit := int64(float64(totalRangeReedBudget) * maxFeedFraction)
	return BudgetFactoryConfig{
		rootMon:                 rootMon,
		provisionalFeedLimit:    feedSizeLimit,
		adjustLimit:             adjustLimit,
		totalRangeReedBudget:    totalRangeReedBudget,
		histogramWindowInterval: histogramWindowInterval,
		settings:                settings,
	}
}

func NewBudgetFactory(ctx context.Context, config BudgetFactoryConfig) *BudgetFactory {
	__antithesis_instrumentation__.Notify(113522)
	if config.empty() {
		__antithesis_instrumentation__.Notify(113524)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(113525)
	}
	__antithesis_instrumentation__.Notify(113523)
	metrics := NewFeedBudgetMetrics(config.histogramWindowInterval)
	systemRangeMonitor := mon.NewMonitorInheritWithLimit("rangefeed-system-monitor",
		systemRangeFeedBudget, config.rootMon)
	systemRangeMonitor.SetMetrics(metrics.SystemBytesCount, nil)
	systemRangeMonitor.Start(ctx, config.rootMon,
		mon.MakeStandaloneBudget(systemRangeFeedBudget))

	rangeFeedPoolMonitor := mon.NewMonitorInheritWithLimit(
		"rangefeed-monitor",
		config.totalRangeReedBudget,
		config.rootMon)
	rangeFeedPoolMonitor.SetMetrics(metrics.SharedBytesCount, nil)
	rangeFeedPoolMonitor.Start(ctx, config.rootMon, mon.BoundAccount{})

	return &BudgetFactory{
		limit:              config.provisionalFeedLimit,
		adjustLimit:        config.adjustLimit,
		feedBytesMon:       rangeFeedPoolMonitor,
		systemFeedBytesMon: systemRangeMonitor,
		settings:           config.settings,
		metrics:            metrics,
	}
}

func (f *BudgetFactory) Stop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(113526)
	if f == nil {
		__antithesis_instrumentation__.Notify(113528)
		return
	} else {
		__antithesis_instrumentation__.Notify(113529)
	}
	__antithesis_instrumentation__.Notify(113527)
	f.systemFeedBytesMon.Stop(ctx)
	f.feedBytesMon.Stop(ctx)
}

func (f *BudgetFactory) CreateBudget(key roachpb.RKey) *FeedBudget {
	__antithesis_instrumentation__.Notify(113530)
	if f == nil {
		__antithesis_instrumentation__.Notify(113534)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(113535)
	}
	__antithesis_instrumentation__.Notify(113531)
	rangeLimit := f.adjustLimit(f.limit)
	if rangeLimit == 0 {
		__antithesis_instrumentation__.Notify(113536)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(113537)
	}
	__antithesis_instrumentation__.Notify(113532)

	if key.Less(roachpb.RKey(keys.SystemSQLCodec.TablePrefix(keys.MaxReservedDescID + 1))) {
		__antithesis_instrumentation__.Notify(113538)
		acc := f.systemFeedBytesMon.MakeBoundAccount()
		return NewFeedBudget(&acc, 0, f.settings)
	} else {
		__antithesis_instrumentation__.Notify(113539)
	}
	__antithesis_instrumentation__.Notify(113533)
	acc := f.feedBytesMon.MakeBoundAccount()
	return NewFeedBudget(&acc, rangeLimit, f.settings)
}

func (f *BudgetFactory) Metrics() *FeedBudgetPoolMetrics {
	__antithesis_instrumentation__.Notify(113540)
	return f.metrics
}
