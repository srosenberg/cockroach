package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

const (
	defaultPushTxnsInterval = 250 * time.Millisecond

	defaultPushTxnsAge = 10 * time.Second

	defaultCheckStreamsInterval = 1 * time.Second
)

func newErrBufferCapacityExceeded() *roachpb.Error {
	__antithesis_instrumentation__.Notify(113627)
	return roachpb.NewError(
		roachpb.NewRangeFeedRetryError(roachpb.RangeFeedRetryError_REASON_SLOW_CONSUMER),
	)
}

type Config struct {
	log.AmbientContext
	Clock   *hlc.Clock
	RangeID roachpb.RangeID
	Span    roachpb.RSpan

	TxnPusher TxnPusher

	PushTxnsInterval time.Duration

	PushTxnsAge time.Duration

	EventChanCap int

	EventChanTimeout time.Duration

	CheckStreamsInterval time.Duration

	Metrics *Metrics

	MemBudget *FeedBudget
}

func (sc *Config) SetDefaults() {
	__antithesis_instrumentation__.Notify(113628)
	if sc.TxnPusher == nil {
		__antithesis_instrumentation__.Notify(113630)
		if sc.PushTxnsInterval != 0 {
			__antithesis_instrumentation__.Notify(113632)
			panic("nil TxnPusher with non-zero PushTxnsInterval")
		} else {
			__antithesis_instrumentation__.Notify(113633)
		}
		__antithesis_instrumentation__.Notify(113631)
		if sc.PushTxnsAge != 0 {
			__antithesis_instrumentation__.Notify(113634)
			panic("nil TxnPusher with non-zero PushTxnsAge")
		} else {
			__antithesis_instrumentation__.Notify(113635)
		}
	} else {
		__antithesis_instrumentation__.Notify(113636)
		if sc.PushTxnsInterval == 0 {
			__antithesis_instrumentation__.Notify(113638)
			sc.PushTxnsInterval = defaultPushTxnsInterval
		} else {
			__antithesis_instrumentation__.Notify(113639)
		}
		__antithesis_instrumentation__.Notify(113637)
		if sc.PushTxnsAge == 0 {
			__antithesis_instrumentation__.Notify(113640)
			sc.PushTxnsAge = defaultPushTxnsAge
		} else {
			__antithesis_instrumentation__.Notify(113641)
		}
	}
	__antithesis_instrumentation__.Notify(113629)
	if sc.CheckStreamsInterval == 0 {
		__antithesis_instrumentation__.Notify(113642)
		sc.CheckStreamsInterval = defaultCheckStreamsInterval
	} else {
		__antithesis_instrumentation__.Notify(113643)
	}
}

type Processor struct {
	Config
	reg registry
	rts resolvedTimestamp

	regC       chan registration
	unregC     chan *registration
	lenReqC    chan struct{}
	lenResC    chan int
	filterReqC chan struct{}
	filterResC chan *Filter
	eventC     chan *event
	spanErrC   chan spanErr
	stopC      chan *roachpb.Error
	stoppedC   chan struct{}
}

var eventSyncPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(113644)
		return new(event)
	},
}

func getPooledEvent(ev event) *event {
	__antithesis_instrumentation__.Notify(113645)
	e := eventSyncPool.Get().(*event)
	*e = ev
	return e
}

func putPooledEvent(ev *event) {
	__antithesis_instrumentation__.Notify(113646)
	*ev = event{}
	eventSyncPool.Put(ev)
}

type event struct {
	ops     []enginepb.MVCCLogicalOp
	ct      hlc.Timestamp
	sst     []byte
	sstSpan roachpb.Span
	sstWTS  hlc.Timestamp
	initRTS bool
	syncC   chan struct{}

	testRegCatchupSpan roachpb.Span

	allocation *SharedBudgetAllocation
}

type spanErr struct {
	span roachpb.Span
	pErr *roachpb.Error
}

