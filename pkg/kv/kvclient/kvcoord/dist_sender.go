package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/multitenant/tenantcostmodel"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var (
	metaDistSenderBatchCount = metric.Metadata{
		Name:        "distsender.batches",
		Help:        "Number of batches processed",
		Measurement: "Batches",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderPartialBatchCount = metric.Metadata{
		Name:        "distsender.batches.partial",
		Help:        "Number of partial batches processed after being divided on range boundaries",
		Measurement: "Partial Batches",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderAsyncSentCount = metric.Metadata{
		Name:        "distsender.batches.async.sent",
		Help:        "Number of partial batches sent asynchronously",
		Measurement: "Partial Batches",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderAsyncThrottledCount = metric.Metadata{
		Name:        "distsender.batches.async.throttled",
		Help:        "Number of partial batches not sent asynchronously due to throttling",
		Measurement: "Partial Batches",
		Unit:        metric.Unit_COUNT,
	}
	metaTransportSentCount = metric.Metadata{
		Name:        "distsender.rpc.sent",
		Help:        "Number of replica-addressed RPCs sent",
		Measurement: "RPCs",
		Unit:        metric.Unit_COUNT,
	}
	metaTransportLocalSentCount = metric.Metadata{
		Name:        "distsender.rpc.sent.local",
		Help:        "Number of replica-addressed RPCs sent through the local-server optimization",
		Measurement: "RPCs",
		Unit:        metric.Unit_COUNT,
	}
	metaTransportSenderNextReplicaErrCount = metric.Metadata{
		Name:        "distsender.rpc.sent.nextreplicaerror",
		Help:        "Number of replica-addressed RPCs sent due to per-replica errors",
		Measurement: "RPCs",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderNotLeaseHolderErrCount = metric.Metadata{
		Name:        "distsender.errors.notleaseholder",
		Help:        "Number of NotLeaseHolderErrors encountered from replica-addressed RPCs",
		Measurement: "Errors",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderInLeaseTransferBackoffsCount = metric.Metadata{
		Name:        "distsender.errors.inleasetransferbackoffs",
		Help:        "Number of times backed off due to NotLeaseHolderErrors during lease transfer",
		Measurement: "Errors",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderRangeLookups = metric.Metadata{
		Name:        "distsender.rangelookups",
		Help:        "Number of range lookups",
		Measurement: "Range Lookups",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderSlowRPCs = metric.Metadata{
		Name: "requests.slow.distsender",
		Help: `Number of replica-bound RPCs currently stuck or retrying for a long time.

Note that this is not a good signal for KV health. The remote side of the
RPCs tracked here may experience contention, so an end user can easily
cause values for this metric to be emitted by leaving a transaction open
for a long time and contending with it using a second transaction.`,
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderMethodCountTmpl = metric.Metadata{
		Name: "distsender.rpc.%s.sent",
		Help: `Number of %s requests processed.

This counts the requests in batches handed to DistSender, not the RPCs
sent to individual Ranges as a result.`,
		Measurement: "RPCs",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderErrCountTmpl = metric.Metadata{
		Name: "distsender.rpc.err.%s",
		Help: `Number of %s errors received replica-bound RPCs

This counts how often error of the specified type was received back from replicas
as part of executing possibly range-spanning requests. Failures to reach the target
replica will be accounted for as 'roachpb.CommunicationErrType' and unclassified
errors as 'roachpb.InternalErrType'.
`,
		Measurement: "Errors",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderRangefeedTotalRanges = metric.Metadata{
		Name: "distsender.rangefeed.total_ranges",
		Help: `Number of ranges executing rangefeed

This counts the number of ranges with an active rangefeed.
`,
		Measurement: "Ranges",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderRangefeedCatchupRanges = metric.Metadata{
		Name: "distsender.rangefeed.catchup_ranges",
		Help: `Number of ranges in catchup mode

This counts the number of ranges with an active rangefeed that are performing catchup scan.
`,
		Measurement: "Ranges",
		Unit:        metric.Unit_COUNT,
	}
	metaDistSenderRangefeedErrorCatchupRanges = metric.Metadata{
		Name:        "distsender.rangefeed.error_catchup_ranges",
		Help:        `Number of ranges in catchup mode which experienced an error`,
		Measurement: "Ranges",
		Unit:        metric.Unit_COUNT,
	}
)

var CanSendToFollower = func(
	_ uuid.UUID,
	_ *cluster.Settings,
	_ *hlc.Clock,
	_ roachpb.RangeClosedTimestampPolicy,
	_ roachpb.BatchRequest,
) bool {
	__antithesis_instrumentation__.Notify(87037)
	return false
}

const (
	defaultSenderConcurrency = 1024

	rangeLookupPrefetchCount = 8

	sameReplicaRetryLimit = 10
)

var rangeDescriptorCacheSize = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.range_descriptor_cache.size",
	"maximum number of entries in the range descriptor cache",
	1e6,
)

var senderConcurrencyLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.dist_sender.concurrency_limit",
	"maximum number of asynchronous send requests",
	max(defaultSenderConcurrency, int64(64*runtime.GOMAXPROCS(0))),
	settings.NonNegativeInt,
)

func max(a, b int64) int64 {
	__antithesis_instrumentation__.Notify(87038)
	if a > b {
		__antithesis_instrumentation__.Notify(87040)
		return a
	} else {
		__antithesis_instrumentation__.Notify(87041)
	}
	__antithesis_instrumentation__.Notify(87039)
	return b
}

type DistSenderMetrics struct {
	BatchCount              *metric.Counter
	PartialBatchCount       *metric.Counter
	AsyncSentCount          *metric.Counter
	AsyncThrottledCount     *metric.Counter
	SentCount               *metric.Counter
	LocalSentCount          *metric.Counter
	NextReplicaErrCount     *metric.Counter
	NotLeaseHolderErrCount  *metric.Counter
	InLeaseTransferBackoffs *metric.Counter
	RangeLookups            *metric.Counter
	SlowRPCs                *metric.Gauge
	RangefeedRanges         *metric.Gauge
	RangefeedCatchupRanges  *metric.Gauge
	RangefeedErrorCatchup   *metric.Counter
	MethodCounts            [roachpb.NumMethods]*metric.Counter
	ErrCounts               [roachpb.NumErrors]*metric.Counter
}

func makeDistSenderMetrics() DistSenderMetrics {
	__antithesis_instrumentation__.Notify(87042)
	m := DistSenderMetrics{
		BatchCount:              metric.NewCounter(metaDistSenderBatchCount),
		PartialBatchCount:       metric.NewCounter(metaDistSenderPartialBatchCount),
		AsyncSentCount:          metric.NewCounter(metaDistSenderAsyncSentCount),
		AsyncThrottledCount:     metric.NewCounter(metaDistSenderAsyncThrottledCount),
		SentCount:               metric.NewCounter(metaTransportSentCount),
		LocalSentCount:          metric.NewCounter(metaTransportLocalSentCount),
		NextReplicaErrCount:     metric.NewCounter(metaTransportSenderNextReplicaErrCount),
		NotLeaseHolderErrCount:  metric.NewCounter(metaDistSenderNotLeaseHolderErrCount),
		InLeaseTransferBackoffs: metric.NewCounter(metaDistSenderInLeaseTransferBackoffsCount),
		RangeLookups:            metric.NewCounter(metaDistSenderRangeLookups),
		SlowRPCs:                metric.NewGauge(metaDistSenderSlowRPCs),
		RangefeedRanges:         metric.NewGauge(metaDistSenderRangefeedTotalRanges),
		RangefeedCatchupRanges:  metric.NewGauge(metaDistSenderRangefeedCatchupRanges),
		RangefeedErrorCatchup:   metric.NewCounter(metaDistSenderRangefeedErrorCatchupRanges),
	}
	for i := range m.MethodCounts {
		__antithesis_instrumentation__.Notify(87045)
		method := roachpb.Method(i).String()
		meta := metaDistSenderMethodCountTmpl
		meta.Name = fmt.Sprintf(meta.Name, strings.ToLower(method))
		meta.Help = fmt.Sprintf(meta.Help, method)
		m.MethodCounts[i] = metric.NewCounter(meta)
	}
	__antithesis_instrumentation__.Notify(87043)
	for i := range m.ErrCounts {
		__antithesis_instrumentation__.Notify(87046)
		errType := roachpb.ErrorDetailType(i).String()
		meta := metaDistSenderErrCountTmpl
		meta.Name = fmt.Sprintf(meta.Name, strings.ToLower(errType))
		meta.Help = fmt.Sprintf(meta.Help, errType)
		m.ErrCounts[i] = metric.NewCounter(meta)
	}
	__antithesis_instrumentation__.Notify(87044)
	return m
}

type FirstRangeProvider interface {
	GetFirstRangeDescriptor() (*roachpb.RangeDescriptor, error)

	OnFirstRangeChanged(func(*roachpb.RangeDescriptor))
}

type DistSender struct {
	log.AmbientContext

	st *cluster.Settings

	nodeDescriptor unsafe.Pointer

	clock *hlc.Clock

	nodeDescs NodeDescStore

	metrics DistSenderMetrics

	rangeCache *rangecache.RangeCache

	firstRangeProvider FirstRangeProvider
	transportFactory   TransportFactory
	rpcContext         *rpc.Context

	nodeDialer      *nodedialer.Dialer
	rpcRetryOptions retry.Options
	asyncSenderSem  *quotapool.IntPool

	logicalClusterID *base.ClusterIDContainer

	kvInterceptor multitenant.TenantSideKVInterceptor

	disableFirstRangeUpdates int32

	disableParallelBatches bool

	latencyFunc LatencyFunc

	dontReorderReplicas bool

	dontConsiderConnHealth bool

	activeRangeFeeds sync.Map
}

var _ kv.Sender = &DistSender{}

type DistSenderConfig struct {
	AmbientCtx log.AmbientContext

	Settings  *cluster.Settings
	Clock     *hlc.Clock
	NodeDescs NodeDescStore

	nodeDescriptor  *roachpb.NodeDescriptor
	RPCRetryOptions *retry.Options
	RPCContext      *rpc.Context

	NodeDialer *nodedialer.Dialer

	FirstRangeProvider FirstRangeProvider
	RangeDescriptorDB  rangecache.RangeDescriptorDB

	KVInterceptor multitenant.TenantSideKVInterceptor

	TestingKnobs ClientTestingKnobs
}

func NewDistSender(cfg DistSenderConfig) *DistSender {
	__antithesis_instrumentation__.Notify(87047)
	ds := &DistSender{
		st:            cfg.Settings,
		clock:         cfg.Clock,
		nodeDescs:     cfg.NodeDescs,
		metrics:       makeDistSenderMetrics(),
		kvInterceptor: cfg.KVInterceptor,
	}
	if ds.st == nil {
		__antithesis_instrumentation__.Notify(87062)
		ds.st = cluster.MakeTestingClusterSettings()
	} else {
		__antithesis_instrumentation__.Notify(87063)
	}
	__antithesis_instrumentation__.Notify(87048)

	ds.AmbientContext = cfg.AmbientCtx
	if ds.AmbientContext.Tracer == nil {
		__antithesis_instrumentation__.Notify(87064)
		panic("no tracer set in AmbientCtx")
	} else {
		__antithesis_instrumentation__.Notify(87065)
	}
	__antithesis_instrumentation__.Notify(87049)

	if cfg.nodeDescriptor != nil {
		__antithesis_instrumentation__.Notify(87066)
		atomic.StorePointer(&ds.nodeDescriptor, unsafe.Pointer(cfg.nodeDescriptor))
	} else {
		__antithesis_instrumentation__.Notify(87067)
	}
	__antithesis_instrumentation__.Notify(87050)
	var rdb rangecache.RangeDescriptorDB
	if cfg.FirstRangeProvider != nil {
		__antithesis_instrumentation__.Notify(87068)
		ds.firstRangeProvider = cfg.FirstRangeProvider
		rdb = ds
	} else {
		__antithesis_instrumentation__.Notify(87069)
	}
	__antithesis_instrumentation__.Notify(87051)
	if cfg.RangeDescriptorDB != nil {
		__antithesis_instrumentation__.Notify(87070)
		rdb = cfg.RangeDescriptorDB
	} else {
		__antithesis_instrumentation__.Notify(87071)
	}
	__antithesis_instrumentation__.Notify(87052)
	if rdb == nil {
		__antithesis_instrumentation__.Notify(87072)
		panic("DistSenderConfig must contain either FirstRangeProvider or RangeDescriptorDB")
	} else {
		__antithesis_instrumentation__.Notify(87073)
	}
	__antithesis_instrumentation__.Notify(87053)
	getRangeDescCacheSize := func() int64 {
		__antithesis_instrumentation__.Notify(87074)
		return rangeDescriptorCacheSize.Get(&ds.st.SV)
	}
	__antithesis_instrumentation__.Notify(87054)
	ds.rangeCache = rangecache.NewRangeCache(ds.st, rdb, getRangeDescCacheSize,
		cfg.RPCContext.Stopper, cfg.AmbientCtx.Tracer)
	if tf := cfg.TestingKnobs.TransportFactory; tf != nil {
		__antithesis_instrumentation__.Notify(87075)
		ds.transportFactory = tf
	} else {
		__antithesis_instrumentation__.Notify(87076)
		ds.transportFactory = GRPCTransportFactory
	}
	__antithesis_instrumentation__.Notify(87055)
	ds.dontReorderReplicas = cfg.TestingKnobs.DontReorderReplicas
	ds.dontConsiderConnHealth = cfg.TestingKnobs.DontConsiderConnHealth
	ds.rpcRetryOptions = base.DefaultRetryOptions()
	if cfg.RPCRetryOptions != nil {
		__antithesis_instrumentation__.Notify(87077)
		ds.rpcRetryOptions = *cfg.RPCRetryOptions
	} else {
		__antithesis_instrumentation__.Notify(87078)
	}
	__antithesis_instrumentation__.Notify(87056)
	if cfg.RPCContext == nil {
		__antithesis_instrumentation__.Notify(87079)
		panic("no RPCContext set in DistSenderConfig")
	} else {
		__antithesis_instrumentation__.Notify(87080)
	}
	__antithesis_instrumentation__.Notify(87057)
	ds.rpcContext = cfg.RPCContext
	ds.nodeDialer = cfg.NodeDialer
	if ds.rpcRetryOptions.Closer == nil {
		__antithesis_instrumentation__.Notify(87081)
		ds.rpcRetryOptions.Closer = ds.rpcContext.Stopper.ShouldQuiesce()
	} else {
		__antithesis_instrumentation__.Notify(87082)
	}
	__antithesis_instrumentation__.Notify(87058)
	ds.logicalClusterID = cfg.RPCContext.LogicalClusterID
	ds.asyncSenderSem = quotapool.NewIntPool("DistSender async concurrency",
		uint64(senderConcurrencyLimit.Get(&cfg.Settings.SV)))
	senderConcurrencyLimit.SetOnChange(&cfg.Settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(87083)
		ds.asyncSenderSem.UpdateCapacity(uint64(senderConcurrencyLimit.Get(&cfg.Settings.SV)))
	})
	__antithesis_instrumentation__.Notify(87059)
	ds.rpcContext.Stopper.AddCloser(ds.asyncSenderSem.Closer("stopper"))

	if ds.firstRangeProvider != nil {
		__antithesis_instrumentation__.Notify(87084)
		ctx := ds.AnnotateCtx(context.Background())
		ds.firstRangeProvider.OnFirstRangeChanged(func(desc *roachpb.RangeDescriptor) {
			__antithesis_instrumentation__.Notify(87085)
			if atomic.LoadInt32(&ds.disableFirstRangeUpdates) == 1 {
				__antithesis_instrumentation__.Notify(87087)
				return
			} else {
				__antithesis_instrumentation__.Notify(87088)
			}
			__antithesis_instrumentation__.Notify(87086)
			log.VEventf(ctx, 1, "gossiped first range descriptor: %+v", desc.Replicas())
			ds.rangeCache.EvictByKey(ctx, roachpb.RKeyMin)
		})
	} else {
		__antithesis_instrumentation__.Notify(87089)
	}
	__antithesis_instrumentation__.Notify(87060)

	if cfg.TestingKnobs.LatencyFunc != nil {
		__antithesis_instrumentation__.Notify(87090)
		ds.latencyFunc = cfg.TestingKnobs.LatencyFunc
	} else {
		__antithesis_instrumentation__.Notify(87091)
		ds.latencyFunc = ds.rpcContext.RemoteClocks.Latency
	}
	__antithesis_instrumentation__.Notify(87061)
	return ds
}

func (ds *DistSender) DisableFirstRangeUpdates() {
	__antithesis_instrumentation__.Notify(87092)
	atomic.StoreInt32(&ds.disableFirstRangeUpdates, 1)
}

func (ds *DistSender) DisableParallelBatches() {
	__antithesis_instrumentation__.Notify(87093)
	ds.disableParallelBatches = true
}

func (ds *DistSender) Metrics() DistSenderMetrics {
	__antithesis_instrumentation__.Notify(87094)
	return ds.metrics
}

func (ds *DistSender) RangeDescriptorCache() *rangecache.RangeCache {
	__antithesis_instrumentation__.Notify(87095)
	return ds.rangeCache
}

func (ds *DistSender) RangeLookup(
	ctx context.Context, key roachpb.RKey, useReverseScan bool,
) ([]roachpb.RangeDescriptor, []roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(87096)
	ds.metrics.RangeLookups.Inc(1)

	rc := roachpb.READ_UNCOMMITTED

	return kv.RangeLookup(ctx, ds, key.AsRawKey(), rc, rangeLookupPrefetchCount, useReverseScan)
}

func (ds *DistSender) FirstRange() (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(87097)
	if ds.firstRangeProvider == nil {
		__antithesis_instrumentation__.Notify(87099)
		panic("with `nil` firstRangeProvider, DistSender must not use itself as RangeDescriptorDB")
	} else {
		__antithesis_instrumentation__.Notify(87100)
	}
	__antithesis_instrumentation__.Notify(87098)
	return ds.firstRangeProvider.GetFirstRangeDescriptor()
}

func (ds *DistSender) getNodeID() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(87101)

	g, ok := ds.nodeDescs.(*gossip.Gossip)
	if !ok {
		__antithesis_instrumentation__.Notify(87103)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(87104)
	}
	__antithesis_instrumentation__.Notify(87102)
	return g.NodeID.Get()
}

func (ds *DistSender) getNodeDescriptor() *roachpb.NodeDescriptor {
	__antithesis_instrumentation__.Notify(87105)
	if desc := atomic.LoadPointer(&ds.nodeDescriptor); desc != nil {
		__antithesis_instrumentation__.Notify(87110)
		return (*roachpb.NodeDescriptor)(desc)
	} else {
		__antithesis_instrumentation__.Notify(87111)
	}
	__antithesis_instrumentation__.Notify(87106)

	g, ok := ds.nodeDescs.(*gossip.Gossip)
	if !ok {
		__antithesis_instrumentation__.Notify(87112)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(87113)
	}
	__antithesis_instrumentation__.Notify(87107)

	ownNodeID := g.NodeID.Get()
	if ownNodeID > 0 {
		__antithesis_instrumentation__.Notify(87114)

		nodeDesc := &roachpb.NodeDescriptor{}
		if err := g.GetInfoProto(gossip.MakeNodeIDKey(ownNodeID), nodeDesc); err == nil {
			__antithesis_instrumentation__.Notify(87115)
			atomic.StorePointer(&ds.nodeDescriptor, unsafe.Pointer(nodeDesc))
			return nodeDesc
		} else {
			__antithesis_instrumentation__.Notify(87116)
		}
	} else {
		__antithesis_instrumentation__.Notify(87117)
	}
	__antithesis_instrumentation__.Notify(87108)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(87118)
		ctx := ds.AnnotateCtx(context.TODO())
		log.Infof(ctx, "unable to determine this node's attributes for replica "+
			"selection; node is most likely bootstrapping")
	} else {
		__antithesis_instrumentation__.Notify(87119)
	}
	__antithesis_instrumentation__.Notify(87109)
	return nil
}

func (ds *DistSender) CountRanges(ctx context.Context, rs roachpb.RSpan) (int64, error) {
	__antithesis_instrumentation__.Notify(87120)
	var count int64
	ri := MakeRangeIterator(ds)
	for ri.Seek(ctx, rs.Key, Ascending); ri.Valid(); ri.Next(ctx) {
		__antithesis_instrumentation__.Notify(87122)
		count++
		if !ri.NeedAnother(rs) {
			__antithesis_instrumentation__.Notify(87123)
			break
		} else {
			__antithesis_instrumentation__.Notify(87124)
		}
	}
	__antithesis_instrumentation__.Notify(87121)
	return count, ri.Error()
}

func (ds *DistSender) getRoutingInfo(
	ctx context.Context,
	descKey roachpb.RKey,
	evictToken rangecache.EvictionToken,
	useReverseScan bool,
) (rangecache.EvictionToken, error) {
	__antithesis_instrumentation__.Notify(87125)
	returnToken, err := ds.rangeCache.LookupWithEvictionToken(
		ctx, descKey, evictToken, useReverseScan,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(87127)
		return rangecache.EvictionToken{}, err
	} else {
		__antithesis_instrumentation__.Notify(87128)
	}

	{
		__antithesis_instrumentation__.Notify(87129)
		containsFn := (*roachpb.RangeDescriptor).ContainsKey
		if useReverseScan {
			__antithesis_instrumentation__.Notify(87131)
			containsFn = (*roachpb.RangeDescriptor).ContainsKeyInverted
		} else {
			__antithesis_instrumentation__.Notify(87132)
		}
		__antithesis_instrumentation__.Notify(87130)
		if !containsFn(returnToken.Desc(), descKey) {
			__antithesis_instrumentation__.Notify(87133)
			log.Fatalf(ctx, "programming error: range resolution returning non-matching descriptor: "+
				"desc: %s, key: %s, reverse: %t", returnToken.Desc(), descKey, redact.Safe(useReverseScan))
		} else {
			__antithesis_instrumentation__.Notify(87134)
		}
	}
	__antithesis_instrumentation__.Notify(87126)

	return returnToken, nil
}

func (ds *DistSender) initAndVerifyBatch(
	ctx context.Context, ba *roachpb.BatchRequest,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(87135)

	if ba.Header.GatewayNodeID == 0 {
		__antithesis_instrumentation__.Notify(87140)
		ba.Header.GatewayNodeID = ds.getNodeID()
	} else {
		__antithesis_instrumentation__.Notify(87141)
	}
	__antithesis_instrumentation__.Notify(87136)

	if ba.ReadConsistency != roachpb.CONSISTENT && func() bool {
		__antithesis_instrumentation__.Notify(87142)
		return ba.Timestamp.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(87143)
		ba.Timestamp = ds.clock.Now()
	} else {
		__antithesis_instrumentation__.Notify(87144)
	}
	__antithesis_instrumentation__.Notify(87137)

	if len(ba.Requests) < 1 {
		__antithesis_instrumentation__.Notify(87145)
		return roachpb.NewErrorf("empty batch")
	} else {
		__antithesis_instrumentation__.Notify(87146)
	}
	__antithesis_instrumentation__.Notify(87138)

	if ba.MaxSpanRequestKeys != 0 || func() bool {
		__antithesis_instrumentation__.Notify(87147)
		return ba.TargetBytes != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(87148)

		isReverse := ba.IsReverse()
		for _, req := range ba.Requests {
			__antithesis_instrumentation__.Notify(87149)
			inner := req.GetInner()
			switch inner.(type) {
			case *roachpb.ScanRequest, *roachpb.ResolveIntentRangeRequest,
				*roachpb.DeleteRangeRequest, *roachpb.RevertRangeRequest,
				*roachpb.ExportRequest, *roachpb.QueryLocksRequest:
				__antithesis_instrumentation__.Notify(87150)

				if isReverse {
					__antithesis_instrumentation__.Notify(87154)
					return roachpb.NewErrorf("batch with limit contains both forward and reverse scans")
				} else {
					__antithesis_instrumentation__.Notify(87155)
				}

			case *roachpb.ReverseScanRequest:
				__antithesis_instrumentation__.Notify(87151)

			case *roachpb.QueryIntentRequest, *roachpb.EndTxnRequest, *roachpb.GetRequest:
				__antithesis_instrumentation__.Notify(87152)

			default:
				__antithesis_instrumentation__.Notify(87153)
				return roachpb.NewErrorf("batch with limit contains %T request", inner)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(87156)
	}
	__antithesis_instrumentation__.Notify(87139)

	return nil
}

var errNo1PCTxn = roachpb.NewErrorf("cannot send 1PC txn to multiple ranges")

func splitBatchAndCheckForRefreshSpans(
	ba *roachpb.BatchRequest, canSplitET bool,
) [][]roachpb.RequestUnion {
	__antithesis_instrumentation__.Notify(87157)
	parts := ba.Split(canSplitET)

	if len(parts) > 1 {
		__antithesis_instrumentation__.Notify(87159)
		unsetCanForwardReadTimestampFlag(ba)
	} else {
		__antithesis_instrumentation__.Notify(87160)
	}
	__antithesis_instrumentation__.Notify(87158)

	return parts
}

func unsetCanForwardReadTimestampFlag(ba *roachpb.BatchRequest) {
	__antithesis_instrumentation__.Notify(87161)
	if !ba.CanForwardReadTimestamp {
		__antithesis_instrumentation__.Notify(87163)

		return
	} else {
		__antithesis_instrumentation__.Notify(87164)
	}
	__antithesis_instrumentation__.Notify(87162)
	for _, req := range ba.Requests {
		__antithesis_instrumentation__.Notify(87165)
		if roachpb.NeedsRefresh(req.GetInner()) {
			__antithesis_instrumentation__.Notify(87166)

			ba.CanForwardReadTimestamp = false
			return
		} else {
			__antithesis_instrumentation__.Notify(87167)
		}
	}
}

func (ds *DistSender) Send(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(87168)
	ds.incrementBatchCounters(&ba)

	if pErr := ds.initAndVerifyBatch(ctx, &ba); pErr != nil {
		__antithesis_instrumentation__.Notify(87176)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(87177)
	}
	__antithesis_instrumentation__.Notify(87169)

	ctx = ds.AnnotateCtx(ctx)
	ctx, sp := tracing.EnsureChildSpan(ctx, ds.AmbientContext.Tracer, "dist sender send")
	defer sp.Finish()

	var reqInfo tenantcostmodel.RequestInfo
	if ds.kvInterceptor != nil {
		__antithesis_instrumentation__.Notify(87178)
		reqInfo = tenantcostmodel.MakeRequestInfo(&ba)
		if err := ds.kvInterceptor.OnRequestWait(ctx, reqInfo); err != nil {
			__antithesis_instrumentation__.Notify(87179)
			return nil, roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(87180)
		}
	} else {
		__antithesis_instrumentation__.Notify(87181)
	}
	__antithesis_instrumentation__.Notify(87170)

	splitET := false
	var require1PC bool
	lastReq := ba.Requests[len(ba.Requests)-1].GetInner()
	if et, ok := lastReq.(*roachpb.EndTxnRequest); ok && func() bool {
		__antithesis_instrumentation__.Notify(87182)
		return et.Require1PC == true
	}() == true {
		__antithesis_instrumentation__.Notify(87183)
		require1PC = true
	} else {
		__antithesis_instrumentation__.Notify(87184)
	}
	__antithesis_instrumentation__.Notify(87171)

	if ba.Txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(87185)
		return ba.Txn.Epoch > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(87186)
		return !require1PC == true
	}() == true {
		__antithesis_instrumentation__.Notify(87187)
		splitET = true
	} else {
		__antithesis_instrumentation__.Notify(87188)
	}
	__antithesis_instrumentation__.Notify(87172)
	parts := splitBatchAndCheckForRefreshSpans(&ba, splitET)
	if len(parts) > 1 && func() bool {
		__antithesis_instrumentation__.Notify(87189)
		return (ba.MaxSpanRequestKeys != 0 || func() bool {
			__antithesis_instrumentation__.Notify(87190)
			return ba.TargetBytes != 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(87191)

		log.Fatalf(ctx, "batch with MaxSpanRequestKeys=%d, TargetBytes=%d needs splitting",
			redact.Safe(ba.MaxSpanRequestKeys), redact.Safe(ba.TargetBytes))
	} else {
		__antithesis_instrumentation__.Notify(87192)
	}
	__antithesis_instrumentation__.Notify(87173)
	var singleRplChunk [1]*roachpb.BatchResponse
	rplChunks := singleRplChunk[:0:1]

	errIdxOffset := 0
	for len(parts) > 0 {
		__antithesis_instrumentation__.Notify(87193)
		part := parts[0]
		ba.Requests = part

		rs, err := keys.Range(ba.Requests)
		if err != nil {
			__antithesis_instrumentation__.Notify(87199)
			return nil, roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(87200)
		}
		__antithesis_instrumentation__.Notify(87194)
		isReverse := ba.IsReverse()

		var withCommit, withParallelCommit bool
		if etArg, ok := ba.GetArg(roachpb.EndTxn); ok {
			__antithesis_instrumentation__.Notify(87201)
			et := etArg.(*roachpb.EndTxnRequest)
			withCommit = et.Commit
			withParallelCommit = et.IsParallelCommit()
		} else {
			__antithesis_instrumentation__.Notify(87202)
		}
		__antithesis_instrumentation__.Notify(87195)

		var rpl *roachpb.BatchResponse
		var pErr *roachpb.Error
		if withParallelCommit {
			__antithesis_instrumentation__.Notify(87203)
			rpl, pErr = ds.divideAndSendParallelCommit(ctx, ba, rs, isReverse, 0)
		} else {
			__antithesis_instrumentation__.Notify(87204)
			rpl, pErr = ds.divideAndSendBatchToRanges(ctx, ba, rs, isReverse, withCommit, 0)
		}
		__antithesis_instrumentation__.Notify(87196)

		if pErr == errNo1PCTxn {
			__antithesis_instrumentation__.Notify(87205)

			if len(parts) != 1 {
				__antithesis_instrumentation__.Notify(87207)
				panic("EndTxn not in last chunk of batch")
			} else {
				__antithesis_instrumentation__.Notify(87208)
				if require1PC {
					__antithesis_instrumentation__.Notify(87209)
					log.Fatalf(ctx, "required 1PC transaction cannot be split: %s", ba)
				} else {
					__antithesis_instrumentation__.Notify(87210)
				}
			}
			__antithesis_instrumentation__.Notify(87206)
			parts = splitBatchAndCheckForRefreshSpans(&ba, true)

			continue
		} else {
			__antithesis_instrumentation__.Notify(87211)
		}
		__antithesis_instrumentation__.Notify(87197)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(87212)
			if pErr.Index != nil && func() bool {
				__antithesis_instrumentation__.Notify(87214)
				return pErr.Index.Index != -1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(87215)
				pErr.Index.Index += int32(errIdxOffset)
			} else {
				__antithesis_instrumentation__.Notify(87216)
			}
			__antithesis_instrumentation__.Notify(87213)
			return nil, pErr
		} else {
			__antithesis_instrumentation__.Notify(87217)
		}
		__antithesis_instrumentation__.Notify(87198)

		errIdxOffset += len(ba.Requests)

		ba.UpdateTxn(rpl.Txn)
		rplChunks = append(rplChunks, rpl)
		parts = parts[1:]
	}
	__antithesis_instrumentation__.Notify(87174)

	var reply *roachpb.BatchResponse
	if len(rplChunks) > 0 {
		__antithesis_instrumentation__.Notify(87218)
		reply = rplChunks[0]
		for _, rpl := range rplChunks[1:] {
			__antithesis_instrumentation__.Notify(87220)
			reply.Responses = append(reply.Responses, rpl.Responses...)
			reply.CollectedSpans = append(reply.CollectedSpans, rpl.CollectedSpans...)
		}
		__antithesis_instrumentation__.Notify(87219)
		lastHeader := rplChunks[len(rplChunks)-1].BatchResponse_Header
		lastHeader.CollectedSpans = reply.CollectedSpans
		reply.BatchResponse_Header = lastHeader

		if ds.kvInterceptor != nil {
			__antithesis_instrumentation__.Notify(87221)
			respInfo := tenantcostmodel.MakeResponseInfo(reply)
			ds.kvInterceptor.OnResponse(ctx, reqInfo, respInfo)
		} else {
			__antithesis_instrumentation__.Notify(87222)
		}
	} else {
		__antithesis_instrumentation__.Notify(87223)
	}
	__antithesis_instrumentation__.Notify(87175)

	return reply, nil
}

func (ds *DistSender) incrementBatchCounters(ba *roachpb.BatchRequest) {
	__antithesis_instrumentation__.Notify(87224)
	ds.metrics.BatchCount.Inc(1)
	for _, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(87225)
		m := ru.GetInner().Method()
		ds.metrics.MethodCounts[m].Inc(1)
	}
}

type response struct {
	reply     *roachpb.BatchResponse
	positions []int
	pErr      *roachpb.Error
}

func (ds *DistSender) divideAndSendParallelCommit(
	ctx context.Context, ba roachpb.BatchRequest, rs roachpb.RSpan, isReverse bool, batchIdx int,
) (br *roachpb.BatchResponse, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(87226)

	swapIdx := -1
	lastIdx := len(ba.Requests) - 1
	for i := lastIdx - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(87236)
		req := ba.Requests[i].GetInner()
		if req.Method() == roachpb.QueryIntent {
			__antithesis_instrumentation__.Notify(87237)
			swapIdx = i
		} else {
			__antithesis_instrumentation__.Notify(87238)
			break
		}
	}
	__antithesis_instrumentation__.Notify(87227)
	if swapIdx == -1 {
		__antithesis_instrumentation__.Notify(87239)

		return ds.divideAndSendBatchToRanges(ctx, ba, rs, isReverse, true, batchIdx)
	} else {
		__antithesis_instrumentation__.Notify(87240)
	}
	__antithesis_instrumentation__.Notify(87228)

	swappedReqs := append([]roachpb.RequestUnion(nil), ba.Requests...)
	swappedReqs[swapIdx], swappedReqs[lastIdx] = swappedReqs[lastIdx], swappedReqs[swapIdx]

	qiBa := ba
	qiBa.Requests = swappedReqs[swapIdx+1:]
	qiRS, err := keys.Range(qiBa.Requests)
	if err != nil {
		__antithesis_instrumentation__.Notify(87241)
		return br, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(87242)
	}
	__antithesis_instrumentation__.Notify(87229)
	qiIsReverse := qiBa.IsReverse()
	qiBatchIdx := batchIdx + 1
	qiResponseCh := make(chan response, 1)
	qiBaCopy := qiBa

	runTask := ds.rpcContext.Stopper.RunAsyncTask
	if ds.disableParallelBatches {
		__antithesis_instrumentation__.Notify(87243)
		runTask = ds.rpcContext.Stopper.RunTask
	} else {
		__antithesis_instrumentation__.Notify(87244)
	}
	__antithesis_instrumentation__.Notify(87230)
	if err := runTask(ctx, "kv.DistSender: sending pre-commit query intents", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(87245)

		positions := make([]int, len(qiBa.Requests))
		positions[len(positions)-1] = swapIdx
		for i := range positions[:len(positions)-1] {
			__antithesis_instrumentation__.Notify(87247)
			positions[i] = swapIdx + 1 + i
		}
		__antithesis_instrumentation__.Notify(87246)

		reply, pErr := ds.divideAndSendBatchToRanges(ctx, qiBa, qiRS, qiIsReverse, true, qiBatchIdx)
		qiResponseCh <- response{reply: reply, positions: positions, pErr: pErr}
	}); err != nil {
		__antithesis_instrumentation__.Notify(87248)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(87249)
	}
	__antithesis_instrumentation__.Notify(87231)

	ba.Requests = swappedReqs[:swapIdx+1]
	rs, err = keys.Range(ba.Requests)
	if err != nil {
		__antithesis_instrumentation__.Notify(87250)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(87251)
	}
	__antithesis_instrumentation__.Notify(87232)
	isReverse = ba.IsReverse()
	br, pErr = ds.divideAndSendBatchToRanges(ctx, ba, rs, isReverse, true, batchIdx)

	qiReply := <-qiResponseCh

	if pErr != nil {
		__antithesis_instrumentation__.Notify(87252)

		if qiReply.reply != nil {
			__antithesis_instrumentation__.Notify(87254)
			pErr.UpdateTxn(qiReply.reply.Txn)
		} else {
			__antithesis_instrumentation__.Notify(87255)
		}
		__antithesis_instrumentation__.Notify(87253)
		maybeSwapErrorIndex(pErr, swapIdx, lastIdx)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(87256)
	}
	__antithesis_instrumentation__.Notify(87233)
	if qiPErr := qiReply.pErr; qiPErr != nil {
		__antithesis_instrumentation__.Notify(87257)

		ignoreMissing := false
		if _, ok := qiPErr.GetDetail().(*roachpb.IntentMissingError); ok {
			__antithesis_instrumentation__.Notify(87260)

			ignoreMissing, err = ds.detectIntentMissingDueToIntentResolution(ctx, br.Txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(87261)
				return nil, roachpb.NewErrorWithTxn(err, br.Txn)
			} else {
				__antithesis_instrumentation__.Notify(87262)
			}
		} else {
			__antithesis_instrumentation__.Notify(87263)
		}
		__antithesis_instrumentation__.Notify(87258)
		if !ignoreMissing {
			__antithesis_instrumentation__.Notify(87264)
			qiPErr.UpdateTxn(br.Txn)
			maybeSwapErrorIndex(qiPErr, swapIdx, lastIdx)
			return nil, qiPErr
		} else {
			__antithesis_instrumentation__.Notify(87265)
		}
		__antithesis_instrumentation__.Notify(87259)

		qiReply.reply = qiBaCopy.CreateReply()
		for _, ru := range qiReply.reply.Responses {
			__antithesis_instrumentation__.Notify(87266)
			ru.GetQueryIntent().FoundIntent = true
		}
	} else {
		__antithesis_instrumentation__.Notify(87267)
	}
	__antithesis_instrumentation__.Notify(87234)

	resps := make([]roachpb.ResponseUnion, len(swappedReqs))
	copy(resps, br.Responses)
	resps[swapIdx], resps[lastIdx] = resps[lastIdx], resps[swapIdx]
	br.Responses = resps
	if err := br.Combine(qiReply.reply, qiReply.positions); err != nil {
		__antithesis_instrumentation__.Notify(87268)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(87269)
	}
	__antithesis_instrumentation__.Notify(87235)
	return br, nil
}

func (ds *DistSender) detectIntentMissingDueToIntentResolution(
	ctx context.Context, txn *roachpb.Transaction,
) (bool, error) {
	__antithesis_instrumentation__.Notify(87270)
	ba := roachpb.BatchRequest{}
	ba.Timestamp = ds.clock.Now()
	ba.Add(&roachpb.QueryTxnRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: txn.TxnMeta.Key,
		},
		Txn: txn.TxnMeta,
	})
	log.VEvent(ctx, 1, "detecting whether missing intent is due to intent resolution")
	br, pErr := ds.Send(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(87272)

		return false, roachpb.NewAmbiguousResultErrorf("error=%s [intent missing]", pErr)
	} else {
		__antithesis_instrumentation__.Notify(87273)
	}
	__antithesis_instrumentation__.Notify(87271)
	resp := br.Responses[0].GetQueryTxn()
	respTxn := &resp.QueriedTxn
	switch respTxn.Status {
	case roachpb.COMMITTED:
		__antithesis_instrumentation__.Notify(87274)

		return true, nil
	case roachpb.ABORTED:
		__antithesis_instrumentation__.Notify(87275)

		if resp.TxnRecordExists {
			__antithesis_instrumentation__.Notify(87278)
			return false, roachpb.NewTransactionAbortedError(roachpb.ABORT_REASON_ABORTED_RECORD_FOUND)
		} else {
			__antithesis_instrumentation__.Notify(87279)
		}
		__antithesis_instrumentation__.Notify(87276)
		return false, roachpb.NewAmbiguousResultErrorf("intent missing and record aborted")
	default:
		__antithesis_instrumentation__.Notify(87277)

		return false, nil
	}
}

func maybeSwapErrorIndex(pErr *roachpb.Error, a, b int) {
	__antithesis_instrumentation__.Notify(87280)
	if pErr.Index == nil {
		__antithesis_instrumentation__.Notify(87282)
		return
	} else {
		__antithesis_instrumentation__.Notify(87283)
	}
	__antithesis_instrumentation__.Notify(87281)
	if pErr.Index.Index == int32(a) {
		__antithesis_instrumentation__.Notify(87284)
		pErr.Index.Index = int32(b)
	} else {
		__antithesis_instrumentation__.Notify(87285)
		if pErr.Index.Index == int32(b) {
			__antithesis_instrumentation__.Notify(87286)
			pErr.Index.Index = int32(a)
		} else {
			__antithesis_instrumentation__.Notify(87287)
		}
	}
}

func mergeErrors(pErr1, pErr2 *roachpb.Error) *roachpb.Error {
	__antithesis_instrumentation__.Notify(87288)
	ret, drop := pErr1, pErr2
	if roachpb.ErrPriority(drop.GoError()) > roachpb.ErrPriority(ret.GoError()) {
		__antithesis_instrumentation__.Notify(87290)
		ret, drop = drop, ret
	} else {
		__antithesis_instrumentation__.Notify(87291)
	}
	__antithesis_instrumentation__.Notify(87289)
	ret.UpdateTxn(drop.GetTxn())
	return ret
}

func (ds *DistSender) divideAndSendBatchToRanges(
	ctx context.Context,
	ba roachpb.BatchRequest,
	rs roachpb.RSpan,
	isReverse bool,
	withCommit bool,
	batchIdx int,
) (br *roachpb.BatchResponse, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(87292)

	if ba.Txn != nil {
		__antithesis_instrumentation__.Notify(87303)
		ba.Txn = ba.Txn.Clone()
	} else {
		__antithesis_instrumentation__.Notify(87304)
	}
	__antithesis_instrumentation__.Notify(87293)

	var scanDir ScanDirection
	var seekKey roachpb.RKey
	if !isReverse {
		__antithesis_instrumentation__.Notify(87305)
		scanDir = Ascending
		seekKey = rs.Key
	} else {
		__antithesis_instrumentation__.Notify(87306)
		scanDir = Descending
		seekKey = rs.EndKey
	}
	__antithesis_instrumentation__.Notify(87294)
	ri := MakeRangeIterator(ds)
	ri.Seek(ctx, seekKey, scanDir)
	if !ri.Valid() {
		__antithesis_instrumentation__.Notify(87307)
		return nil, roachpb.NewError(ri.Error())
	} else {
		__antithesis_instrumentation__.Notify(87308)
	}
	__antithesis_instrumentation__.Notify(87295)

	if !ri.NeedAnother(rs) {
		__antithesis_instrumentation__.Notify(87309)
		resp := ds.sendPartialBatch(
			ctx, ba, rs, isReverse, withCommit, batchIdx, ri.Token(), false,
		)
		return resp.reply, resp.pErr
	} else {
		__antithesis_instrumentation__.Notify(87310)
	}
	__antithesis_instrumentation__.Notify(87296)

	if ba.IsUnsplittable() {
		__antithesis_instrumentation__.Notify(87311)
		mismatch := roachpb.NewRangeKeyMismatchErrorWithCTPolicy(ctx,
			rs.Key.AsRawKey(),
			rs.EndKey.AsRawKey(),
			ri.Desc(),
			nil,
			ri.ClosedTimestampPolicy(),
		)
		return nil, roachpb.NewError(mismatch)
	} else {
		__antithesis_instrumentation__.Notify(87312)
	}
	__antithesis_instrumentation__.Notify(87297)

	if ba.Txn == nil && func() bool {
		__antithesis_instrumentation__.Notify(87313)
		return ba.IsTransactional() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(87314)
		return ba.ReadConsistency == roachpb.CONSISTENT == true
	}() == true {
		__antithesis_instrumentation__.Notify(87315)
		return nil, roachpb.NewError(&roachpb.OpRequiresTxnError{})
	} else {
		__antithesis_instrumentation__.Notify(87316)
	}
	__antithesis_instrumentation__.Notify(87298)

	if withCommit {
		__antithesis_instrumentation__.Notify(87317)
		etArg, ok := ba.GetArg(roachpb.EndTxn)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(87318)
			return !etArg.(*roachpb.EndTxnRequest).IsParallelCommit() == true
		}() == true {
			__antithesis_instrumentation__.Notify(87319)
			return nil, errNo1PCTxn
		} else {
			__antithesis_instrumentation__.Notify(87320)
		}
	} else {
		__antithesis_instrumentation__.Notify(87321)
	}
	__antithesis_instrumentation__.Notify(87299)

	unsetCanForwardReadTimestampFlag(&ba)

	br = &roachpb.BatchResponse{
		Responses: make([]roachpb.ResponseUnion, len(ba.Requests)),
	}

	var responseChs []chan response

	var couldHaveSkippedResponses bool

	var resumeReason roachpb.ResumeReason
	defer func() {
		__antithesis_instrumentation__.Notify(87322)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(87325)

			panic(r)
		} else {
			__antithesis_instrumentation__.Notify(87326)
		}
		__antithesis_instrumentation__.Notify(87323)

		for _, responseCh := range responseChs {
			__antithesis_instrumentation__.Notify(87327)
			resp := <-responseCh
			if resp.pErr != nil {
				__antithesis_instrumentation__.Notify(87329)
				if pErr == nil {
					__antithesis_instrumentation__.Notify(87331)
					pErr = resp.pErr

					pErr.UpdateTxn(br.Txn)
				} else {
					__antithesis_instrumentation__.Notify(87332)

					pErr = mergeErrors(pErr, resp.pErr)
				}
				__antithesis_instrumentation__.Notify(87330)
				continue
			} else {
				__antithesis_instrumentation__.Notify(87333)
			}
			__antithesis_instrumentation__.Notify(87328)

			if pErr == nil {
				__antithesis_instrumentation__.Notify(87334)
				if err := br.Combine(resp.reply, resp.positions); err != nil {
					__antithesis_instrumentation__.Notify(87335)
					pErr = roachpb.NewError(err)
				} else {
					__antithesis_instrumentation__.Notify(87336)
				}
			} else {
				__antithesis_instrumentation__.Notify(87337)

				pErr.UpdateTxn(resp.reply.Txn)
			}
		}
		__antithesis_instrumentation__.Notify(87324)

		if pErr == nil && func() bool {
			__antithesis_instrumentation__.Notify(87338)
			return couldHaveSkippedResponses == true
		}() == true {
			__antithesis_instrumentation__.Notify(87339)
			fillSkippedResponses(ba, br, seekKey, resumeReason)
		} else {
			__antithesis_instrumentation__.Notify(87340)
		}
	}()
	__antithesis_instrumentation__.Notify(87300)

	canParallelize := ba.Header.MaxSpanRequestKeys == 0 && func() bool {
		__antithesis_instrumentation__.Notify(87341)
		return ba.Header.TargetBytes == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(87342)
		return !ba.Header.ReturnOnRangeBoundary == true
	}() == true
	if ba.IsSingleCheckConsistencyRequest() {
		__antithesis_instrumentation__.Notify(87343)

		isExpensive := ba.Requests[0].GetCheckConsistency().Mode == roachpb.ChecksumMode_CHECK_FULL
		canParallelize = canParallelize && func() bool {
			__antithesis_instrumentation__.Notify(87344)
			return !isExpensive == true
		}() == true
	} else {
		__antithesis_instrumentation__.Notify(87345)
	}
	__antithesis_instrumentation__.Notify(87301)

	for ; ri.Valid(); ri.Seek(ctx, seekKey, scanDir) {
		__antithesis_instrumentation__.Notify(87346)
		responseCh := make(chan response, 1)
		responseChs = append(responseChs, responseCh)

		var err error
		nextRS := rs
		if scanDir == Descending {
			__antithesis_instrumentation__.Notify(87351)

			seekKey, err = prev(ba.Requests, ri.Desc().StartKey)
			nextRS.EndKey = seekKey
		} else {
			__antithesis_instrumentation__.Notify(87352)

			seekKey, err = Next(ba.Requests, ri.Desc().EndKey)
			nextRS.Key = seekKey
		}
		__antithesis_instrumentation__.Notify(87347)
		if err != nil {
			__antithesis_instrumentation__.Notify(87353)
			responseCh <- response{pErr: roachpb.NewError(err)}
			return
		} else {
			__antithesis_instrumentation__.Notify(87354)
		}
		__antithesis_instrumentation__.Notify(87348)

		lastRange := !ri.NeedAnother(rs)

		if canParallelize && func() bool {
			__antithesis_instrumentation__.Notify(87355)
			return !lastRange == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(87356)
			return !ds.disableParallelBatches == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(87357)
			return ds.sendPartialBatchAsync(ctx, ba, rs, isReverse, withCommit, batchIdx, ri.Token(), responseCh) == true
		}() == true {
			__antithesis_instrumentation__.Notify(87358)

		} else {
			__antithesis_instrumentation__.Notify(87359)
			resp := ds.sendPartialBatch(
				ctx, ba, rs, isReverse, withCommit, batchIdx, ri.Token(), true,
			)
			responseCh <- resp
			if resp.pErr != nil {
				__antithesis_instrumentation__.Notify(87362)
				return
			} else {
				__antithesis_instrumentation__.Notify(87363)
			}
			__antithesis_instrumentation__.Notify(87360)

			if !lastRange {
				__antithesis_instrumentation__.Notify(87364)
				ba.UpdateTxn(resp.reply.Txn)
			} else {
				__antithesis_instrumentation__.Notify(87365)
			}
			__antithesis_instrumentation__.Notify(87361)

			mightStopEarly := ba.MaxSpanRequestKeys > 0 || func() bool {
				__antithesis_instrumentation__.Notify(87366)
				return ba.TargetBytes > 0 == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(87367)
				return ba.ReturnOnRangeBoundary == true
			}() == true

			if mightStopEarly {
				__antithesis_instrumentation__.Notify(87368)
				var replyKeys int64
				var replyBytes int64
				for _, r := range resp.reply.Responses {
					__antithesis_instrumentation__.Notify(87372)
					h := r.GetInner().Header()
					replyKeys += h.NumKeys
					replyBytes += h.NumBytes
					if h.ResumeSpan != nil {
						__antithesis_instrumentation__.Notify(87373)
						couldHaveSkippedResponses = true
						resumeReason = h.ResumeReason
						return
					} else {
						__antithesis_instrumentation__.Notify(87374)
					}
				}
				__antithesis_instrumentation__.Notify(87369)

				if ba.MaxSpanRequestKeys > 0 {
					__antithesis_instrumentation__.Notify(87375)
					ba.MaxSpanRequestKeys -= replyKeys
					if ba.MaxSpanRequestKeys <= 0 {
						__antithesis_instrumentation__.Notify(87376)
						couldHaveSkippedResponses = true
						resumeReason = roachpb.RESUME_KEY_LIMIT
						return
					} else {
						__antithesis_instrumentation__.Notify(87377)
					}
				} else {
					__antithesis_instrumentation__.Notify(87378)
				}
				__antithesis_instrumentation__.Notify(87370)
				if ba.TargetBytes > 0 {
					__antithesis_instrumentation__.Notify(87379)
					ba.TargetBytes -= replyBytes
					if ba.TargetBytes <= 0 {
						__antithesis_instrumentation__.Notify(87380)
						couldHaveSkippedResponses = true
						resumeReason = roachpb.RESUME_BYTE_LIMIT
						return
					} else {
						__antithesis_instrumentation__.Notify(87381)
					}
				} else {
					__antithesis_instrumentation__.Notify(87382)
				}
				__antithesis_instrumentation__.Notify(87371)

				if ba.Header.ReturnOnRangeBoundary && func() bool {
					__antithesis_instrumentation__.Notify(87383)
					return replyKeys > 0 == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(87384)
					return !lastRange == true
				}() == true {
					__antithesis_instrumentation__.Notify(87385)
					couldHaveSkippedResponses = true
					resumeReason = roachpb.RESUME_RANGE_BOUNDARY
					return
				} else {
					__antithesis_instrumentation__.Notify(87386)
				}
			} else {
				__antithesis_instrumentation__.Notify(87387)
			}
		}
		__antithesis_instrumentation__.Notify(87349)

		if lastRange || func() bool {
			__antithesis_instrumentation__.Notify(87388)
			return !nextRS.Key.Less(nextRS.EndKey) == true
		}() == true {
			__antithesis_instrumentation__.Notify(87389)
			return
		} else {
			__antithesis_instrumentation__.Notify(87390)
		}
		__antithesis_instrumentation__.Notify(87350)
		batchIdx++
		rs = nextRS
	}
	__antithesis_instrumentation__.Notify(87302)

	responseCh := make(chan response, 1)
	responseCh <- response{pErr: roachpb.NewError(ri.Error())}
	responseChs = append(responseChs, responseCh)
	return
}

