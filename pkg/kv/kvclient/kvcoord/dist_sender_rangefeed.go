package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"math"
	"sync"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/limit"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type singleRangeInfo struct {
	rs        roachpb.RSpan
	startFrom hlc.Timestamp
	token     rangecache.EvictionToken
}

var useDedicatedRangefeedConnectionClass = settings.RegisterBoolSetting(
	settings.SystemOnly,
	"kv.rangefeed.use_dedicated_connection_class.enabled",
	"uses dedicated connection when running rangefeeds",
	util.ConstantWithMetamorphicTestBool(
		"kv.rangefeed.use_dedicated_connection_class.enabled", false),
)

var catchupScanConcurrency = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.rangefeed.catchup_scan_concurrency",
	"number of catchup scans that a single rangefeed can execute concurrently; 0 implies unlimited",
	8,
	settings.NonNegativeInt,
)

func maxConcurrentCatchupScans(sv *settings.Values) int {
	__antithesis_instrumentation__.Notify(87662)
	l := catchupScanConcurrency.Get(sv)
	if l == 0 {
		__antithesis_instrumentation__.Notify(87664)
		return math.MaxInt
	} else {
		__antithesis_instrumentation__.Notify(87665)
	}
	__antithesis_instrumentation__.Notify(87663)
	return int(l)
}

func (ds *DistSender) RangeFeed(
	ctx context.Context,
	spans []roachpb.Span,
	startFrom hlc.Timestamp,
	withDiff bool,
	eventCh chan<- *roachpb.RangeFeedEvent,
) error {
	__antithesis_instrumentation__.Notify(87666)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(87670)
		return errors.AssertionFailedf("expected at least 1 span, got none")
	} else {
		__antithesis_instrumentation__.Notify(87671)
	}
	__antithesis_instrumentation__.Notify(87667)

	ctx = ds.AnnotateCtx(ctx)
	ctx, sp := tracing.EnsureChildSpan(ctx, ds.AmbientContext.Tracer, "dist sender")
	defer sp.Finish()

	rr := newRangeFeedRegistry(ctx, startFrom, withDiff)
	ds.activeRangeFeeds.Store(rr, nil)
	defer ds.activeRangeFeeds.Delete(rr)

	catchupSem := limit.MakeConcurrentRequestLimiter(
		"distSenderCatchupLimit", maxConcurrentCatchupScans(&ds.st.SV))

	g := ctxgroup.WithContext(ctx)

	rangeCh := make(chan singleRangeInfo, 16)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(87672)
		for {
			__antithesis_instrumentation__.Notify(87673)
			select {
			case sri := <-rangeCh:
				__antithesis_instrumentation__.Notify(87674)

				g.GoCtx(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(87676)
					return ds.partialRangeFeed(ctx, rr, sri.rs, sri.startFrom, sri.token, withDiff, &catchupSem, rangeCh, eventCh)
				})
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(87675)
				return ctx.Err()
			}
		}
	})
	__antithesis_instrumentation__.Notify(87668)

	for _, span := range spans {
		__antithesis_instrumentation__.Notify(87677)
		rs, err := keys.SpanAddr(span)
		if err != nil {
			__antithesis_instrumentation__.Notify(87679)
			return err
		} else {
			__antithesis_instrumentation__.Notify(87680)
		}
		__antithesis_instrumentation__.Notify(87678)
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(87681)
			return ds.divideAndSendRangeFeedToRanges(ctx, rs, startFrom, rangeCh)
		})
	}
	__antithesis_instrumentation__.Notify(87669)
	return g.Wait()
}

type RangeFeedContext struct {
	ID      int64
	CtxTags string

	StartFrom hlc.Timestamp
	WithDiff  bool
}

type PartialRangeFeed struct {
	Span              roachpb.Span
	StartTS           hlc.Timestamp
	NodeID            roachpb.NodeID
	RangeID           roachpb.RangeID
	CreatedTime       time.Time
	LastValueReceived time.Time
	Resolved          hlc.Timestamp
}

type ActiveRangeFeedIterFn func(rfCtx RangeFeedContext, feed PartialRangeFeed) error

