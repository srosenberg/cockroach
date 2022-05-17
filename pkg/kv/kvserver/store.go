package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/sidetransport"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/idalloc"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/intentresolver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/raftentry"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/tenantrate"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/tscache"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/txnrecovery"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/txnwait"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigstore"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/limit"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/shuffle"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
	raft "go.etcd.io/etcd/raft/v3"
	"golang.org/x/time/rate"
)

const (
	rangeIDAllocCount = 10

	defaultRaftEntryCacheSize = 1 << 24

	replicaRequestQueueSize = 100

	defaultGossipWhenCapacityDeltaExceedsFraction = 0.01

	systemDataGossipInterval = 1 * time.Minute
)

var storeSchedulerConcurrency = envutil.EnvOrDefaultInt(

	"COCKROACH_SCHEDULER_CONCURRENCY", min(8*runtime.GOMAXPROCS(0), 96))

var logSSTInfoTicks = envutil.EnvOrDefaultInt(
	"COCKROACH_LOG_SST_INFO_TICKS_INTERVAL", 60)

var bulkIOWriteLimit = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"kv.bulk_io_write.max_rate",
	"the rate limit (bytes/sec) to use for writes to disk on behalf of bulk io ops",
	1<<40,
).WithPublic()

var addSSTableRequestLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.bulk_io_write.concurrent_addsstable_requests",
	"number of concurrent AddSSTable requests per store before queueing",
	1,
	settings.PositiveInt,
)

var addSSTableAsWritesRequestLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.bulk_io_write.concurrent_addsstable_as_writes_requests",
	"number of concurrent AddSSTable requests ingested as writes per store before queueing",
	10,
	settings.PositiveInt,
)

var concurrentRangefeedItersLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.rangefeed.concurrent_catchup_iterators",
	"number of rangefeeds catchup iterators a store will allow concurrently before queueing",
	16,
	settings.PositiveInt,
)

var concurrentscanInterleavedIntentsLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.migration.concurrent_scan_interleaved_intents",
	"number of scan interleaved intents requests a store will handle concurrently before queueing",
	1,
	settings.PositiveInt,
)

var queueAdditionOnSystemConfigUpdateRate = settings.RegisterFloatSetting(
	settings.TenantWritable,
	"kv.store.system_config_update.queue_add_rate",
	"the rate (per second) at which the store will add, all replicas to the split and merge queue due to system config gossip",
	.5,
	settings.NonNegativeFloat,
)

var queueAdditionOnSystemConfigUpdateBurst = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.store.system_config_update.queue_add_burst",
	"the burst rate at which the store will add all replicas to the split and merge queue due to system config gossip",
	32,
	settings.NonNegativeInt,
)

var leaseTransferWait = func() *settings.DurationSetting {
	__antithesis_instrumentation__.Notify(123821)
	s := settings.RegisterDurationSetting(
		settings.TenantWritable,
		leaseTransferWaitSettingName,
		"the timeout for a single iteration of the range lease transfer phase of draining "+
			"(note that the --drain-wait parameter for cockroach node drain may need adjustment "+
			"after changing this setting)",
		5*time.Second,
		func(v time.Duration) error {
			__antithesis_instrumentation__.Notify(123823)
			if v < 0 {
				__antithesis_instrumentation__.Notify(123825)
				return errors.Errorf("cannot set %s to a negative duration: %s",
					leaseTransferWaitSettingName, v)
			} else {
				__antithesis_instrumentation__.Notify(123826)
			}
			__antithesis_instrumentation__.Notify(123824)
			return nil
		},
	)
	__antithesis_instrumentation__.Notify(123822)
	s.SetVisibility(settings.Public)
	return s
}()

const leaseTransferWaitSettingName = "server.shutdown.lease_transfer_wait"

var ExportRequestsLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.bulk_io_write.concurrent_export_requests",
	"number of export requests a store will handle concurrently before queuing",
	3,
	settings.PositiveInt,
)

func TestStoreConfig(clock *hlc.Clock) StoreConfig {
	__antithesis_instrumentation__.Notify(123827)
	return testStoreConfig(clock, clusterversion.TestingBinaryVersion)
}

func testStoreConfig(clock *hlc.Clock, version roachpb.Version) StoreConfig {
	__antithesis_instrumentation__.Notify(123828)
	if clock == nil {
		__antithesis_instrumentation__.Notify(123830)
		clock = hlc.NewClock(hlc.UnixNano, time.Nanosecond)
	} else {
		__antithesis_instrumentation__.Notify(123831)
	}
	__antithesis_instrumentation__.Notify(123829)
	st := cluster.MakeTestingClusterSettingsWithVersions(version, version, true)
	tracer := tracing.NewTracerWithOpt(context.TODO(), tracing.WithClusterSettings(&st.SV))
	sc := StoreConfig{
		DefaultSpanConfig:           zonepb.DefaultZoneConfigRef().AsSpanConfig(),
		Settings:                    st,
		AmbientCtx:                  log.MakeTestingAmbientContext(tracer),
		Clock:                       clock,
		CoalescedHeartbeatsInterval: 50 * time.Millisecond,
		ScanInterval:                10 * time.Minute,
		HistogramWindowInterval:     metric.TestSampleInterval,
		ProtectedTimestampReader:    spanconfig.EmptyProtectedTSReader(clock),

		SystemConfigProvider: config.NewConstantSystemConfigProvider(
			config.NewSystemConfig(zonepb.DefaultZoneConfigRef()),
		),
	}

	sc.RaftHeartbeatIntervalTicks = 1
	sc.RaftElectionTimeoutTicks = 3
	sc.RaftTickInterval = 100 * time.Millisecond
	sc.SetDefaults()
	return sc
}

func newRaftConfig(
	strg raft.Storage, id uint64, appliedIndex uint64, storeCfg StoreConfig, logger raft.Logger,
) *raft.Config {
	__antithesis_instrumentation__.Notify(123832)
	return &raft.Config{
		ID:                        id,
		Applied:                   appliedIndex,
		ElectionTick:              storeCfg.RaftElectionTimeoutTicks,
		HeartbeatTick:             storeCfg.RaftHeartbeatIntervalTicks,
		MaxUncommittedEntriesSize: storeCfg.RaftMaxUncommittedEntriesSize,
		MaxCommittedSizePerReady:  storeCfg.RaftMaxCommittedSizePerReady,
		MaxSizePerMsg:             storeCfg.RaftMaxSizePerMsg,
		MaxInflightMsgs:           storeCfg.RaftMaxInflightMsgs,
		Storage:                   strg,
		Logger:                    logger,

		PreVote: true,
	}
}

func verifyKeys(start, end roachpb.Key, checkEndKey bool) error {
	__antithesis_instrumentation__.Notify(123833)
	if bytes.Compare(start, roachpb.KeyMax) >= 0 {
		__antithesis_instrumentation__.Notify(123838)
		return errors.Errorf("start key %q must be less than KeyMax", start)
	} else {
		__antithesis_instrumentation__.Notify(123839)
	}
	__antithesis_instrumentation__.Notify(123834)
	if !checkEndKey {
		__antithesis_instrumentation__.Notify(123840)
		if len(end) != 0 {
			__antithesis_instrumentation__.Notify(123842)
			return errors.Errorf("end key %q should not be specified for this operation", end)
		} else {
			__antithesis_instrumentation__.Notify(123843)
		}
		__antithesis_instrumentation__.Notify(123841)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(123844)
	}
	__antithesis_instrumentation__.Notify(123835)
	if end == nil {
		__antithesis_instrumentation__.Notify(123845)
		return errors.Errorf("end key must be specified")
	} else {
		__antithesis_instrumentation__.Notify(123846)
	}
	__antithesis_instrumentation__.Notify(123836)
	if bytes.Compare(roachpb.KeyMax, end) < 0 {
		__antithesis_instrumentation__.Notify(123847)
		return errors.Errorf("end key %q must be less than or equal to KeyMax", end)
	} else {
		__antithesis_instrumentation__.Notify(123848)
	}
	{
		__antithesis_instrumentation__.Notify(123849)
		sAddr, err := keys.Addr(start)
		if err != nil {
			__antithesis_instrumentation__.Notify(123853)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123854)
		}
		__antithesis_instrumentation__.Notify(123850)
		eAddr, err := keys.Addr(end)
		if err != nil {
			__antithesis_instrumentation__.Notify(123855)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123856)
		}
		__antithesis_instrumentation__.Notify(123851)
		if !sAddr.Less(eAddr) {
			__antithesis_instrumentation__.Notify(123857)
			return errors.Errorf("end key %q must be greater than start %q", end, start)
		} else {
			__antithesis_instrumentation__.Notify(123858)
		}
		__antithesis_instrumentation__.Notify(123852)
		if !bytes.Equal(sAddr, start) {
			__antithesis_instrumentation__.Notify(123859)
			if bytes.Equal(eAddr, end) {
				__antithesis_instrumentation__.Notify(123860)
				return errors.Errorf("start key is range-local, but end key is not")
			} else {
				__antithesis_instrumentation__.Notify(123861)
			}
		} else {
			__antithesis_instrumentation__.Notify(123862)
			if bytes.Compare(start, keys.LocalMax) < 0 {
				__antithesis_instrumentation__.Notify(123863)

				return errors.Errorf("start key in [%q,%q) must be greater than LocalMax", start, end)
			} else {
				__antithesis_instrumentation__.Notify(123864)
			}
		}
	}
	__antithesis_instrumentation__.Notify(123837)

	return nil
}

type NotBootstrappedError struct{}

func (e *NotBootstrappedError) Error() string {
	__antithesis_instrumentation__.Notify(123865)
	return "store has not been bootstrapped"
}

type storeReplicaVisitor struct {
	store   *Store
	repls   []*Replica
	ordered bool
	visited int
}

func (rs storeReplicaVisitor) Len() int {
	__antithesis_instrumentation__.Notify(123866)
	return len(rs.repls)
}

func (rs storeReplicaVisitor) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(123867)
	return rs.repls[i].RangeID < rs.repls[j].RangeID
}

func (rs storeReplicaVisitor) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(123868)
	rs.repls[i], rs.repls[j] = rs.repls[j], rs.repls[i]
}

func newStoreReplicaVisitor(store *Store) *storeReplicaVisitor {
	__antithesis_instrumentation__.Notify(123869)
	return &storeReplicaVisitor{
		store:   store,
		visited: -1,
	}
}

func (rs *storeReplicaVisitor) InOrder() *storeReplicaVisitor {
	__antithesis_instrumentation__.Notify(123870)
	rs.ordered = true
	return rs
}

func (rs *storeReplicaVisitor) Visit(visitor func(*Replica) bool) {
	__antithesis_instrumentation__.Notify(123871)

	rs.repls = nil
	rs.store.mu.replicasByRangeID.Range(func(repl *Replica) {
		__antithesis_instrumentation__.Notify(123875)
		rs.repls = append(rs.repls, repl)
	})
	__antithesis_instrumentation__.Notify(123872)

	if rs.ordered {
		__antithesis_instrumentation__.Notify(123876)

		sort.Sort(rs)
	} else {
		__antithesis_instrumentation__.Notify(123877)

		shuffle.Shuffle(rs)
	}
	__antithesis_instrumentation__.Notify(123873)

	rs.visited = 0
	for _, repl := range rs.repls {
		__antithesis_instrumentation__.Notify(123878)

		rs.visited++
		repl.mu.RLock()
		destroyed := repl.mu.destroyStatus
		initialized := repl.isInitializedRLocked()
		repl.mu.RUnlock()
		if initialized && func() bool {
			__antithesis_instrumentation__.Notify(123879)
			return destroyed.IsAlive() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(123880)
			return !visitor(repl) == true
		}() == true {
			__antithesis_instrumentation__.Notify(123881)
			break
		} else {
			__antithesis_instrumentation__.Notify(123882)
		}
	}
	__antithesis_instrumentation__.Notify(123874)
	rs.visited = 0
}

func (rs *storeReplicaVisitor) EstimatedCount() int {
	__antithesis_instrumentation__.Notify(123883)
	if rs.visited <= 0 {
		__antithesis_instrumentation__.Notify(123885)
		return rs.store.ReplicaCount()
	} else {
		__antithesis_instrumentation__.Notify(123886)
	}
	__antithesis_instrumentation__.Notify(123884)
	return len(rs.repls) - rs.visited
}

type Store struct {
	Ident           *roachpb.StoreIdent
	cfg             StoreConfig
	db              *kv.DB
	engine          storage.Engine
	tsCache         tscache.Cache
	allocator       Allocator
	replRankings    *replicaRankings
	storeRebalancer *StoreRebalancer
	rangeIDAlloc    *idalloc.Allocator
	mvccGCQueue     *mvccGCQueue
	mergeQueue      *mergeQueue
	splitQueue      *splitQueue
	replicateQueue  *replicateQueue
	replicaGCQueue  *replicaGCQueue
	raftLogQueue    *raftLogQueue

	raftTruncator      *raftLogTruncator
	raftSnapshotQueue  *raftSnapshotQueue
	tsMaintenanceQueue *timeSeriesMaintenanceQueue
	scanner            *replicaScanner
	consistencyQueue   *consistencyQueue
	consistencyLimiter *quotapool.RateLimiter
	consistencySem     *quotapool.IntPool
	metrics            *StoreMetrics
	intentResolver     *intentresolver.IntentResolver
	recoveryMgr        txnrecovery.Manager
	raftEntryCache     *raftentry.Cache
	limiters           batcheval.Limiters
	txnWaitMetrics     *txnwait.Metrics
	sstSnapshotStorage SSTSnapshotStorage
	protectedtsReader  spanconfig.ProtectedTSReader
	ctSender           *sidetransport.Sender

	gossipRangeCountdown int32
	gossipLeaseCountdown int32

	gossipQueriesPerSecondVal syncutil.AtomicFloat64
	gossipWritesPerSecondVal  syncutil.AtomicFloat64

	coalescedMu struct {
		syncutil.Mutex
		heartbeats         map[roachpb.StoreIdent][]kvserverpb.RaftHeartbeat
		heartbeatResponses map[roachpb.StoreIdent][]kvserverpb.RaftHeartbeat
	}

	started int32
	stopper *stop.Stopper

	startedAt    int64
	nodeDesc     *roachpb.NodeDescriptor
	initComplete sync.WaitGroup

	snapshotApplySem chan struct{}

	renewableLeases       syncutil.IntMap
	renewableLeasesSignal chan struct{}

	draining atomic.Value

	mu struct {
		syncutil.RWMutex

		replicasByRangeID rangeIDReplicaMap

		replicasByKey *storeReplicaBTree

		uninitReplicas map[roachpb.RangeID]*Replica

		replicaPlaceholders map[roachpb.RangeID]*ReplicaPlaceholder
	}

	unquiescedReplicas struct {
		syncutil.Mutex
		m map[roachpb.RangeID]struct{}
	}

	rangefeedReplicas struct {
		syncutil.Mutex
		m map[roachpb.RangeID]struct{}
	}

	replicaQueues syncutil.IntMap

	scheduler *raftScheduler

	livenessMap atomic.Value

	cachedCapacity struct {
		syncutil.Mutex
		roachpb.StoreCapacity
	}

	counts struct {
		failedPlaceholders int32

		filledPlaceholders int32

		droppedPlaceholders int32
	}

	tenantRateLimiters *tenantrate.LimiterFactory

	computeInitialMetrics              sync.Once
	systemConfigUpdateQueueRateLimiter *quotapool.RateLimiter
	spanConfigUpdateQueueRateLimiter   *quotapool.RateLimiter
}

var _ kv.Sender = &Store{}

type StoreConfig struct {
	AmbientCtx log.AmbientContext
	base.RaftConfig

	DefaultSpanConfig    roachpb.SpanConfig
	Settings             *cluster.Settings
	Clock                *hlc.Clock
	DB                   *kv.DB
	Gossip               *gossip.Gossip
	NodeLiveness         *liveness.NodeLiveness
	StorePool            *StorePool
	Transport            *RaftTransport
	NodeDialer           *nodedialer.Dialer
	RPCContext           *rpc.Context
	RangeDescriptorCache *rangecache.RangeCache

	ClosedTimestampSender   *sidetransport.Sender
	ClosedTimestampReceiver sidetransportReceiver

	SQLExecutor sqlutil.InternalExecutor

	TimeSeriesDataStore TimeSeriesDataStore

	CoalescedHeartbeatsInterval time.Duration

	ScanInterval time.Duration

	ScanMinIdleTime time.Duration

	ScanMaxIdleTime time.Duration

	LogRangeEvents bool

	RaftEntryCacheSize uint64

	IntentResolverTaskLimit int

	TestingKnobs StoreTestingKnobs

	concurrentSnapshotApplyLimit int

	HistogramWindowInterval time.Duration

	ExternalStorage        cloud.ExternalStorageFactory
	ExternalStorageFromURI cloud.ExternalStorageFromURIFactory

	ProtectedTimestampReader spanconfig.ProtectedTSReader

	KVMemoryMonitor        *mon.BytesMonitor
	RangefeedBudgetFactory *rangefeed.BudgetFactory

	SpanConfigsDisabled bool

	SpanConfigSubscriber spanconfig.KVSubscriber

	KVAdmissionController KVAdmissionController

	SystemConfigProvider config.SystemConfigProvider
}

type ConsistencyTestingKnobs struct {
	OnBadChecksumFatal func(roachpb.StoreIdent)

	BadChecksumReportDiff      func(roachpb.StoreIdent, ReplicaSnapshotDiffSlice)
	ConsistencyQueueResultHook func(response roachpb.CheckConsistencyResponse)
}