func (ds *DistSender) sendPartialBatchAsync(
	ctx context.Context,
	ba roachpb.BatchRequest,
	rs roachpb.RSpan,
	isReverse bool,
	withCommit bool,
	batchIdx int,
	routing rangecache.EvictionToken,
	responseCh chan response,
) bool {
	__antithesis_instrumentation__.Notify(87391)
	if err := ds.rpcContext.Stopper.RunAsyncTaskEx(
		ctx,
		stop.TaskOpts{
			TaskName:   "kv.DistSender: sending partial batch",
			SpanOpt:    stop.ChildSpan,
			Sem:        ds.asyncSenderSem,
			WaitForSem: false,
		},
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(87393)
			ds.metrics.AsyncSentCount.Inc(1)
			responseCh <- ds.sendPartialBatch(
				ctx, ba, rs, isReverse, withCommit, batchIdx, routing, true,
			)
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(87394)
		ds.metrics.AsyncThrottledCount.Inc(1)
		return false
	} else {
		__antithesis_instrumentation__.Notify(87395)
	}
	__antithesis_instrumentation__.Notify(87392)
	return true
}

func slowRangeRPCWarningStr(
	s *redact.StringBuilder,
	ba roachpb.BatchRequest,
	dur time.Duration,
	attempts int64,
	desc *roachpb.RangeDescriptor,
	err error,
	br *roachpb.BatchResponse,
) {
	__antithesis_instrumentation__.Notify(87396)
	resp := interface{}(err)
	if resp == nil {
		__antithesis_instrumentation__.Notify(87398)
		resp = br
	} else {
		__antithesis_instrumentation__.Notify(87399)
	}
	__antithesis_instrumentation__.Notify(87397)
	s.Printf("have been waiting %.2fs (%d attempts) for RPC %s to %s; resp: %s",
		dur.Seconds(), attempts, ba, desc, resp)
}

