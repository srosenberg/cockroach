package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/intentresolver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

var RangefeedEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.rangefeed.enabled",
	"if set, rangefeed registration is enabled",
	false,
).WithPublic()

var RangeFeedRefreshInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.rangefeed.closed_timestamp_refresh_interval",
	"the interval at which closed-timestamp updates"+
		"are delivered to rangefeeds; set to 0 to use kv.closed_timestamp.side_transport_interval",
	0,
	settings.NonNegativeDuration,
)

var RangefeedTBIEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.rangefeed.catchup_scan_iterator_optimization.enabled",
	"if true, rangefeeds will use time-bound iterators for catchup-scans when possible",
	util.ConstantWithMetamorphicTestBool("kv.rangefeed.catchup_scan_iterator_optimization.enabled", true),
)

type lockedRangefeedStream struct {
	wrapped rangefeed.Stream
	sendMu  syncutil.Mutex
}

func (s *lockedRangefeedStream) Context() context.Context {
	__antithesis_instrumentation__.Notify(119901)
	return s.wrapped.Context()
}

func (s *lockedRangefeedStream) Send(e *roachpb.RangeFeedEvent) error {
	__antithesis_instrumentation__.Notify(119902)
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	return s.wrapped.Send(e)
}

type rangefeedTxnPusher struct {
	ir *intentresolver.IntentResolver
	r  *Replica
}

func (tp *rangefeedTxnPusher) PushTxns(
	ctx context.Context, txns []enginepb.TxnMeta, ts hlc.Timestamp,
) ([]*roachpb.Transaction, error) {
	__antithesis_instrumentation__.Notify(119903)
	pushTxnMap := make(map[uuid.UUID]*enginepb.TxnMeta, len(txns))
	for i := range txns {
		__antithesis_instrumentation__.Notify(119907)
		txn := &txns[i]
		pushTxnMap[txn.ID] = txn
	}
	__antithesis_instrumentation__.Notify(119904)

	h := roachpb.Header{
		Timestamp: ts,
		Txn: &roachpb.Transaction{
			TxnMeta: enginepb.TxnMeta{
				Priority: enginepb.MaxTxnPriority,
			},
		},
	}

	pushedTxnMap, pErr := tp.ir.MaybePushTransactions(
		ctx, pushTxnMap, h, roachpb.PUSH_TIMESTAMP, false,
	)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(119908)
		return nil, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(119909)
	}
	__antithesis_instrumentation__.Notify(119905)

	pushedTxns := make([]*roachpb.Transaction, 0, len(pushedTxnMap))
	for _, txn := range pushedTxnMap {
		__antithesis_instrumentation__.Notify(119910)
		pushedTxns = append(pushedTxns, txn)
	}
	__antithesis_instrumentation__.Notify(119906)
	return pushedTxns, nil
}

func (tp *rangefeedTxnPusher) ResolveIntents(
	ctx context.Context, intents []roachpb.LockUpdate,
) error {
	__antithesis_instrumentation__.Notify(119911)
	return tp.ir.ResolveIntents(ctx, intents,

		intentresolver.ResolveOptions{Poison: true},
	).GoError()
}

func (r *Replica) RangeFeed(
	args *roachpb.RangeFeedRequest, stream rangefeed.Stream,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(119912)
	return r.rangeFeedWithRangeID(r.RangeID, args, stream)
}