func (sc *StoreConfig) Valid() bool {
	__antithesis_instrumentation__.Notify(123887)
	return sc.Clock != nil && func() bool {
		__antithesis_instrumentation__.Notify(123888)
		return sc.Transport != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(123889)
		return sc.RaftTickInterval != 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(123890)
		return sc.RaftHeartbeatIntervalTicks > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(123891)
		return sc.RaftElectionTimeoutTicks > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(123892)
		return sc.ScanInterval >= 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(123893)
		return sc.AmbientCtx.Tracer != nil == true
	}() == true
}

func (sc *StoreConfig) SetDefaults() {
	__antithesis_instrumentation__.Notify(123894)
	sc.RaftConfig.SetDefaults()

	if sc.CoalescedHeartbeatsInterval == 0 {
		__antithesis_instrumentation__.Notify(123898)
		sc.CoalescedHeartbeatsInterval = sc.RaftTickInterval / 2
	} else {
		__antithesis_instrumentation__.Notify(123899)
	}
	__antithesis_instrumentation__.Notify(123895)
	if sc.RaftEntryCacheSize == 0 {
		__antithesis_instrumentation__.Notify(123900)
		sc.RaftEntryCacheSize = defaultRaftEntryCacheSize
	} else {
		__antithesis_instrumentation__.Notify(123901)
	}
	__antithesis_instrumentation__.Notify(123896)
	if sc.concurrentSnapshotApplyLimit == 0 {
		__antithesis_instrumentation__.Notify(123902)

		sc.concurrentSnapshotApplyLimit =
			envutil.EnvOrDefaultInt("COCKROACH_CONCURRENT_SNAPSHOT_APPLY_LIMIT", 1)
	} else {
		__antithesis_instrumentation__.Notify(123903)
	}
	__antithesis_instrumentation__.Notify(123897)

	if sc.TestingKnobs.GossipWhenCapacityDeltaExceedsFraction == 0 {
		__antithesis_instrumentation__.Notify(123904)
		sc.TestingKnobs.GossipWhenCapacityDeltaExceedsFraction = defaultGossipWhenCapacityDeltaExceedsFraction
	} else {
		__antithesis_instrumentation__.Notify(123905)
	}
}

func (s *Store) GetStoreConfig() *StoreConfig {
	__antithesis_instrumentation__.Notify(123906)
	return &s.cfg
}

func (sc *StoreConfig) LeaseExpiration() int64 {
	__antithesis_instrumentation__.Notify(123907)

	maxOffset := sc.Clock.MaxOffset()
	return 2 * (sc.RangeLeaseActiveDuration() + maxOffset).Nanoseconds()
}

func NewStore(
	ctx context.Context, cfg StoreConfig, eng storage.Engine, nodeDesc *roachpb.NodeDescriptor,
) *Store {
	__antithesis_instrumentation__.Notify(123908)

	cfg.SetDefaults()

	if !cfg.Valid() {
		__antithesis_instrumentation__.Notify(123932)
		log.Fatalf(ctx, "invalid store configuration: %+v", &cfg)
	} else {
		__antithesis_instrumentation__.Notify(123933)
	}
	__antithesis_instrumentation__.Notify(123909)
	s := &Store{
		cfg:      cfg,
		db:       cfg.DB,
		engine:   eng,
		nodeDesc: nodeDesc,
		metrics:  newStoreMetrics(cfg.HistogramWindowInterval),
		ctSender: cfg.ClosedTimestampSender,
	}
	if cfg.RPCContext != nil {
		__antithesis_instrumentation__.Notify(123934)
		s.allocator = MakeAllocator(
			cfg.StorePool,
			cfg.RPCContext.RemoteClocks.Latency,
			cfg.TestingKnobs.AllocatorKnobs,
			s.metrics,
		)
	} else {
		__antithesis_instrumentation__.Notify(123935)
		s.allocator = MakeAllocator(
			cfg.StorePool, func(string) (time.Duration, bool) {
				__antithesis_instrumentation__.Notify(123936)
				return 0, false
			}, cfg.TestingKnobs.AllocatorKnobs, s.metrics,
		)
	}
	__antithesis_instrumentation__.Notify(123910)
	s.replRankings = newReplicaRankings()

	s.draining.Store(false)
	s.scheduler = newRaftScheduler(cfg.AmbientCtx, s.metrics, s, storeSchedulerConcurrency)

	s.raftEntryCache = raftentry.NewCache(cfg.RaftEntryCacheSize)
	s.metrics.registry.AddMetricStruct(s.raftEntryCache.Metrics())

	s.coalescedMu.Lock()
	s.coalescedMu.heartbeats = map[roachpb.StoreIdent][]kvserverpb.RaftHeartbeat{}
	s.coalescedMu.heartbeatResponses = map[roachpb.StoreIdent][]kvserverpb.RaftHeartbeat{}
	s.coalescedMu.Unlock()

	s.mu.Lock()
	s.mu.replicaPlaceholders = map[roachpb.RangeID]*ReplicaPlaceholder{}
	s.mu.replicasByKey = newStoreReplicaBTree()
	s.mu.uninitReplicas = map[roachpb.RangeID]*Replica{}
	s.mu.Unlock()

	s.unquiescedReplicas.Lock()
	s.unquiescedReplicas.m = map[roachpb.RangeID]struct{}{}
	s.unquiescedReplicas.Unlock()

	s.rangefeedReplicas.Lock()
	s.rangefeedReplicas.m = map[roachpb.RangeID]struct{}{}
	s.rangefeedReplicas.Unlock()

	s.tsCache = tscache.New(cfg.Clock)
	s.metrics.registry.AddMetricStruct(s.tsCache.Metrics())

	s.txnWaitMetrics = txnwait.NewMetrics(cfg.HistogramWindowInterval)
	s.metrics.registry.AddMetricStruct(s.txnWaitMetrics)
	s.snapshotApplySem = make(chan struct{}, cfg.concurrentSnapshotApplyLimit)

	if ch := s.cfg.TestingKnobs.LeaseRenewalSignalChan; ch != nil {
		__antithesis_instrumentation__.Notify(123937)
		s.renewableLeasesSignal = ch
	} else {
		__antithesis_instrumentation__.Notify(123938)
		s.renewableLeasesSignal = make(chan struct{}, 1)
	}
	__antithesis_instrumentation__.Notify(123911)

	s.limiters.BulkIOWriteRate = rate.NewLimiter(rate.Limit(bulkIOWriteLimit.Get(&cfg.Settings.SV)), bulkIOWriteBurst)
	bulkIOWriteLimit.SetOnChange(&cfg.Settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(123939)
		s.limiters.BulkIOWriteRate.SetLimit(rate.Limit(bulkIOWriteLimit.Get(&cfg.Settings.SV)))
	})
	__antithesis_instrumentation__.Notify(123912)
	s.limiters.ConcurrentExportRequests = limit.MakeConcurrentRequestLimiter(
		"exportRequestLimiter", int(ExportRequestsLimit.Get(&cfg.Settings.SV)),
	)
	s.limiters.ConcurrentScanInterleavedIntents = limit.MakeConcurrentRequestLimiter(
		"scanInterleavedIntentsLimiter", int(concurrentscanInterleavedIntentsLimit.Get(&cfg.Settings.SV)))
	concurrentscanInterleavedIntentsLimit.SetOnChange(&cfg.Settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(123940)
		s.limiters.ConcurrentScanInterleavedIntents.SetLimit(int(concurrentscanInterleavedIntentsLimit.Get(&cfg.Settings.SV)))
	})
	__antithesis_instrumentation__.Notify(123913)

	s.sstSnapshotStorage = NewSSTSnapshotStorage(s.engine, s.limiters.BulkIOWriteRate)
	if err := s.sstSnapshotStorage.Clear(); err != nil {
		__antithesis_instrumentation__.Notify(123941)
		log.Warningf(ctx, "failed to clear snapshot storage: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(123942)
	}
	__antithesis_instrumentation__.Notify(123914)
	s.protectedtsReader = cfg.ProtectedTimestampReader

	exportCores := runtime.GOMAXPROCS(0) - 1
	if exportCores < 1 {
		__antithesis_instrumentation__.Notify(123943)
		exportCores = 1
	} else {
		__antithesis_instrumentation__.Notify(123944)
	}
	__antithesis_instrumentation__.Notify(123915)
	ExportRequestsLimit.SetOnChange(&cfg.Settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(123945)
		limit := int(ExportRequestsLimit.Get(&cfg.Settings.SV))
		if limit > exportCores {
			__antithesis_instrumentation__.Notify(123947)
			limit = exportCores
		} else {
			__antithesis_instrumentation__.Notify(123948)
		}
		__antithesis_instrumentation__.Notify(123946)
		s.limiters.ConcurrentExportRequests.SetLimit(limit)
	})
	__antithesis_instrumentation__.Notify(123916)
	s.limiters.ConcurrentAddSSTableRequests = limit.MakeConcurrentRequestLimiter(
		"addSSTableRequestLimiter", int(addSSTableRequestLimit.Get(&cfg.Settings.SV)),
	)
	addSSTableRequestLimit.SetOnChange(&cfg.Settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(123949)
		s.limiters.ConcurrentAddSSTableRequests.SetLimit(
			int(addSSTableRequestLimit.Get(&cfg.Settings.SV)))
	})
	__antithesis_instrumentation__.Notify(123917)
	s.limiters.ConcurrentAddSSTableAsWritesRequests = limit.MakeConcurrentRequestLimiter(
		"addSSTableAsWritesRequestLimiter", int(addSSTableAsWritesRequestLimit.Get(&cfg.Settings.SV)),
	)
	addSSTableAsWritesRequestLimit.SetOnChange(&cfg.Settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(123950)
		s.limiters.ConcurrentAddSSTableAsWritesRequests.SetLimit(
			int(addSSTableAsWritesRequestLimit.Get(&cfg.Settings.SV)))
	})
	__antithesis_instrumentation__.Notify(123918)
	s.limiters.ConcurrentRangefeedIters = limit.MakeConcurrentRequestLimiter(
		"rangefeedIterLimiter", int(concurrentRangefeedItersLimit.Get(&cfg.Settings.SV)),
	)
	concurrentRangefeedItersLimit.SetOnChange(&cfg.Settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(123951)
		s.limiters.ConcurrentRangefeedIters.SetLimit(
			int(concurrentRangefeedItersLimit.Get(&cfg.Settings.SV)))
	})
	__antithesis_instrumentation__.Notify(123919)

	s.tenantRateLimiters = tenantrate.NewLimiterFactory(&cfg.Settings.SV, &cfg.TestingKnobs.TenantRateKnobs)
	s.metrics.registry.AddMetricStruct(s.tenantRateLimiters.Metrics())

	s.systemConfigUpdateQueueRateLimiter = quotapool.NewRateLimiter(
		"SystemConfigUpdateQueue",
		quotapool.Limit(queueAdditionOnSystemConfigUpdateRate.Get(&cfg.Settings.SV)),
		queueAdditionOnSystemConfigUpdateBurst.Get(&cfg.Settings.SV))
	updateSystemConfigUpdateQueueLimits := func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(123952)
		s.systemConfigUpdateQueueRateLimiter.UpdateLimit(
			quotapool.Limit(queueAdditionOnSystemConfigUpdateRate.Get(&cfg.Settings.SV)),
			queueAdditionOnSystemConfigUpdateBurst.Get(&cfg.Settings.SV))
	}
	__antithesis_instrumentation__.Notify(123920)
	queueAdditionOnSystemConfigUpdateRate.SetOnChange(&cfg.Settings.SV,
		updateSystemConfigUpdateQueueLimits)
	queueAdditionOnSystemConfigUpdateBurst.SetOnChange(&cfg.Settings.SV,
		updateSystemConfigUpdateQueueLimits)

	if s.cfg.Gossip != nil {
		__antithesis_instrumentation__.Notify(123953)

		s.scanner = newReplicaScanner(
			s.cfg.AmbientCtx, s.cfg.Clock, cfg.ScanInterval,
			cfg.ScanMinIdleTime, cfg.ScanMaxIdleTime, newStoreReplicaVisitor(s),
		)
		s.mvccGCQueue = newMVCCGCQueue(s)
		s.mergeQueue = newMergeQueue(s, s.db)
		s.splitQueue = newSplitQueue(s, s.db)
		s.replicateQueue = newReplicateQueue(s, s.allocator)
		s.replicaGCQueue = newReplicaGCQueue(s, s.db)
		s.raftLogQueue = newRaftLogQueue(s, s.db)
		s.raftSnapshotQueue = newRaftSnapshotQueue(s)
		s.consistencyQueue = newConsistencyQueue(s)

		s.scanner.AddQueues(
			s.mvccGCQueue, s.mergeQueue, s.splitQueue, s.replicateQueue, s.replicaGCQueue,
			s.raftLogQueue, s.raftSnapshotQueue, s.consistencyQueue)
		tsDS := s.cfg.TimeSeriesDataStore
		if s.cfg.TestingKnobs.TimeSeriesDataStore != nil {
			__antithesis_instrumentation__.Notify(123955)
			tsDS = s.cfg.TestingKnobs.TimeSeriesDataStore
		} else {
			__antithesis_instrumentation__.Notify(123956)
		}
		__antithesis_instrumentation__.Notify(123954)
		if tsDS != nil {
			__antithesis_instrumentation__.Notify(123957)
			s.tsMaintenanceQueue = newTimeSeriesMaintenanceQueue(
				s, s.db, tsDS,
			)
			s.scanner.AddQueues(s.tsMaintenanceQueue)
		} else {
			__antithesis_instrumentation__.Notify(123958)
		}
	} else {
		__antithesis_instrumentation__.Notify(123959)
	}
	__antithesis_instrumentation__.Notify(123921)

	if cfg.TestingKnobs.DisableGCQueue {
		__antithesis_instrumentation__.Notify(123960)
		s.setGCQueueActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123961)
	}
	__antithesis_instrumentation__.Notify(123922)
	if cfg.TestingKnobs.DisableMergeQueue {
		__antithesis_instrumentation__.Notify(123962)
		s.setMergeQueueActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123963)
	}
	__antithesis_instrumentation__.Notify(123923)
	if cfg.TestingKnobs.DisableRaftLogQueue {
		__antithesis_instrumentation__.Notify(123964)
		s.setRaftLogQueueActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123965)
	}
	__antithesis_instrumentation__.Notify(123924)
	if cfg.TestingKnobs.DisableReplicaGCQueue {
		__antithesis_instrumentation__.Notify(123966)
		s.setReplicaGCQueueActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123967)
	}
	__antithesis_instrumentation__.Notify(123925)
	if cfg.TestingKnobs.DisableReplicateQueue {
		__antithesis_instrumentation__.Notify(123968)
		s.SetReplicateQueueActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123969)
	}
	__antithesis_instrumentation__.Notify(123926)
	if cfg.TestingKnobs.DisableSplitQueue {
		__antithesis_instrumentation__.Notify(123970)
		s.setSplitQueueActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123971)
	}
	__antithesis_instrumentation__.Notify(123927)
	if cfg.TestingKnobs.DisableTimeSeriesMaintenanceQueue {
		__antithesis_instrumentation__.Notify(123972)
		s.setTimeSeriesMaintenanceQueueActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123973)
	}
	__antithesis_instrumentation__.Notify(123928)
	if cfg.TestingKnobs.DisableRaftSnapshotQueue {
		__antithesis_instrumentation__.Notify(123974)
		s.setRaftSnapshotQueueActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123975)
	}
	__antithesis_instrumentation__.Notify(123929)
	if cfg.TestingKnobs.DisableConsistencyQueue {
		__antithesis_instrumentation__.Notify(123976)
		s.setConsistencyQueueActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123977)
	}
	__antithesis_instrumentation__.Notify(123930)
	if cfg.TestingKnobs.DisableScanner {
		__antithesis_instrumentation__.Notify(123978)
		s.setScannerActive(false)
	} else {
		__antithesis_instrumentation__.Notify(123979)
	}
	__antithesis_instrumentation__.Notify(123931)

	return s
}

func (s *Store) String() string {
	__antithesis_instrumentation__.Notify(123980)
	return redact.StringWithoutMarkers(s)
}

func (s *Store) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(123981)
	w.Printf("[n%d,s%d]", s.Ident.NodeID, s.Ident.StoreID)
}

func (s *Store) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(123982)
	return s.cfg.Settings
}

func (s *Store) AnnotateCtx(ctx context.Context) context.Context {
	__antithesis_instrumentation__.Notify(123983)
	return s.cfg.AmbientCtx.AnnotateCtx(ctx)
}