func NewProcessor(cfg Config) *Processor {
	__antithesis_instrumentation__.Notify(113647)
	cfg.SetDefaults()
	cfg.AmbientContext.AddLogTag("rangefeed", nil)
	p := &Processor{
		Config: cfg,
		reg:    makeRegistry(),
		rts:    makeResolvedTimestamp(),

		regC:       make(chan registration),
		unregC:     make(chan *registration),
		lenReqC:    make(chan struct{}),
		lenResC:    make(chan int),
		filterReqC: make(chan struct{}),
		filterResC: make(chan *Filter),
		eventC:     make(chan *event, cfg.EventChanCap),
		spanErrC:   make(chan spanErr),
		stopC:      make(chan *roachpb.Error, 1),
		stoppedC:   make(chan struct{}),
	}
	return p
}

type IntentScannerConstructor func() IntentScanner

type CatchUpIteratorConstructor func() *CatchUpIterator

func (p *Processor) Start(stopper *stop.Stopper, rtsIterFunc IntentScannerConstructor) error {
	__antithesis_instrumentation__.Notify(113648)
	ctx := p.AnnotateCtx(context.Background())
	if err := stopper.RunAsyncTask(ctx, "rangefeed.Processor", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(113650)
		p.run(ctx, p.RangeID, rtsIterFunc, stopper)
	}); err != nil {
		__antithesis_instrumentation__.Notify(113651)
		p.reg.DisconnectWithErr(all, roachpb.NewError(err))
		close(p.stoppedC)
		return err
	} else {
		__antithesis_instrumentation__.Notify(113652)
	}
	__antithesis_instrumentation__.Notify(113649)
	return nil
}