func (ds *DistSender) ForEachActiveRangeFeed(fn ActiveRangeFeedIterFn) (iterErr error) {
	__antithesis_instrumentation__.Notify(87682)
	const continueIter = true
	const stopIter = false

	partialRangeFeed := func(active *activeRangeFeed) PartialRangeFeed {
		__antithesis_instrumentation__.Notify(87686)
		active.Lock()
		defer active.Unlock()
		return active.PartialRangeFeed
	}
	__antithesis_instrumentation__.Notify(87683)

	ds.activeRangeFeeds.Range(func(k, v interface{}) bool {
		__antithesis_instrumentation__.Notify(87687)
		r := k.(*rangeFeedRegistry)
		r.ranges.Range(func(k, v interface{}) bool {
			__antithesis_instrumentation__.Notify(87689)
			active := k.(*activeRangeFeed)
			if err := fn(r.RangeFeedContext, partialRangeFeed(active)); err != nil {
				__antithesis_instrumentation__.Notify(87691)
				iterErr = err
				return stopIter
			} else {
				__antithesis_instrumentation__.Notify(87692)
			}
			__antithesis_instrumentation__.Notify(87690)
			return continueIter
		})
		__antithesis_instrumentation__.Notify(87688)
		return iterErr == nil
	})
	__antithesis_instrumentation__.Notify(87684)

	if iterutil.Done(iterErr) {
		__antithesis_instrumentation__.Notify(87693)
		iterErr = nil
	} else {
		__antithesis_instrumentation__.Notify(87694)
	}
	__antithesis_instrumentation__.Notify(87685)

	return
}

type activeRangeFeed struct {
	syncutil.Mutex
	PartialRangeFeed
}

func (a *activeRangeFeed) onRangeEvent(
	nodeID roachpb.NodeID, rangeID roachpb.RangeID, event *roachpb.RangeFeedEvent,
) {
	__antithesis_instrumentation__.Notify(87695)
	a.Lock()
	defer a.Unlock()
	if event.Val != nil || func() bool {
		__antithesis_instrumentation__.Notify(87697)
		return event.SST != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(87698)
		a.LastValueReceived = timeutil.Now()
	} else {
		__antithesis_instrumentation__.Notify(87699)
		if event.Checkpoint != nil {
			__antithesis_instrumentation__.Notify(87700)
			a.Resolved = event.Checkpoint.ResolvedTS
		} else {
			__antithesis_instrumentation__.Notify(87701)
		}
	}
	__antithesis_instrumentation__.Notify(87696)

	a.NodeID = nodeID
	a.RangeID = rangeID
}

type rangeFeedRegistry struct {
	RangeFeedContext
	ranges sync.Map
}

func newRangeFeedRegistry(
	ctx context.Context, startFrom hlc.Timestamp, withDiff bool,
) *rangeFeedRegistry {
	__antithesis_instrumentation__.Notify(87702)
	rr := &rangeFeedRegistry{
		RangeFeedContext: RangeFeedContext{
			StartFrom: startFrom,
			WithDiff:  withDiff,
		},
	}
	rr.ID = *(*int64)(unsafe.Pointer(&rr))

	if b := logtags.FromContext(ctx); b != nil {
		__antithesis_instrumentation__.Notify(87704)
		rr.CtxTags = b.String()
	} else {
		__antithesis_instrumentation__.Notify(87705)
	}
	__antithesis_instrumentation__.Notify(87703)
	return rr
}

func (ds *DistSender) divideAndSendRangeFeedToRanges(
	ctx context.Context, rs roachpb.RSpan, startFrom hlc.Timestamp, rangeCh chan<- singleRangeInfo,
) error {
	__antithesis_instrumentation__.Notify(87706)

	nextRS := rs
	ri := MakeRangeIterator(ds)
	for ri.Seek(ctx, nextRS.Key, Ascending); ri.Valid(); ri.Next(ctx) {
		__antithesis_instrumentation__.Notify(87708)
		desc := ri.Desc()
		partialRS, err := nextRS.Intersect(desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(87711)
			return err
		} else {
			__antithesis_instrumentation__.Notify(87712)
		}
		__antithesis_instrumentation__.Notify(87709)
		nextRS.Key = partialRS.EndKey
		select {
		case rangeCh <- singleRangeInfo{
			rs:        partialRS,
			startFrom: startFrom,
			token:     ri.Token(),
		}:
			__antithesis_instrumentation__.Notify(87713)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(87714)
			return ctx.Err()
		}
		__antithesis_instrumentation__.Notify(87710)
		if !ri.NeedAnother(nextRS) {
			__antithesis_instrumentation__.Notify(87715)
			break
		} else {
			__antithesis_instrumentation__.Notify(87716)
		}
	}
	__antithesis_instrumentation__.Notify(87707)
	return ri.Error()
}