func (s *Store) SetDraining(drain bool, reporter func(int, redact.SafeString), verbose bool) {
	__antithesis_instrumentation__.Notify(123984)
	s.draining.Store(drain)
	if !drain {
		__antithesis_instrumentation__.Notify(123988)
		return
	} else {
		__antithesis_instrumentation__.Notify(123989)
	}
	__antithesis_instrumentation__.Notify(123985)

	baseCtx := logtags.AddTag(context.Background(), "drain", nil)

	ctx, cancelFn := s.stopper.WithCancelOnQuiesce(baseCtx)
	defer cancelFn()

	var wg sync.WaitGroup

	transferAllAway := func(transferCtx context.Context) int {
		__antithesis_instrumentation__.Notify(123990)

		const leaseTransferConcurrency = 100
		sem := quotapool.NewIntPool("Store.SetDraining", leaseTransferConcurrency)

		var numTransfersAttempted int32
		newStoreReplicaVisitor(s).Visit(func(r *Replica) bool {
			__antithesis_instrumentation__.Notify(123992)

			atomic.AddInt32(&numTransfersAttempted, 1)
			wg.Add(1)
			if err := s.stopper.RunAsyncTaskEx(
				r.AnnotateCtx(ctx),
				stop.TaskOpts{
					TaskName:   "storage.Store: draining replica",
					Sem:        sem,
					WaitForSem: true,
				},
				func(ctx context.Context) {
					__antithesis_instrumentation__.Notify(123994)
					defer wg.Done()

					select {
					case <-transferCtx.Done():
						__antithesis_instrumentation__.Notify(123999)

						if verbose || func() bool {
							__antithesis_instrumentation__.Notify(124002)
							return log.V(1) == true
						}() == true {
							__antithesis_instrumentation__.Notify(124003)
							log.Infof(ctx, "lease transfer aborted due to exceeded timeout")
						} else {
							__antithesis_instrumentation__.Notify(124004)
						}
						__antithesis_instrumentation__.Notify(124000)
						return
					default:
						__antithesis_instrumentation__.Notify(124001)
					}
					__antithesis_instrumentation__.Notify(123995)

					now := s.Clock().NowAsClockTimestamp()
					var drainingLeaseStatus kvserverpb.LeaseStatus
					for {
						__antithesis_instrumentation__.Notify(124005)
						var llHandle *leaseRequestHandle
						r.mu.Lock()
						drainingLeaseStatus = r.leaseStatusAtRLocked(ctx, now)
						_, nextLease := r.getLeaseRLocked()
						if nextLease != (roachpb.Lease{}) && func() bool {
							__antithesis_instrumentation__.Notify(124008)
							return nextLease.OwnedBy(s.StoreID()) == true
						}() == true {
							__antithesis_instrumentation__.Notify(124009)
							llHandle = r.mu.pendingLeaseRequest.JoinRequest()
						} else {
							__antithesis_instrumentation__.Notify(124010)
						}
						__antithesis_instrumentation__.Notify(124006)
						r.mu.Unlock()

						if llHandle != nil {
							__antithesis_instrumentation__.Notify(124011)
							<-llHandle.C()
							continue
						} else {
							__antithesis_instrumentation__.Notify(124012)
						}
						__antithesis_instrumentation__.Notify(124007)
						break
					}
					__antithesis_instrumentation__.Notify(123996)

					needsLeaseTransfer := len(r.Desc().Replicas().VoterDescriptors()) > 1 && func() bool {
						__antithesis_instrumentation__.Notify(124013)
						return drainingLeaseStatus.IsValid() == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(124014)
						return drainingLeaseStatus.OwnedBy(s.StoreID()) == true
					}() == true

					if !needsLeaseTransfer {
						__antithesis_instrumentation__.Notify(124015)
						if verbose || func() bool {
							__antithesis_instrumentation__.Notify(124017)
							return log.V(1) == true
						}() == true {
							__antithesis_instrumentation__.Notify(124018)

							log.Info(ctx, "not moving out")
						} else {
							__antithesis_instrumentation__.Notify(124019)
						}
						__antithesis_instrumentation__.Notify(124016)
						atomic.AddInt32(&numTransfersAttempted, -1)
						return
					} else {
						__antithesis_instrumentation__.Notify(124020)
					}
					__antithesis_instrumentation__.Notify(123997)

					desc, conf := r.DescAndSpanConfig()

					if verbose || func() bool {
						__antithesis_instrumentation__.Notify(124021)
						return log.V(1) == true
					}() == true {
						__antithesis_instrumentation__.Notify(124022)

						log.Infof(ctx, "attempting to transfer lease %v for range %s", drainingLeaseStatus.Lease, desc)
					} else {
						__antithesis_instrumentation__.Notify(124023)
					}
					__antithesis_instrumentation__.Notify(123998)

					start := timeutil.Now()
					transferStatus, err := s.replicateQueue.shedLease(
						ctx,
						r,
						desc,
						conf,
						transferLeaseOptions{},
					)
					duration := timeutil.Since(start).Microseconds()

					if transferStatus != transferOK {
						__antithesis_instrumentation__.Notify(124024)
						const failFormat = "failed to transfer lease %s for range %s when draining: %v"
						const durationFailFormat = "blocked for %d microseconds on transfer attempt"

						infoArgs := []interface{}{
							drainingLeaseStatus.Lease,
							desc,
						}
						if err != nil {
							__antithesis_instrumentation__.Notify(124026)
							infoArgs = append(infoArgs, err)
						} else {
							__antithesis_instrumentation__.Notify(124027)
							infoArgs = append(infoArgs, transferStatus)
						}
						__antithesis_instrumentation__.Notify(124025)

						if verbose {
							__antithesis_instrumentation__.Notify(124028)
							log.Dev.Infof(ctx, failFormat, infoArgs...)
							log.Dev.Infof(ctx, durationFailFormat, duration)
						} else {
							__antithesis_instrumentation__.Notify(124029)
							log.VErrEventf(ctx, 1, failFormat, infoArgs...)
							log.VErrEventf(ctx, 1, durationFailFormat, duration)
						}
					} else {
						__antithesis_instrumentation__.Notify(124030)
					}
				}); err != nil {
				__antithesis_instrumentation__.Notify(124031)
				if verbose || func() bool {
					__antithesis_instrumentation__.Notify(124033)
					return log.V(1) == true
				}() == true {
					__antithesis_instrumentation__.Notify(124034)
					log.Errorf(ctx, "error running draining task: %+v", err)
				} else {
					__antithesis_instrumentation__.Notify(124035)
				}
				__antithesis_instrumentation__.Notify(124032)
				wg.Done()
				return false
			} else {
				__antithesis_instrumentation__.Notify(124036)
			}
			__antithesis_instrumentation__.Notify(123993)
			return true
		})
		__antithesis_instrumentation__.Notify(123991)
		wg.Wait()
		return int(numTransfersAttempted)
	}
	__antithesis_instrumentation__.Notify(123986)

	if numRemaining := transferAllAway(ctx); numRemaining > 0 {
		__antithesis_instrumentation__.Notify(124037)

		if reporter != nil {
			__antithesis_instrumentation__.Notify(124038)
			reporter(numRemaining, "range lease iterations")
		} else {
			__antithesis_instrumentation__.Notify(124039)
		}
	} else {
		__antithesis_instrumentation__.Notify(124040)

		return
	}
	__antithesis_instrumentation__.Notify(123987)

	transferTimeout := leaseTransferWait.Get(&s.cfg.Settings.SV)

	drainLeasesOp := "transfer range leases"
	if err := contextutil.RunWithTimeout(ctx, drainLeasesOp, transferTimeout,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(124041)
			opts := retry.Options{
				InitialBackoff: 10 * time.Millisecond,
				MaxBackoff:     time.Second,
				Multiplier:     2,
			}
			everySecond := log.Every(time.Second)
			var err error

			for r := retry.StartWithCtx(ctx, opts); r.Next(); {
				__antithesis_instrumentation__.Notify(124043)
				err = nil
				if numRemaining := transferAllAway(ctx); numRemaining > 0 {
					__antithesis_instrumentation__.Notify(124045)

					if reporter != nil {
						__antithesis_instrumentation__.Notify(124047)
						reporter(numRemaining, "range lease iterations")
					} else {
						__antithesis_instrumentation__.Notify(124048)
					}
					__antithesis_instrumentation__.Notify(124046)
					err = errors.Errorf("waiting for %d replicas to transfer their lease away", numRemaining)
					if everySecond.ShouldLog() {
						__antithesis_instrumentation__.Notify(124049)
						log.Infof(ctx, "%v", err)
					} else {
						__antithesis_instrumentation__.Notify(124050)
					}
				} else {
					__antithesis_instrumentation__.Notify(124051)
				}
				__antithesis_instrumentation__.Notify(124044)
				if err == nil {
					__antithesis_instrumentation__.Notify(124052)

					break
				} else {
					__antithesis_instrumentation__.Notify(124053)
				}
			}
			__antithesis_instrumentation__.Notify(124042)

			return errors.CombineErrors(err, ctx.Err())
		}); err != nil {
		__antithesis_instrumentation__.Notify(124054)
		if tErr := (*contextutil.TimeoutError)(nil); errors.As(err, &tErr) && func() bool {
			__antithesis_instrumentation__.Notify(124055)
			return tErr.Operation() == drainLeasesOp == true
		}() == true {
			__antithesis_instrumentation__.Notify(124056)

			log.Warningf(ctx, "unable to drain cleanly within %s (cluster setting %s), "+
				"service might briefly deteriorate if the node is terminated: %s",
				transferTimeout, leaseTransferWaitSettingName, tErr.Cause())
		} else {
			__antithesis_instrumentation__.Notify(124057)
			log.Warningf(ctx, "drain error: %+v", err)
		}
	} else {
		__antithesis_instrumentation__.Notify(124058)
	}
}

func (s *Store) IsStarted() bool {
	__antithesis_instrumentation__.Notify(124059)
	return atomic.LoadInt32(&s.started) == 1
}

func IterateIDPrefixKeys(
	ctx context.Context,
	reader storage.Reader,
	keyFn func(roachpb.RangeID) roachpb.Key,
	msg protoutil.Message,
	f func(_ roachpb.RangeID) error,
) error {
	__antithesis_instrumentation__.Notify(124060)
	rangeID := roachpb.RangeID(1)

	iter := reader.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{
		UpperBound: keys.LocalRangeIDPrefix.PrefixEnd().AsRawKey(),
	})
	defer iter.Close()

	for {
		__antithesis_instrumentation__.Notify(124061)
		bumped := false
		mvccKey := storage.MakeMVCCMetadataKey(keyFn(rangeID))
		iter.SeekGE(mvccKey)

		if ok, err := iter.Valid(); !ok {
			__antithesis_instrumentation__.Notify(124070)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124071)
		}
		__antithesis_instrumentation__.Notify(124062)

		unsafeKey := iter.UnsafeKey()

		if !bytes.HasPrefix(unsafeKey.Key, keys.LocalRangeIDPrefix) {
			__antithesis_instrumentation__.Notify(124072)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(124073)
		}
		__antithesis_instrumentation__.Notify(124063)

		curRangeID, _, _, _, err := keys.DecodeRangeIDKey(unsafeKey.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(124074)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124075)
		}
		__antithesis_instrumentation__.Notify(124064)

		if curRangeID > rangeID {
			__antithesis_instrumentation__.Notify(124076)

			if !bumped {
				__antithesis_instrumentation__.Notify(124078)
				rangeID = curRangeID
				bumped = true
			} else {
				__antithesis_instrumentation__.Notify(124079)
			}
			__antithesis_instrumentation__.Notify(124077)
			mvccKey = storage.MakeMVCCMetadataKey(keyFn(rangeID))
		} else {
			__antithesis_instrumentation__.Notify(124080)
		}
		__antithesis_instrumentation__.Notify(124065)

		if !unsafeKey.Key.Equal(mvccKey.Key) {
			__antithesis_instrumentation__.Notify(124081)
			if !bumped {
				__antithesis_instrumentation__.Notify(124083)

				rangeID++
				bumped = true
			} else {
				__antithesis_instrumentation__.Notify(124084)
			}
			__antithesis_instrumentation__.Notify(124082)
			continue
		} else {
			__antithesis_instrumentation__.Notify(124085)
		}
		__antithesis_instrumentation__.Notify(124066)

		ok, err := storage.MVCCGetProto(
			ctx, reader, unsafeKey.Key, hlc.Timestamp{}, msg, storage.MVCCGetOptions{})
		if err != nil {
			__antithesis_instrumentation__.Notify(124086)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124087)
		}
		__antithesis_instrumentation__.Notify(124067)
		if !ok {
			__antithesis_instrumentation__.Notify(124088)
			return errors.Errorf("unable to unmarshal %s into %T", unsafeKey.Key, msg)
		} else {
			__antithesis_instrumentation__.Notify(124089)
		}
		__antithesis_instrumentation__.Notify(124068)

		if err := f(rangeID); err != nil {
			__antithesis_instrumentation__.Notify(124090)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(124092)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(124093)
			}
			__antithesis_instrumentation__.Notify(124091)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124094)
		}
		__antithesis_instrumentation__.Notify(124069)
		rangeID++
	}
}

func IterateRangeDescriptorsFromDisk(
	ctx context.Context, reader storage.Reader, fn func(desc roachpb.RangeDescriptor) error,
) error {
	__antithesis_instrumentation__.Notify(124095)
	log.Event(ctx, "beginning range descriptor iteration")

	start := keys.RangeDescriptorKey(roachpb.RKeyMin)
	end := keys.RangeDescriptorKey(roachpb.RKeyMax)

	allCount := 0
	matchCount := 0
	bySuffix := make(map[string]int)
	kvToDesc := func(kv roachpb.KeyValue) error {
		__antithesis_instrumentation__.Notify(124097)
		allCount++

		_, suffix, _, err := keys.DecodeRangeKey(kv.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(124102)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124103)
		}
		__antithesis_instrumentation__.Notify(124098)
		bySuffix[string(suffix)]++
		if !bytes.Equal(suffix, keys.LocalRangeDescriptorSuffix) {
			__antithesis_instrumentation__.Notify(124104)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(124105)
		}
		__antithesis_instrumentation__.Notify(124099)
		var desc roachpb.RangeDescriptor
		if err := kv.Value.GetProto(&desc); err != nil {
			__antithesis_instrumentation__.Notify(124106)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124107)
		}
		__antithesis_instrumentation__.Notify(124100)
		matchCount++
		if err := fn(desc); iterutil.Done(err) {
			__antithesis_instrumentation__.Notify(124108)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(124109)
		}
		__antithesis_instrumentation__.Notify(124101)
		return err
	}
	__antithesis_instrumentation__.Notify(124096)

	_, err := storage.MVCCIterate(ctx, reader, start, end, hlc.MaxTimestamp,
		storage.MVCCScanOptions{Inconsistent: true}, kvToDesc)
	log.Eventf(ctx, "iterated over %d keys to find %d range descriptors (by suffix: %v)",
		allCount, matchCount, bySuffix)
	return err
}

func ReadStoreIdent(ctx context.Context, eng storage.Engine) (roachpb.StoreIdent, error) {
	__antithesis_instrumentation__.Notify(124110)
	var ident roachpb.StoreIdent
	ok, err := storage.MVCCGetProto(
		ctx, eng, keys.StoreIdentKey(), hlc.Timestamp{}, &ident, storage.MVCCGetOptions{})
	if err != nil {
		__antithesis_instrumentation__.Notify(124112)
		return roachpb.StoreIdent{}, err
	} else {
		__antithesis_instrumentation__.Notify(124113)
		if !ok {
			__antithesis_instrumentation__.Notify(124114)
			return roachpb.StoreIdent{}, &NotBootstrappedError{}
		} else {
			__antithesis_instrumentation__.Notify(124115)
		}
	}
	__antithesis_instrumentation__.Notify(124111)
	return ident, err
}