func (p *Processor) run(
	ctx context.Context,
	_forStacks roachpb.RangeID,
	rtsIterFunc IntentScannerConstructor,
	stopper *stop.Stopper,
) {
	__antithesis_instrumentation__.Notify(113653)
	defer close(p.stoppedC)
	ctx, cancelOutputLoops := context.WithCancel(ctx)
	defer cancelOutputLoops()
	defer p.MemBudget.Close(ctx)

	if rtsIterFunc != nil {
		__antithesis_instrumentation__.Notify(113656)
		rtsIter := rtsIterFunc()
		initScan := newInitResolvedTSScan(p, rtsIter)
		err := stopper.RunAsyncTask(ctx, "rangefeed: init resolved ts", initScan.Run)
		if err != nil {
			__antithesis_instrumentation__.Notify(113657)
			initScan.Cancel()
		} else {
			__antithesis_instrumentation__.Notify(113658)
		}
	} else {
		__antithesis_instrumentation__.Notify(113659)
		p.initResolvedTS(ctx)
	}
	__antithesis_instrumentation__.Notify(113654)

	var txnPushTicker *time.Ticker
	var txnPushTickerC <-chan time.Time
	var txnPushAttemptC chan struct{}
	if p.PushTxnsInterval > 0 {
		__antithesis_instrumentation__.Notify(113660)
		txnPushTicker = time.NewTicker(p.PushTxnsInterval)
		txnPushTickerC = txnPushTicker.C
		defer txnPushTicker.Stop()
	} else {
		__antithesis_instrumentation__.Notify(113661)
	}
	__antithesis_instrumentation__.Notify(113655)

	for {
		__antithesis_instrumentation__.Notify(113662)
		select {

		case r := <-p.regC:
			__antithesis_instrumentation__.Notify(113663)
			if !p.Span.AsRawSpanWithNoLocals().Contains(r.span) {
				__antithesis_instrumentation__.Notify(113676)
				log.Fatalf(ctx, "registration %s not in Processor's key range %v", r, p.Span)
			} else {
				__antithesis_instrumentation__.Notify(113677)
			}
			__antithesis_instrumentation__.Notify(113664)

			r.maybeConstructCatchUpIter()

			p.reg.Register(&r)

			p.filterResC <- p.reg.NewFilter()

			r.publish(ctx, p.newCheckpointEvent(), nil)

			runOutputLoop := func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(113678)
				r.runOutputLoop(ctx, p.RangeID)
				select {
				case p.unregC <- &r:
					__antithesis_instrumentation__.Notify(113679)
				case <-p.stoppedC:
					__antithesis_instrumentation__.Notify(113680)
				}
			}
			__antithesis_instrumentation__.Notify(113665)
			if err := stopper.RunAsyncTask(ctx, "rangefeed: output loop", runOutputLoop); err != nil {
				__antithesis_instrumentation__.Notify(113681)
				r.disconnect(roachpb.NewError(err))
				p.reg.Unregister(ctx, &r)
			} else {
				__antithesis_instrumentation__.Notify(113682)
			}

		case r := <-p.unregC:
			__antithesis_instrumentation__.Notify(113666)
			p.reg.Unregister(ctx, r)

		case e := <-p.spanErrC:
			__antithesis_instrumentation__.Notify(113667)
			p.reg.DisconnectWithErr(e.span, e.pErr)

		case <-p.lenReqC:
			__antithesis_instrumentation__.Notify(113668)
			p.lenResC <- p.reg.Len()

		case <-p.filterReqC:
			__antithesis_instrumentation__.Notify(113669)
			p.filterResC <- p.reg.NewFilter()

		case e := <-p.eventC:
			__antithesis_instrumentation__.Notify(113670)
			p.consumeEvent(ctx, e)
			e.allocation.Release(ctx)
			putPooledEvent(e)

		case <-txnPushTickerC:
			__antithesis_instrumentation__.Notify(113671)

			if !p.rts.IsInit() {
				__antithesis_instrumentation__.Notify(113683)
				continue
			} else {
				__antithesis_instrumentation__.Notify(113684)
			}
			__antithesis_instrumentation__.Notify(113672)

			now := p.Clock.Now()
			before := now.Add(-p.PushTxnsAge.Nanoseconds(), 0)
			oldTxns := p.rts.intentQ.Before(before)

			if len(oldTxns) > 0 {
				__antithesis_instrumentation__.Notify(113685)
				toPush := make([]enginepb.TxnMeta, len(oldTxns))
				for i, txn := range oldTxns {
					__antithesis_instrumentation__.Notify(113687)
					toPush[i] = txn.asTxnMeta()
				}
				__antithesis_instrumentation__.Notify(113686)

				txnPushTickerC = nil
				txnPushAttemptC = make(chan struct{})

				pushTxns := newTxnPushAttempt(p, toPush, now, txnPushAttemptC)
				err := stopper.RunAsyncTask(ctx, "rangefeed: pushing old txns", pushTxns.Run)
				if err != nil {
					__antithesis_instrumentation__.Notify(113688)
					pushTxns.Cancel()
				} else {
					__antithesis_instrumentation__.Notify(113689)
				}
			} else {
				__antithesis_instrumentation__.Notify(113690)
			}

		case <-txnPushAttemptC:
			__antithesis_instrumentation__.Notify(113673)

			txnPushTickerC = txnPushTicker.C
			txnPushAttemptC = nil

		case pErr := <-p.stopC:
			__antithesis_instrumentation__.Notify(113674)
			p.reg.DisconnectWithErr(all, pErr)
			return

		case <-stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(113675)
			pErr := roachpb.NewError(&roachpb.NodeUnavailableError{})
			p.reg.DisconnectWithErr(all, pErr)
			return
		}
	}
}

func (p *Processor) Stop() {
	__antithesis_instrumentation__.Notify(113691)
	p.StopWithErr(nil)
}

func (p *Processor) StopWithErr(pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(113692)
	if p == nil {
		__antithesis_instrumentation__.Notify(113694)
		return
	} else {
		__antithesis_instrumentation__.Notify(113695)
	}
	__antithesis_instrumentation__.Notify(113693)

	p.syncEventC()

	p.sendStop(pErr)
}