func slowRangeRPCReturnWarningStr(s *redact.StringBuilder, dur time.Duration, attempts int64) {
	__antithesis_instrumentation__.Notify(87400)
	s.Printf("slow RPC finished after %.2fs (%d attempts)", dur.Seconds(), attempts)
}

func (ds *DistSender) sendPartialBatch(
	ctx context.Context,
	ba roachpb.BatchRequest,
	rs roachpb.RSpan,
	isReverse bool,
	withCommit bool,
	batchIdx int,
	routingTok rangecache.EvictionToken,
	needsTruncate bool,
) response {
	__antithesis_instrumentation__.Notify(87401)
	if batchIdx == 1 {
		__antithesis_instrumentation__.Notify(87406)
		ds.metrics.PartialBatchCount.Inc(2)
	} else {
		__antithesis_instrumentation__.Notify(87407)
		if batchIdx > 1 {
			__antithesis_instrumentation__.Notify(87408)
			ds.metrics.PartialBatchCount.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(87409)
		}
	}
	__antithesis_instrumentation__.Notify(87402)
	var reply *roachpb.BatchResponse
	var pErr *roachpb.Error
	var err error
	var positions []int

	if needsTruncate {
		__antithesis_instrumentation__.Notify(87410)

		rs, err = rs.Intersect(routingTok.Desc())
		if err != nil {
			__antithesis_instrumentation__.Notify(87413)
			return response{pErr: roachpb.NewError(err)}
		} else {
			__antithesis_instrumentation__.Notify(87414)
		}
		__antithesis_instrumentation__.Notify(87411)
		ba.Requests, positions, err = Truncate(ba.Requests, rs)
		if len(positions) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(87415)
			return err == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(87416)

			return response{
				pErr: roachpb.NewErrorf("truncation resulted in empty batch on %s: %s", rs, ba),
			}
		} else {
			__antithesis_instrumentation__.Notify(87417)
		}
		__antithesis_instrumentation__.Notify(87412)
		if err != nil {
			__antithesis_instrumentation__.Notify(87418)
			return response{pErr: roachpb.NewError(err)}
		} else {
			__antithesis_instrumentation__.Notify(87419)
		}
	} else {
		__antithesis_instrumentation__.Notify(87420)
	}
	__antithesis_instrumentation__.Notify(87403)

	tBegin, attempts := timeutil.Now(), int64(0)

	var prevTok rangecache.EvictionToken
	for r := retry.StartWithCtx(ctx, ds.rpcRetryOptions); r.Next(); {
		__antithesis_instrumentation__.Notify(87421)
		attempts++
		pErr = nil

		if !routingTok.Valid() {
			__antithesis_instrumentation__.Notify(87428)
			var descKey roachpb.RKey
			if isReverse {
				__antithesis_instrumentation__.Notify(87432)
				descKey = rs.EndKey
			} else {
				__antithesis_instrumentation__.Notify(87433)
				descKey = rs.Key
			}
			__antithesis_instrumentation__.Notify(87429)
			routingTok, err = ds.getRoutingInfo(ctx, descKey, prevTok, isReverse)
			if err != nil {
				__antithesis_instrumentation__.Notify(87434)
				log.VErrEventf(ctx, 1, "range descriptor re-lookup failed: %s", err)

				pErr = roachpb.NewError(err)
				if !rangecache.IsRangeLookupErrorRetryable(err) {
					__antithesis_instrumentation__.Notify(87436)
					return response{pErr: pErr}
				} else {
					__antithesis_instrumentation__.Notify(87437)
				}
				__antithesis_instrumentation__.Notify(87435)
				continue
			} else {
				__antithesis_instrumentation__.Notify(87438)
			}
			__antithesis_instrumentation__.Notify(87430)

			intersection, err := rs.Intersect(routingTok.Desc())
			if err != nil {
				__antithesis_instrumentation__.Notify(87439)
				return response{pErr: roachpb.NewError(err)}
			} else {
				__antithesis_instrumentation__.Notify(87440)
			}
			__antithesis_instrumentation__.Notify(87431)
			if !intersection.Equal(rs) {
				__antithesis_instrumentation__.Notify(87441)
				log.Eventf(ctx, "range shrunk; sub-dividing the request")
				reply, pErr = ds.divideAndSendBatchToRanges(ctx, ba, rs, isReverse, withCommit, batchIdx)
				return response{reply: reply, positions: positions, pErr: pErr}
			} else {
				__antithesis_instrumentation__.Notify(87442)
			}
		} else {
			__antithesis_instrumentation__.Notify(87443)
		}
		__antithesis_instrumentation__.Notify(87422)

		prevTok = routingTok
		reply, err = ds.sendToReplicas(ctx, ba, routingTok, withCommit)

		const slowDistSenderThreshold = time.Minute
		if dur := timeutil.Since(tBegin); dur > slowDistSenderThreshold && func() bool {
			__antithesis_instrumentation__.Notify(87444)
			return !tBegin.IsZero() == true
		}() == true {
			__antithesis_instrumentation__.Notify(87445)
			{
				__antithesis_instrumentation__.Notify(87448)
				var s redact.StringBuilder
				slowRangeRPCWarningStr(&s, ba, dur, attempts, routingTok.Desc(), err, reply)
				log.Warningf(ctx, "slow range RPC: %v", &s)
			}
			__antithesis_instrumentation__.Notify(87446)

			if err != nil || func() bool {
				__antithesis_instrumentation__.Notify(87449)
				return reply.Error != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(87450)
				ds.metrics.SlowRPCs.Inc(1)
				defer func(tBegin time.Time, attempts int64) {
					__antithesis_instrumentation__.Notify(87451)
					ds.metrics.SlowRPCs.Dec(1)
					var s redact.StringBuilder
					slowRangeRPCReturnWarningStr(&s, timeutil.Since(tBegin), attempts)
					log.Warningf(ctx, "slow RPC response: %v", &s)
				}(tBegin, attempts)
			} else {
				__antithesis_instrumentation__.Notify(87452)
			}
			__antithesis_instrumentation__.Notify(87447)
			tBegin = time.Time{}
		} else {
			__antithesis_instrumentation__.Notify(87453)
		}
		__antithesis_instrumentation__.Notify(87423)

		if err != nil {
			__antithesis_instrumentation__.Notify(87454)

			pErr = roachpb.NewError(err)
			switch {
			case errors.HasType(err, sendError{}):
				__antithesis_instrumentation__.Notify(87456)

				log.VEventf(ctx, 1, "evicting range desc %s after %s", routingTok, err)
				routingTok.Evict(ctx)
				continue
			default:
				__antithesis_instrumentation__.Notify(87457)
			}
			__antithesis_instrumentation__.Notify(87455)
			break
		} else {
			__antithesis_instrumentation__.Notify(87458)
		}
		__antithesis_instrumentation__.Notify(87424)

		if reply.Error == nil {
			__antithesis_instrumentation__.Notify(87459)
			return response{reply: reply, positions: positions}
		} else {
			__antithesis_instrumentation__.Notify(87460)
		}
		__antithesis_instrumentation__.Notify(87425)

		pErr = reply.Error
		reply.Error = nil

		if pErr.Index != nil && func() bool {
			__antithesis_instrumentation__.Notify(87461)
			return pErr.Index.Index != -1 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(87462)
			return positions != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(87463)
			pErr.Index.Index = int32(positions[pErr.Index.Index])
		} else {
			__antithesis_instrumentation__.Notify(87464)
		}
		__antithesis_instrumentation__.Notify(87426)

		log.VErrEventf(ctx, 2, "reply error %s: %s", ba, pErr)

		switch tErr := pErr.GetDetail().(type) {
		case *roachpb.RangeKeyMismatchError:
			__antithesis_instrumentation__.Notify(87465)

			for _, ri := range tErr.Ranges {
				__antithesis_instrumentation__.Notify(87467)

				if routingTok.Desc().RSpan().Equal(ri.Desc.RSpan()) {
					__antithesis_instrumentation__.Notify(87468)
					return response{pErr: roachpb.NewError(errors.AssertionFailedf(
						"mismatched range suggestion not different from original desc. desc: %s. suggested: %s. err: %s",
						routingTok.Desc(), ri.Desc, pErr))}
				} else {
					__antithesis_instrumentation__.Notify(87469)
				}
			}
			__antithesis_instrumentation__.Notify(87466)
			routingTok.EvictAndReplace(ctx, tErr.Ranges...)

			log.VEventf(ctx, 1, "likely split; will resend. Got new descriptors: %s", tErr.Ranges)
			reply, pErr = ds.divideAndSendBatchToRanges(ctx, ba, rs, isReverse, withCommit, batchIdx)
			return response{reply: reply, positions: positions, pErr: pErr}
		}
		__antithesis_instrumentation__.Notify(87427)
		break
	}
	__antithesis_instrumentation__.Notify(87404)

	if pErr == nil {
		__antithesis_instrumentation__.Notify(87470)
		if err := ds.deduceRetryEarlyExitError(ctx); err == nil {
			__antithesis_instrumentation__.Notify(87471)
			log.Fatal(ctx, "exited retry loop without an error")
		} else {
			__antithesis_instrumentation__.Notify(87472)
			pErr = roachpb.NewError(err)
		}
	} else {
		__antithesis_instrumentation__.Notify(87473)
	}
	__antithesis_instrumentation__.Notify(87405)

	return response{pErr: pErr}
}