func (s *Store) Start(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(124116)
	s.stopper = stopper

	ident, err := ReadStoreIdent(ctx, s.engine)
	if err != nil {
		__antithesis_instrumentation__.Notify(124134)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124135)
	}
	__antithesis_instrumentation__.Notify(124117)
	s.Ident = &ident

	s.cfg.AmbientCtx.AddLogTag("s", s.StoreID())
	ctx = s.AnnotateCtx(ctx)
	log.Event(ctx, "read store identity")

	if logSetter, ok := s.engine.(storage.StoreIDSetter); ok {
		__antithesis_instrumentation__.Notify(124136)
		logSetter.SetStoreID(ctx, int32(s.StoreID()))
	} else {
		__antithesis_instrumentation__.Notify(124137)
	}
	__antithesis_instrumentation__.Notify(124118)

	if s.scanner != nil {
		__antithesis_instrumentation__.Notify(124138)
		s.scanner.AmbientContext.AddLogTag("s", s.StoreID())

		s.scanner.stopper = s.stopper
	} else {
		__antithesis_instrumentation__.Notify(124139)
	}
	__antithesis_instrumentation__.Notify(124119)

	if s.nodeDesc.NodeID != 0 && func() bool {
		__antithesis_instrumentation__.Notify(124140)
		return s.Ident.NodeID != s.nodeDesc.NodeID == true
	}() == true {
		__antithesis_instrumentation__.Notify(124141)
		return errors.Errorf("node id:%d does not equal the one in node descriptor:%d", s.Ident.NodeID, s.nodeDesc.NodeID)
	} else {
		__antithesis_instrumentation__.Notify(124142)
	}
	__antithesis_instrumentation__.Notify(124120)

	if s.cfg.Gossip != nil {
		__antithesis_instrumentation__.Notify(124143)
		s.cfg.Gossip.NodeID.Set(ctx, s.Ident.NodeID)
	} else {
		__antithesis_instrumentation__.Notify(124144)
	}
	__antithesis_instrumentation__.Notify(124121)

	idAlloc, err := idalloc.NewAllocator(idalloc.Options{
		AmbientCtx:  s.cfg.AmbientCtx,
		Key:         keys.RangeIDGenerator,
		Incrementer: idalloc.DBIncrementer(s.db),
		BlockSize:   rangeIDAllocCount,
		Stopper:     s.stopper,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(124145)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124146)
	}
	__antithesis_instrumentation__.Notify(124122)

	var intentResolverRangeCache intentresolver.RangeCache
	rngCache := s.cfg.RangeDescriptorCache
	if s.cfg.RangeDescriptorCache != nil {
		__antithesis_instrumentation__.Notify(124147)
		intentResolverRangeCache = rngCache
	} else {
		__antithesis_instrumentation__.Notify(124148)
	}
	__antithesis_instrumentation__.Notify(124123)

	s.intentResolver = intentresolver.New(intentresolver.Config{
		Clock:                s.cfg.Clock,
		DB:                   s.db,
		Stopper:              stopper,
		TaskLimit:            s.cfg.IntentResolverTaskLimit,
		AmbientCtx:           s.cfg.AmbientCtx,
		TestingKnobs:         s.cfg.TestingKnobs.IntentResolverKnobs,
		RangeDescriptorCache: intentResolverRangeCache,
	})
	s.metrics.registry.AddMetricStruct(s.intentResolver.Metrics)

	s.raftTruncator = makeRaftLogTruncator(s.cfg.AmbientCtx, (*storeForTruncatorImpl)(s), stopper)
	{
		__antithesis_instrumentation__.Notify(124149)
		truncator := s.raftTruncator
		s.engine.RegisterFlushCompletedCallback(func() {
			__antithesis_instrumentation__.Notify(124150)
			truncator.durabilityAdvancedCallback()
		})
	}
	__antithesis_instrumentation__.Notify(124124)

	s.recoveryMgr = txnrecovery.NewManager(
		s.cfg.AmbientCtx, s.cfg.Clock, s.db, stopper,
	)
	s.metrics.registry.AddMetricStruct(s.recoveryMgr.Metrics())

	s.rangeIDAlloc = idAlloc

	now := s.cfg.Clock.Now()
	s.startedAt = now.WallTime

	err = IterateRangeDescriptorsFromDisk(ctx, s.engine,
		func(desc roachpb.RangeDescriptor) error {
			__antithesis_instrumentation__.Notify(124151)
			if !desc.IsInitialized() {
				__antithesis_instrumentation__.Notify(124158)
				return errors.Errorf("found uninitialized RangeDescriptor: %+v", desc)
			} else {
				__antithesis_instrumentation__.Notify(124159)
			}
			__antithesis_instrumentation__.Notify(124152)
			replicaDesc, found := desc.GetReplicaDescriptor(s.StoreID())
			if !found {
				__antithesis_instrumentation__.Notify(124160)

				return errors.AssertionFailedf(
					"found RangeDescriptor for range %d at generation %d which does not"+
						" contain this store %d",
					redact.Safe(desc.RangeID),
					redact.Safe(desc.Generation),
					redact.Safe(s.StoreID()))
			} else {
				__antithesis_instrumentation__.Notify(124161)
			}
			__antithesis_instrumentation__.Notify(124153)

			rep, err := newReplica(ctx, &desc, s, replicaDesc.ReplicaID)
			if err != nil {
				__antithesis_instrumentation__.Notify(124162)
				return err
			} else {
				__antithesis_instrumentation__.Notify(124163)
			}
			__antithesis_instrumentation__.Notify(124154)

			s.mu.Lock()
			err = s.addReplicaInternalLocked(rep)
			s.mu.Unlock()
			if err != nil {
				__antithesis_instrumentation__.Notify(124164)
				return err
			} else {
				__antithesis_instrumentation__.Notify(124165)
			}
			__antithesis_instrumentation__.Notify(124155)

			s.metrics.ReplicaCount.Inc(1)
			if _, ok := rep.TenantID(); ok {
				__antithesis_instrumentation__.Notify(124166)

				s.metrics.addMVCCStats(ctx, rep.tenantMetricsRef, rep.GetMVCCStats())
			} else {
				__antithesis_instrumentation__.Notify(124167)
				return errors.AssertionFailedf("found newly constructed replica"+
					" for range %d at generation %d with an invalid tenant ID in store %d",
					redact.Safe(desc.RangeID),
					redact.Safe(desc.Generation),
					redact.Safe(s.StoreID()))
			}
			__antithesis_instrumentation__.Notify(124156)

			if _, ok := desc.GetReplicaDescriptor(s.StoreID()); !ok {
				__antithesis_instrumentation__.Notify(124168)

				s.replicaGCQueue.AddAsync(ctx, rep, replicaGCPriorityRemoved)
			} else {
				__antithesis_instrumentation__.Notify(124169)
			}
			__antithesis_instrumentation__.Notify(124157)

			return nil
		})
	__antithesis_instrumentation__.Notify(124125)
	if err != nil {
		__antithesis_instrumentation__.Notify(124170)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124171)
	}
	__antithesis_instrumentation__.Notify(124126)

	s.cfg.Transport.Listen(s.StoreID(), s)
	s.processRaft(ctx)

	if s.cfg.NodeLiveness != nil {
		__antithesis_instrumentation__.Notify(124172)
		s.cfg.NodeLiveness.RegisterCallback(s.nodeIsLiveCallback)
	} else {
		__antithesis_instrumentation__.Notify(124173)
	}
	__antithesis_instrumentation__.Notify(124127)

	if scp := s.cfg.SystemConfigProvider; scp != nil {
		__antithesis_instrumentation__.Notify(124174)
		systemCfgUpdateC, _ := scp.RegisterSystemConfigChannel()
		_ = s.stopper.RunAsyncTask(ctx, "syscfg-listener", func(context.Context) {
			__antithesis_instrumentation__.Notify(124175)
			for {
				__antithesis_instrumentation__.Notify(124176)
				select {
				case <-systemCfgUpdateC:
					__antithesis_instrumentation__.Notify(124177)
					cfg := scp.GetSystemConfig()
					s.systemGossipUpdate(cfg)
				case <-s.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(124178)
					return
				}
			}
		})
	} else {
		__antithesis_instrumentation__.Notify(124179)
	}
	__antithesis_instrumentation__.Notify(124128)

	if s.cfg.Gossip != nil {
		__antithesis_instrumentation__.Notify(124180)

		s.startGossip()

		_ = s.stopper.RunAsyncTask(ctx, "scanner", func(context.Context) {
			__antithesis_instrumentation__.Notify(124181)
			select {
			case <-s.cfg.Gossip.Connected:
				__antithesis_instrumentation__.Notify(124182)
				s.scanner.Start()
			case <-s.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(124183)
				return
			}
		})
	} else {
		__antithesis_instrumentation__.Notify(124184)
	}
	__antithesis_instrumentation__.Notify(124129)

	if !s.cfg.SpanConfigsDisabled {
		__antithesis_instrumentation__.Notify(124185)
		s.cfg.SpanConfigSubscriber.Subscribe(func(ctx context.Context, update roachpb.Span) {
			__antithesis_instrumentation__.Notify(124187)
			s.onSpanConfigUpdate(ctx, update)
		})
		__antithesis_instrumentation__.Notify(124186)

		spanconfigstore.EnabledSetting.SetOnChange(&s.ClusterSettings().SV, func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(124188)
			enabled := spanconfigstore.EnabledSetting.Get(&s.ClusterSettings().SV)
			if enabled {
				__antithesis_instrumentation__.Notify(124189)
				s.applyAllFromSpanConfigStore(ctx)
			} else {
				__antithesis_instrumentation__.Notify(124190)
				if scp := s.cfg.SystemConfigProvider; scp != nil {
					__antithesis_instrumentation__.Notify(124191)
					if sc := scp.GetSystemConfig(); sc != nil {
						__antithesis_instrumentation__.Notify(124192)
						s.systemGossipUpdate(sc)
					} else {
						__antithesis_instrumentation__.Notify(124193)
					}
				} else {
					__antithesis_instrumentation__.Notify(124194)
				}
			}
		})
	} else {
		__antithesis_instrumentation__.Notify(124195)
	}
	__antithesis_instrumentation__.Notify(124130)

	if !s.cfg.TestingKnobs.DisableAutomaticLeaseRenewal {
		__antithesis_instrumentation__.Notify(124196)
		s.startLeaseRenewer(ctx)
	} else {
		__antithesis_instrumentation__.Notify(124197)
	}
	__antithesis_instrumentation__.Notify(124131)

	s.startRangefeedUpdater(ctx)

	if s.replicateQueue != nil {
		__antithesis_instrumentation__.Notify(124198)
		s.storeRebalancer = NewStoreRebalancer(
			s.cfg.AmbientCtx, s.cfg.Settings, s.replicateQueue, s.replRankings)
		s.storeRebalancer.Start(ctx, s.stopper)
	} else {
		__antithesis_instrumentation__.Notify(124199)
	}
	__antithesis_instrumentation__.Notify(124132)

	s.consistencyLimiter = quotapool.NewRateLimiter(
		"ConsistencyQueue",
		quotapool.Limit(consistencyCheckRate.Get(&s.ClusterSettings().SV)),
		consistencyCheckRate.Get(&s.ClusterSettings().SV)*consistencyCheckRateBurstFactor,
		quotapool.WithMinimumWait(consistencyCheckRateMinWait))

	consistencyCheckRate.SetOnChange(&s.ClusterSettings().SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(124200)
		rate := consistencyCheckRate.Get(&s.ClusterSettings().SV)
		s.consistencyLimiter.UpdateLimit(quotapool.Limit(rate), rate*consistencyCheckRateBurstFactor)
	})
	__antithesis_instrumentation__.Notify(124133)
	s.consistencySem = quotapool.NewIntPool("concurrent async consistency checks",
		consistencyCheckAsyncConcurrency)
	s.stopper.AddCloser(s.consistencySem.Closer("stopper"))

	atomic.StoreInt32(&s.started, 1)

	return nil
}

func (s *Store) WaitForInit() {
	__antithesis_instrumentation__.Notify(124201)
	s.initComplete.Wait()
}

var errPeriodicGossipsDisabled = errors.New("periodic gossip is disabled")

func (s *Store) startGossip() {
	__antithesis_instrumentation__.Notify(124202)
	wakeReplica := func(ctx context.Context, repl *Replica) error {
		__antithesis_instrumentation__.Notify(124205)

		_, pErr := repl.getLeaseForGossip(ctx)
		return pErr.GoError()
	}
	__antithesis_instrumentation__.Notify(124203)
	gossipFns := []struct {
		key         roachpb.Key
		fn          func(context.Context, *Replica) error
		description redact.SafeString
		interval    time.Duration
	}{
		{
			key: roachpb.KeyMin,
			fn: func(ctx context.Context, repl *Replica) error {
				__antithesis_instrumentation__.Notify(124206)

				return repl.maybeGossipFirstRange(ctx).GoError()
			},
			description: "first range descriptor",
			interval:    s.cfg.SentinelGossipTTL() / 2,
		},
		{
			key:         keys.SystemConfigSpan.Key,
			fn:          wakeReplica,
			description: "system config",
			interval:    systemDataGossipInterval,
		},
		{
			key:         keys.NodeLivenessSpan.Key,
			fn:          wakeReplica,
			description: "node liveness",
			interval:    systemDataGossipInterval,
		},
	}
	__antithesis_instrumentation__.Notify(124204)

	cannotGossipEvery := log.Every(time.Minute)
	cannotGossipEvery.ShouldLog()

	s.initComplete.Add(len(gossipFns))
	for _, gossipFn := range gossipFns {
		__antithesis_instrumentation__.Notify(124207)
		gossipFn := gossipFn
		bgCtx := s.AnnotateCtx(context.Background())
		if err := s.stopper.RunAsyncTask(bgCtx, "store-gossip", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(124208)
			ticker := time.NewTicker(gossipFn.interval)
			defer ticker.Stop()
			for first := true; ; {
				__antithesis_instrumentation__.Notify(124209)

				retryOptions := base.DefaultRetryOptions()
				retryOptions.Closer = s.stopper.ShouldQuiesce()
				for r := retry.Start(retryOptions); r.Next(); {
					__antithesis_instrumentation__.Notify(124212)
					if repl := s.LookupReplica(roachpb.RKey(gossipFn.key)); repl != nil {
						__antithesis_instrumentation__.Notify(124214)
						annotatedCtx := repl.AnnotateCtx(ctx)
						if err := gossipFn.fn(annotatedCtx, repl); err != nil {
							__antithesis_instrumentation__.Notify(124215)
							if cannotGossipEvery.ShouldLog() {
								__antithesis_instrumentation__.Notify(124217)
								log.Infof(annotatedCtx, "could not gossip %s: %v", gossipFn.description, err)
							} else {
								__antithesis_instrumentation__.Notify(124218)
							}
							__antithesis_instrumentation__.Notify(124216)
							if !errors.Is(err, errPeriodicGossipsDisabled) {
								__antithesis_instrumentation__.Notify(124219)
								continue
							} else {
								__antithesis_instrumentation__.Notify(124220)
							}
						} else {
							__antithesis_instrumentation__.Notify(124221)
						}
					} else {
						__antithesis_instrumentation__.Notify(124222)
					}
					__antithesis_instrumentation__.Notify(124213)
					break
				}
				__antithesis_instrumentation__.Notify(124210)
				if first {
					__antithesis_instrumentation__.Notify(124223)
					first = false
					s.initComplete.Done()
				} else {
					__antithesis_instrumentation__.Notify(124224)
				}
				__antithesis_instrumentation__.Notify(124211)
				select {
				case <-ticker.C:
					__antithesis_instrumentation__.Notify(124225)
				case <-s.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(124226)
					return
				}
			}
		}); err != nil {
			__antithesis_instrumentation__.Notify(124227)
			s.initComplete.Done()
		} else {
			__antithesis_instrumentation__.Notify(124228)
		}
	}
}

var errSysCfgUnavailable = errors.New("system config not available in gossip")

func (s *Store) GetConfReader(ctx context.Context) (spanconfig.StoreReader, error) {
	__antithesis_instrumentation__.Notify(124229)
	if s.cfg.TestingKnobs.MakeSystemConfigSpanUnavailableToQueues {
		__antithesis_instrumentation__.Notify(124232)
		return nil, errSysCfgUnavailable
	} else {
		__antithesis_instrumentation__.Notify(124233)
	}
	__antithesis_instrumentation__.Notify(124230)

	_ = clusterversion.EnsureSpanConfigReconciliation

	_ = clusterversion.EnsureSpanConfigSubscription

	_ = clusterversion.EnableSpanConfigStore

	if s.cfg.SpanConfigsDisabled || func() bool {
		__antithesis_instrumentation__.Notify(124234)
		return !spanconfigstore.EnabledSetting.Get(&s.ClusterSettings().SV) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(124235)
		return !s.cfg.Settings.Version.IsActive(ctx, clusterversion.EnableSpanConfigStore) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(124236)
		return s.TestingKnobs().UseSystemConfigSpanForQueues == true
	}() == true {
		__antithesis_instrumentation__.Notify(124237)

		sysCfg := s.cfg.SystemConfigProvider.GetSystemConfig()
		if sysCfg == nil {
			__antithesis_instrumentation__.Notify(124239)
			return nil, errSysCfgUnavailable
		} else {
			__antithesis_instrumentation__.Notify(124240)
		}
		__antithesis_instrumentation__.Notify(124238)
		return sysCfg, nil
	} else {
		__antithesis_instrumentation__.Notify(124241)
	}
	__antithesis_instrumentation__.Notify(124231)

	return s.cfg.SpanConfigSubscriber, nil
}

func (s *Store) startLeaseRenewer(ctx context.Context) {
	__antithesis_instrumentation__.Notify(124242)

	_ = s.stopper.RunAsyncTask(ctx, "lease-renewer", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(124243)
		timer := timeutil.NewTimer()
		defer timer.Stop()

		renewalDuration := s.cfg.RangeLeaseActiveDuration() / 5
		if d := s.cfg.TestingKnobs.LeaseRenewalDurationOverride; d > 0 {
			__antithesis_instrumentation__.Notify(124245)
			renewalDuration = d
		} else {
			__antithesis_instrumentation__.Notify(124246)
		}
		__antithesis_instrumentation__.Notify(124244)
		for {
			__antithesis_instrumentation__.Notify(124247)
			var numRenewableLeases int
			s.renewableLeases.Range(func(k int64, v unsafe.Pointer) bool {
				__antithesis_instrumentation__.Notify(124251)
				numRenewableLeases++
				repl := (*Replica)(v)
				annotatedCtx := repl.AnnotateCtx(ctx)
				if _, pErr := repl.redirectOnOrAcquireLease(annotatedCtx); pErr != nil {
					__antithesis_instrumentation__.Notify(124253)
					if _, ok := pErr.GetDetail().(*roachpb.NotLeaseHolderError); !ok {
						__antithesis_instrumentation__.Notify(124255)
						log.Warningf(annotatedCtx, "failed to proactively renew lease: %s", pErr)
					} else {
						__antithesis_instrumentation__.Notify(124256)
					}
					__antithesis_instrumentation__.Notify(124254)
					s.renewableLeases.Delete(k)
				} else {
					__antithesis_instrumentation__.Notify(124257)
				}
				__antithesis_instrumentation__.Notify(124252)
				return true
			})
			__antithesis_instrumentation__.Notify(124248)

			if numRenewableLeases > 0 {
				__antithesis_instrumentation__.Notify(124258)
				timer.Reset(renewalDuration)
			} else {
				__antithesis_instrumentation__.Notify(124259)
			}
			__antithesis_instrumentation__.Notify(124249)
			if fn := s.cfg.TestingKnobs.LeaseRenewalOnPostCycle; fn != nil {
				__antithesis_instrumentation__.Notify(124260)
				fn()
			} else {
				__antithesis_instrumentation__.Notify(124261)
			}
			__antithesis_instrumentation__.Notify(124250)
			select {
			case <-s.renewableLeasesSignal:
				__antithesis_instrumentation__.Notify(124262)
			case <-timer.C:
				__antithesis_instrumentation__.Notify(124263)
				timer.Read = true
			case <-s.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(124264)
				return
			}
		}
	})
}