func (p *Processor) DisconnectSpanWithErr(span roachpb.Span, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(113696)
	if p == nil {
		__antithesis_instrumentation__.Notify(113698)
		return
	} else {
		__antithesis_instrumentation__.Notify(113699)
	}
	__antithesis_instrumentation__.Notify(113697)
	select {
	case p.spanErrC <- spanErr{span: span, pErr: pErr}:
		__antithesis_instrumentation__.Notify(113700)
	case <-p.stoppedC:
		__antithesis_instrumentation__.Notify(113701)

	}
}

func (p *Processor) sendStop(pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(113702)
	select {
	case p.stopC <- pErr:
		__antithesis_instrumentation__.Notify(113703)

	case <-p.stoppedC:
		__antithesis_instrumentation__.Notify(113704)

	}
}

func (p *Processor) Register(
	span roachpb.RSpan,
	startTS hlc.Timestamp,
	catchUpIterConstructor CatchUpIteratorConstructor,
	withDiff bool,
	stream Stream,
	errC chan<- *roachpb.Error,
) (bool, *Filter) {
	__antithesis_instrumentation__.Notify(113705)

	p.syncEventC()

	r := newRegistration(
		span.AsRawSpanWithNoLocals(), startTS, catchUpIterConstructor, withDiff,
		p.Config.EventChanCap, p.Metrics, stream, errC,
	)
	select {
	case p.regC <- r:
		__antithesis_instrumentation__.Notify(113706)

		return true, <-p.filterResC
	case <-p.stoppedC:
		__antithesis_instrumentation__.Notify(113707)
		return false, nil
	}
}

func (p *Processor) Len() int {
	__antithesis_instrumentation__.Notify(113708)
	if p == nil {
		__antithesis_instrumentation__.Notify(113710)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(113711)
	}
	__antithesis_instrumentation__.Notify(113709)

	select {
	case p.lenReqC <- struct{}{}:
		__antithesis_instrumentation__.Notify(113712)

		return <-p.lenResC
	case <-p.stoppedC:
		__antithesis_instrumentation__.Notify(113713)
		return 0
	}
}

func (p *Processor) Filter() *Filter {
	__antithesis_instrumentation__.Notify(113714)
	if p == nil {
		__antithesis_instrumentation__.Notify(113716)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(113717)
	}
	__antithesis_instrumentation__.Notify(113715)

	select {
	case p.filterReqC <- struct{}{}:
		__antithesis_instrumentation__.Notify(113718)

		return <-p.filterResC
	case <-p.stoppedC:
		__antithesis_instrumentation__.Notify(113719)
		return nil
	}
}

func (p *Processor) ConsumeLogicalOps(ctx context.Context, ops ...enginepb.MVCCLogicalOp) bool {
	__antithesis_instrumentation__.Notify(113720)
	if p == nil {
		__antithesis_instrumentation__.Notify(113723)
		return true
	} else {
		__antithesis_instrumentation__.Notify(113724)
	}
	__antithesis_instrumentation__.Notify(113721)
	if len(ops) == 0 {
		__antithesis_instrumentation__.Notify(113725)
		return true
	} else {
		__antithesis_instrumentation__.Notify(113726)
	}
	__antithesis_instrumentation__.Notify(113722)
	return p.sendEvent(ctx, event{ops: ops}, p.EventChanTimeout)
}

func (p *Processor) ConsumeSSTable(
	ctx context.Context, sst []byte, sstSpan roachpb.Span, writeTS hlc.Timestamp,
) bool {
	__antithesis_instrumentation__.Notify(113727)
	if p == nil {
		__antithesis_instrumentation__.Notify(113729)
		return true
	} else {
		__antithesis_instrumentation__.Notify(113730)
	}
	__antithesis_instrumentation__.Notify(113728)
	return p.sendEvent(ctx, event{sst: sst, sstSpan: sstSpan, sstWTS: writeTS}, p.EventChanTimeout)
}