func (ds *DistSender) partialRangeFeed(
	ctx context.Context,
	rr *rangeFeedRegistry,
	rs roachpb.RSpan,
	startFrom hlc.Timestamp,
	token rangecache.EvictionToken,
	withDiff bool,
	catchupSem *limit.ConcurrentRequestLimiter,
	rangeCh chan<- singleRangeInfo,
	eventCh chan<- *roachpb.RangeFeedEvent,
) error {
	__antithesis_instrumentation__.Notify(87717)

	span := rs.AsRawSpanWithNoLocals()

	active := &activeRangeFeed{
		PartialRangeFeed: PartialRangeFeed{
			Span:        span,
			StartTS:     startFrom,
			CreatedTime: timeutil.Now(),
		},
	}
	rr.ranges.Store(active, nil)
	ds.metrics.RangefeedRanges.Inc(1)
	defer rr.ranges.Delete(active)
	defer ds.metrics.RangefeedRanges.Dec(1)

	for r := retry.StartWithCtx(ctx, ds.rpcRetryOptions); r.Next(); {
		__antithesis_instrumentation__.Notify(87719)

		if !token.Valid() {
			__antithesis_instrumentation__.Notify(87721)
			var err error
			ri, err := ds.getRoutingInfo(ctx, rs.Key, rangecache.EvictionToken{}, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(87723)
				log.VErrEventf(ctx, 1, "range descriptor re-lookup failed: %s", err)
				if !rangecache.IsRangeLookupErrorRetryable(err) {
					__antithesis_instrumentation__.Notify(87725)
					return err
				} else {
					__antithesis_instrumentation__.Notify(87726)
				}
				__antithesis_instrumentation__.Notify(87724)
				continue
			} else {
				__antithesis_instrumentation__.Notify(87727)
			}
			__antithesis_instrumentation__.Notify(87722)
			token = ri
		} else {
			__antithesis_instrumentation__.Notify(87728)
		}
		__antithesis_instrumentation__.Notify(87720)

		maxTS, err := ds.singleRangeFeed(ctx, span, startFrom, withDiff, token.Desc(),
			catchupSem, eventCh, active.onRangeEvent)

		startFrom.Forward(maxTS)

		if err != nil {
			__antithesis_instrumentation__.Notify(87729)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(87731)
				log.Infof(ctx, "RangeFeed %s disconnected with last checkpoint %s ago: %v",
					span, timeutil.Since(startFrom.GoTime()), err)
			} else {
				__antithesis_instrumentation__.Notify(87732)
			}
			__antithesis_instrumentation__.Notify(87730)
			switch {
			case errors.HasType(err, (*roachpb.StoreNotFoundError)(nil)) || func() bool {
				__antithesis_instrumentation__.Notify(87739)
				return errors.HasType(err, (*roachpb.NodeUnavailableError)(nil)) == true
			}() == true:
				__antithesis_instrumentation__.Notify(87733)

			case IsSendError(err), errors.HasType(err, (*roachpb.RangeNotFoundError)(nil)):
				__antithesis_instrumentation__.Notify(87734)

				token.Evict(ctx)
				token = rangecache.EvictionToken{}
				continue
			case errors.HasType(err, (*roachpb.RangeKeyMismatchError)(nil)):
				__antithesis_instrumentation__.Notify(87735)

				token.Evict(ctx)
				return ds.divideAndSendRangeFeedToRanges(ctx, rs, startFrom, rangeCh)
			case errors.HasType(err, (*roachpb.RangeFeedRetryError)(nil)):
				__antithesis_instrumentation__.Notify(87736)
				var t *roachpb.RangeFeedRetryError
				if ok := errors.As(err, &t); !ok {
					__antithesis_instrumentation__.Notify(87740)
					return errors.AssertionFailedf("wrong error type: %T", err)
				} else {
					__antithesis_instrumentation__.Notify(87741)
				}
				__antithesis_instrumentation__.Notify(87737)
				switch t.Reason {
				case roachpb.RangeFeedRetryError_REASON_REPLICA_REMOVED,
					roachpb.RangeFeedRetryError_REASON_RAFT_SNAPSHOT,
					roachpb.RangeFeedRetryError_REASON_LOGICAL_OPS_MISSING,
					roachpb.RangeFeedRetryError_REASON_SLOW_CONSUMER:
					__antithesis_instrumentation__.Notify(87742)

					continue
				case roachpb.RangeFeedRetryError_REASON_RANGE_SPLIT,
					roachpb.RangeFeedRetryError_REASON_RANGE_MERGED,
					roachpb.RangeFeedRetryError_REASON_NO_LEASEHOLDER:
					__antithesis_instrumentation__.Notify(87743)

					token.Evict(ctx)
					return ds.divideAndSendRangeFeedToRanges(ctx, rs, startFrom, rangeCh)
				default:
					__antithesis_instrumentation__.Notify(87744)
					return errors.AssertionFailedf("unrecognized retryable error type: %T", err)
				}
			default:
				__antithesis_instrumentation__.Notify(87738)
				return err
			}
		} else {
			__antithesis_instrumentation__.Notify(87745)
		}
	}
	__antithesis_instrumentation__.Notify(87718)
	return ctx.Err()
}