func (s *Store) startRangefeedUpdater(ctx context.Context) {
	__antithesis_instrumentation__.Notify(124265)
	_ = s.stopper.RunAsyncTaskEx(ctx,
		stop.TaskOpts{
			TaskName: "closedts-rangefeed-updater",
			SpanOpt:  stop.SterileRootSpan,
		}, func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(124266)
			timer := timeutil.NewTimer()
			defer timer.Stop()
			var replIDs []roachpb.RangeID
			st := s.cfg.Settings

			confCh := make(chan struct{}, 1)
			confChanged := func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(124269)
				select {
				case confCh <- struct{}{}:
					__antithesis_instrumentation__.Notify(124270)
				default:
					__antithesis_instrumentation__.Notify(124271)
				}
			}
			__antithesis_instrumentation__.Notify(124267)
			closedts.SideTransportCloseInterval.SetOnChange(&st.SV, confChanged)
			RangeFeedRefreshInterval.SetOnChange(&st.SV, confChanged)

			getInterval := func() time.Duration {
				__antithesis_instrumentation__.Notify(124272)
				refresh := RangeFeedRefreshInterval.Get(&st.SV)
				if refresh != 0 {
					__antithesis_instrumentation__.Notify(124274)
					return refresh
				} else {
					__antithesis_instrumentation__.Notify(124275)
				}
				__antithesis_instrumentation__.Notify(124273)
				return closedts.SideTransportCloseInterval.Get(&st.SV)
			}
			__antithesis_instrumentation__.Notify(124268)

			for {
				__antithesis_instrumentation__.Notify(124276)
				interval := getInterval()
				if interval > 0 {
					__antithesis_instrumentation__.Notify(124278)
					timer.Reset(interval)
				} else {
					__antithesis_instrumentation__.Notify(124279)

					timer.Stop()
					timer = timeutil.NewTimer()
				}
				__antithesis_instrumentation__.Notify(124277)
				select {
				case <-timer.C:
					__antithesis_instrumentation__.Notify(124280)
					timer.Read = true
					s.rangefeedReplicas.Lock()
					replIDs = replIDs[:0]
					for replID := range s.rangefeedReplicas.m {
						__antithesis_instrumentation__.Notify(124284)
						replIDs = append(replIDs, replID)
					}
					__antithesis_instrumentation__.Notify(124281)
					s.rangefeedReplicas.Unlock()

					for _, replID := range replIDs {
						__antithesis_instrumentation__.Notify(124285)
						r := s.GetReplicaIfExists(replID)
						if r == nil {
							__antithesis_instrumentation__.Notify(124287)
							continue
						} else {
							__antithesis_instrumentation__.Notify(124288)
						}
						__antithesis_instrumentation__.Notify(124286)
						r.handleClosedTimestampUpdate(ctx, r.GetClosedTimestamp(ctx))
					}
				case <-confCh:
					__antithesis_instrumentation__.Notify(124282)

					continue
				case <-s.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(124283)
					return
				}

			}
		})
}

func (s *Store) addReplicaWithRangefeed(rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(124289)
	s.rangefeedReplicas.Lock()
	s.rangefeedReplicas.m[rangeID] = struct{}{}
	s.rangefeedReplicas.Unlock()
}

func (s *Store) removeReplicaWithRangefeed(rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(124290)
	s.rangefeedReplicas.Lock()
	delete(s.rangefeedReplicas.m, rangeID)
	s.rangefeedReplicas.Unlock()
}

func (s *Store) systemGossipUpdate(sysCfg *config.SystemConfig) {
	__antithesis_instrumentation__.Notify(124291)
	if !s.cfg.SpanConfigsDisabled && func() bool {
		__antithesis_instrumentation__.Notify(124294)
		return spanconfigstore.EnabledSetting.Get(&s.ClusterSettings().SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(124295)
		return
	} else {
		__antithesis_instrumentation__.Notify(124296)
	}
	__antithesis_instrumentation__.Notify(124292)

	ctx := s.AnnotateCtx(context.Background())
	s.computeInitialMetrics.Do(func() {
		__antithesis_instrumentation__.Notify(124297)

		if err := s.ComputeMetrics(ctx, -1); err != nil {
			__antithesis_instrumentation__.Notify(124299)
			log.Infof(ctx, "%s: failed initial metrics computation: %s", s, err)
		} else {
			__antithesis_instrumentation__.Notify(124300)
		}
		__antithesis_instrumentation__.Notify(124298)
		log.Event(ctx, "computed initial metrics")
	})
	__antithesis_instrumentation__.Notify(124293)

	shouldQueue := s.systemConfigUpdateQueueRateLimiter.AdmitN(1)

	now := s.cfg.Clock.NowAsClockTimestamp()
	newStoreReplicaVisitor(s).Visit(func(repl *Replica) bool {
		__antithesis_instrumentation__.Notify(124301)
		key := repl.Desc().StartKey
		conf, err := sysCfg.GetSpanConfigForKey(ctx, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(124305)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(124307)
				log.Infof(context.TODO(), "failed to get span config for key %s", key)
			} else {
				__antithesis_instrumentation__.Notify(124308)
			}
			__antithesis_instrumentation__.Notify(124306)
			conf = s.cfg.DefaultSpanConfig
		} else {
			__antithesis_instrumentation__.Notify(124309)
		}
		__antithesis_instrumentation__.Notify(124302)

		if s.cfg.SpanConfigsDisabled || func() bool {
			__antithesis_instrumentation__.Notify(124310)
			return !spanconfigstore.EnabledSetting.Get(&s.ClusterSettings().SV) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(124311)
			return !s.cfg.Settings.Version.IsActive(ctx, clusterversion.EnableSpanConfigStore) == true
		}() == true {
			__antithesis_instrumentation__.Notify(124312)
			repl.SetSpanConfig(conf)
		} else {
			__antithesis_instrumentation__.Notify(124313)
		}
		__antithesis_instrumentation__.Notify(124303)

		if shouldQueue {
			__antithesis_instrumentation__.Notify(124314)
			s.splitQueue.Async(ctx, "gossip update", true, func(ctx context.Context, h queueHelper) {
				__antithesis_instrumentation__.Notify(124316)
				h.MaybeAdd(ctx, repl, now)
			})
			__antithesis_instrumentation__.Notify(124315)
			s.mergeQueue.Async(ctx, "gossip update", true, func(ctx context.Context, h queueHelper) {
				__antithesis_instrumentation__.Notify(124317)
				h.MaybeAdd(ctx, repl, now)
			})
		} else {
			__antithesis_instrumentation__.Notify(124318)
		}
		__antithesis_instrumentation__.Notify(124304)
		return true
	})
}