func (ds *DistSender) deduceRetryEarlyExitError(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(87474)
	select {
	case <-ds.rpcRetryOptions.Closer:
		__antithesis_instrumentation__.Notify(87476)

		return &roachpb.NodeUnavailableError{}
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(87477)

		return errors.Wrap(ctx.Err(), "aborted in DistSender")
	default:
		__antithesis_instrumentation__.Notify(87478)
	}
	__antithesis_instrumentation__.Notify(87475)
	return nil
}

func fillSkippedResponses(
	ba roachpb.BatchRequest,
	br *roachpb.BatchResponse,
	nextKey roachpb.RKey,
	resumeReason roachpb.ResumeReason,
) {
	__antithesis_instrumentation__.Notify(87479)

	var scratchBA roachpb.BatchRequest
	for i := range br.Responses {
		__antithesis_instrumentation__.Notify(87481)
		if br.Responses[i] != (roachpb.ResponseUnion{}) {
			__antithesis_instrumentation__.Notify(87484)
			continue
		} else {
			__antithesis_instrumentation__.Notify(87485)
		}
		__antithesis_instrumentation__.Notify(87482)
		req := ba.Requests[i].GetInner()

		if scratchBA.Requests == nil {
			__antithesis_instrumentation__.Notify(87486)
			scratchBA.Requests = make([]roachpb.RequestUnion, 1)
		} else {
			__antithesis_instrumentation__.Notify(87487)
		}
		__antithesis_instrumentation__.Notify(87483)
		scratchBA.Requests[0].MustSetInner(req)
		br.Responses[i] = scratchBA.CreateReply().Responses[0]
	}
	__antithesis_instrumentation__.Notify(87480)

	isReverse := ba.IsReverse()
	for i, resp := range br.Responses {
		__antithesis_instrumentation__.Notify(87488)
		req := ba.Requests[i].GetInner()
		hdr := resp.GetInner().Header()
		maybeSetResumeSpan(req, &hdr, nextKey, isReverse)
		if hdr.ResumeSpan != nil {
			__antithesis_instrumentation__.Notify(87490)
			hdr.ResumeReason = resumeReason
		} else {
			__antithesis_instrumentation__.Notify(87491)
		}
		__antithesis_instrumentation__.Notify(87489)
		br.Responses[i].GetInner().SetHeader(hdr)
	}
}