func (r *Replica) rangeFeedWithRangeID(
	_forStacks roachpb.RangeID, args *roachpb.RangeFeedRequest, stream rangefeed.Stream,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(119913)
	if !r.isRangefeedEnabled() && func() bool {
		__antithesis_instrumentation__.Notify(119921)
		return !RangefeedEnabled.Get(&r.store.cfg.Settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(119922)
		return roachpb.NewErrorf("rangefeeds require the kv.rangefeed.enabled setting. See %s",
			docs.URL(`change-data-capture.html#enable-rangefeeds-to-reduce-latency`))
	} else {
		__antithesis_instrumentation__.Notify(119923)
	}
	__antithesis_instrumentation__.Notify(119914)
	ctx := r.AnnotateCtx(stream.Context())

	rSpan, err := keys.SpanAddr(args.Span)
	if err != nil {
		__antithesis_instrumentation__.Notify(119924)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(119925)
	}
	__antithesis_instrumentation__.Notify(119915)

	if err := r.ensureClosedTimestampStarted(ctx); err != nil {
		__antithesis_instrumentation__.Notify(119926)
		if err := stream.Send(&roachpb.RangeFeedEvent{Error: &roachpb.RangeFeedError{
			Error: *err,
		}}); err != nil {
			__antithesis_instrumentation__.Notify(119928)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(119929)
		}
		__antithesis_instrumentation__.Notify(119927)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(119930)
	}
	__antithesis_instrumentation__.Notify(119916)

	checkTS := args.Timestamp
	if checkTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(119931)

		checkTS = r.Clock().Now()
	} else {
		__antithesis_instrumentation__.Notify(119932)
	}
	__antithesis_instrumentation__.Notify(119917)

	lockedStream := &lockedRangefeedStream{wrapped: stream}
	errC := make(chan *roachpb.Error, 1)

	usingCatchUpIter := false
	var iterSemRelease func()
	if !args.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(119933)
		usingCatchUpIter = true
		alloc, err := r.store.limiters.ConcurrentRangefeedIters.Begin(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(119936)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(119937)
		}
		__antithesis_instrumentation__.Notify(119934)

		var iterSemReleaseOnce sync.Once
		iterSemRelease = func() {
			__antithesis_instrumentation__.Notify(119938)
			iterSemReleaseOnce.Do(alloc.Release)
		}
		__antithesis_instrumentation__.Notify(119935)
		defer iterSemRelease()
	} else {
		__antithesis_instrumentation__.Notify(119939)
	}
	__antithesis_instrumentation__.Notify(119918)

	r.raftMu.Lock()
	if err := r.checkExecutionCanProceedForRangeFeed(ctx, rSpan, checkTS); err != nil {
		__antithesis_instrumentation__.Notify(119940)
		r.raftMu.Unlock()
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(119941)
	}
	__antithesis_instrumentation__.Notify(119919)

	var catchUpIterFunc rangefeed.CatchUpIteratorConstructor
	if usingCatchUpIter {
		__antithesis_instrumentation__.Notify(119942)
		catchUpIterFunc = func() *rangefeed.CatchUpIterator {
			__antithesis_instrumentation__.Notify(119943)

			r.raftMu.AssertHeld()
			return rangefeed.NewCatchUpIterator(r.Engine(),
				args, RangefeedTBIEnabled.Get(&r.store.cfg.Settings.SV), iterSemRelease)
		}
	} else {
		__antithesis_instrumentation__.Notify(119944)
	}
	__antithesis_instrumentation__.Notify(119920)
	p := r.registerWithRangefeedRaftMuLocked(
		ctx, rSpan, args.Timestamp, catchUpIterFunc, args.WithDiff, lockedStream, errC,
	)
	r.raftMu.Unlock()

	defer r.maybeDisconnectEmptyRangefeed(p)

	return <-errC
}

func (r *Replica) getRangefeedProcessorAndFilter() (*rangefeed.Processor, *rangefeed.Filter) {
	__antithesis_instrumentation__.Notify(119945)
	r.rangefeedMu.RLock()
	defer r.rangefeedMu.RUnlock()
	return r.rangefeedMu.proc, r.rangefeedMu.opFilter
}

func (r *Replica) getRangefeedProcessor() *rangefeed.Processor {
	__antithesis_instrumentation__.Notify(119946)
	p, _ := r.getRangefeedProcessorAndFilter()
	return p
}

func (r *Replica) setRangefeedProcessor(p *rangefeed.Processor) {
	__antithesis_instrumentation__.Notify(119947)
	r.rangefeedMu.Lock()
	defer r.rangefeedMu.Unlock()
	r.rangefeedMu.proc = p
	r.store.addReplicaWithRangefeed(r.RangeID)
}

func (r *Replica) unsetRangefeedProcessorLocked(p *rangefeed.Processor) {
	__antithesis_instrumentation__.Notify(119948)
	if r.rangefeedMu.proc != p {
		__antithesis_instrumentation__.Notify(119950)

		return
	} else {
		__antithesis_instrumentation__.Notify(119951)
	}
	__antithesis_instrumentation__.Notify(119949)
	r.rangefeedMu.proc = nil
	r.rangefeedMu.opFilter = nil
	r.store.removeReplicaWithRangefeed(r.RangeID)
}