func (s *Store) onSpanConfigUpdate(ctx context.Context, updated roachpb.Span) {
	__antithesis_instrumentation__.Notify(124319)
	if !spanconfigstore.EnabledSetting.Get(&s.ClusterSettings().SV) {
		__antithesis_instrumentation__.Notify(124322)
		return
	} else {
		__antithesis_instrumentation__.Notify(124323)
	}
	__antithesis_instrumentation__.Notify(124320)

	sp, err := keys.SpanAddr(updated)
	if err != nil {
		__antithesis_instrumentation__.Notify(124324)
		log.Errorf(ctx, "skipped applying update (%s), unexpected error resolving span address: %v",
			updated, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(124325)
	}
	__antithesis_instrumentation__.Notify(124321)

	now := s.cfg.Clock.NowAsClockTimestamp()

	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.mu.replicasByKey.VisitKeyRange(ctx, sp.Key, sp.EndKey, AscendingKeyOrder,
		func(ctx context.Context, it replicaOrPlaceholder) error {
			__antithesis_instrumentation__.Notify(124326)
			repl := it.repl
			if repl == nil {
				__antithesis_instrumentation__.Notify(124331)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(124332)
			}
			__antithesis_instrumentation__.Notify(124327)

			replCtx := repl.AnnotateCtx(ctx)
			startKey := repl.Desc().StartKey
			if sp.ContainsKey(startKey) {
				__antithesis_instrumentation__.Notify(124333)

				conf, err := s.cfg.SpanConfigSubscriber.GetSpanConfigForKey(replCtx, startKey)
				if err != nil {
					__antithesis_instrumentation__.Notify(124335)
					log.Errorf(ctx, "skipped applying update, unexpected error reading from subscriber: %v", err)
					return err
				} else {
					__antithesis_instrumentation__.Notify(124336)
				}
				__antithesis_instrumentation__.Notify(124334)
				repl.SetSpanConfig(conf)
			} else {
				__antithesis_instrumentation__.Notify(124337)
			}
			__antithesis_instrumentation__.Notify(124328)

			s.splitQueue.Async(replCtx, "span config update", true, func(ctx context.Context, h queueHelper) {
				__antithesis_instrumentation__.Notify(124338)
				h.MaybeAdd(ctx, repl, now)
			})
			__antithesis_instrumentation__.Notify(124329)
			s.mergeQueue.Async(replCtx, "span config update", true, func(ctx context.Context, h queueHelper) {
				__antithesis_instrumentation__.Notify(124339)
				h.MaybeAdd(ctx, repl, now)
			})
			__antithesis_instrumentation__.Notify(124330)
			return nil
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(124340)

		log.Errorf(ctx, "unexpected error visiting replicas: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(124341)
	}
}

func (s *Store) applyAllFromSpanConfigStore(ctx context.Context) {
	__antithesis_instrumentation__.Notify(124342)
	now := s.cfg.Clock.NowAsClockTimestamp()
	newStoreReplicaVisitor(s).Visit(func(repl *Replica) bool {
		__antithesis_instrumentation__.Notify(124343)
		replCtx := repl.AnnotateCtx(ctx)
		key := repl.Desc().StartKey
		conf, err := s.cfg.SpanConfigSubscriber.GetSpanConfigForKey(replCtx, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(124347)
			log.Errorf(ctx, "skipped applying config update, unexpected error reading from subscriber: %v", err)
			return true
		} else {
			__antithesis_instrumentation__.Notify(124348)
		}
		__antithesis_instrumentation__.Notify(124344)

		repl.SetSpanConfig(conf)
		s.splitQueue.Async(replCtx, "span config update", true, func(ctx context.Context, h queueHelper) {
			__antithesis_instrumentation__.Notify(124349)
			h.MaybeAdd(ctx, repl, now)
		})
		__antithesis_instrumentation__.Notify(124345)
		s.mergeQueue.Async(replCtx, "span config update", true, func(ctx context.Context, h queueHelper) {
			__antithesis_instrumentation__.Notify(124350)
			h.MaybeAdd(ctx, repl, now)
		})
		__antithesis_instrumentation__.Notify(124346)
		return true
	})
}

func (s *Store) asyncGossipStore(ctx context.Context, reason string, useCached bool) {
	__antithesis_instrumentation__.Notify(124351)
	if err := s.stopper.RunAsyncTask(
		ctx, fmt.Sprintf("storage.Store: gossip on %s", reason),
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(124352)
			if err := s.GossipStore(ctx, useCached); err != nil {
				__antithesis_instrumentation__.Notify(124353)
				log.Warningf(ctx, "error gossiping on %s: %+v", reason, err)
			} else {
				__antithesis_instrumentation__.Notify(124354)
			}
		}); err != nil {
		__antithesis_instrumentation__.Notify(124355)
		log.Warningf(ctx, "unable to gossip on %s: %+v", reason, err)
	} else {
		__antithesis_instrumentation__.Notify(124356)
	}
}

func (s *Store) GossipStore(ctx context.Context, useCached bool) error {
	__antithesis_instrumentation__.Notify(124357)

	syncutil.StoreFloat64(&s.gossipQueriesPerSecondVal, -1)
	syncutil.StoreFloat64(&s.gossipWritesPerSecondVal, -1)

	storeDesc, err := s.Descriptor(ctx, useCached)
	if err != nil {
		__antithesis_instrumentation__.Notify(124359)
		return errors.Wrapf(err, "problem getting store descriptor for store %+v", s.Ident)
	} else {
		__antithesis_instrumentation__.Notify(124360)
	}
	__antithesis_instrumentation__.Notify(124358)

	rangeCountdown := float64(storeDesc.Capacity.RangeCount) * s.cfg.TestingKnobs.GossipWhenCapacityDeltaExceedsFraction
	atomic.StoreInt32(&s.gossipRangeCountdown, int32(math.Ceil(math.Min(rangeCountdown, 3))))
	leaseCountdown := float64(storeDesc.Capacity.LeaseCount) * s.cfg.TestingKnobs.GossipWhenCapacityDeltaExceedsFraction
	atomic.StoreInt32(&s.gossipLeaseCountdown, int32(math.Ceil(math.Max(leaseCountdown, 1))))
	syncutil.StoreFloat64(&s.gossipQueriesPerSecondVal, storeDesc.Capacity.QueriesPerSecond)
	syncutil.StoreFloat64(&s.gossipWritesPerSecondVal, storeDesc.Capacity.WritesPerSecond)

	gossipStoreKey := gossip.MakeStoreKey(storeDesc.StoreID)

	return s.cfg.Gossip.AddInfoProto(gossipStoreKey, storeDesc, gossip.StoreTTL)
}

type capacityChangeEvent int

const (
	rangeAddEvent capacityChangeEvent = iota
	rangeRemoveEvent
	leaseAddEvent
	leaseRemoveEvent
)

func (s *Store) maybeGossipOnCapacityChange(ctx context.Context, cce capacityChangeEvent) {
	__antithesis_instrumentation__.Notify(124361)
	if s.cfg.TestingKnobs.DisableLeaseCapacityGossip && func() bool {
		__antithesis_instrumentation__.Notify(124364)
		return (cce == leaseAddEvent || func() bool {
			__antithesis_instrumentation__.Notify(124365)
			return cce == leaseRemoveEvent == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(124366)
		return
	} else {
		__antithesis_instrumentation__.Notify(124367)
	}
	__antithesis_instrumentation__.Notify(124362)

	s.cachedCapacity.Lock()
	switch cce {
	case rangeAddEvent:
		__antithesis_instrumentation__.Notify(124368)
		s.cachedCapacity.RangeCount++
	case rangeRemoveEvent:
		__antithesis_instrumentation__.Notify(124369)
		s.cachedCapacity.RangeCount--
	case leaseAddEvent:
		__antithesis_instrumentation__.Notify(124370)
		s.cachedCapacity.LeaseCount++
	case leaseRemoveEvent:
		__antithesis_instrumentation__.Notify(124371)
		s.cachedCapacity.LeaseCount--
	default:
		__antithesis_instrumentation__.Notify(124372)
	}
	__antithesis_instrumentation__.Notify(124363)
	s.cachedCapacity.Unlock()

	if ((cce == rangeAddEvent || func() bool {
		__antithesis_instrumentation__.Notify(124373)
		return cce == rangeRemoveEvent == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(124374)
		return atomic.AddInt32(&s.gossipRangeCountdown, -1) == 0 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(124375)
		return ((cce == leaseAddEvent || func() bool {
			__antithesis_instrumentation__.Notify(124376)
			return cce == leaseRemoveEvent == true
		}() == true) && func() bool {
			__antithesis_instrumentation__.Notify(124377)
			return atomic.AddInt32(&s.gossipLeaseCountdown, -1) == 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(124378)

		atomic.StoreInt32(&s.gossipRangeCountdown, 0)
		atomic.StoreInt32(&s.gossipLeaseCountdown, 0)
		s.asyncGossipStore(ctx, "capacity change", true)
	} else {
		__antithesis_instrumentation__.Notify(124379)
	}
}

func (s *Store) recordNewPerSecondStats(newQPS, newWPS float64) {
	__antithesis_instrumentation__.Notify(124380)
	oldQPS := syncutil.LoadFloat64(&s.gossipQueriesPerSecondVal)
	oldWPS := syncutil.LoadFloat64(&s.gossipWritesPerSecondVal)
	if oldQPS == -1 || func() bool {
		__antithesis_instrumentation__.Notify(124384)
		return oldWPS == -1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(124385)

		return
	} else {
		__antithesis_instrumentation__.Notify(124386)
	}
	__antithesis_instrumentation__.Notify(124381)

	const minAbsoluteChange = 100
	updateForQPS := (newQPS < oldQPS*.5 || func() bool {
		__antithesis_instrumentation__.Notify(124387)
		return newQPS > oldQPS*1.5 == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(124388)
		return math.Abs(newQPS-oldQPS) > minAbsoluteChange == true
	}() == true
	updateForWPS := (newWPS < oldWPS*.5 || func() bool {
		__antithesis_instrumentation__.Notify(124389)
		return newWPS > oldWPS*1.5 == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(124390)
		return math.Abs(newWPS-oldWPS) > minAbsoluteChange == true
	}() == true

	if !updateForQPS && func() bool {
		__antithesis_instrumentation__.Notify(124391)
		return !updateForWPS == true
	}() == true {
		__antithesis_instrumentation__.Notify(124392)
		return
	} else {
		__antithesis_instrumentation__.Notify(124393)
	}
	__antithesis_instrumentation__.Notify(124382)

	var message string
	if updateForQPS && func() bool {
		__antithesis_instrumentation__.Notify(124394)
		return updateForWPS == true
	}() == true {
		__antithesis_instrumentation__.Notify(124395)
		message = "queries-per-second and writes-per-second change"
	} else {
		__antithesis_instrumentation__.Notify(124396)
		if updateForQPS {
			__antithesis_instrumentation__.Notify(124397)
			message = "queries-per-second change"
		} else {
			__antithesis_instrumentation__.Notify(124398)
			message = "writes-per-second change"
		}
	}
	__antithesis_instrumentation__.Notify(124383)

	s.asyncGossipStore(context.TODO(), message, false)
}

type VisitReplicasOption func(*storeReplicaVisitor)

func WithReplicasInOrder() VisitReplicasOption {
	__antithesis_instrumentation__.Notify(124399)
	return func(visitor *storeReplicaVisitor) {
		__antithesis_instrumentation__.Notify(124400)
		visitor.InOrder()
	}
}

func (s *Store) VisitReplicas(visitor func(*Replica) (wantMore bool), opts ...VisitReplicasOption) {
	__antithesis_instrumentation__.Notify(124401)
	v := newStoreReplicaVisitor(s)
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(124403)
		opt(v)
	}
	__antithesis_instrumentation__.Notify(124402)
	v.Visit(visitor)
}

func (s *Store) visitReplicasByKey(
	ctx context.Context,
	startKey, endKey roachpb.RKey,
	order IterationOrder,
	visitor func(context.Context, *Replica) error,
) error {
	__antithesis_instrumentation__.Notify(124404)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mu.replicasByKey.VisitKeyRange(ctx, startKey, endKey, order, func(ctx context.Context, it replicaOrPlaceholder) error {
		__antithesis_instrumentation__.Notify(124405)
		if it.repl != nil {
			__antithesis_instrumentation__.Notify(124407)
			return visitor(ctx, it.repl)
		} else {
			__antithesis_instrumentation__.Notify(124408)
		}
		__antithesis_instrumentation__.Notify(124406)
		return nil
	})
}

func (s *Store) WriteLastUpTimestamp(ctx context.Context, time hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(124409)
	ctx = s.AnnotateCtx(ctx)
	return storage.MVCCPutProto(
		ctx,
		s.engine,
		nil,
		keys.StoreLastUpKey(),
		hlc.Timestamp{},
		nil,
		&time,
	)
}

func (s *Store) ReadLastUpTimestamp(ctx context.Context) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(124410)
	var timestamp hlc.Timestamp
	ok, err := storage.MVCCGetProto(ctx, s.Engine(), keys.StoreLastUpKey(), hlc.Timestamp{},
		&timestamp, storage.MVCCGetOptions{})
	if err != nil {
		__antithesis_instrumentation__.Notify(124412)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(124413)
		if !ok {
			__antithesis_instrumentation__.Notify(124414)
			return hlc.Timestamp{}, nil
		} else {
			__antithesis_instrumentation__.Notify(124415)
		}
	}
	__antithesis_instrumentation__.Notify(124411)
	return timestamp, nil
}

func (s *Store) WriteHLCUpperBound(ctx context.Context, time int64) error {
	__antithesis_instrumentation__.Notify(124416)
	ctx = s.AnnotateCtx(ctx)
	ts := hlc.Timestamp{WallTime: time}
	batch := s.Engine().NewBatch()

	defer batch.Close()
	if err := storage.MVCCPutProto(
		ctx,
		batch,
		nil,
		keys.StoreHLCUpperBoundKey(),
		hlc.Timestamp{},
		nil,
		&ts,
	); err != nil {
		__antithesis_instrumentation__.Notify(124419)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124420)
	}
	__antithesis_instrumentation__.Notify(124417)

	if err := batch.Commit(true); err != nil {
		__antithesis_instrumentation__.Notify(124421)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124422)
	}
	__antithesis_instrumentation__.Notify(124418)
	return nil
}

func ReadHLCUpperBound(ctx context.Context, e storage.Engine) (int64, error) {
	__antithesis_instrumentation__.Notify(124423)
	var timestamp hlc.Timestamp
	ok, err := storage.MVCCGetProto(ctx, e, keys.StoreHLCUpperBoundKey(), hlc.Timestamp{},
		&timestamp, storage.MVCCGetOptions{})
	if err != nil {
		__antithesis_instrumentation__.Notify(124425)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(124426)
		if !ok {
			__antithesis_instrumentation__.Notify(124427)
			return 0, nil
		} else {
			__antithesis_instrumentation__.Notify(124428)
		}
	}
	__antithesis_instrumentation__.Notify(124424)
	return timestamp.WallTime, nil
}

func ReadMaxHLCUpperBound(ctx context.Context, engines []storage.Engine) (int64, error) {
	__antithesis_instrumentation__.Notify(124429)
	var hlcUpperBound int64
	for _, e := range engines {
		__antithesis_instrumentation__.Notify(124431)
		engineHLCUpperBound, err := ReadHLCUpperBound(ctx, e)
		if err != nil {
			__antithesis_instrumentation__.Notify(124433)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(124434)
		}
		__antithesis_instrumentation__.Notify(124432)
		if engineHLCUpperBound > hlcUpperBound {
			__antithesis_instrumentation__.Notify(124435)
			hlcUpperBound = engineHLCUpperBound
		} else {
			__antithesis_instrumentation__.Notify(124436)
		}
	}
	__antithesis_instrumentation__.Notify(124430)
	return hlcUpperBound, nil
}

func checkCanInitializeEngine(ctx context.Context, eng storage.Engine) error {
	__antithesis_instrumentation__.Notify(124437)

	ident, err := ReadStoreIdent(ctx, eng)
	if err == nil {
		__antithesis_instrumentation__.Notify(124444)
		return errors.Errorf("engine already initialized as %s", ident.String())
	} else {
		__antithesis_instrumentation__.Notify(124445)
		if !errors.HasType(err, (*NotBootstrappedError)(nil)) {
			__antithesis_instrumentation__.Notify(124446)
			return errors.Wrap(err, "unable to read store ident")
		} else {
			__antithesis_instrumentation__.Notify(124447)
		}
	}
	__antithesis_instrumentation__.Notify(124438)

	iter := eng.NewEngineIterator(storage.IterOptions{UpperBound: roachpb.KeyMax})
	defer iter.Close()
	valid, err := iter.SeekEngineKeyGE(storage.EngineKey{Key: roachpb.KeyMin})
	if !valid {
		__antithesis_instrumentation__.Notify(124448)
		if err == nil {
			__antithesis_instrumentation__.Notify(124450)
			return errors.New("no cluster version found on uninitialized engine")
		} else {
			__antithesis_instrumentation__.Notify(124451)
		}
		__antithesis_instrumentation__.Notify(124449)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124452)
	}
	__antithesis_instrumentation__.Notify(124439)
	getMVCCKey := func() (storage.MVCCKey, error) {
		__antithesis_instrumentation__.Notify(124453)
		var k storage.EngineKey
		k, err = iter.EngineKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(124456)
			return storage.MVCCKey{}, err
		} else {
			__antithesis_instrumentation__.Notify(124457)
		}
		__antithesis_instrumentation__.Notify(124454)
		if !k.IsMVCCKey() {
			__antithesis_instrumentation__.Notify(124458)
			return storage.MVCCKey{}, errors.Errorf("found non-mvcc key: %s", k)
		} else {
			__antithesis_instrumentation__.Notify(124459)
		}
		__antithesis_instrumentation__.Notify(124455)
		return k.ToMVCCKey()
	}
	__antithesis_instrumentation__.Notify(124440)
	var k storage.MVCCKey
	if k, err = getMVCCKey(); err != nil {
		__antithesis_instrumentation__.Notify(124460)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124461)
	}
	__antithesis_instrumentation__.Notify(124441)
	if !k.Key.Equal(keys.StoreClusterVersionKey()) {
		__antithesis_instrumentation__.Notify(124462)
		return errors.New("no cluster version found on uninitialized engine")
	} else {
		__antithesis_instrumentation__.Notify(124463)
	}
	__antithesis_instrumentation__.Notify(124442)
	valid, err = iter.NextEngineKey()
	for valid {
		__antithesis_instrumentation__.Notify(124464)

		if k, err = getMVCCKey(); err != nil {
			__antithesis_instrumentation__.Notify(124467)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124468)
		}
		__antithesis_instrumentation__.Notify(124465)
		if _, err := keys.DecodeStoreCachedSettingsKey(k.Key); err != nil {
			__antithesis_instrumentation__.Notify(124469)
			return errors.Errorf("engine cannot be bootstrapped, contains key:\n%s", k.String())
		} else {
			__antithesis_instrumentation__.Notify(124470)
		}
		__antithesis_instrumentation__.Notify(124466)

		valid, err = iter.NextEngineKey()
	}
	__antithesis_instrumentation__.Notify(124443)
	return err
}

func (s *Store) GetReplica(rangeID roachpb.RangeID) (*Replica, error) {
	__antithesis_instrumentation__.Notify(124471)
	if r := s.GetReplicaIfExists(rangeID); r != nil {
		__antithesis_instrumentation__.Notify(124473)
		return r, nil
	} else {
		__antithesis_instrumentation__.Notify(124474)
	}
	__antithesis_instrumentation__.Notify(124472)
	return nil, roachpb.NewRangeNotFoundError(rangeID, s.StoreID())
}

func (s *Store) GetReplicaIfExists(rangeID roachpb.RangeID) *Replica {
	__antithesis_instrumentation__.Notify(124475)
	if repl, ok := s.mu.replicasByRangeID.Load(rangeID); ok {
		__antithesis_instrumentation__.Notify(124477)
		return repl
	} else {
		__antithesis_instrumentation__.Notify(124478)
	}
	__antithesis_instrumentation__.Notify(124476)
	return nil
}

func (s *Store) LookupReplica(key roachpb.RKey) *Replica {
	__antithesis_instrumentation__.Notify(124479)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mu.replicasByKey.LookupReplica(context.Background(), key)
}

func (s *Store) lookupPrecedingReplica(key roachpb.RKey) *Replica {
	__antithesis_instrumentation__.Notify(124480)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mu.replicasByKey.LookupPrecedingReplica(context.Background(), key)
}

func (s *Store) getOverlappingKeyRangeLocked(
	rngDesc *roachpb.RangeDescriptor,
) replicaOrPlaceholder {
	__antithesis_instrumentation__.Notify(124481)
	var it replicaOrPlaceholder
	if err := s.mu.replicasByKey.VisitKeyRange(
		context.Background(), rngDesc.StartKey, rngDesc.EndKey, AscendingKeyOrder,
		func(ctx context.Context, iit replicaOrPlaceholder) error {
			__antithesis_instrumentation__.Notify(124483)
			it = iit
			return iterutil.StopIteration()
		}); err != nil {
		__antithesis_instrumentation__.Notify(124484)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(124485)
	}
	__antithesis_instrumentation__.Notify(124482)

	return it
}

func (s *Store) RaftStatus(rangeID roachpb.RangeID) *raft.Status {
	__antithesis_instrumentation__.Notify(124486)
	if repl, ok := s.mu.replicasByRangeID.Load(rangeID); ok {
		__antithesis_instrumentation__.Notify(124488)
		return repl.RaftStatus()
	} else {
		__antithesis_instrumentation__.Notify(124489)
	}
	__antithesis_instrumentation__.Notify(124487)
	return nil
}

func (s *Store) ClusterID() uuid.UUID {
	__antithesis_instrumentation__.Notify(124490)
	return s.Ident.ClusterID
}

func (s *Store) NodeID() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(124491)
	return s.Ident.NodeID
}

func (s *Store) StoreID() roachpb.StoreID {
	__antithesis_instrumentation__.Notify(124492)
	return s.Ident.StoreID
}

func (s *Store) Clock() *hlc.Clock { __antithesis_instrumentation__.Notify(124493); return s.cfg.Clock }

func (s *Store) Engine() storage.Engine {
	__antithesis_instrumentation__.Notify(124494)
	return s.engine
}

func (s *Store) DB() *kv.DB { __antithesis_instrumentation__.Notify(124495); return s.cfg.DB }

func (s *Store) Gossip() *gossip.Gossip {
	__antithesis_instrumentation__.Notify(124496)
	return s.cfg.Gossip
}

func (s *Store) Stopper() *stop.Stopper {
	__antithesis_instrumentation__.Notify(124497)
	return s.stopper
}

func (s *Store) TestingKnobs() *StoreTestingKnobs {
	__antithesis_instrumentation__.Notify(124498)
	return &s.cfg.TestingKnobs
}

func (s *Store) IsDraining() bool {
	__antithesis_instrumentation__.Notify(124499)
	return s.draining.Load().(bool)
}

func (s *Store) AllocateRangeID(ctx context.Context) (roachpb.RangeID, error) {
	__antithesis_instrumentation__.Notify(124500)
	id, err := s.rangeIDAlloc.Allocate(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(124502)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(124503)
	}
	__antithesis_instrumentation__.Notify(124501)
	return roachpb.RangeID(id), nil
}

func (s *Store) Attrs() roachpb.Attributes {
	__antithesis_instrumentation__.Notify(124504)
	return s.engine.Attrs()
}

func (s *Store) Properties() roachpb.StoreProperties {
	__antithesis_instrumentation__.Notify(124505)
	return s.engine.Properties()
}

func (s *Store) Capacity(ctx context.Context, useCached bool) (roachpb.StoreCapacity, error) {
	__antithesis_instrumentation__.Notify(124506)
	if useCached {
		__antithesis_instrumentation__.Notify(124510)
		s.cachedCapacity.Lock()
		capacity := s.cachedCapacity.StoreCapacity
		s.cachedCapacity.Unlock()
		if capacity != (roachpb.StoreCapacity{}) {
			__antithesis_instrumentation__.Notify(124511)
			return capacity, nil
		} else {
			__antithesis_instrumentation__.Notify(124512)
		}
	} else {
		__antithesis_instrumentation__.Notify(124513)
	}
	__antithesis_instrumentation__.Notify(124507)

	capacity, err := s.engine.Capacity()
	if err != nil {
		__antithesis_instrumentation__.Notify(124514)
		return capacity, err
	} else {
		__antithesis_instrumentation__.Notify(124515)
	}
	__antithesis_instrumentation__.Notify(124508)

	now := s.cfg.Clock.NowAsClockTimestamp()
	var leaseCount int32
	var rangeCount int32
	var logicalBytes int64
	var totalQueriesPerSecond float64
	var totalWritesPerSecond float64
	replicaCount := s.metrics.ReplicaCount.Value()
	bytesPerReplica := make([]float64, 0, replicaCount)
	writesPerReplica := make([]float64, 0, replicaCount)
	rankingsAccumulator := s.replRankings.newAccumulator()
	newStoreReplicaVisitor(s).Visit(func(r *Replica) bool {
		__antithesis_instrumentation__.Notify(124516)
		rangeCount++
		if r.OwnsValidLease(ctx, now) {
			__antithesis_instrumentation__.Notify(124520)
			leaseCount++
		} else {
			__antithesis_instrumentation__.Notify(124521)
		}
		__antithesis_instrumentation__.Notify(124517)
		mvccStats := r.GetMVCCStats()
		logicalBytes += mvccStats.Total()
		bytesPerReplica = append(bytesPerReplica, float64(mvccStats.Total()))

		var qps float64
		if avgQPS, dur := r.leaseholderStats.avgQPS(); dur >= MinStatsDuration {
			__antithesis_instrumentation__.Notify(124522)
			qps = avgQPS
			totalQueriesPerSecond += avgQPS

		} else {
			__antithesis_instrumentation__.Notify(124523)
		}
		__antithesis_instrumentation__.Notify(124518)
		if wps, dur := r.writeStats.avgQPS(); dur >= MinStatsDuration {
			__antithesis_instrumentation__.Notify(124524)
			totalWritesPerSecond += wps
			writesPerReplica = append(writesPerReplica, wps)
		} else {
			__antithesis_instrumentation__.Notify(124525)
		}
		__antithesis_instrumentation__.Notify(124519)
		rankingsAccumulator.addReplica(replicaWithStats{
			repl: r,
			qps:  qps,
		})
		return true
	})
	__antithesis_instrumentation__.Notify(124509)
	capacity.RangeCount = rangeCount
	capacity.LeaseCount = leaseCount
	capacity.LogicalBytes = logicalBytes
	capacity.QueriesPerSecond = totalQueriesPerSecond
	capacity.WritesPerSecond = totalWritesPerSecond
	capacity.L0Sublevels = s.metrics.RdbL0Sublevels.Value()
	capacity.BytesPerReplica = roachpb.PercentilesFromData(bytesPerReplica)
	capacity.WritesPerReplica = roachpb.PercentilesFromData(writesPerReplica)
	s.recordNewPerSecondStats(totalQueriesPerSecond, totalWritesPerSecond)
	s.replRankings.update(rankingsAccumulator)

	s.cachedCapacity.Lock()
	s.cachedCapacity.StoreCapacity = capacity
	s.cachedCapacity.Unlock()

	return capacity, nil
}

func (s *Store) ReplicaCount() int {
	__antithesis_instrumentation__.Notify(124526)
	var count int
	s.mu.replicasByRangeID.Range(func(*Replica) {
		__antithesis_instrumentation__.Notify(124528)
		count++
	})
	__antithesis_instrumentation__.Notify(124527)
	return count
}

func (s *Store) Registry() *metric.Registry {
	__antithesis_instrumentation__.Notify(124529)
	return s.metrics.registry
}

func (s *Store) Metrics() *StoreMetrics {
	__antithesis_instrumentation__.Notify(124530)
	return s.metrics
}

func (s *Store) ReplicateQueueMetrics() ReplicateQueueMetrics {
	__antithesis_instrumentation__.Notify(124531)
	return s.replicateQueue.metrics
}

func (s *Store) Descriptor(ctx context.Context, useCached bool) (*roachpb.StoreDescriptor, error) {
	__antithesis_instrumentation__.Notify(124532)
	capacity, err := s.Capacity(ctx, useCached)
	if err != nil {
		__antithesis_instrumentation__.Notify(124534)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(124535)
	}
	__antithesis_instrumentation__.Notify(124533)

	return &roachpb.StoreDescriptor{
		StoreID:    s.Ident.StoreID,
		Attrs:      s.Attrs(),
		Node:       *s.nodeDesc,
		Capacity:   capacity,
		Properties: s.Properties(),
	}, nil
}

func (s *Store) RangeFeed(
	args *roachpb.RangeFeedRequest, stream roachpb.Internal_RangeFeedServer,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(124536)

	if filter := s.TestingKnobs().TestingRangefeedFilter; filter != nil {
		__antithesis_instrumentation__.Notify(124541)
		if pErr := filter(args, stream); pErr != nil {
			__antithesis_instrumentation__.Notify(124542)
			return pErr
		} else {
			__antithesis_instrumentation__.Notify(124543)
		}
	} else {
		__antithesis_instrumentation__.Notify(124544)
	}
	__antithesis_instrumentation__.Notify(124537)

	if err := verifyKeys(args.Span.Key, args.Span.EndKey, true); err != nil {
		__antithesis_instrumentation__.Notify(124545)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(124546)
	}
	__antithesis_instrumentation__.Notify(124538)

	repl, err := s.GetReplica(args.RangeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(124547)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(124548)
	}
	__antithesis_instrumentation__.Notify(124539)
	if !repl.IsInitialized() {
		__antithesis_instrumentation__.Notify(124549)

		return roachpb.NewError(roachpb.NewRangeNotFoundError(args.RangeID, s.StoreID()))
	} else {
		__antithesis_instrumentation__.Notify(124550)
	}
	__antithesis_instrumentation__.Notify(124540)
	return repl.RangeFeed(args, stream)
}

func (s *Store) updateReplicationGauges(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(124551)
	var (
		raftLeaderCount               int64
		leaseHolderCount              int64
		leaseExpirationCount          int64
		leaseEpochCount               int64
		raftLeaderNotLeaseHolderCount int64
		quiescentCount                int64
		uninitializedCount            int64
		averageQueriesPerSecond       float64
		averageWritesPerSecond        float64

		rangeCount                int64
		unavailableRangeCount     int64
		underreplicatedRangeCount int64
		overreplicatedRangeCount  int64
		behindCount               int64

		locks                          int64
		totalLockHoldDurationNanos     int64
		maxLockHoldDurationNanos       int64
		locksWithWaitQueues            int64
		lockWaitQueueWaiters           int64
		totalLockWaitDurationNanos     int64
		maxLockWaitDurationNanos       int64
		maxLockWaitQueueWaitersForLock int64

		minMaxClosedTS hlc.Timestamp
	)

	now := s.cfg.Clock.NowAsClockTimestamp()
	var livenessMap liveness.IsLiveMap
	if s.cfg.NodeLiveness != nil {
		__antithesis_instrumentation__.Notify(124557)
		livenessMap = s.cfg.NodeLiveness.GetIsLiveMap()
	} else {
		__antithesis_instrumentation__.Notify(124558)
	}
	__antithesis_instrumentation__.Notify(124552)
	clusterNodes := s.ClusterNodeCount()

	s.mu.RLock()
	uninitializedCount = int64(len(s.mu.uninitReplicas))
	s.mu.RUnlock()

	newStoreReplicaVisitor(s).Visit(func(rep *Replica) bool {
		__antithesis_instrumentation__.Notify(124559)
		metrics := rep.Metrics(ctx, now, livenessMap, clusterNodes)
		if metrics.Leader {
			__antithesis_instrumentation__.Notify(124570)
			raftLeaderCount++
			if metrics.LeaseValid && func() bool {
				__antithesis_instrumentation__.Notify(124571)
				return !metrics.Leaseholder == true
			}() == true {
				__antithesis_instrumentation__.Notify(124572)
				raftLeaderNotLeaseHolderCount++
			} else {
				__antithesis_instrumentation__.Notify(124573)
			}
		} else {
			__antithesis_instrumentation__.Notify(124574)
		}
		__antithesis_instrumentation__.Notify(124560)
		if metrics.Leaseholder {
			__antithesis_instrumentation__.Notify(124575)
			leaseHolderCount++
			switch metrics.LeaseType {
			case roachpb.LeaseNone:
				__antithesis_instrumentation__.Notify(124576)
			case roachpb.LeaseExpiration:
				__antithesis_instrumentation__.Notify(124577)
				leaseExpirationCount++
			case roachpb.LeaseEpoch:
				__antithesis_instrumentation__.Notify(124578)
				leaseEpochCount++
			default:
				__antithesis_instrumentation__.Notify(124579)
			}
		} else {
			__antithesis_instrumentation__.Notify(124580)
		}
		__antithesis_instrumentation__.Notify(124561)
		if metrics.Quiescent {
			__antithesis_instrumentation__.Notify(124581)
			quiescentCount++
		} else {
			__antithesis_instrumentation__.Notify(124582)
		}
		__antithesis_instrumentation__.Notify(124562)
		if metrics.RangeCounter {
			__antithesis_instrumentation__.Notify(124583)
			rangeCount++
			if metrics.Unavailable {
				__antithesis_instrumentation__.Notify(124586)
				unavailableRangeCount++
			} else {
				__antithesis_instrumentation__.Notify(124587)
			}
			__antithesis_instrumentation__.Notify(124584)
			if metrics.Underreplicated {
				__antithesis_instrumentation__.Notify(124588)
				underreplicatedRangeCount++
			} else {
				__antithesis_instrumentation__.Notify(124589)
			}
			__antithesis_instrumentation__.Notify(124585)
			if metrics.Overreplicated {
				__antithesis_instrumentation__.Notify(124590)
				overreplicatedRangeCount++
			} else {
				__antithesis_instrumentation__.Notify(124591)
			}
		} else {
			__antithesis_instrumentation__.Notify(124592)
		}
		__antithesis_instrumentation__.Notify(124563)
		behindCount += metrics.BehindCount
		if qps, dur := rep.leaseholderStats.avgQPS(); dur >= MinStatsDuration {
			__antithesis_instrumentation__.Notify(124593)
			averageQueriesPerSecond += qps
		} else {
			__antithesis_instrumentation__.Notify(124594)
		}
		__antithesis_instrumentation__.Notify(124564)
		if wps, dur := rep.writeStats.avgQPS(); dur >= MinStatsDuration {
			__antithesis_instrumentation__.Notify(124595)
			averageWritesPerSecond += wps
		} else {
			__antithesis_instrumentation__.Notify(124596)
		}
		__antithesis_instrumentation__.Notify(124565)
		locks += metrics.LockTableMetrics.Locks
		totalLockHoldDurationNanos += metrics.LockTableMetrics.TotalLockHoldDurationNanos
		locksWithWaitQueues += metrics.LockTableMetrics.LocksWithWaitQueues
		lockWaitQueueWaiters += metrics.LockTableMetrics.Waiters
		totalLockWaitDurationNanos += metrics.LockTableMetrics.TotalWaitDurationNanos
		if w := metrics.LockTableMetrics.TopKLocksByWaiters[0].Waiters; w > maxLockWaitQueueWaitersForLock {
			__antithesis_instrumentation__.Notify(124597)
			maxLockWaitQueueWaitersForLock = w
		} else {
			__antithesis_instrumentation__.Notify(124598)
		}
		__antithesis_instrumentation__.Notify(124566)
		if w := metrics.LockTableMetrics.TopKLocksByHoldDuration[0].HoldDurationNanos; w > maxLockHoldDurationNanos {
			__antithesis_instrumentation__.Notify(124599)
			maxLockHoldDurationNanos = w
		} else {
			__antithesis_instrumentation__.Notify(124600)
		}
		__antithesis_instrumentation__.Notify(124567)
		if w := metrics.LockTableMetrics.TopKLocksByWaitDuration[0].MaxWaitDurationNanos; w > maxLockWaitDurationNanos {
			__antithesis_instrumentation__.Notify(124601)
			maxLockWaitDurationNanos = w
		} else {
			__antithesis_instrumentation__.Notify(124602)
		}
		__antithesis_instrumentation__.Notify(124568)
		mc := rep.GetClosedTimestamp(ctx)
		if minMaxClosedTS.IsEmpty() || func() bool {
			__antithesis_instrumentation__.Notify(124603)
			return mc.Less(minMaxClosedTS) == true
		}() == true {
			__antithesis_instrumentation__.Notify(124604)
			minMaxClosedTS = mc
		} else {
			__antithesis_instrumentation__.Notify(124605)
		}
		__antithesis_instrumentation__.Notify(124569)
		return true
	})
	__antithesis_instrumentation__.Notify(124553)

	s.metrics.RaftLeaderCount.Update(raftLeaderCount)
	s.metrics.RaftLeaderNotLeaseHolderCount.Update(raftLeaderNotLeaseHolderCount)
	s.metrics.LeaseHolderCount.Update(leaseHolderCount)
	s.metrics.LeaseExpirationCount.Update(leaseExpirationCount)
	s.metrics.LeaseEpochCount.Update(leaseEpochCount)
	s.metrics.QuiescentCount.Update(quiescentCount)
	s.metrics.UninitializedCount.Update(uninitializedCount)
	s.metrics.AverageQueriesPerSecond.Update(averageQueriesPerSecond)
	s.metrics.AverageWritesPerSecond.Update(averageWritesPerSecond)
	s.recordNewPerSecondStats(averageQueriesPerSecond, averageWritesPerSecond)

	s.metrics.RangeCount.Update(rangeCount)
	s.metrics.UnavailableRangeCount.Update(unavailableRangeCount)
	s.metrics.UnderReplicatedRangeCount.Update(underreplicatedRangeCount)
	s.metrics.OverReplicatedRangeCount.Update(overreplicatedRangeCount)
	s.metrics.RaftLogFollowerBehindCount.Update(behindCount)

	var averageLockHoldDurationNanos int64
	var averageLockWaitDurationNanos int64
	if locks > 0 {
		__antithesis_instrumentation__.Notify(124606)
		averageLockHoldDurationNanos = totalLockHoldDurationNanos / locks
	} else {
		__antithesis_instrumentation__.Notify(124607)
	}
	__antithesis_instrumentation__.Notify(124554)
	if lockWaitQueueWaiters > 0 {
		__antithesis_instrumentation__.Notify(124608)
		averageLockWaitDurationNanos = totalLockWaitDurationNanos / lockWaitQueueWaiters
	} else {
		__antithesis_instrumentation__.Notify(124609)
	}
	__antithesis_instrumentation__.Notify(124555)

	s.metrics.Locks.Update(locks)
	s.metrics.AverageLockHoldDurationNanos.Update(averageLockHoldDurationNanos)
	s.metrics.MaxLockHoldDurationNanos.Update(maxLockHoldDurationNanos)
	s.metrics.LocksWithWaitQueues.Update(locksWithWaitQueues)
	s.metrics.LockWaitQueueWaiters.Update(lockWaitQueueWaiters)
	s.metrics.AverageLockWaitDurationNanos.Update(averageLockWaitDurationNanos)
	s.metrics.MaxLockWaitDurationNanos.Update(maxLockWaitDurationNanos)
	s.metrics.MaxLockWaitQueueWaitersForLock.Update(maxLockWaitQueueWaitersForLock)

	if !minMaxClosedTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(124610)
		nanos := timeutil.Since(minMaxClosedTS.GoTime()).Nanoseconds()
		s.metrics.ClosedTimestampMaxBehindNanos.Update(nanos)
	} else {
		__antithesis_instrumentation__.Notify(124611)
	}
	__antithesis_instrumentation__.Notify(124556)

	s.metrics.RaftEnqueuedPending.Update(s.cfg.Transport.queuedMessageCount())

	return nil
}