func maybeSetResumeSpan(
	req roachpb.Request, hdr *roachpb.ResponseHeader, nextKey roachpb.RKey, isReverse bool,
) {
	__antithesis_instrumentation__.Notify(87492)
	if _, ok := req.(*roachpb.GetRequest); ok {
		__antithesis_instrumentation__.Notify(87495)

		if hdr.ResumeSpan != nil {
			__antithesis_instrumentation__.Notify(87498)

			return
		} else {
			__antithesis_instrumentation__.Notify(87499)
		}
		__antithesis_instrumentation__.Notify(87496)
		key := req.Header().Span().Key
		if isReverse {
			__antithesis_instrumentation__.Notify(87500)
			if !nextKey.Less(roachpb.RKey(key)) {
				__antithesis_instrumentation__.Notify(87501)

				hdr.ResumeSpan = &roachpb.Span{Key: key}
			} else {
				__antithesis_instrumentation__.Notify(87502)
			}
		} else {
			__antithesis_instrumentation__.Notify(87503)
			if !roachpb.RKey(key).Less(nextKey) {
				__antithesis_instrumentation__.Notify(87504)

				hdr.ResumeSpan = &roachpb.Span{Key: key}
			} else {
				__antithesis_instrumentation__.Notify(87505)
			}
		}
		__antithesis_instrumentation__.Notify(87497)
		return
	} else {
		__antithesis_instrumentation__.Notify(87506)
	}
	__antithesis_instrumentation__.Notify(87493)

	if !roachpb.IsRange(req) {
		__antithesis_instrumentation__.Notify(87507)
		return
	} else {
		__antithesis_instrumentation__.Notify(87508)
	}
	__antithesis_instrumentation__.Notify(87494)

	origSpan := req.Header().Span()
	if isReverse {
		__antithesis_instrumentation__.Notify(87509)
		if hdr.ResumeSpan != nil {
			__antithesis_instrumentation__.Notify(87510)

			hdr.ResumeSpan.Key = origSpan.Key
		} else {
			__antithesis_instrumentation__.Notify(87511)
			if roachpb.RKey(origSpan.Key).Less(nextKey) {
				__antithesis_instrumentation__.Notify(87512)

				hdr.ResumeSpan = new(roachpb.Span)
				*hdr.ResumeSpan = origSpan
				if nextKey.Less(roachpb.RKey(origSpan.EndKey)) {
					__antithesis_instrumentation__.Notify(87513)

					hdr.ResumeSpan.EndKey = nextKey.AsRawKey()
				} else {
					__antithesis_instrumentation__.Notify(87514)
				}
			} else {
				__antithesis_instrumentation__.Notify(87515)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(87516)
		if hdr.ResumeSpan != nil {
			__antithesis_instrumentation__.Notify(87517)

			hdr.ResumeSpan.EndKey = origSpan.EndKey
		} else {
			__antithesis_instrumentation__.Notify(87518)

			if nextKey.Less(roachpb.RKey(origSpan.EndKey)) {
				__antithesis_instrumentation__.Notify(87519)

				hdr.ResumeSpan = new(roachpb.Span)
				*hdr.ResumeSpan = origSpan
				if roachpb.RKey(origSpan.Key).Less(nextKey) {
					__antithesis_instrumentation__.Notify(87520)

					hdr.ResumeSpan.Key = nextKey.AsRawKey()
				} else {
					__antithesis_instrumentation__.Notify(87521)
				}
			} else {
				__antithesis_instrumentation__.Notify(87522)
			}
		}
	}
}

func noMoreReplicasErr(ambiguousErr, lastAttemptErr error) error {
	__antithesis_instrumentation__.Notify(87523)
	if ambiguousErr != nil {
		__antithesis_instrumentation__.Notify(87525)
		return roachpb.NewAmbiguousResultErrorf("error=%s [exhausted]", ambiguousErr)
	} else {
		__antithesis_instrumentation__.Notify(87526)
	}
	__antithesis_instrumentation__.Notify(87524)

	return newSendError(fmt.Sprintf("sending to all replicas failed; last error: %s", lastAttemptErr))
}

func (ds *DistSender) sendToReplicas(
	ctx context.Context, ba roachpb.BatchRequest, routing rangecache.EvictionToken, withCommit bool,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(87527)
	desc := routing.Desc()
	ba.RangeID = desc.RangeID

	if ba.RoutingPolicy == roachpb.RoutingPolicy_LEASEHOLDER && func() bool {
		__antithesis_instrumentation__.Notify(87533)
		return CanSendToFollower(ds.logicalClusterID.Get(), ds.st, ds.clock, routing.ClosedTimestampPolicy(), ba) == true
	}() == true {
		__antithesis_instrumentation__.Notify(87534)
		ba.RoutingPolicy = roachpb.RoutingPolicy_NEAREST
	} else {
		__antithesis_instrumentation__.Notify(87535)
	}
	__antithesis_instrumentation__.Notify(87528)

	var replicaFilter ReplicaSliceFilter
	switch ba.RoutingPolicy {
	case roachpb.RoutingPolicy_LEASEHOLDER:
		__antithesis_instrumentation__.Notify(87536)
		replicaFilter = OnlyPotentialLeaseholders
	case roachpb.RoutingPolicy_NEAREST:
		__antithesis_instrumentation__.Notify(87537)
		replicaFilter = AllExtantReplicas
	default:
		__antithesis_instrumentation__.Notify(87538)
		log.Fatalf(ctx, "unknown routing policy: %s", ba.RoutingPolicy)
	}
	__antithesis_instrumentation__.Notify(87529)
	leaseholder := routing.Leaseholder()
	replicas, err := NewReplicaSlice(ctx, ds.nodeDescs, desc, leaseholder, replicaFilter)
	if err != nil {
		__antithesis_instrumentation__.Notify(87539)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(87540)
	}
	__antithesis_instrumentation__.Notify(87530)

	var leaseholderFirst bool
	switch ba.RoutingPolicy {
	case roachpb.RoutingPolicy_LEASEHOLDER:
		__antithesis_instrumentation__.Notify(87541)

		if !ds.dontReorderReplicas {
			__antithesis_instrumentation__.Notify(87546)
			replicas.OptimizeReplicaOrder(ds.getNodeDescriptor(), ds.latencyFunc)
		} else {
			__antithesis_instrumentation__.Notify(87547)
		}
		__antithesis_instrumentation__.Notify(87542)

		idx := -1
		if leaseholder != nil {
			__antithesis_instrumentation__.Notify(87548)
			idx = replicas.Find(leaseholder.ReplicaID)
		} else {
			__antithesis_instrumentation__.Notify(87549)
		}
		__antithesis_instrumentation__.Notify(87543)
		if idx != -1 {
			__antithesis_instrumentation__.Notify(87550)
			replicas.MoveToFront(idx)
			leaseholderFirst = true
		} else {
			__antithesis_instrumentation__.Notify(87551)

			log.VEvent(ctx, 2, "routing to nearest replica; leaseholder not known")
		}

	case roachpb.RoutingPolicy_NEAREST:
		__antithesis_instrumentation__.Notify(87544)

		log.VEvent(ctx, 2, "routing to nearest replica; leaseholder not required")
		replicas.OptimizeReplicaOrder(ds.getNodeDescriptor(), ds.latencyFunc)

	default:
		__antithesis_instrumentation__.Notify(87545)
		log.Fatalf(ctx, "unknown routing policy: %s", ba.RoutingPolicy)
	}
	__antithesis_instrumentation__.Notify(87531)

	opts := SendOptions{
		class:                  rpc.ConnectionClassForKey(desc.RSpan().Key),
		metrics:                &ds.metrics,
		dontConsiderConnHealth: ds.dontConsiderConnHealth,
	}
	transport, err := ds.transportFactory(opts, ds.nodeDialer, replicas)
	if err != nil {
		__antithesis_instrumentation__.Notify(87552)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(87553)
	}
	__antithesis_instrumentation__.Notify(87532)
	defer transport.Release()

	inTransferRetry := retry.StartWithCtx(ctx, ds.rpcRetryOptions)
	inTransferRetry.Next()
	var sameReplicaRetries int
	var prevReplica roachpb.ReplicaDescriptor

	var ambiguousError error
	var br *roachpb.BatchResponse
	for first := true; ; first = false {
		__antithesis_instrumentation__.Notify(87554)
		if !first {
			__antithesis_instrumentation__.Notify(87560)
			ds.metrics.NextReplicaErrCount.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(87561)
		}
		__antithesis_instrumentation__.Notify(87555)

		lastErr := err
		if lastErr == nil && func() bool {
			__antithesis_instrumentation__.Notify(87562)
			return br != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(87563)
			lastErr = br.Error.GoError()
		} else {
			__antithesis_instrumentation__.Notify(87564)
		}
		__antithesis_instrumentation__.Notify(87556)
		err = skipStaleReplicas(transport, routing, ambiguousError, lastErr)
		if err != nil {
			__antithesis_instrumentation__.Notify(87565)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(87566)
		}
		__antithesis_instrumentation__.Notify(87557)
		curReplica := transport.NextReplica()
		if first {
			__antithesis_instrumentation__.Notify(87567)
			if log.ExpensiveLogEnabled(ctx, 2) {
				__antithesis_instrumentation__.Notify(87568)
				log.VEventf(ctx, 2, "r%d: sending batch %s to %s", desc.RangeID, ba.Summary(), curReplica)
			} else {
				__antithesis_instrumentation__.Notify(87569)
			}
		} else {
			__antithesis_instrumentation__.Notify(87570)
			log.VEventf(ctx, 2, "trying next peer %s", curReplica.String())
			if prevReplica == curReplica {
				__antithesis_instrumentation__.Notify(87571)
				sameReplicaRetries++
			} else {
				__antithesis_instrumentation__.Notify(87572)
				sameReplicaRetries = 0
			}
		}
		__antithesis_instrumentation__.Notify(87558)
		prevReplica = curReplica

		ba.ClientRangeInfo = roachpb.ClientRangeInfo{

			DescriptorGeneration: routing.Desc().Generation,

			LeaseSequence: routing.LeaseSeq(),

			ClosedTimestampPolicy: routing.ClosedTimestampPolicy(),

			ExplicitlyRequested: ba.ClientRangeInfo.ExplicitlyRequested,
		}
		br, err = transport.SendNext(ctx, ba)
		ds.maybeIncrementErrCounters(br, err)

		if err != nil {
			__antithesis_instrumentation__.Notify(87573)
			if grpcutil.IsAuthError(err) {
				__antithesis_instrumentation__.Notify(87576)

				if ambiguousError != nil {
					__antithesis_instrumentation__.Notify(87578)
					return nil, roachpb.NewAmbiguousResultErrorf("error=%s [propagate]", ambiguousError)
				} else {
					__antithesis_instrumentation__.Notify(87579)
				}
				__antithesis_instrumentation__.Notify(87577)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(87580)
			}
			__antithesis_instrumentation__.Notify(87574)

			if withCommit && func() bool {
				__antithesis_instrumentation__.Notify(87581)
				return !grpcutil.RequestDidNotStart(err) == true
			}() == true {
				__antithesis_instrumentation__.Notify(87582)
				ambiguousError = err
			} else {
				__antithesis_instrumentation__.Notify(87583)
			}
			__antithesis_instrumentation__.Notify(87575)
			log.VErrEventf(ctx, 2, "RPC error: %s", err)

			if ctx.Err() == nil {
				__antithesis_instrumentation__.Notify(87584)
				if lh := routing.Leaseholder(); lh != nil && func() bool {
					__antithesis_instrumentation__.Notify(87585)
					return *lh == curReplica == true
				}() == true {
					__antithesis_instrumentation__.Notify(87586)
					routing.EvictLease(ctx)
				} else {
					__antithesis_instrumentation__.Notify(87587)
				}
			} else {
				__antithesis_instrumentation__.Notify(87588)
			}
		} else {
			__antithesis_instrumentation__.Notify(87589)

			if br.Error != nil {
				__antithesis_instrumentation__.Notify(87593)
				log.VErrEventf(ctx, 2, "%v", br.Error)
				if !br.Error.Now.IsEmpty() {
					__antithesis_instrumentation__.Notify(87594)
					ds.clock.Update(br.Error.Now)
				} else {
					__antithesis_instrumentation__.Notify(87595)
				}
			} else {
				__antithesis_instrumentation__.Notify(87596)
				if !br.Now.IsEmpty() {
					__antithesis_instrumentation__.Notify(87597)
					ds.clock.Update(br.Now)
				} else {
					__antithesis_instrumentation__.Notify(87598)
				}
			}
			__antithesis_instrumentation__.Notify(87590)

			if br.Error == nil {
				__antithesis_instrumentation__.Notify(87599)

				if len(br.RangeInfos) > 0 {
					__antithesis_instrumentation__.Notify(87601)
					log.VEventf(ctx, 2, "received updated range info: %s", br.RangeInfos)
					routing.EvictAndReplace(ctx, br.RangeInfos...)
					if !ba.Header.ClientRangeInfo.ExplicitlyRequested {
						__antithesis_instrumentation__.Notify(87602)

						br.RangeInfos = nil
					} else {
						__antithesis_instrumentation__.Notify(87603)
					}
				} else {
					__antithesis_instrumentation__.Notify(87604)
				}
				__antithesis_instrumentation__.Notify(87600)
				return br, nil
			} else {
				__antithesis_instrumentation__.Notify(87605)
			}
			__antithesis_instrumentation__.Notify(87591)

			switch tErr := br.Error.GetDetail().(type) {
			case *roachpb.StoreNotFoundError, *roachpb.NodeUnavailableError:
				__antithesis_instrumentation__.Notify(87606)

			case *roachpb.RangeNotFoundError:
				__antithesis_instrumentation__.Notify(87607)

			case *roachpb.NotLeaseHolderError:
				__antithesis_instrumentation__.Notify(87608)
				ds.metrics.NotLeaseHolderErrCount.Inc(1)

				if tErr.Lease != nil || func() bool {
					__antithesis_instrumentation__.Notify(87611)
					return tErr.LeaseHolder != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(87612)

					var updatedLeaseholder bool
					if tErr.Lease != nil {
						__antithesis_instrumentation__.Notify(87616)
						updatedLeaseholder = routing.UpdateLease(ctx, tErr.Lease, tErr.RangeDesc.Generation)
					} else {
						__antithesis_instrumentation__.Notify(87617)
						if tErr.LeaseHolder != nil {
							__antithesis_instrumentation__.Notify(87618)

							routing.UpdateLeaseholder(ctx, *tErr.LeaseHolder, tErr.RangeDesc.Generation)
							updatedLeaseholder = true
						} else {
							__antithesis_instrumentation__.Notify(87619)
						}
					}
					__antithesis_instrumentation__.Notify(87613)

					if lh := routing.Leaseholder(); lh != nil {
						__antithesis_instrumentation__.Notify(87620)

						if *lh != curReplica || func() bool {
							__antithesis_instrumentation__.Notify(87621)
							return sameReplicaRetries < sameReplicaRetryLimit == true
						}() == true {
							__antithesis_instrumentation__.Notify(87622)
							transport.MoveToFront(*lh)
						} else {
							__antithesis_instrumentation__.Notify(87623)
						}
					} else {
						__antithesis_instrumentation__.Notify(87624)
					}
					__antithesis_instrumentation__.Notify(87614)

					intentionallySentToFollower := first && func() bool {
						__antithesis_instrumentation__.Notify(87625)
						return !leaseholderFirst == true
					}() == true

					shouldBackoff := !updatedLeaseholder && func() bool {
						__antithesis_instrumentation__.Notify(87626)
						return !intentionallySentToFollower == true
					}() == true
					if shouldBackoff {
						__antithesis_instrumentation__.Notify(87627)
						ds.metrics.InLeaseTransferBackoffs.Inc(1)
						log.VErrEventf(ctx, 2, "backing off due to NotLeaseHolderErr with stale info")
					} else {
						__antithesis_instrumentation__.Notify(87628)
						inTransferRetry.Reset()
					}
					__antithesis_instrumentation__.Notify(87615)
					inTransferRetry.Next()
				} else {
					__antithesis_instrumentation__.Notify(87629)
				}
			default:
				__antithesis_instrumentation__.Notify(87609)
				if ambiguousError != nil {
					__antithesis_instrumentation__.Notify(87630)
					return nil, roachpb.NewAmbiguousResultErrorf("error=%s [propagate]", ambiguousError)
				} else {
					__antithesis_instrumentation__.Notify(87631)
				}
				__antithesis_instrumentation__.Notify(87610)

				return br, nil
			}
			__antithesis_instrumentation__.Notify(87592)

			log.VErrEventf(ctx, 1, "application error: %s", br.Error)
		}
		__antithesis_instrumentation__.Notify(87559)

		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(87632)

			if ambiguousError != nil {
				__antithesis_instrumentation__.Notify(87634)
				err = roachpb.NewAmbiguousResultError(errors.Wrapf(ambiguousError, "context done during DistSender.Send"))
			} else {
				__antithesis_instrumentation__.Notify(87635)
				err = errors.Wrap(ctx.Err(), "aborted during DistSender.Send")
			}
			__antithesis_instrumentation__.Notify(87633)
			log.Eventf(ctx, "%v", err)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(87636)
		}
	}
}

func (ds *DistSender) maybeIncrementErrCounters(br *roachpb.BatchResponse, err error) {
	__antithesis_instrumentation__.Notify(87637)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(87639)
		return br.Error == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(87640)
		return
	} else {
		__antithesis_instrumentation__.Notify(87641)
	}
	__antithesis_instrumentation__.Notify(87638)
	if err != nil {
		__antithesis_instrumentation__.Notify(87642)
		ds.metrics.ErrCounts[roachpb.CommunicationErrType].Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(87643)
		typ := roachpb.InternalErrType
		if detail := br.Error.GetDetail(); detail != nil {
			__antithesis_instrumentation__.Notify(87645)
			typ = detail.Type()
		} else {
			__antithesis_instrumentation__.Notify(87646)
		}
		__antithesis_instrumentation__.Notify(87644)
		ds.metrics.ErrCounts[typ].Inc(1)
	}
}