func (p *Processor) ForwardClosedTS(ctx context.Context, closedTS hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(113731)
	if p == nil {
		__antithesis_instrumentation__.Notify(113734)
		return true
	} else {
		__antithesis_instrumentation__.Notify(113735)
	}
	__antithesis_instrumentation__.Notify(113732)
	if closedTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(113736)
		return true
	} else {
		__antithesis_instrumentation__.Notify(113737)
	}
	__antithesis_instrumentation__.Notify(113733)
	return p.sendEvent(ctx, event{ct: closedTS}, p.EventChanTimeout)
}

func (p *Processor) sendEvent(ctx context.Context, e event, timeout time.Duration) bool {
	__antithesis_instrumentation__.Notify(113738)

	var allocation *SharedBudgetAllocation
	if p.MemBudget != nil {
		__antithesis_instrumentation__.Notify(113741)
		size := calculateDateEventSize(e)
		if size > 0 {
			__antithesis_instrumentation__.Notify(113742)
			var err error

			allocation, err = p.MemBudget.TryGet(ctx, size)
			if err != nil {
				__antithesis_instrumentation__.Notify(113745)

				if timeout > 0 {
					__antithesis_instrumentation__.Notify(113747)
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, timeout)
					defer cancel()

					timeout = 0
				} else {
					__antithesis_instrumentation__.Notify(113748)
				}
				__antithesis_instrumentation__.Notify(113746)
				p.Metrics.RangeFeedBudgetBlocked.Inc(1)
				allocation, err = p.MemBudget.WaitAndGet(ctx, size)
			} else {
				__antithesis_instrumentation__.Notify(113749)
			}
			__antithesis_instrumentation__.Notify(113743)
			if err != nil {
				__antithesis_instrumentation__.Notify(113750)
				p.Metrics.RangeFeedBudgetExhausted.Inc(1)
				p.sendStop(newErrBufferCapacityExceeded())
				return false
			} else {
				__antithesis_instrumentation__.Notify(113751)
			}
			__antithesis_instrumentation__.Notify(113744)
			defer func() {
				__antithesis_instrumentation__.Notify(113752)
				allocation.Release(ctx)
			}()
		} else {
			__antithesis_instrumentation__.Notify(113753)
		}
	} else {
		__antithesis_instrumentation__.Notify(113754)
	}
	__antithesis_instrumentation__.Notify(113739)
	ev := getPooledEvent(e)
	ev.allocation = allocation
	if timeout == 0 {
		__antithesis_instrumentation__.Notify(113755)

		select {
		case p.eventC <- ev:
			__antithesis_instrumentation__.Notify(113756)

			allocation = nil
		case <-p.stoppedC:
			__antithesis_instrumentation__.Notify(113757)

		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(113758)
			p.sendStop(newErrBufferCapacityExceeded())
			return false
		}
	} else {
		__antithesis_instrumentation__.Notify(113759)

		select {
		case p.eventC <- ev:
			__antithesis_instrumentation__.Notify(113760)

			allocation = nil
		case <-p.stoppedC:
			__antithesis_instrumentation__.Notify(113761)

		default:
			__antithesis_instrumentation__.Notify(113762)

			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
			select {
			case p.eventC <- ev:
				__antithesis_instrumentation__.Notify(113763)

				allocation = nil
			case <-p.stoppedC:
				__antithesis_instrumentation__.Notify(113764)

			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(113765)

				p.sendStop(newErrBufferCapacityExceeded())
				return false
			}
		}
	}
	__antithesis_instrumentation__.Notify(113740)
	return true
}

func (p *Processor) setResolvedTSInitialized(ctx context.Context) {
	__antithesis_instrumentation__.Notify(113766)
	p.sendEvent(ctx, event{initRTS: true}, 0)
}

func (p *Processor) syncEventC() {
	__antithesis_instrumentation__.Notify(113767)
	syncC := make(chan struct{})
	ev := getPooledEvent(event{syncC: syncC})
	select {
	case p.eventC <- ev:
		__antithesis_instrumentation__.Notify(113768)
		select {
		case <-syncC:
			__antithesis_instrumentation__.Notify(113770)

		case <-p.stoppedC:
			__antithesis_instrumentation__.Notify(113771)

		}
	case <-p.stoppedC:
		__antithesis_instrumentation__.Notify(113769)

	}
}