func (s *Store) checkpoint(ctx context.Context, tag string) (string, error) {
	__antithesis_instrumentation__.Notify(124612)
	checkpointBase := filepath.Join(s.engine.GetAuxiliaryDir(), "checkpoints")
	_ = s.engine.MkdirAll(checkpointBase)

	checkpointDir := filepath.Join(checkpointBase, tag)
	if err := s.engine.CreateCheckpoint(checkpointDir); err != nil {
		__antithesis_instrumentation__.Notify(124614)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(124615)
	}
	__antithesis_instrumentation__.Notify(124613)

	return checkpointDir, nil
}

func (s *Store) ComputeMetrics(ctx context.Context, tick int) error {
	__antithesis_instrumentation__.Notify(124616)
	ctx = s.AnnotateCtx(ctx)
	if err := s.updateCapacityGauges(ctx); err != nil {
		__antithesis_instrumentation__.Notify(124621)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124622)
	}
	__antithesis_instrumentation__.Notify(124617)
	if err := s.updateReplicationGauges(ctx); err != nil {
		__antithesis_instrumentation__.Notify(124623)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124624)
	}
	__antithesis_instrumentation__.Notify(124618)

	m := s.engine.GetMetrics()
	s.metrics.updateEngineMetrics(m)

	envStats, err := s.engine.GetEnvStats()
	if err != nil {
		__antithesis_instrumentation__.Notify(124625)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124626)
	}
	__antithesis_instrumentation__.Notify(124619)
	s.metrics.updateEnvStats(*envStats)

	if tick%logSSTInfoTicks == 1 {
		__antithesis_instrumentation__.Notify(124627)

		log.Infof(ctx, "\n%s", m.Metrics)
	} else {
		__antithesis_instrumentation__.Notify(124628)
	}
	__antithesis_instrumentation__.Notify(124620)
	return nil
}

func (s *Store) ClusterNodeCount() int {
	__antithesis_instrumentation__.Notify(124629)
	return s.cfg.StorePool.ClusterNodeCount()
}

type HotReplicaInfo struct {
	Desc *roachpb.RangeDescriptor
	QPS  float64
}

func (s *Store) HottestReplicas() []HotReplicaInfo {
	__antithesis_instrumentation__.Notify(124630)
	topQPS := s.replRankings.topQPS()
	hotRepls := make([]HotReplicaInfo, len(topQPS))
	for i := range topQPS {
		__antithesis_instrumentation__.Notify(124632)
		hotRepls[i].Desc = topQPS[i].repl.Desc()
		hotRepls[i].QPS = topQPS[i].qps
	}
	__antithesis_instrumentation__.Notify(124631)
	return hotRepls
}

type StoreKeySpanStats struct {
	ReplicaCount         int
	MVCC                 enginepb.MVCCStats
	ApproximateDiskBytes uint64
}

func (s *Store) ComputeStatsForKeySpan(startKey, endKey roachpb.RKey) (StoreKeySpanStats, error) {
	__antithesis_instrumentation__.Notify(124633)
	var result StoreKeySpanStats

	newStoreReplicaVisitor(s).Visit(func(repl *Replica) bool {
		__antithesis_instrumentation__.Notify(124635)
		desc := repl.Desc()
		if bytes.Compare(startKey, desc.EndKey) >= 0 || func() bool {
			__antithesis_instrumentation__.Notify(124637)
			return bytes.Compare(desc.StartKey, endKey) >= 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(124638)
			return true
		} else {
			__antithesis_instrumentation__.Notify(124639)
		}
		__antithesis_instrumentation__.Notify(124636)
		result.MVCC.Add(repl.GetMVCCStats())
		result.ReplicaCount++
		return true
	})
	__antithesis_instrumentation__.Notify(124634)

	var err error
	result.ApproximateDiskBytes, err = s.engine.ApproximateDiskBytes(startKey.AsRawKey(), endKey.AsRawKey())
	return result, err
}

func (s *Store) AllocatorDryRun(ctx context.Context, repl *Replica) (tracing.Recording, error) {
	__antithesis_instrumentation__.Notify(124640)
	ctx, collectAndFinish := tracing.ContextWithRecordingSpan(ctx, s.cfg.AmbientCtx.Tracer, "allocator dry run")
	defer collectAndFinish()
	canTransferLease := func(ctx context.Context, repl *Replica) bool {
		__antithesis_instrumentation__.Notify(124643)
		return true
	}
	__antithesis_instrumentation__.Notify(124641)
	_, err := s.replicateQueue.processOneChange(
		ctx, repl, canTransferLease, false, true,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(124644)
		log.Eventf(ctx, "error simulating allocator on replica %s: %s", repl, err)
	} else {
		__antithesis_instrumentation__.Notify(124645)
	}
	__antithesis_instrumentation__.Notify(124642)
	return collectAndFinish(), nil
}