func (r *Replica) unsetRangefeedProcessor(p *rangefeed.Processor) {
	__antithesis_instrumentation__.Notify(119952)
	r.rangefeedMu.Lock()
	defer r.rangefeedMu.Unlock()
	r.unsetRangefeedProcessorLocked(p)
}

func (r *Replica) setRangefeedFilterLocked(f *rangefeed.Filter) {
	__antithesis_instrumentation__.Notify(119953)
	if f == nil {
		__antithesis_instrumentation__.Notify(119955)
		panic("filter nil")
	} else {
		__antithesis_instrumentation__.Notify(119956)
	}
	__antithesis_instrumentation__.Notify(119954)
	r.rangefeedMu.opFilter = f
}

func (r *Replica) updateRangefeedFilterLocked() bool {
	__antithesis_instrumentation__.Notify(119957)
	f := r.rangefeedMu.proc.Filter()

	if f != nil {
		__antithesis_instrumentation__.Notify(119959)
		r.setRangefeedFilterLocked(f)
		return true
	} else {
		__antithesis_instrumentation__.Notify(119960)
	}
	__antithesis_instrumentation__.Notify(119958)
	return false
}

const defaultEventChanCap = 4096

func logSlowRangefeedRegistration(ctx context.Context) func() {
	__antithesis_instrumentation__.Notify(119961)
	const slowRaftMuWarnThreshold = 20 * time.Millisecond
	start := timeutil.Now()
	return func() {
		__antithesis_instrumentation__.Notify(119962)
		elapsed := timeutil.Since(start)
		if elapsed >= slowRaftMuWarnThreshold {
			__antithesis_instrumentation__.Notify(119963)
			log.Warningf(ctx, "rangefeed registration took %s", elapsed)
		} else {
			__antithesis_instrumentation__.Notify(119964)
		}
	}
}