func (p *Processor) consumeEvent(ctx context.Context, e *event) {
	__antithesis_instrumentation__.Notify(113772)
	switch {
	case len(e.ops) > 0:
		__antithesis_instrumentation__.Notify(113773)
		p.consumeLogicalOps(ctx, e.ops, e.allocation)
	case len(e.sst) > 0:
		__antithesis_instrumentation__.Notify(113774)
		p.consumeSSTable(ctx, e.sst, e.sstSpan, e.sstWTS, e.allocation)
	case !e.ct.IsEmpty():
		__antithesis_instrumentation__.Notify(113775)
		p.forwardClosedTS(ctx, e.ct)
	case e.initRTS:
		__antithesis_instrumentation__.Notify(113776)
		p.initResolvedTS(ctx)
	case e.syncC != nil:
		__antithesis_instrumentation__.Notify(113777)
		if e.testRegCatchupSpan.Valid() {
			__antithesis_instrumentation__.Notify(113780)
			if err := p.reg.waitForCaughtUp(e.testRegCatchupSpan); err != nil {
				__antithesis_instrumentation__.Notify(113781)
				log.Errorf(
					ctx,
					"error waiting for registries to catch up during test, results might be impacted: %s",
					err,
				)
			} else {
				__antithesis_instrumentation__.Notify(113782)
			}
		} else {
			__antithesis_instrumentation__.Notify(113783)
		}
		__antithesis_instrumentation__.Notify(113778)
		close(e.syncC)
	default:
		__antithesis_instrumentation__.Notify(113779)
		panic(fmt.Sprintf("missing event variant: %+v", e))
	}
}

func (p *Processor) consumeLogicalOps(
	ctx context.Context, ops []enginepb.MVCCLogicalOp, allocation *SharedBudgetAllocation,
) {
	__antithesis_instrumentation__.Notify(113784)
	for _, op := range ops {
		__antithesis_instrumentation__.Notify(113785)

		switch t := op.GetValue().(type) {
		case *enginepb.MVCCWriteValueOp:
			__antithesis_instrumentation__.Notify(113787)

			p.publishValue(ctx, t.Key, t.Timestamp, t.Value, t.PrevValue, allocation)

		case *enginepb.MVCCWriteIntentOp:
			__antithesis_instrumentation__.Notify(113788)

		case *enginepb.MVCCUpdateIntentOp:
			__antithesis_instrumentation__.Notify(113789)

		case *enginepb.MVCCCommitIntentOp:
			__antithesis_instrumentation__.Notify(113790)

			p.publishValue(ctx, t.Key, t.Timestamp, t.Value, t.PrevValue, allocation)

		case *enginepb.MVCCAbortIntentOp:
			__antithesis_instrumentation__.Notify(113791)

		case *enginepb.MVCCAbortTxnOp:
			__antithesis_instrumentation__.Notify(113792)

		default:
			__antithesis_instrumentation__.Notify(113793)
			panic(errors.AssertionFailedf("unknown logical op %T", t))
		}
		__antithesis_instrumentation__.Notify(113786)

		if p.rts.ConsumeLogicalOp(op) {
			__antithesis_instrumentation__.Notify(113794)
			p.publishCheckpoint(ctx)
		} else {
			__antithesis_instrumentation__.Notify(113795)
		}
	}
}

func (p *Processor) consumeSSTable(
	ctx context.Context,
	sst []byte,
	sstSpan roachpb.Span,
	sstWTS hlc.Timestamp,
	allocation *SharedBudgetAllocation,
) {
	__antithesis_instrumentation__.Notify(113796)
	p.publishSSTable(ctx, sst, sstSpan, sstWTS, allocation)
}