func skipStaleReplicas(
	transport Transport, routing rangecache.EvictionToken, ambiguousError error, lastErr error,
) error {
	__antithesis_instrumentation__.Notify(87647)

	if !routing.Valid() {
		__antithesis_instrumentation__.Notify(87649)
		return noMoreReplicasErr(
			ambiguousError,
			errors.Wrap(lastErr, "routing information detected to be stale"))
	} else {
		__antithesis_instrumentation__.Notify(87650)
	}
	__antithesis_instrumentation__.Notify(87648)

	for {
		__antithesis_instrumentation__.Notify(87651)
		if transport.IsExhausted() {
			__antithesis_instrumentation__.Notify(87654)
			return noMoreReplicasErr(ambiguousError, lastErr)
		} else {
			__antithesis_instrumentation__.Notify(87655)
		}
		__antithesis_instrumentation__.Notify(87652)

		if _, ok := routing.Desc().GetReplicaDescriptorByID(transport.NextReplica().ReplicaID); ok {
			__antithesis_instrumentation__.Notify(87656)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(87657)
		}
		__antithesis_instrumentation__.Notify(87653)
		transport.SkipReplica()
	}
}

type sendError struct {
	message string
}

func newSendError(msg string) error {
	__antithesis_instrumentation__.Notify(87658)
	return sendError{message: msg}
}

func TestNewSendError(msg string) error {
	__antithesis_instrumentation__.Notify(87659)
	return newSendError(msg)
}

func (s sendError) Error() string {
	__antithesis_instrumentation__.Notify(87660)
	return "failed to send RPC: " + s.message
}

func IsSendError(err error) bool {
	__antithesis_instrumentation__.Notify(87661)
	return errors.HasType(err, sendError{})
}