func (r *Replica) registerWithRangefeedRaftMuLocked(
	ctx context.Context,
	span roachpb.RSpan,
	startTS hlc.Timestamp,
	catchUpIter rangefeed.CatchUpIteratorConstructor,
	withDiff bool,
	stream rangefeed.Stream,
	errC chan<- *roachpb.Error,
) *rangefeed.Processor {
	__antithesis_instrumentation__.Notify(119965)
	defer logSlowRangefeedRegistration(ctx)()

	r.rangefeedMu.Lock()
	p := r.rangefeedMu.proc
	if p != nil {
		__antithesis_instrumentation__.Notify(119970)
		reg, filter := p.Register(span, startTS, catchUpIter, withDiff, stream, errC)
		if reg {
			__antithesis_instrumentation__.Notify(119972)

			r.setRangefeedFilterLocked(filter)
			r.rangefeedMu.Unlock()
			return p
		} else {
			__antithesis_instrumentation__.Notify(119973)
		}
		__antithesis_instrumentation__.Notify(119971)

		r.unsetRangefeedProcessorLocked(p)
		p = nil
	} else {
		__antithesis_instrumentation__.Notify(119974)
	}
	__antithesis_instrumentation__.Notify(119966)
	r.rangefeedMu.Unlock()

	feedBudget := r.store.GetStoreConfig().RangefeedBudgetFactory.CreateBudget(r.startKey)

	desc := r.Desc()
	tp := rangefeedTxnPusher{ir: r.store.intentResolver, r: r}
	cfg := rangefeed.Config{
		AmbientContext:   r.AmbientContext,
		Clock:            r.Clock(),
		RangeID:          r.RangeID,
		Span:             desc.RSpan(),
		TxnPusher:        &tp,
		PushTxnsInterval: r.store.TestingKnobs().RangeFeedPushTxnsInterval,
		PushTxnsAge:      r.store.TestingKnobs().RangeFeedPushTxnsAge,
		EventChanCap:     defaultEventChanCap,
		EventChanTimeout: 50 * time.Millisecond,
		Metrics:          r.store.metrics.RangeFeedMetrics,
		MemBudget:        feedBudget,
	}
	p = rangefeed.NewProcessor(cfg)

	rtsIter := func() rangefeed.IntentScanner {
		__antithesis_instrumentation__.Notify(119975)

		r.raftMu.AssertHeld()

		lowerBound, _ := keys.LockTableSingleKey(desc.StartKey.AsRawKey(), nil)
		upperBound, _ := keys.LockTableSingleKey(desc.EndKey.AsRawKey(), nil)
		iter := r.Engine().NewEngineIterator(storage.IterOptions{
			LowerBound: lowerBound,
			UpperBound: upperBound,
		})
		return rangefeed.NewSeparatedIntentScanner(iter)
	}
	__antithesis_instrumentation__.Notify(119967)

	if err := p.Start(r.store.Stopper(), rtsIter); err != nil {
		__antithesis_instrumentation__.Notify(119976)
		errC <- roachpb.NewError(err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(119977)
	}
	__antithesis_instrumentation__.Notify(119968)

	reg, filter := p.Register(span, startTS, catchUpIter, withDiff, stream, errC)
	if !reg {
		__antithesis_instrumentation__.Notify(119978)
		select {
		case <-r.store.Stopper().ShouldQuiesce():
			__antithesis_instrumentation__.Notify(119979)
			errC <- roachpb.NewError(&roachpb.NodeUnavailableError{})
			return nil
		default:
			__antithesis_instrumentation__.Notify(119980)
			panic("unexpected Stopped processor")
		}
	} else {
		__antithesis_instrumentation__.Notify(119981)
	}
	__antithesis_instrumentation__.Notify(119969)

	r.setRangefeedProcessor(p)
	r.setRangefeedFilterLocked(filter)

	r.handleClosedTimestampUpdateRaftMuLocked(ctx, r.GetClosedTimestamp(ctx))

	return p
}

func (r *Replica) maybeDisconnectEmptyRangefeed(p *rangefeed.Processor) {
	__antithesis_instrumentation__.Notify(119982)
	r.rangefeedMu.Lock()
	defer r.rangefeedMu.Unlock()
	if p == nil || func() bool {
		__antithesis_instrumentation__.Notify(119984)
		return p != r.rangefeedMu.proc == true
	}() == true {
		__antithesis_instrumentation__.Notify(119985)

		return
	} else {
		__antithesis_instrumentation__.Notify(119986)
	}
	__antithesis_instrumentation__.Notify(119983)
	if p.Len() == 0 || func() bool {
		__antithesis_instrumentation__.Notify(119987)
		return !r.updateRangefeedFilterLocked() == true
	}() == true {
		__antithesis_instrumentation__.Notify(119988)

		p.Stop()
		r.unsetRangefeedProcessorLocked(p)
	} else {
		__antithesis_instrumentation__.Notify(119989)
	}
}

func (r *Replica) disconnectRangefeedWithErr(p *rangefeed.Processor, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(119990)
	p.StopWithErr(pErr)
	r.unsetRangefeedProcessor(p)
}

func (r *Replica) disconnectRangefeedSpanWithErr(span roachpb.Span, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(119991)
	p := r.getRangefeedProcessor()
	if p == nil {
		__antithesis_instrumentation__.Notify(119993)
		return
	} else {
		__antithesis_instrumentation__.Notify(119994)
	}
	__antithesis_instrumentation__.Notify(119992)
	p.DisconnectSpanWithErr(span, pErr)
	r.maybeDisconnectEmptyRangefeed(p)
}

func (r *Replica) disconnectRangefeedWithReason(reason roachpb.RangeFeedRetryError_Reason) {
	__antithesis_instrumentation__.Notify(119995)
	p := r.getRangefeedProcessor()
	if p == nil {
		__antithesis_instrumentation__.Notify(119997)
		return
	} else {
		__antithesis_instrumentation__.Notify(119998)
	}
	__antithesis_instrumentation__.Notify(119996)
	pErr := roachpb.NewError(roachpb.NewRangeFeedRetryError(reason))
	r.disconnectRangefeedWithErr(p, pErr)
}

func (r *Replica) numRangefeedRegistrations() int {
	__antithesis_instrumentation__.Notify(119999)
	p := r.getRangefeedProcessor()
	if p == nil {
		__antithesis_instrumentation__.Notify(120001)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(120002)
	}
	__antithesis_instrumentation__.Notify(120000)
	return p.Len()
}

func (r *Replica) populatePrevValsInLogicalOpLogRaftMuLocked(
	ctx context.Context, ops *kvserverpb.LogicalOpLog, prevReader storage.Reader,
) {
	__antithesis_instrumentation__.Notify(120003)
	p, filter := r.getRangefeedProcessorAndFilter()
	if p == nil {
		__antithesis_instrumentation__.Notify(120005)
		return
	} else {
		__antithesis_instrumentation__.Notify(120006)
	}
	__antithesis_instrumentation__.Notify(120004)

	for _, op := range ops.Ops {
		__antithesis_instrumentation__.Notify(120007)
		var key []byte
		var ts hlc.Timestamp
		var prevValPtr *[]byte
		switch t := op.GetValue().(type) {
		case *enginepb.MVCCWriteValueOp:
			__antithesis_instrumentation__.Notify(120011)
			key, ts, prevValPtr = t.Key, t.Timestamp, &t.PrevValue
		case *enginepb.MVCCCommitIntentOp:
			__antithesis_instrumentation__.Notify(120012)
			key, ts, prevValPtr = t.Key, t.Timestamp, &t.PrevValue
		case *enginepb.MVCCWriteIntentOp,
			*enginepb.MVCCUpdateIntentOp,
			*enginepb.MVCCAbortIntentOp,
			*enginepb.MVCCAbortTxnOp:
			__antithesis_instrumentation__.Notify(120013)

			continue
		default:
			__antithesis_instrumentation__.Notify(120014)
			panic(errors.AssertionFailedf("unknown logical op %T", t))
		}
		__antithesis_instrumentation__.Notify(120008)

		if !filter.NeedPrevVal(roachpb.Span{Key: key}) {
			__antithesis_instrumentation__.Notify(120015)
			continue
		} else {
			__antithesis_instrumentation__.Notify(120016)
		}
		__antithesis_instrumentation__.Notify(120009)

		prevVal, _, err := storage.MVCCGet(
			ctx, prevReader, key, ts, storage.MVCCGetOptions{Tombstones: true, Inconsistent: true},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(120017)
			r.disconnectRangefeedWithErr(p, roachpb.NewErrorf(
				"error consuming %T for key %v @ ts %v: %v", op, key, ts, err,
			))
			return
		} else {
			__antithesis_instrumentation__.Notify(120018)
		}
		__antithesis_instrumentation__.Notify(120010)
		if prevVal != nil {
			__antithesis_instrumentation__.Notify(120019)
			*prevValPtr = prevVal.RawBytes
		} else {
			__antithesis_instrumentation__.Notify(120020)
			*prevValPtr = nil
		}
	}
}

func (r *Replica) handleLogicalOpLogRaftMuLocked(
	ctx context.Context, ops *kvserverpb.LogicalOpLog, reader storage.Reader,
) {
	__antithesis_instrumentation__.Notify(120021)
	p, filter := r.getRangefeedProcessorAndFilter()
	if p == nil {
		__antithesis_instrumentation__.Notify(120026)
		return
	} else {
		__antithesis_instrumentation__.Notify(120027)
	}
	__antithesis_instrumentation__.Notify(120022)
	if ops == nil {
		__antithesis_instrumentation__.Notify(120028)

		r.disconnectRangefeedWithReason(roachpb.RangeFeedRetryError_REASON_LOGICAL_OPS_MISSING)
		return
	} else {
		__antithesis_instrumentation__.Notify(120029)
	}
	__antithesis_instrumentation__.Notify(120023)
	if len(ops.Ops) == 0 {
		__antithesis_instrumentation__.Notify(120030)
		return
	} else {
		__antithesis_instrumentation__.Notify(120031)
	}
	__antithesis_instrumentation__.Notify(120024)

	for _, op := range ops.Ops {
		__antithesis_instrumentation__.Notify(120032)
		var key []byte
		var ts hlc.Timestamp
		var valPtr *[]byte
		switch t := op.GetValue().(type) {
		case *enginepb.MVCCWriteValueOp:
			__antithesis_instrumentation__.Notify(120037)
			key, ts, valPtr = t.Key, t.Timestamp, &t.Value
		case *enginepb.MVCCCommitIntentOp:
			__antithesis_instrumentation__.Notify(120038)
			key, ts, valPtr = t.Key, t.Timestamp, &t.Value
		case *enginepb.MVCCWriteIntentOp,
			*enginepb.MVCCUpdateIntentOp,
			*enginepb.MVCCAbortIntentOp,
			*enginepb.MVCCAbortTxnOp:
			__antithesis_instrumentation__.Notify(120039)

			continue
		default:
			__antithesis_instrumentation__.Notify(120040)
			panic(errors.AssertionFailedf("unknown logical op %T", t))
		}
		__antithesis_instrumentation__.Notify(120033)

		if !filter.NeedVal(roachpb.Span{Key: key}) {
			__antithesis_instrumentation__.Notify(120041)
			continue
		} else {
			__antithesis_instrumentation__.Notify(120042)
		}
		__antithesis_instrumentation__.Notify(120034)

		val, _, err := storage.MVCCGet(ctx, reader, key, ts, storage.MVCCGetOptions{Tombstones: true})
		if val == nil && func() bool {
			__antithesis_instrumentation__.Notify(120043)
			return err == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(120044)
			err = errors.New("value missing in reader")
		} else {
			__antithesis_instrumentation__.Notify(120045)
		}
		__antithesis_instrumentation__.Notify(120035)
		if err != nil {
			__antithesis_instrumentation__.Notify(120046)
			r.disconnectRangefeedWithErr(p, roachpb.NewErrorf(
				"error consuming %T for key %v @ ts %v: %v", op, key, ts, err,
			))
			return
		} else {
			__antithesis_instrumentation__.Notify(120047)
		}
		__antithesis_instrumentation__.Notify(120036)
		*valPtr = val.RawBytes
	}
	__antithesis_instrumentation__.Notify(120025)

	if !p.ConsumeLogicalOps(ctx, ops.Ops...) {
		__antithesis_instrumentation__.Notify(120048)

		r.unsetRangefeedProcessor(p)
	} else {
		__antithesis_instrumentation__.Notify(120049)
	}
}

func (r *Replica) handleSSTableRaftMuLocked(
	ctx context.Context, sst []byte, sstSpan roachpb.Span, writeTS hlc.Timestamp,
) {
	__antithesis_instrumentation__.Notify(120050)
	p, _ := r.getRangefeedProcessorAndFilter()
	if p == nil {
		__antithesis_instrumentation__.Notify(120052)
		return
	} else {
		__antithesis_instrumentation__.Notify(120053)
	}
	__antithesis_instrumentation__.Notify(120051)
	if !p.ConsumeSSTable(ctx, sst, sstSpan, writeTS) {
		__antithesis_instrumentation__.Notify(120054)
		r.unsetRangefeedProcessor(p)
	} else {
		__antithesis_instrumentation__.Notify(120055)
	}
}

func (r *Replica) handleClosedTimestampUpdate(ctx context.Context, closedTS hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(120056)
	ctx = r.AnnotateCtx(ctx)
	r.raftMu.Lock()
	defer r.raftMu.Unlock()
	r.handleClosedTimestampUpdateRaftMuLocked(ctx, closedTS)
}

func (r *Replica) handleClosedTimestampUpdateRaftMuLocked(
	ctx context.Context, closedTS hlc.Timestamp,
) {
	__antithesis_instrumentation__.Notify(120057)
	p := r.getRangefeedProcessor()
	if p == nil {
		__antithesis_instrumentation__.Notify(120061)
		return
	} else {
		__antithesis_instrumentation__.Notify(120062)
	}
	__antithesis_instrumentation__.Notify(120058)

	behind := r.Clock().PhysicalTime().Sub(closedTS.GoTime())
	slowClosedTSThresh := 5 * closedts.TargetDuration.Get(&r.store.cfg.Settings.SV)
	if behind > slowClosedTSThresh {
		__antithesis_instrumentation__.Notify(120063)
		m := r.store.metrics.RangeFeedMetrics
		if m.RangeFeedSlowClosedTimestampLogN.ShouldLog() {
			__antithesis_instrumentation__.Notify(120066)
			if closedTS.IsEmpty() {
				__antithesis_instrumentation__.Notify(120067)
				log.Infof(ctx, "RangeFeed closed timestamp is empty")
			} else {
				__antithesis_instrumentation__.Notify(120068)
				log.Infof(ctx, "RangeFeed closed timestamp %s is behind by %s", closedTS, behind)
			}
		} else {
			__antithesis_instrumentation__.Notify(120069)
		}
		__antithesis_instrumentation__.Notify(120064)

		key := fmt.Sprintf(`rangefeed-slow-closed-timestamp-nudge-r%d`, r.RangeID)

		taskCtx, sp := tracing.EnsureForkSpan(ctx, r.AmbientContext.Tracer, key)
		_, leader := m.RangeFeedSlowClosedTimestampNudge.DoChan(key, func() (interface{}, error) {
			__antithesis_instrumentation__.Notify(120070)
			defer sp.Finish()

			_ = r.store.stopper.RunTask(taskCtx, key, func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(120072)

				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(120075)

					return
				case m.RangeFeedSlowClosedTimestampNudgeSem <- struct{}{}:
					__antithesis_instrumentation__.Notify(120076)
				}
				__antithesis_instrumentation__.Notify(120073)
				defer func() { __antithesis_instrumentation__.Notify(120077); <-m.RangeFeedSlowClosedTimestampNudgeSem }()
				__antithesis_instrumentation__.Notify(120074)
				if err := r.ensureClosedTimestampStarted(ctx); err != nil {
					__antithesis_instrumentation__.Notify(120078)
					log.Infof(ctx, `RangeFeed failed to nudge: %s`, err)
				} else {
					__antithesis_instrumentation__.Notify(120079)
				}
			})
			__antithesis_instrumentation__.Notify(120071)
			return nil, nil
		})
		__antithesis_instrumentation__.Notify(120065)
		if !leader {
			__antithesis_instrumentation__.Notify(120080)

			sp.Finish()
		} else {
			__antithesis_instrumentation__.Notify(120081)
		}
	} else {
		__antithesis_instrumentation__.Notify(120082)
	}
	__antithesis_instrumentation__.Notify(120059)

	if closedTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(120083)
		return
	} else {
		__antithesis_instrumentation__.Notify(120084)
	}
	__antithesis_instrumentation__.Notify(120060)
	if !p.ForwardClosedTS(ctx, closedTS) {
		__antithesis_instrumentation__.Notify(120085)

		r.unsetRangefeedProcessor(p)
	} else {
		__antithesis_instrumentation__.Notify(120086)
	}
}