type onRangeEventCb func(nodeID roachpb.NodeID, rangeID roachpb.RangeID, event *roachpb.RangeFeedEvent)

func (ds *DistSender) singleRangeFeed(
	ctx context.Context,
	span roachpb.Span,
	startFrom hlc.Timestamp,
	withDiff bool,
	desc *roachpb.RangeDescriptor,
	catchupSem *limit.ConcurrentRequestLimiter,
	eventCh chan<- *roachpb.RangeFeedEvent,
	onRangeEvent onRangeEventCb,
) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(87746)
	args := roachpb.RangeFeedRequest{
		Span: span,
		Header: roachpb.Header{
			Timestamp: startFrom,
			RangeID:   desc.RangeID,
		},
		WithDiff: withDiff,
	}

	var latencyFn LatencyFunc
	if ds.rpcContext != nil {
		__antithesis_instrumentation__.Notify(87752)
		latencyFn = ds.rpcContext.RemoteClocks.Latency
	} else {
		__antithesis_instrumentation__.Notify(87753)
	}
	__antithesis_instrumentation__.Notify(87747)
	replicas, err := NewReplicaSlice(ctx, ds.nodeDescs, desc, nil, AllExtantReplicas)
	if err != nil {
		__antithesis_instrumentation__.Notify(87754)
		return args.Timestamp, err
	} else {
		__antithesis_instrumentation__.Notify(87755)
	}
	__antithesis_instrumentation__.Notify(87748)
	replicas.OptimizeReplicaOrder(ds.getNodeDescriptor(), latencyFn)

	opts := SendOptions{class: connectionClass(&ds.st.SV)}
	transport, err := ds.transportFactory(opts, ds.nodeDialer, replicas)
	if err != nil {
		__antithesis_instrumentation__.Notify(87756)
		return args.Timestamp, err
	} else {
		__antithesis_instrumentation__.Notify(87757)
	}
	__antithesis_instrumentation__.Notify(87749)
	defer transport.Release()

	catchupSem.SetLimit(maxConcurrentCatchupScans(&ds.st.SV))
	catchupRes, err := catchupSem.Begin(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(87758)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(87759)
	}
	__antithesis_instrumentation__.Notify(87750)
	ds.metrics.RangefeedCatchupRanges.Inc(1)
	finishCatchupScan := func() {
		__antithesis_instrumentation__.Notify(87760)
		if catchupRes != nil {
			__antithesis_instrumentation__.Notify(87761)
			catchupRes.Release()
			ds.metrics.RangefeedCatchupRanges.Dec(1)
			catchupRes = nil
		} else {
			__antithesis_instrumentation__.Notify(87762)
		}
	}
	__antithesis_instrumentation__.Notify(87751)

	defer finishCatchupScan()

	for {
		__antithesis_instrumentation__.Notify(87763)
		if transport.IsExhausted() {
			__antithesis_instrumentation__.Notify(87767)
			return args.Timestamp, newSendError(
				fmt.Sprintf("sending to all %d replicas failed", len(replicas)))
		} else {
			__antithesis_instrumentation__.Notify(87768)
		}
		__antithesis_instrumentation__.Notify(87764)

		args.Replica = transport.NextReplica()
		clientCtx, client, err := transport.NextInternalClient(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(87769)
			log.VErrEventf(ctx, 2, "RPC error: %s", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(87770)
		}
		__antithesis_instrumentation__.Notify(87765)

		log.VEventf(ctx, 3, "attempting to create a RangeFeed over replica %s", args.Replica)
		stream, err := client.RangeFeed(clientCtx, &args)
		if err != nil {
			__antithesis_instrumentation__.Notify(87771)
			log.VErrEventf(ctx, 2, "RPC error: %s", err)
			if grpcutil.IsAuthError(err) {
				__antithesis_instrumentation__.Notify(87773)

				return args.Timestamp, err
			} else {
				__antithesis_instrumentation__.Notify(87774)
			}
			__antithesis_instrumentation__.Notify(87772)
			continue
		} else {
			__antithesis_instrumentation__.Notify(87775)
		}
		__antithesis_instrumentation__.Notify(87766)

		for {
			__antithesis_instrumentation__.Notify(87776)
			event, err := stream.Recv()
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(87780)
				return args.Timestamp, nil
			} else {
				__antithesis_instrumentation__.Notify(87781)
			}
			__antithesis_instrumentation__.Notify(87777)
			if err != nil {
				__antithesis_instrumentation__.Notify(87782)
				return args.Timestamp, err
			} else {
				__antithesis_instrumentation__.Notify(87783)
			}
			__antithesis_instrumentation__.Notify(87778)
			switch t := event.GetValue().(type) {
			case *roachpb.RangeFeedCheckpoint:
				__antithesis_instrumentation__.Notify(87784)
				if t.Span.Contains(args.Span) {
					__antithesis_instrumentation__.Notify(87787)

					if !t.ResolvedTS.IsEmpty() && func() bool {
						__antithesis_instrumentation__.Notify(87789)
						return catchupRes != nil == true
					}() == true {
						__antithesis_instrumentation__.Notify(87790)
						finishCatchupScan()
					} else {
						__antithesis_instrumentation__.Notify(87791)
					}
					__antithesis_instrumentation__.Notify(87788)
					args.Timestamp.Forward(t.ResolvedTS.Next())
				} else {
					__antithesis_instrumentation__.Notify(87792)
				}
			case *roachpb.RangeFeedError:
				__antithesis_instrumentation__.Notify(87785)
				log.VErrEventf(ctx, 2, "RangeFeedError: %s", t.Error.GoError())
				if catchupRes != nil {
					__antithesis_instrumentation__.Notify(87793)
					ds.metrics.RangefeedErrorCatchup.Inc(1)
				} else {
					__antithesis_instrumentation__.Notify(87794)
				}
				__antithesis_instrumentation__.Notify(87786)
				return args.Timestamp, t.Error.GoError()
			}
			__antithesis_instrumentation__.Notify(87779)
			onRangeEvent(args.Replica.NodeID, desc.RangeID, event)

			select {
			case eventCh <- event:
				__antithesis_instrumentation__.Notify(87795)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(87796)
				return args.Timestamp, ctx.Err()
			}
		}
	}
}

func connectionClass(sv *settings.Values) rpc.ConnectionClass {
	__antithesis_instrumentation__.Notify(87797)
	if useDedicatedRangefeedConnectionClass.Get(sv) {
		__antithesis_instrumentation__.Notify(87799)
		return rpc.RangefeedClass
	} else {
		__antithesis_instrumentation__.Notify(87800)
	}
	__antithesis_instrumentation__.Notify(87798)
	return rpc.DefaultClass
}