func (p *Processor) forwardClosedTS(ctx context.Context, newClosedTS hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(113797)
	if p.rts.ForwardClosedTS(newClosedTS) {
		__antithesis_instrumentation__.Notify(113798)
		p.publishCheckpoint(ctx)
	} else {
		__antithesis_instrumentation__.Notify(113799)
	}
}

func (p *Processor) initResolvedTS(ctx context.Context) {
	__antithesis_instrumentation__.Notify(113800)
	if p.rts.Init() {
		__antithesis_instrumentation__.Notify(113801)
		p.publishCheckpoint(ctx)
	} else {
		__antithesis_instrumentation__.Notify(113802)
	}
}

func (p *Processor) publishValue(
	ctx context.Context,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value, prevValue []byte,
	allocation *SharedBudgetAllocation,
) {
	__antithesis_instrumentation__.Notify(113803)
	if !p.Span.ContainsKey(roachpb.RKey(key)) {
		__antithesis_instrumentation__.Notify(113806)
		log.Fatalf(ctx, "key %v not in Processor's key range %v", key, p.Span)
	} else {
		__antithesis_instrumentation__.Notify(113807)
	}
	__antithesis_instrumentation__.Notify(113804)

	var prevVal roachpb.Value
	if prevValue != nil {
		__antithesis_instrumentation__.Notify(113808)
		prevVal.RawBytes = prevValue
	} else {
		__antithesis_instrumentation__.Notify(113809)
	}
	__antithesis_instrumentation__.Notify(113805)
	var event roachpb.RangeFeedEvent
	event.MustSetValue(&roachpb.RangeFeedValue{
		Key: key,
		Value: roachpb.Value{
			RawBytes:  value,
			Timestamp: timestamp,
		},
		PrevValue: prevVal,
	})
	p.reg.PublishToOverlapping(ctx, roachpb.Span{Key: key}, &event, allocation)
}

func (p *Processor) publishSSTable(
	ctx context.Context,
	sst []byte,
	sstSpan roachpb.Span,
	sstWTS hlc.Timestamp,
	allocation *SharedBudgetAllocation,
) {
	__antithesis_instrumentation__.Notify(113810)
	if sstSpan.Equal(roachpb.Span{}) {
		__antithesis_instrumentation__.Notify(113813)
		panic(errors.AssertionFailedf("received SSTable without span"))
	} else {
		__antithesis_instrumentation__.Notify(113814)
	}
	__antithesis_instrumentation__.Notify(113811)
	if sstWTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(113815)
		panic(errors.AssertionFailedf("received SSTable without write timestamp"))
	} else {
		__antithesis_instrumentation__.Notify(113816)
	}
	__antithesis_instrumentation__.Notify(113812)
	p.reg.PublishToOverlapping(ctx, sstSpan, &roachpb.RangeFeedEvent{
		SST: &roachpb.RangeFeedSSTable{
			Data:    sst,
			Span:    sstSpan,
			WriteTS: sstWTS,
		},
	}, allocation)
}

func (p *Processor) publishCheckpoint(ctx context.Context) {
	__antithesis_instrumentation__.Notify(113817)

	event := p.newCheckpointEvent()
	p.reg.PublishToOverlapping(ctx, all, event, nil)
}

func (p *Processor) newCheckpointEvent() *roachpb.RangeFeedEvent {
	__antithesis_instrumentation__.Notify(113818)

	var event roachpb.RangeFeedEvent
	event.MustSetValue(&roachpb.RangeFeedCheckpoint{
		Span:       p.Span.AsRawSpanWithNoLocals(),
		ResolvedTS: p.rts.Get(),
	})
	return &event
}

func calculateDateEventSize(e event) int64 {
	__antithesis_instrumentation__.Notify(113819)
	var size int64
	for _, op := range e.ops {
		__antithesis_instrumentation__.Notify(113821)
		size += int64(op.Size())
	}
	__antithesis_instrumentation__.Notify(113820)
	size += int64(len(e.sst))
	return size
}