func (r *Replica) ensureClosedTimestampStarted(ctx context.Context) *roachpb.Error {
	__antithesis_instrumentation__.Notify(120087)

	lease := r.CurrentLeaseStatus(ctx)

	if !lease.IsValid() {
		__antithesis_instrumentation__.Notify(120089)

		log.VEventf(ctx, 2, "ensuring lease for rangefeed range. current lease invalid: %s", lease.Lease)
		err := contextutil.RunWithTimeout(ctx, "read forcing lease acquisition", 5*time.Second,
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(120091)
				var b kv.Batch
				liReq := &roachpb.LeaseInfoRequest{}
				liReq.Key = r.Desc().StartKey.AsRawKey()
				b.AddRawRequest(liReq)
				return r.store.DB().Run(ctx, &b)
			})
		__antithesis_instrumentation__.Notify(120090)
		if err != nil {
			__antithesis_instrumentation__.Notify(120092)
			if errors.HasType(err, (*contextutil.TimeoutError)(nil)) {
				__antithesis_instrumentation__.Notify(120094)
				err = &roachpb.RangeFeedRetryError{
					Reason: roachpb.RangeFeedRetryError_REASON_NO_LEASEHOLDER,
				}
			} else {
				__antithesis_instrumentation__.Notify(120095)
			}
			__antithesis_instrumentation__.Notify(120093)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(120096)
		}
	} else {
		__antithesis_instrumentation__.Notify(120097)
	}
	__antithesis_instrumentation__.Notify(120088)
	return nil
}