func (s *Store) ManuallyEnqueue(
	ctx context.Context, queueName string, repl *Replica, skipShouldQueue bool,
) (recording tracing.Recording, processError error, enqueueError error) {
	__antithesis_instrumentation__.Notify(124646)
	ctx = repl.AnnotateCtx(ctx)

	if !repl.IsInitialized() {
		__antithesis_instrumentation__.Notify(124653)
		return nil, nil, errors.Errorf("not enqueueing uninitialized replica %s", repl)
	} else {
		__antithesis_instrumentation__.Notify(124654)
	}
	__antithesis_instrumentation__.Notify(124647)

	var queue queueImpl
	var needsLease bool
	for _, replicaQueue := range s.scanner.queues {
		__antithesis_instrumentation__.Notify(124655)
		if strings.EqualFold(replicaQueue.Name(), queueName) {
			__antithesis_instrumentation__.Notify(124656)
			queue = replicaQueue.(queueImpl)
			needsLease = replicaQueue.NeedsLease()
		} else {
			__antithesis_instrumentation__.Notify(124657)
		}
	}
	__antithesis_instrumentation__.Notify(124648)
	if queue == nil {
		__antithesis_instrumentation__.Notify(124658)
		return nil, nil, errors.Errorf("unknown queue type %q", queueName)
	} else {
		__antithesis_instrumentation__.Notify(124659)
	}
	__antithesis_instrumentation__.Notify(124649)

	confReader, err := s.GetConfReader(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(124660)
		return nil, nil, errors.Wrap(err,
			"unable to retrieve conf reader, cannot run queue; make sure "+
				"the cluster has been initialized and all nodes connected to it")
	} else {
		__antithesis_instrumentation__.Notify(124661)
	}
	__antithesis_instrumentation__.Notify(124650)

	if needsLease {
		__antithesis_instrumentation__.Notify(124662)
		hasLease, pErr := repl.getLeaseForGossip(ctx)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(124664)
			return nil, nil, pErr.GoError()
		} else {
			__antithesis_instrumentation__.Notify(124665)
		}
		__antithesis_instrumentation__.Notify(124663)
		if !hasLease {
			__antithesis_instrumentation__.Notify(124666)
			return nil, errors.Newf("replica %v does not have the range lease", repl), nil
		} else {
			__antithesis_instrumentation__.Notify(124667)
		}
	} else {
		__antithesis_instrumentation__.Notify(124668)
	}
	__antithesis_instrumentation__.Notify(124651)

	ctx, collectAndFinish := tracing.ContextWithRecordingSpan(
		ctx, s.cfg.AmbientCtx.Tracer, fmt.Sprintf("manual %s queue run", queueName))
	defer collectAndFinish()

	if !skipShouldQueue {
		__antithesis_instrumentation__.Notify(124669)
		log.Eventf(ctx, "running %s.shouldQueue", queueName)
		shouldQueue, priority := queue.shouldQueue(ctx, s.cfg.Clock.NowAsClockTimestamp(), repl, confReader)
		log.Eventf(ctx, "shouldQueue=%v, priority=%f", shouldQueue, priority)
		if !shouldQueue {
			__antithesis_instrumentation__.Notify(124670)
			return collectAndFinish(), nil, nil
		} else {
			__antithesis_instrumentation__.Notify(124671)
		}
	} else {
		__antithesis_instrumentation__.Notify(124672)
	}
	__antithesis_instrumentation__.Notify(124652)

	log.Eventf(ctx, "running %s.process", queueName)
	processed, processErr := queue.process(ctx, repl, confReader)
	log.Eventf(ctx, "processed: %t (err: %v)", processed, processErr)
	return collectAndFinish(), processErr, nil
}

func (s *Store) PurgeOutdatedReplicas(ctx context.Context, version roachpb.Version) error {
	__antithesis_instrumentation__.Notify(124673)
	if interceptor := s.TestingKnobs().PurgeOutdatedReplicasInterceptor; interceptor != nil {
		__antithesis_instrumentation__.Notify(124676)
		interceptor()
	} else {
		__antithesis_instrumentation__.Notify(124677)
	}
	__antithesis_instrumentation__.Notify(124674)

	qp := quotapool.NewIntPool("purge-outdated-replicas", 50)
	g := ctxgroup.WithContext(ctx)
	s.VisitReplicas(func(repl *Replica) (wantMore bool) {
		__antithesis_instrumentation__.Notify(124678)
		if !repl.Version().Less(version) {
			__antithesis_instrumentation__.Notify(124682)

			return true
		} else {
			__antithesis_instrumentation__.Notify(124683)
		}
		__antithesis_instrumentation__.Notify(124679)

		alloc, err := qp.Acquire(ctx, 1)
		if err != nil {
			__antithesis_instrumentation__.Notify(124684)
			g.GoCtx(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(124686)
				return err
			})
			__antithesis_instrumentation__.Notify(124685)
			return false
		} else {
			__antithesis_instrumentation__.Notify(124687)
		}
		__antithesis_instrumentation__.Notify(124680)

		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(124688)
			defer alloc.Release()

			processed, err := s.replicaGCQueue.process(ctx, repl, nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(124691)
				return errors.Wrapf(err, "on %s", repl.Desc())
			} else {
				__antithesis_instrumentation__.Notify(124692)
			}
			__antithesis_instrumentation__.Notify(124689)
			if !processed {
				__antithesis_instrumentation__.Notify(124693)

				return errors.Newf("unable to gc %s", repl.Desc())
			} else {
				__antithesis_instrumentation__.Notify(124694)
			}
			__antithesis_instrumentation__.Notify(124690)
			return nil
		})
		__antithesis_instrumentation__.Notify(124681)

		return true
	})
	__antithesis_instrumentation__.Notify(124675)

	return g.Wait()
}

func (s *Store) WaitForSpanConfigSubscription(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(124695)
	if s.cfg.SpanConfigsDisabled {
		__antithesis_instrumentation__.Notify(124698)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(124699)
	}
	__antithesis_instrumentation__.Notify(124696)

	for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
		__antithesis_instrumentation__.Notify(124700)
		if !s.cfg.SpanConfigSubscriber.LastUpdated().IsEmpty() {
			__antithesis_instrumentation__.Notify(124702)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(124703)
		}
		__antithesis_instrumentation__.Notify(124701)

		log.Warningf(ctx, "waiting for span config subscription...")
		continue
	}
	__antithesis_instrumentation__.Notify(124697)

	return errors.Newf("unable to subscribe to span configs")
}

func (s *Store) registerLeaseholder(
	ctx context.Context, r *Replica, leaseSeq roachpb.LeaseSequence,
) {
	__antithesis_instrumentation__.Notify(124704)
	if s.ctSender != nil {
		__antithesis_instrumentation__.Notify(124705)
		s.ctSender.RegisterLeaseholder(ctx, r, leaseSeq)
	} else {
		__antithesis_instrumentation__.Notify(124706)
	}
}

func (s *Store) unregisterLeaseholder(ctx context.Context, r *Replica) {
	__antithesis_instrumentation__.Notify(124707)
	s.unregisterLeaseholderByID(ctx, r.RangeID)
}

func (s *Store) unregisterLeaseholderByID(ctx context.Context, rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(124708)
	if s.ctSender != nil {
		__antithesis_instrumentation__.Notify(124709)
		s.ctSender.UnregisterLeaseholder(ctx, s.StoreID(), rangeID)
	} else {
		__antithesis_instrumentation__.Notify(124710)
	}
}

func (s *Store) getRootMemoryMonitorForKV() *mon.BytesMonitor {
	__antithesis_instrumentation__.Notify(124711)
	return s.cfg.KVMemoryMonitor
}

type storeForTruncatorImpl Store

var _ storeForTruncator = &storeForTruncatorImpl{}

func (s *storeForTruncatorImpl) acquireReplicaForTruncator(
	rangeID roachpb.RangeID,
) replicaForTruncator {
	__antithesis_instrumentation__.Notify(124712)
	r, err := (*Store)(s).GetReplica(rangeID)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(124715)
		return r == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(124716)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(124717)
	}
	__antithesis_instrumentation__.Notify(124713)
	r.raftMu.Lock()
	if isAlive := func() bool {
		__antithesis_instrumentation__.Notify(124718)
		r.mu.Lock()
		defer r.mu.Unlock()
		return r.mu.destroyStatus.IsAlive()
	}(); !isAlive {
		__antithesis_instrumentation__.Notify(124719)
		r.raftMu.Unlock()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(124720)
	}
	__antithesis_instrumentation__.Notify(124714)
	return (*raftTruncatorReplica)(r)
}

func (s *storeForTruncatorImpl) releaseReplicaForTruncator(r replicaForTruncator) {
	__antithesis_instrumentation__.Notify(124721)
	replica := r.(*raftTruncatorReplica)
	replica.raftMu.Unlock()
}

func (s *storeForTruncatorImpl) getEngine() storage.Engine {
	__antithesis_instrumentation__.Notify(124722)
	return (*Store)(s).engine
}

func WriteClusterVersion(
	ctx context.Context, eng storage.Engine, cv clusterversion.ClusterVersion,
) error {
	__antithesis_instrumentation__.Notify(124723)
	err := storage.MVCCPutProto(ctx, eng, nil, keys.StoreClusterVersionKey(), hlc.Timestamp{}, nil, &cv)
	if err != nil {
		__antithesis_instrumentation__.Notify(124725)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124726)
	}
	__antithesis_instrumentation__.Notify(124724)

	return eng.SetMinVersion(cv.Version)
}

func ReadClusterVersion(
	ctx context.Context, reader storage.Reader,
) (clusterversion.ClusterVersion, error) {
	__antithesis_instrumentation__.Notify(124727)
	var cv clusterversion.ClusterVersion
	_, err := storage.MVCCGetProto(ctx, reader, keys.StoreClusterVersionKey(), hlc.Timestamp{},
		&cv, storage.MVCCGetOptions{})
	return cv, err
}

func init() {
	tracing.RegisterTagRemapping("s", "store")
}

func min(a, b int) int {
	__antithesis_instrumentation__.Notify(124728)
	if a < b {
		__antithesis_instrumentation__.Notify(124730)
		return a
	} else {
		__antithesis_instrumentation__.Notify(124731)
	}
	__antithesis_instrumentation__.Notify(124729)
	return b
}

type KVAdmissionController interface {
	AdmitKVWork(
		ctx context.Context, tenantID roachpb.TenantID, ba *roachpb.BatchRequest,
	) (handle interface{}, err error)

	AdmittedKVWorkDone(handle interface{})

	SetTenantWeightProvider(provider TenantWeightProvider, stopper *stop.Stopper)
}

type TenantWeightProvider interface {
	GetTenantWeights() TenantWeights
}

type TenantWeights struct {
	Node map[uint64]uint32

	Stores []TenantWeightsForStore
}

type TenantWeightsForStore struct {
	roachpb.StoreID

	Weights map[uint64]uint32
}

type KVAdmissionControllerImpl struct {
	kvAdmissionQ     *admission.WorkQueue
	storeGrantCoords *admission.StoreGrantCoordinators
	settings         *cluster.Settings
}

var _ KVAdmissionController = KVAdmissionControllerImpl{}

type admissionHandle struct {
	tenantID                           roachpb.TenantID
	callAdmittedWorkDoneOnKVAdmissionQ bool
	storeAdmissionQ                    *admission.WorkQueue
}

func MakeKVAdmissionController(
	kvAdmissionQ *admission.WorkQueue,
	storeGrantCoords *admission.StoreGrantCoordinators,
	settings *cluster.Settings,
) KVAdmissionController {
	__antithesis_instrumentation__.Notify(124732)
	return KVAdmissionControllerImpl{
		kvAdmissionQ:     kvAdmissionQ,
		storeGrantCoords: storeGrantCoords,
		settings:         settings,
	}
}

func (n KVAdmissionControllerImpl) AdmitKVWork(
	ctx context.Context, tenantID roachpb.TenantID, ba *roachpb.BatchRequest,
) (handle interface{}, err error) {
	__antithesis_instrumentation__.Notify(124733)
	ah := admissionHandle{tenantID: tenantID}
	if n.kvAdmissionQ != nil {
		__antithesis_instrumentation__.Notify(124735)
		bypassAdmission := ba.IsAdmin()
		source := ba.AdmissionHeader.Source
		if !roachpb.IsSystemTenantID(tenantID.ToUint64()) {
			__antithesis_instrumentation__.Notify(124741)

			bypassAdmission = false
			source = roachpb.AdmissionHeader_FROM_SQL
		} else {
			__antithesis_instrumentation__.Notify(124742)
		}
		__antithesis_instrumentation__.Notify(124736)
		if source == roachpb.AdmissionHeader_OTHER {
			__antithesis_instrumentation__.Notify(124743)
			bypassAdmission = true
		} else {
			__antithesis_instrumentation__.Notify(124744)
		}
		__antithesis_instrumentation__.Notify(124737)
		createTime := ba.AdmissionHeader.CreateTime
		if !bypassAdmission && func() bool {
			__antithesis_instrumentation__.Notify(124745)
			return createTime == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(124746)

			createTime = timeutil.Now().UnixNano()
		} else {
			__antithesis_instrumentation__.Notify(124747)
		}
		__antithesis_instrumentation__.Notify(124738)
		admissionInfo := admission.WorkInfo{
			TenantID:        tenantID,
			Priority:        admission.WorkPriority(ba.AdmissionHeader.Priority),
			CreateTime:      createTime,
			BypassAdmission: bypassAdmission,
		}
		var err error

		if ba.IsWrite() && func() bool {
			__antithesis_instrumentation__.Notify(124748)
			return !ba.IsSingleHeartbeatTxnRequest() == true
		}() == true {
			__antithesis_instrumentation__.Notify(124749)
			ah.storeAdmissionQ = n.storeGrantCoords.TryGetQueueForStore(int32(ba.Replica.StoreID))
		} else {
			__antithesis_instrumentation__.Notify(124750)
		}
		__antithesis_instrumentation__.Notify(124739)
		admissionEnabled := true
		if ah.storeAdmissionQ != nil {
			__antithesis_instrumentation__.Notify(124751)
			if admissionEnabled, err = ah.storeAdmissionQ.Admit(ctx, admissionInfo); err != nil {
				__antithesis_instrumentation__.Notify(124753)
				return admissionHandle{}, err
			} else {
				__antithesis_instrumentation__.Notify(124754)
			}
			__antithesis_instrumentation__.Notify(124752)
			if !admissionEnabled {
				__antithesis_instrumentation__.Notify(124755)

				ah.storeAdmissionQ = nil
			} else {
				__antithesis_instrumentation__.Notify(124756)
			}
		} else {
			__antithesis_instrumentation__.Notify(124757)
		}
		__antithesis_instrumentation__.Notify(124740)
		if admissionEnabled {
			__antithesis_instrumentation__.Notify(124758)
			ah.callAdmittedWorkDoneOnKVAdmissionQ, err = n.kvAdmissionQ.Admit(ctx, admissionInfo)
			if err != nil {
				__antithesis_instrumentation__.Notify(124759)
				return admissionHandle{}, err
			} else {
				__antithesis_instrumentation__.Notify(124760)
			}
		} else {
			__antithesis_instrumentation__.Notify(124761)
		}
	} else {
		__antithesis_instrumentation__.Notify(124762)
	}
	__antithesis_instrumentation__.Notify(124734)
	return ah, nil
}

func (n KVAdmissionControllerImpl) AdmittedKVWorkDone(handle interface{}) {
	__antithesis_instrumentation__.Notify(124763)
	ah := handle.(admissionHandle)
	if ah.callAdmittedWorkDoneOnKVAdmissionQ {
		__antithesis_instrumentation__.Notify(124765)
		n.kvAdmissionQ.AdmittedWorkDone(ah.tenantID)
	} else {
		__antithesis_instrumentation__.Notify(124766)
	}
	__antithesis_instrumentation__.Notify(124764)
	if ah.storeAdmissionQ != nil {
		__antithesis_instrumentation__.Notify(124767)
		ah.storeAdmissionQ.AdmittedWorkDone(ah.tenantID)
	} else {
		__antithesis_instrumentation__.Notify(124768)
	}
}

func (n KVAdmissionControllerImpl) SetTenantWeightProvider(
	provider TenantWeightProvider, stopper *stop.Stopper,
) {
	__antithesis_instrumentation__.Notify(124769)
	go func() {
		__antithesis_instrumentation__.Notify(124770)
		const weightCalculationPeriod = 10 * time.Minute
		ticker := time.NewTicker(weightCalculationPeriod)

		allWeightsDisabled := false
		for {
			__antithesis_instrumentation__.Notify(124771)
			select {
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(124772)
				kvDisabled := !admission.KVTenantWeightsEnabled.Get(&n.settings.SV)
				kvStoresDisabled := !admission.KVStoresTenantWeightsEnabled.Get(&n.settings.SV)
				if allWeightsDisabled && func() bool {
					__antithesis_instrumentation__.Notify(124777)
					return kvDisabled == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(124778)
					return kvStoresDisabled == true
				}() == true {
					__antithesis_instrumentation__.Notify(124779)

					continue
				} else {
					__antithesis_instrumentation__.Notify(124780)
				}
				__antithesis_instrumentation__.Notify(124773)
				weights := provider.GetTenantWeights()
				if kvDisabled {
					__antithesis_instrumentation__.Notify(124781)
					weights.Node = nil
				} else {
					__antithesis_instrumentation__.Notify(124782)
				}
				__antithesis_instrumentation__.Notify(124774)
				n.kvAdmissionQ.SetTenantWeights(weights.Node)
				for _, storeWeights := range weights.Stores {
					__antithesis_instrumentation__.Notify(124783)
					q := n.storeGrantCoords.TryGetQueueForStore(int32(storeWeights.StoreID))
					if q != nil {
						__antithesis_instrumentation__.Notify(124784)
						if kvStoresDisabled {
							__antithesis_instrumentation__.Notify(124786)
							storeWeights.Weights = nil
						} else {
							__antithesis_instrumentation__.Notify(124787)
						}
						__antithesis_instrumentation__.Notify(124785)
						q.SetTenantWeights(storeWeights.Weights)
					} else {
						__antithesis_instrumentation__.Notify(124788)
					}
				}
				__antithesis_instrumentation__.Notify(124775)
				allWeightsDisabled = kvDisabled && func() bool {
					__antithesis_instrumentation__.Notify(124789)
					return kvStoresDisabled == true
				}() == true
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(124776)
				ticker.Stop()
				return
			}
		}
	}()
}
