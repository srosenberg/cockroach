package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/cdcutils"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeeddist"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvfeed"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/schemafeed"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/span"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
)

type changeAggregator struct {
	execinfra.ProcessorBase

	flowCtx *execinfra.FlowCtx
	spec    execinfrapb.ChangeAggregatorSpec
	memAcc  mon.BoundAccount

	cancel func()

	errCh chan error

	kvFeedDoneCh chan struct{}
	kvFeedMemMon *mon.BytesMonitor

	encoder Encoder

	sink Sink

	changedRowBuf *encDatumRowBuffer

	resolvedSpanBuf encDatumRowBuffer

	recentKVCount uint64

	eventProducer kvevent.Reader

	eventConsumer kvEventConsumer

	lastFlush      time.Time
	flushFrequency time.Duration

	frontier *schemaChangeFrontier

	metrics    *Metrics
	sliMetrics *sliMetrics
	knobs      TestingKnobs
	topicNamer *TopicNamer
}

type timestampLowerBoundOracle interface {
	inclusiveLowerBoundTS() hlc.Timestamp
}

type changeAggregatorLowerBoundOracle struct {
	sf                         *span.Frontier
	initialInclusiveLowerBound hlc.Timestamp
}

func (o *changeAggregatorLowerBoundOracle) inclusiveLowerBoundTS() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(15445)
	if frontier := o.sf.Frontier(); !frontier.IsEmpty() {
		__antithesis_instrumentation__.Notify(15447)

		return frontier.Next()
	} else {
		__antithesis_instrumentation__.Notify(15448)
	}
	__antithesis_instrumentation__.Notify(15446)

	return o.initialInclusiveLowerBound
}

var _ execinfra.Processor = &changeAggregator{}
var _ execinfra.RowSource = &changeAggregator{}

func newChangeAggregatorProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.ChangeAggregatorSpec,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(15449)
	ctx := flowCtx.EvalCtx.Ctx()
	memMonitor := execinfra.NewMonitor(ctx, flowCtx.EvalCtx.Mon, "changeagg-mem")
	ca := &changeAggregator{
		flowCtx: flowCtx,
		spec:    spec,
		memAcc:  memMonitor.MakeBoundAccount(),
	}
	if err := ca.Init(
		ca,
		post,
		changefeeddist.ChangefeedResultTypes,
		flowCtx,
		processorID,
		output,
		memMonitor,
		execinfra.ProcStateOpts{
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(15454)
				ca.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(15455)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15456)
	}
	__antithesis_instrumentation__.Notify(15450)

	var err error
	if ca.encoder, err = getEncoder(ca.spec.Feed.Opts, AllTargets(ca.spec.Feed)); err != nil {
		__antithesis_instrumentation__.Notify(15457)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15458)
	}
	__antithesis_instrumentation__.Notify(15451)

	if _, needTopics := ca.spec.Feed.Opts[changefeedbase.OptTopicInValue]; needTopics {
		__antithesis_instrumentation__.Notify(15459)
		ca.topicNamer, err = MakeTopicNamer(ca.spec.Feed.TargetSpecifications)
		if err != nil {
			__antithesis_instrumentation__.Notify(15460)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(15461)
		}
	} else {
		__antithesis_instrumentation__.Notify(15462)
	}
	__antithesis_instrumentation__.Notify(15452)

	if r, ok := ca.spec.Feed.Opts[changefeedbase.OptMinCheckpointFrequency]; ok && func() bool {
		__antithesis_instrumentation__.Notify(15463)
		return r != `` == true
	}() == true {
		__antithesis_instrumentation__.Notify(15464)
		ca.flushFrequency, err = time.ParseDuration(r)
		if err != nil {
			__antithesis_instrumentation__.Notify(15465)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(15466)
		}
	} else {
		__antithesis_instrumentation__.Notify(15467)
		ca.flushFrequency = changefeedbase.DefaultMinCheckpointFrequency
	}
	__antithesis_instrumentation__.Notify(15453)

	return ca, nil
}

func (ca *changeAggregator) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(15468)
	return true
}

func (ca *changeAggregator) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(15469)
	if ca.spec.JobID != 0 {
		__antithesis_instrumentation__.Notify(15479)
		ctx = logtags.AddTag(ctx, "job", ca.spec.JobID)
	} else {
		__antithesis_instrumentation__.Notify(15480)
	}
	__antithesis_instrumentation__.Notify(15470)
	ctx = ca.StartInternal(ctx, changeAggregatorProcName)

	ctx, ca.cancel = context.WithCancel(ctx)
	ca.Ctx = ctx

	initialHighWater, needsInitialScan := getKVFeedInitialParameters(ca.spec)

	frontierHighWater := initialHighWater
	if needsInitialScan {
		__antithesis_instrumentation__.Notify(15481)

		frontierHighWater = hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(15482)
	}
	__antithesis_instrumentation__.Notify(15471)
	spans, err := ca.setupSpansAndFrontier(frontierHighWater)

	endTime := ca.spec.Feed.EndTime

	if err != nil {
		__antithesis_instrumentation__.Notify(15483)
		ca.MoveToDraining(err)
		ca.cancel()
		return
	} else {
		__antithesis_instrumentation__.Notify(15484)
	}
	__antithesis_instrumentation__.Notify(15472)
	timestampOracle := &changeAggregatorLowerBoundOracle{
		sf:                         ca.frontier.SpanFrontier(),
		initialInclusiveLowerBound: ca.spec.Feed.StatementTime,
	}

	if cfKnobs, ok := ca.flowCtx.TestingKnobs().Changefeed.(*TestingKnobs); ok {
		__antithesis_instrumentation__.Notify(15485)
		ca.knobs = *cfKnobs
	} else {
		__antithesis_instrumentation__.Notify(15486)
	}
	__antithesis_instrumentation__.Notify(15473)

	pool := ca.flowCtx.Cfg.BackfillerMonitor
	if ca.knobs.MemMonitor != nil {
		__antithesis_instrumentation__.Notify(15487)
		pool = ca.knobs.MemMonitor
	} else {
		__antithesis_instrumentation__.Notify(15488)
	}
	__antithesis_instrumentation__.Notify(15474)
	limit := changefeedbase.PerChangefeedMemLimit.Get(&ca.flowCtx.Cfg.Settings.SV)
	kvFeedMemMon := mon.NewMonitorInheritWithLimit("kvFeed", limit, pool)
	kvFeedMemMon.Start(ctx, pool, mon.BoundAccount{})
	ca.kvFeedMemMon = kvFeedMemMon

	ca.metrics = ca.flowCtx.Cfg.JobRegistry.MetricsStruct().Changefeed.(*Metrics)
	ca.sliMetrics, err = ca.metrics.getSLIMetrics(ca.spec.Feed.Opts[changefeedbase.OptMetricsScope])
	if err != nil {
		__antithesis_instrumentation__.Notify(15489)
		ca.MoveToDraining(err)
		ca.cancel()
		return
	} else {
		__antithesis_instrumentation__.Notify(15490)
	}
	__antithesis_instrumentation__.Notify(15475)

	ca.sink, err = getSink(ctx, ca.flowCtx.Cfg, ca.spec.Feed, timestampOracle,
		ca.spec.User(), ca.spec.JobID, ca.sliMetrics)

	if err != nil {
		__antithesis_instrumentation__.Notify(15491)
		err = changefeedbase.MarkRetryableError(err)

		ca.MoveToDraining(err)
		ca.cancel()
		return
	} else {
		__antithesis_instrumentation__.Notify(15492)
	}
	__antithesis_instrumentation__.Notify(15476)

	if b, ok := ca.sink.(*bufferSink); ok {
		__antithesis_instrumentation__.Notify(15493)
		ca.changedRowBuf = &b.buf
	} else {
		__antithesis_instrumentation__.Notify(15494)
	}
	__antithesis_instrumentation__.Notify(15477)

	ca.sink = &errorWrapperSink{wrapped: ca.sink}

	ca.eventProducer, err = ca.startKVFeed(ctx, spans, initialHighWater, needsInitialScan, endTime, ca.sliMetrics)
	if err != nil {
		__antithesis_instrumentation__.Notify(15495)

		ca.MoveToDraining(err)
		ca.cancel()
		return
	} else {
		__antithesis_instrumentation__.Notify(15496)
	}
	__antithesis_instrumentation__.Notify(15478)

	if ca.spec.Feed.Opts[changefeedbase.OptFormat] == string(changefeedbase.OptFormatNative) {
		__antithesis_instrumentation__.Notify(15497)
		ca.eventConsumer = newNativeKVConsumer(ca.sink)
	} else {
		__antithesis_instrumentation__.Notify(15498)
		ca.eventConsumer = newKVEventToRowConsumer(
			ctx, ca.flowCtx.Cfg, ca.frontier.SpanFrontier(), initialHighWater,
			ca.sink, ca.encoder, ca.spec.Feed, ca.knobs, ca.topicNamer)
	}
}

func (ca *changeAggregator) startKVFeed(
	ctx context.Context,
	spans []roachpb.Span,
	initialHighWater hlc.Timestamp,
	needsInitialScan bool,
	endTime hlc.Timestamp,
	sm *sliMetrics,
) (kvevent.Reader, error) {
	__antithesis_instrumentation__.Notify(15499)
	cfg := ca.flowCtx.Cfg
	buf := kvevent.NewThrottlingBuffer(
		kvevent.NewMemBuffer(ca.kvFeedMemMon.MakeBoundAccount(), &cfg.Settings.SV, &ca.metrics.KVFeedMetrics),
		cdcutils.NodeLevelThrottler(&cfg.Settings.SV, &ca.metrics.ThrottleMetrics))

	kvfeedCfg := ca.makeKVFeedCfg(ctx, spans, buf, initialHighWater, needsInitialScan, endTime, sm)

	ca.errCh = make(chan error, 2)
	ca.kvFeedDoneCh = make(chan struct{})
	if err := ca.flowCtx.Stopper().RunAsyncTask(ctx, "changefeed-poller", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(15501)
		defer close(ca.kvFeedDoneCh)

		ca.errCh <- kvfeed.Run(ctx, kvfeedCfg)
		ca.cancel()
	}); err != nil {
		__antithesis_instrumentation__.Notify(15502)

		close(ca.kvFeedDoneCh)
		ca.errCh <- err
		ca.cancel()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15503)
	}
	__antithesis_instrumentation__.Notify(15500)

	return buf, nil
}

func (ca *changeAggregator) makeKVFeedCfg(
	ctx context.Context,
	spans []roachpb.Span,
	buf kvevent.Writer,
	initialHighWater hlc.Timestamp,
	needsInitialScan bool,
	endTime hlc.Timestamp,
	sm *sliMetrics,
) kvfeed.Config {
	__antithesis_instrumentation__.Notify(15504)
	schemaChangeEvents := changefeedbase.SchemaChangeEventClass(
		ca.spec.Feed.Opts[changefeedbase.OptSchemaChangeEvents])
	schemaChangePolicy := changefeedbase.SchemaChangePolicy(
		ca.spec.Feed.Opts[changefeedbase.OptSchemaChangePolicy])
	_, withDiff := ca.spec.Feed.Opts[changefeedbase.OptDiff]
	cfg := ca.flowCtx.Cfg

	var sf schemafeed.SchemaFeed

	initialScanOnly := endTime.EqOrdering(initialHighWater)

	if schemaChangePolicy == changefeedbase.OptSchemaChangePolicyIgnore || func() bool {
		__antithesis_instrumentation__.Notify(15506)
		return initialScanOnly == true
	}() == true {
		__antithesis_instrumentation__.Notify(15507)
		sf = schemafeed.DoNothingSchemaFeed
	} else {
		__antithesis_instrumentation__.Notify(15508)
		sf = schemafeed.New(ctx, cfg, schemaChangeEvents, AllTargets(ca.spec.Feed),
			initialHighWater, &ca.metrics.SchemaFeedMetrics, ca.spec.Feed.Opts)
	}
	__antithesis_instrumentation__.Notify(15505)

	return kvfeed.Config{
		Writer:                  buf,
		Settings:                cfg.Settings,
		DB:                      cfg.DB,
		Codec:                   cfg.Codec,
		Clock:                   cfg.DB.Clock(),
		Gossip:                  cfg.Gossip,
		Spans:                   spans,
		BackfillCheckpoint:      ca.spec.Checkpoint.Spans,
		Targets:                 AllTargets(ca.spec.Feed),
		Metrics:                 &ca.metrics.KVFeedMetrics,
		OnBackfillCallback:      ca.sliMetrics.getBackfillCallback(),
		OnBackfillRangeCallback: ca.sliMetrics.getBackfillRangeCallback(),
		MM:                      ca.kvFeedMemMon,
		InitialHighWater:        initialHighWater,
		EndTime:                 endTime,
		WithDiff:                withDiff,
		NeedsInitialScan:        needsInitialScan,
		SchemaChangeEvents:      schemaChangeEvents,
		SchemaChangePolicy:      schemaChangePolicy,
		SchemaFeed:              sf,
		Knobs:                   ca.knobs.FeedKnobs,
	}
}

func getKVFeedInitialParameters(
	spec execinfrapb.ChangeAggregatorSpec,
) (initialHighWater hlc.Timestamp, needsInitialScan bool) {
	__antithesis_instrumentation__.Notify(15509)
	for _, watch := range spec.Watches {
		__antithesis_instrumentation__.Notify(15512)
		if initialHighWater.IsEmpty() || func() bool {
			__antithesis_instrumentation__.Notify(15513)
			return watch.InitialResolved.Less(initialHighWater) == true
		}() == true {
			__antithesis_instrumentation__.Notify(15514)
			initialHighWater = watch.InitialResolved
		} else {
			__antithesis_instrumentation__.Notify(15515)
		}
	}
	__antithesis_instrumentation__.Notify(15510)

	if needsInitialScan = initialHighWater.IsEmpty(); needsInitialScan {
		__antithesis_instrumentation__.Notify(15516)
		initialHighWater = spec.Feed.StatementTime
	} else {
		__antithesis_instrumentation__.Notify(15517)
	}
	__antithesis_instrumentation__.Notify(15511)
	return initialHighWater, needsInitialScan
}

func (ca *changeAggregator) setupSpansAndFrontier(
	initialHighWater hlc.Timestamp,
) (spans []roachpb.Span, err error) {
	__antithesis_instrumentation__.Notify(15518)
	spans = make([]roachpb.Span, 0, len(ca.spec.Watches))
	for _, watch := range ca.spec.Watches {
		__antithesis_instrumentation__.Notify(15523)
		spans = append(spans, watch.Span)
	}
	__antithesis_instrumentation__.Notify(15519)

	ca.frontier, err = makeSchemaChangeFrontier(initialHighWater, spans...)
	if err != nil {
		__antithesis_instrumentation__.Notify(15524)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15525)
	}
	__antithesis_instrumentation__.Notify(15520)

	var checkpointedSpanTs hlc.Timestamp
	if initialHighWater.IsEmpty() {
		__antithesis_instrumentation__.Notify(15526)
		checkpointedSpanTs = ca.spec.Feed.StatementTime
	} else {
		__antithesis_instrumentation__.Notify(15527)
		checkpointedSpanTs = initialHighWater.Next()
	}
	__antithesis_instrumentation__.Notify(15521)
	for _, checkpointedSpan := range ca.spec.Checkpoint.Spans {
		__antithesis_instrumentation__.Notify(15528)
		if _, err := ca.frontier.Forward(checkpointedSpan, checkpointedSpanTs); err != nil {
			__antithesis_instrumentation__.Notify(15529)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(15530)
		}
	}
	__antithesis_instrumentation__.Notify(15522)
	return spans, nil
}

func (ca *changeAggregator) close() {
	__antithesis_instrumentation__.Notify(15531)
	if ca.Closed {
		__antithesis_instrumentation__.Notify(15536)
		return
	} else {
		__antithesis_instrumentation__.Notify(15537)
	}
	__antithesis_instrumentation__.Notify(15532)
	ca.cancel()

	if ca.kvFeedDoneCh != nil {
		__antithesis_instrumentation__.Notify(15538)
		<-ca.kvFeedDoneCh
	} else {
		__antithesis_instrumentation__.Notify(15539)
	}
	__antithesis_instrumentation__.Notify(15533)
	if ca.sink != nil {
		__antithesis_instrumentation__.Notify(15540)
		if err := ca.sink.Close(); err != nil {
			__antithesis_instrumentation__.Notify(15541)
			log.Warningf(ca.Ctx, `error closing sink. goroutines may have leaked: %v`, err)
		} else {
			__antithesis_instrumentation__.Notify(15542)
		}
	} else {
		__antithesis_instrumentation__.Notify(15543)
	}
	__antithesis_instrumentation__.Notify(15534)

	ca.memAcc.Close(ca.Ctx)
	if ca.kvFeedMemMon != nil {
		__antithesis_instrumentation__.Notify(15544)
		ca.kvFeedMemMon.Stop(ca.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(15545)
	}
	__antithesis_instrumentation__.Notify(15535)
	ca.MemMonitor.Stop(ca.Ctx)
	ca.InternalClose()
}

func (ca *changeAggregator) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(15546)
	for ca.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(15548)
		if !ca.changedRowBuf.IsEmpty() {
			__antithesis_instrumentation__.Notify(15550)
			return ca.ProcessRowHelper(ca.changedRowBuf.Pop()), nil
		} else {
			__antithesis_instrumentation__.Notify(15551)
			if !ca.resolvedSpanBuf.IsEmpty() {
				__antithesis_instrumentation__.Notify(15552)
				return ca.ProcessRowHelper(ca.resolvedSpanBuf.Pop()), nil
			} else {
				__antithesis_instrumentation__.Notify(15553)
			}
		}
		__antithesis_instrumentation__.Notify(15549)

		if err := ca.tick(); err != nil {
			__antithesis_instrumentation__.Notify(15554)
			var e kvevent.ErrBufferClosed
			if errors.As(err, &e) {
				__antithesis_instrumentation__.Notify(15556)

				err = e.Unwrap()
				if errors.Is(err, kvevent.ErrNormalRestartReason) {
					__antithesis_instrumentation__.Notify(15557)
					err = nil
				} else {
					__antithesis_instrumentation__.Notify(15558)
				}
			} else {
				__antithesis_instrumentation__.Notify(15559)

				select {

				case err = <-ca.errCh:
					__antithesis_instrumentation__.Notify(15560)
				default:
					__antithesis_instrumentation__.Notify(15561)
				}
			}
			__antithesis_instrumentation__.Notify(15555)

			ca.cancel()

			ca.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(15562)
		}
	}
	__antithesis_instrumentation__.Notify(15547)
	return nil, ca.DrainHelper()
}

func (ca *changeAggregator) tick() error {
	__antithesis_instrumentation__.Notify(15563)
	event, err := ca.eventProducer.Get(ca.Ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(15566)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15567)
	}
	__antithesis_instrumentation__.Notify(15564)

	queuedNanos := timeutil.Since(event.BufferAddTimestamp()).Nanoseconds()
	ca.metrics.QueueTimeNanos.Inc(queuedNanos)

	switch event.Type() {
	case kvevent.TypeKV:
		__antithesis_instrumentation__.Notify(15568)

		if event.BackfillTimestamp().IsEmpty() {
			__antithesis_instrumentation__.Notify(15573)
			ca.sliMetrics.AdmitLatency.RecordValue(timeutil.Since(event.Timestamp().GoTime()).Nanoseconds())
		} else {
			__antithesis_instrumentation__.Notify(15574)
		}
		__antithesis_instrumentation__.Notify(15569)
		ca.recentKVCount++
		return ca.eventConsumer.ConsumeEvent(ca.Ctx, event)
	case kvevent.TypeResolved:
		__antithesis_instrumentation__.Notify(15570)
		a := event.DetachAlloc()
		a.Release(ca.Ctx)
		resolved := event.Resolved()
		if ca.knobs.ShouldSkipResolved == nil || func() bool {
			__antithesis_instrumentation__.Notify(15575)
			return !ca.knobs.ShouldSkipResolved(resolved) == true
		}() == true {
			__antithesis_instrumentation__.Notify(15576)
			return ca.noteResolvedSpan(resolved)
		} else {
			__antithesis_instrumentation__.Notify(15577)
		}
	case kvevent.TypeFlush:
		__antithesis_instrumentation__.Notify(15571)
		return ca.sink.Flush(ca.Ctx)
	default:
		__antithesis_instrumentation__.Notify(15572)
	}
	__antithesis_instrumentation__.Notify(15565)

	return nil
}

func (ca *changeAggregator) noteResolvedSpan(resolved *jobspb.ResolvedSpan) error {
	__antithesis_instrumentation__.Notify(15578)
	advanced, err := ca.frontier.ForwardResolvedSpan(*resolved)
	if err != nil {
		__antithesis_instrumentation__.Notify(15581)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15582)
	}
	__antithesis_instrumentation__.Notify(15579)

	forceFlush := resolved.BoundaryType != jobspb.ResolvedSpan_NONE

	checkpointFrontier := advanced && func() bool {
		__antithesis_instrumentation__.Notify(15583)
		return (forceFlush || func() bool {
			__antithesis_instrumentation__.Notify(15584)
			return timeutil.Since(ca.lastFlush) > ca.flushFrequency == true
		}() == true) == true
	}() == true

	checkpointBackfill := ca.spec.JobID != 0 && func() bool {
		__antithesis_instrumentation__.Notify(15585)
		return resolved.Timestamp.Equal(ca.frontier.BackfillTS()) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(15586)
		return canCheckpointBackfill(&ca.flowCtx.Cfg.Settings.SV, ca.lastFlush) == true
	}() == true

	if checkpointFrontier || func() bool {
		__antithesis_instrumentation__.Notify(15587)
		return checkpointBackfill == true
	}() == true {
		__antithesis_instrumentation__.Notify(15588)
		defer func() {
			__antithesis_instrumentation__.Notify(15590)
			ca.lastFlush = timeutil.Now()
		}()
		__antithesis_instrumentation__.Notify(15589)
		return ca.flushFrontier()
	} else {
		__antithesis_instrumentation__.Notify(15591)
	}
	__antithesis_instrumentation__.Notify(15580)

	return nil
}

func (ca *changeAggregator) flushFrontier() error {
	__antithesis_instrumentation__.Notify(15592)

	if err := ca.sink.Flush(ca.Ctx); err != nil {
		__antithesis_instrumentation__.Notify(15595)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15596)
	}
	__antithesis_instrumentation__.Notify(15593)

	var batch jobspb.ResolvedSpans
	ca.frontier.Entries(func(s roachpb.Span, ts hlc.Timestamp) span.OpResult {
		__antithesis_instrumentation__.Notify(15597)
		boundaryType := jobspb.ResolvedSpan_NONE
		if ca.frontier.boundaryTime.Equal(ts) {
			__antithesis_instrumentation__.Notify(15599)
			boundaryType = ca.frontier.boundaryType
		} else {
			__antithesis_instrumentation__.Notify(15600)
		}
		__antithesis_instrumentation__.Notify(15598)

		batch.ResolvedSpans = append(batch.ResolvedSpans, jobspb.ResolvedSpan{
			Span:         s,
			Timestamp:    ts,
			BoundaryType: boundaryType,
		})
		return span.ContinueMatch
	})
	__antithesis_instrumentation__.Notify(15594)

	return ca.emitResolved(batch)
}

func (ca *changeAggregator) emitResolved(batch jobspb.ResolvedSpans) error {
	__antithesis_instrumentation__.Notify(15601)

	if !ca.flowCtx.Cfg.Settings.Version.IsActive(ca.Ctx, clusterversion.ChangefeedIdleness) {
		__antithesis_instrumentation__.Notify(15604)
		for _, resolved := range batch.ResolvedSpans {
			__antithesis_instrumentation__.Notify(15606)
			resolvedBytes, err := protoutil.Marshal(&resolved)
			if err != nil {
				__antithesis_instrumentation__.Notify(15608)
				return err
			} else {
				__antithesis_instrumentation__.Notify(15609)
			}
			__antithesis_instrumentation__.Notify(15607)

			ca.resolvedSpanBuf.Push(rowenc.EncDatumRow{
				rowenc.EncDatum{Datum: tree.NewDBytes(tree.DBytes(resolvedBytes))},
				rowenc.EncDatum{Datum: tree.DNull},
				rowenc.EncDatum{Datum: tree.DNull},
				rowenc.EncDatum{Datum: tree.DNull},
			})
			ca.metrics.ResolvedMessages.Inc(1)
		}
		__antithesis_instrumentation__.Notify(15605)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(15610)
	}
	__antithesis_instrumentation__.Notify(15602)

	progressUpdate := jobspb.ResolvedSpans{
		ResolvedSpans: batch.ResolvedSpans,
		Stats: jobspb.ResolvedSpans_Stats{
			RecentKvCount: ca.recentKVCount,
		},
	}
	updateBytes, err := protoutil.Marshal(&progressUpdate)
	if err != nil {
		__antithesis_instrumentation__.Notify(15611)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15612)
	}
	__antithesis_instrumentation__.Notify(15603)
	ca.resolvedSpanBuf.Push(rowenc.EncDatumRow{
		rowenc.EncDatum{Datum: tree.NewDBytes(tree.DBytes(updateBytes))},
		rowenc.EncDatum{Datum: tree.DNull},
		rowenc.EncDatum{Datum: tree.DNull},
		rowenc.EncDatum{Datum: tree.DNull},
	})

	ca.recentKVCount = 0
	return nil
}

func (ca *changeAggregator) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(15613)

	ca.close()
}

type kvEventConsumer interface {
	ConsumeEvent(ctx context.Context, event kvevent.Event) error
}

type kvEventToRowConsumer struct {
	frontier             *span.Frontier
	encoder              Encoder
	scratch              bufalloc.ByteAllocator
	sink                 Sink
	cursor               hlc.Timestamp
	knobs                TestingKnobs
	rfCache              *rowFetcherCache
	details              jobspb.ChangefeedDetails
	kvFetcher            row.SpanKVFetcher
	topicDescriptorCache map[TopicIdentifier]TopicDescriptor
	topicNamer           *TopicNamer
}

var _ kvEventConsumer = &kvEventToRowConsumer{}

func newKVEventToRowConsumer(
	ctx context.Context,
	cfg *execinfra.ServerConfig,
	frontier *span.Frontier,
	cursor hlc.Timestamp,
	sink Sink,
	encoder Encoder,
	details jobspb.ChangefeedDetails,
	knobs TestingKnobs,
	topicNamer *TopicNamer,
) kvEventConsumer {
	__antithesis_instrumentation__.Notify(15614)
	rfCache := newRowFetcherCache(
		ctx,
		cfg.Codec,
		cfg.LeaseManager.(*lease.Manager),
		cfg.CollectionFactory,
		cfg.DB,
		details,
	)

	return &kvEventToRowConsumer{
		frontier:             frontier,
		encoder:              encoder,
		sink:                 sink,
		cursor:               cursor,
		rfCache:              rfCache,
		details:              details,
		knobs:                knobs,
		topicDescriptorCache: make(map[TopicIdentifier]TopicDescriptor),
		topicNamer:           topicNamer,
	}
}

type tableDescriptorTopic struct {
	tableDesc           catalog.TableDescriptor
	spec                jobspb.ChangefeedTargetSpecification
	nameComponentsCache []string
	identifierCache     TopicIdentifier
}

func (tdt *tableDescriptorTopic) GetNameComponents() []string {
	__antithesis_instrumentation__.Notify(15615)
	if len(tdt.nameComponentsCache) == 0 {
		__antithesis_instrumentation__.Notify(15617)
		tdt.nameComponentsCache = []string{tdt.spec.StatementTimeName}
	} else {
		__antithesis_instrumentation__.Notify(15618)
	}
	__antithesis_instrumentation__.Notify(15616)
	return tdt.nameComponentsCache
}

func (tdt *tableDescriptorTopic) GetTopicIdentifier() TopicIdentifier {
	__antithesis_instrumentation__.Notify(15619)
	if tdt.identifierCache.TableID == 0 {
		__antithesis_instrumentation__.Notify(15621)
		tdt.identifierCache = TopicIdentifier{
			TableID: tdt.tableDesc.GetID(),
		}
	} else {
		__antithesis_instrumentation__.Notify(15622)
	}
	__antithesis_instrumentation__.Notify(15620)
	return tdt.identifierCache
}

func (tdt *tableDescriptorTopic) GetVersion() descpb.DescriptorVersion {
	__antithesis_instrumentation__.Notify(15623)
	return tdt.tableDesc.GetVersion()
}

func (tdt *tableDescriptorTopic) GetTargetSpecification() jobspb.ChangefeedTargetSpecification {
	__antithesis_instrumentation__.Notify(15624)
	return tdt.spec
}

var _ TopicDescriptor = &tableDescriptorTopic{}

type columnFamilyTopic struct {
	tableDesc           catalog.TableDescriptor
	familyDesc          descpb.ColumnFamilyDescriptor
	spec                jobspb.ChangefeedTargetSpecification
	nameComponentsCache []string
	identifierCache     TopicIdentifier
}

func (cft *columnFamilyTopic) GetNameComponents() []string {
	__antithesis_instrumentation__.Notify(15625)
	if len(cft.nameComponentsCache) == 0 {
		__antithesis_instrumentation__.Notify(15627)
		cft.nameComponentsCache = []string{
			cft.spec.StatementTimeName,
			cft.familyDesc.Name,
		}
	} else {
		__antithesis_instrumentation__.Notify(15628)
	}
	__antithesis_instrumentation__.Notify(15626)
	return cft.nameComponentsCache
}

func (cft *columnFamilyTopic) GetTopicIdentifier() TopicIdentifier {
	__antithesis_instrumentation__.Notify(15629)
	if cft.identifierCache.TableID == 0 {
		__antithesis_instrumentation__.Notify(15631)
		cft.identifierCache = TopicIdentifier{
			TableID:  cft.tableDesc.GetID(),
			FamilyID: cft.familyDesc.ID,
		}
	} else {
		__antithesis_instrumentation__.Notify(15632)
	}
	__antithesis_instrumentation__.Notify(15630)
	return cft.identifierCache
}

func (cft *columnFamilyTopic) GetVersion() descpb.DescriptorVersion {
	__antithesis_instrumentation__.Notify(15633)
	return cft.tableDesc.GetVersion()
}

func (cft *columnFamilyTopic) GetTargetSpecification() jobspb.ChangefeedTargetSpecification {
	__antithesis_instrumentation__.Notify(15634)
	return cft.spec
}

var _ TopicDescriptor = &columnFamilyTopic{}

type noTopic struct{}

func (n noTopic) GetNameComponents() []string {
	__antithesis_instrumentation__.Notify(15635)
	return []string{}
}

func (n noTopic) GetTopicIdentifier() TopicIdentifier {
	__antithesis_instrumentation__.Notify(15636)
	return TopicIdentifier{}
}

func (n noTopic) GetVersion() descpb.DescriptorVersion {
	__antithesis_instrumentation__.Notify(15637)
	return 0
}

func (n noTopic) GetTargetSpecification() jobspb.ChangefeedTargetSpecification {
	__antithesis_instrumentation__.Notify(15638)
	return jobspb.ChangefeedTargetSpecification{}
}

var _ TopicDescriptor = &noTopic{}

func makeTopicDescriptorFromSpecForRow(
	s jobspb.ChangefeedTargetSpecification, r encodeRow,
) (TopicDescriptor, error) {
	__antithesis_instrumentation__.Notify(15639)
	switch s.Type {
	case jobspb.ChangefeedTargetSpecification_PRIMARY_FAMILY_ONLY:
		__antithesis_instrumentation__.Notify(15640)
		return &tableDescriptorTopic{
			tableDesc: r.tableDesc,
			spec:      s,
		}, nil
	case jobspb.ChangefeedTargetSpecification_EACH_FAMILY, jobspb.ChangefeedTargetSpecification_COLUMN_FAMILY:
		__antithesis_instrumentation__.Notify(15641)
		familyDesc, err := r.tableDesc.FindFamilyByID(r.familyID)
		if err != nil {
			__antithesis_instrumentation__.Notify(15644)
			return noTopic{}, err
		} else {
			__antithesis_instrumentation__.Notify(15645)
		}
		__antithesis_instrumentation__.Notify(15642)
		return &columnFamilyTopic{
			tableDesc:  r.tableDesc,
			spec:       s,
			familyDesc: *familyDesc,
		}, nil
	default:
		__antithesis_instrumentation__.Notify(15643)
		return noTopic{}, errors.AssertionFailedf("Unsupported target type %s", s.Type)
	}
}

func (c *kvEventToRowConsumer) topicForRow(r encodeRow) (TopicDescriptor, error) {
	__antithesis_instrumentation__.Notify(15646)
	if topic, ok := c.topicDescriptorCache[TopicIdentifier{TableID: r.tableDesc.GetID(), FamilyID: r.familyID}]; ok {
		__antithesis_instrumentation__.Notify(15650)
		if topic.GetVersion() == r.tableDesc.GetVersion() {
			__antithesis_instrumentation__.Notify(15651)
			return topic, nil
		} else {
			__antithesis_instrumentation__.Notify(15652)
		}
	} else {
		__antithesis_instrumentation__.Notify(15653)
	}
	__antithesis_instrumentation__.Notify(15647)
	family, err := r.tableDesc.FindFamilyByID(r.familyID)
	if err != nil {
		__antithesis_instrumentation__.Notify(15654)
		return noTopic{}, err
	} else {
		__antithesis_instrumentation__.Notify(15655)
	}
	__antithesis_instrumentation__.Notify(15648)
	for _, s := range c.details.TargetSpecifications {
		__antithesis_instrumentation__.Notify(15656)
		if s.TableID == r.tableDesc.GetID() && func() bool {
			__antithesis_instrumentation__.Notify(15657)
			return (s.FamilyName == "" || func() bool {
				__antithesis_instrumentation__.Notify(15658)
				return s.FamilyName == family.Name == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(15659)
			topic, err := makeTopicDescriptorFromSpecForRow(s, r)
			if err != nil {
				__antithesis_instrumentation__.Notify(15661)
				return noTopic{}, err
			} else {
				__antithesis_instrumentation__.Notify(15662)
			}
			__antithesis_instrumentation__.Notify(15660)
			c.topicDescriptorCache[topic.GetTopicIdentifier()] = topic
			return topic, nil
		} else {
			__antithesis_instrumentation__.Notify(15663)
		}
	}
	__antithesis_instrumentation__.Notify(15649)
	return noTopic{}, errors.AssertionFailedf("no TargetSpecification for row %v", r)
}

func (c *kvEventToRowConsumer) ConsumeEvent(ctx context.Context, ev kvevent.Event) error {
	__antithesis_instrumentation__.Notify(15664)
	if ev.Type() != kvevent.TypeKV {
		__antithesis_instrumentation__.Notify(15675)
		return errors.AssertionFailedf("expected kv ev, got %v", ev.Type())
	} else {
		__antithesis_instrumentation__.Notify(15676)
	}
	__antithesis_instrumentation__.Notify(15665)

	r, err := c.eventToRow(ctx, ev)
	if err != nil {
		__antithesis_instrumentation__.Notify(15677)

		if errors.Is(err, ErrUnwatchedFamily) {
			__antithesis_instrumentation__.Notify(15679)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(15680)
		}
		__antithesis_instrumentation__.Notify(15678)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15681)
	}
	__antithesis_instrumentation__.Notify(15666)

	topic, err := c.topicForRow(r)
	if err != nil {
		__antithesis_instrumentation__.Notify(15682)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15683)
	}
	__antithesis_instrumentation__.Notify(15667)
	if c.topicNamer != nil {
		__antithesis_instrumentation__.Notify(15684)
		r.topic, err = c.topicNamer.Name(topic)
		if err != nil {
			__antithesis_instrumentation__.Notify(15685)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15686)
		}
	} else {
		__antithesis_instrumentation__.Notify(15687)
	}
	__antithesis_instrumentation__.Notify(15668)

	if r.updated.LessEq(c.frontier.Frontier()) && func() bool {
		__antithesis_instrumentation__.Notify(15688)
		return !r.updated.Equal(c.cursor) == true
	}() == true {
		__antithesis_instrumentation__.Notify(15689)
		log.Errorf(ctx, "cdc ux violation: detected timestamp %s that is less than "+
			"or equal to the local frontier %s.", r.updated, c.frontier.Frontier())
		return nil
	} else {
		__antithesis_instrumentation__.Notify(15690)
	}
	__antithesis_instrumentation__.Notify(15669)
	var keyCopy, valueCopy []byte
	encodedKey, err := c.encoder.EncodeKey(ctx, r)
	if err != nil {
		__antithesis_instrumentation__.Notify(15691)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15692)
	}
	__antithesis_instrumentation__.Notify(15670)
	c.scratch, keyCopy = c.scratch.Copy(encodedKey, 0)
	encodedValue, err := c.encoder.EncodeValue(ctx, r)
	if err != nil {
		__antithesis_instrumentation__.Notify(15693)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15694)
	}
	__antithesis_instrumentation__.Notify(15671)
	c.scratch, valueCopy = c.scratch.Copy(encodedValue, 0)

	if c.knobs.BeforeEmitRow != nil {
		__antithesis_instrumentation__.Notify(15695)
		if err := c.knobs.BeforeEmitRow(ctx); err != nil {
			__antithesis_instrumentation__.Notify(15696)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15697)
		}
	} else {
		__antithesis_instrumentation__.Notify(15698)
	}
	__antithesis_instrumentation__.Notify(15672)
	if err := c.sink.EmitRow(
		ctx, topic,
		keyCopy, valueCopy, r.updated, r.mvccTimestamp, ev.DetachAlloc(),
	); err != nil {
		__antithesis_instrumentation__.Notify(15699)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15700)
	}
	__antithesis_instrumentation__.Notify(15673)
	if log.V(3) {
		__antithesis_instrumentation__.Notify(15701)
		log.Infof(ctx, `r %s: %s -> %s`, r.tableDesc.GetName(), keyCopy, valueCopy)
	} else {
		__antithesis_instrumentation__.Notify(15702)
	}
	__antithesis_instrumentation__.Notify(15674)
	return nil
}

func (c *kvEventToRowConsumer) eventToRow(
	ctx context.Context, event kvevent.Event,
) (encodeRow, error) {
	__antithesis_instrumentation__.Notify(15703)
	var r encodeRow
	schemaTimestamp := event.KV().Value.Timestamp
	prevSchemaTimestamp := schemaTimestamp
	mvccTimestamp := event.MVCCTimestamp()

	if backfillTs := event.BackfillTimestamp(); !backfillTs.IsEmpty() {
		__antithesis_instrumentation__.Notify(15713)
		schemaTimestamp = backfillTs
		prevSchemaTimestamp = schemaTimestamp.Prev()
	} else {
		__antithesis_instrumentation__.Notify(15714)
	}
	__antithesis_instrumentation__.Notify(15704)

	desc, family, err := c.rfCache.TableDescForKey(ctx, event.KV().Key, schemaTimestamp)
	if err != nil {
		__antithesis_instrumentation__.Notify(15715)
		return r, err
	} else {
		__antithesis_instrumentation__.Notify(15716)
	}
	__antithesis_instrumentation__.Notify(15705)

	r.tableDesc = desc
	r.familyID = family
	var rf *row.Fetcher
	rf, err = c.rfCache.RowFetcherForColumnFamily(desc, family)
	if err != nil {
		__antithesis_instrumentation__.Notify(15717)
		return r, err
	} else {
		__antithesis_instrumentation__.Notify(15718)
	}
	__antithesis_instrumentation__.Notify(15706)

	c.kvFetcher.KVs = c.kvFetcher.KVs[:0]
	c.kvFetcher.KVs = append(c.kvFetcher.KVs, event.KV())
	if err := rf.StartScanFrom(ctx, &c.kvFetcher, false); err != nil {
		__antithesis_instrumentation__.Notify(15719)
		return r, err
	} else {
		__antithesis_instrumentation__.Notify(15720)
	}
	__antithesis_instrumentation__.Notify(15707)

	r.datums, err = rf.NextRow(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(15721)
		return r, err
	} else {
		__antithesis_instrumentation__.Notify(15722)
	}
	__antithesis_instrumentation__.Notify(15708)
	if r.datums == nil {
		__antithesis_instrumentation__.Notify(15723)
		return r, errors.AssertionFailedf("unexpected empty datums")
	} else {
		__antithesis_instrumentation__.Notify(15724)
	}
	__antithesis_instrumentation__.Notify(15709)
	r.datums = append(rowenc.EncDatumRow(nil), r.datums...)
	r.deleted = rf.RowIsDeleted()
	r.updated = schemaTimestamp
	r.mvccTimestamp = mvccTimestamp

	nextRow := encodeRow{
		tableDesc: desc,
	}
	nextRow.datums, err = rf.NextRow(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(15725)
		return r, err
	} else {
		__antithesis_instrumentation__.Notify(15726)
	}
	__antithesis_instrumentation__.Notify(15710)
	if nextRow.datums != nil {
		__antithesis_instrumentation__.Notify(15727)
		return r, errors.AssertionFailedf("unexpected non-empty datums")
	} else {
		__antithesis_instrumentation__.Notify(15728)
	}
	__antithesis_instrumentation__.Notify(15711)

	_, withDiff := c.details.Opts[changefeedbase.OptDiff]
	if withDiff {
		__antithesis_instrumentation__.Notify(15729)
		prevRF := rf
		r.prevTableDesc = r.tableDesc
		r.prevFamilyID = r.familyID
		if prevSchemaTimestamp != schemaTimestamp {
			__antithesis_instrumentation__.Notify(15735)

			prevDesc, family, err := c.rfCache.TableDescForKey(ctx, event.KV().Key, prevSchemaTimestamp)

			if err != nil {
				__antithesis_instrumentation__.Notify(15737)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(15738)
			}
			__antithesis_instrumentation__.Notify(15736)
			r.prevTableDesc = prevDesc
			r.prevFamilyID = family
			prevRF, err = c.rfCache.RowFetcherForColumnFamily(prevDesc, family)

			if err != nil {
				__antithesis_instrumentation__.Notify(15739)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(15740)
			}
		} else {
			__antithesis_instrumentation__.Notify(15741)
		}
		__antithesis_instrumentation__.Notify(15730)

		prevKV := roachpb.KeyValue{Key: event.KV().Key, Value: event.PrevValue()}

		c.kvFetcher.KVs = c.kvFetcher.KVs[:0]
		c.kvFetcher.KVs = append(c.kvFetcher.KVs, prevKV)
		if err := prevRF.StartScanFrom(ctx, &c.kvFetcher, false); err != nil {
			__antithesis_instrumentation__.Notify(15742)
			return r, err
		} else {
			__antithesis_instrumentation__.Notify(15743)
		}
		__antithesis_instrumentation__.Notify(15731)
		r.prevDatums, err = prevRF.NextRow(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(15744)
			return r, err
		} else {
			__antithesis_instrumentation__.Notify(15745)
		}
		__antithesis_instrumentation__.Notify(15732)
		if r.prevDatums == nil {
			__antithesis_instrumentation__.Notify(15746)
			return r, errors.AssertionFailedf("unexpected empty datums")
		} else {
			__antithesis_instrumentation__.Notify(15747)
		}
		__antithesis_instrumentation__.Notify(15733)
		r.prevDatums = append(rowenc.EncDatumRow(nil), r.prevDatums...)
		r.prevDeleted = prevRF.RowIsDeleted()

		nextRow := encodeRow{
			prevTableDesc: r.prevTableDesc,
		}
		nextRow.prevDatums, err = prevRF.NextRow(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(15748)
			return r, err
		} else {
			__antithesis_instrumentation__.Notify(15749)
		}
		__antithesis_instrumentation__.Notify(15734)
		if nextRow.prevDatums != nil {
			__antithesis_instrumentation__.Notify(15750)
			return r, errors.AssertionFailedf("unexpected non-empty datums")
		} else {
			__antithesis_instrumentation__.Notify(15751)
		}
	} else {
		__antithesis_instrumentation__.Notify(15752)
	}
	__antithesis_instrumentation__.Notify(15712)

	return r, nil
}

type nativeKVConsumer struct {
	sink Sink
}

var _ kvEventConsumer = &nativeKVConsumer{}

func newNativeKVConsumer(sink Sink) kvEventConsumer {
	__antithesis_instrumentation__.Notify(15753)
	return &nativeKVConsumer{sink: sink}
}

func (c *nativeKVConsumer) ConsumeEvent(ctx context.Context, ev kvevent.Event) error {
	__antithesis_instrumentation__.Notify(15754)
	if ev.Type() != kvevent.TypeKV {
		__antithesis_instrumentation__.Notify(15757)
		return errors.AssertionFailedf("expected kv ev, got %v", ev.Type())
	} else {
		__antithesis_instrumentation__.Notify(15758)
	}
	__antithesis_instrumentation__.Notify(15755)
	keyBytes := []byte(ev.KV().Key)
	val := ev.KV().Value
	valBytes, err := protoutil.Marshal(&val)
	if err != nil {
		__antithesis_instrumentation__.Notify(15759)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15760)
	}
	__antithesis_instrumentation__.Notify(15756)

	return c.sink.EmitRow(
		ctx, &noTopic{}, keyBytes, valBytes, val.Timestamp, val.Timestamp, ev.DetachAlloc())
}

const (
	emitAllResolved = 0
	emitNoResolved  = -1
)

type changeFrontier struct {
	execinfra.ProcessorBase

	flowCtx *execinfra.FlowCtx
	spec    execinfrapb.ChangeFrontierSpec
	memAcc  mon.BoundAccount
	a       tree.DatumAlloc

	input execinfra.RowSource

	frontier *schemaChangeFrontier

	encoder Encoder

	sink Sink

	freqEmitResolved time.Duration

	lastEmitResolved time.Time

	slowLogEveryN log.EveryN

	lastProtectedTimestampUpdate time.Time

	js *jobState

	highWaterAtStart hlc.Timestamp

	passthroughBuf encDatumRowBuffer

	resolvedBuf *encDatumRowBuffer

	metrics    *Metrics
	sliMetrics *sliMetrics

	metricsID int
}

const (
	runStatusUpdateFrequency time.Duration = time.Minute
	slowSpanMaxFrequency                   = 10 * time.Second
)

type jobState struct {
	job      *jobs.Job
	settings *cluster.Settings
	metrics  *Metrics
	ts       timeutil.TimeSource

	lastRunStatusUpdate time.Time

	lastProgressUpdate time.Time

	checkpointDuration time.Duration

	progressUpdatesSkipped bool
}

func newJobState(
	j *jobs.Job, st *cluster.Settings, metrics *Metrics, ts timeutil.TimeSource,
) *jobState {
	__antithesis_instrumentation__.Notify(15761)
	return &jobState{
		job:                j,
		settings:           st,
		metrics:            metrics,
		ts:                 ts,
		lastProgressUpdate: ts.Now(),
	}
}

func canCheckpointBackfill(sv *settings.Values, lastCheckpoint time.Time) bool {
	__antithesis_instrumentation__.Notify(15762)
	freq := changefeedbase.FrontierCheckpointFrequency.Get(sv)
	if freq == 0 {
		__antithesis_instrumentation__.Notify(15764)
		return false
	} else {
		__antithesis_instrumentation__.Notify(15765)
	}
	__antithesis_instrumentation__.Notify(15763)
	return timeutil.Since(lastCheckpoint) > freq
}

func (j *jobState) canCheckpointBackfill() bool {
	__antithesis_instrumentation__.Notify(15766)
	return canCheckpointBackfill(&j.settings.SV, j.lastProgressUpdate)
}

func (j *jobState) canCheckpointHighWatermark(frontierChanged bool) bool {
	__antithesis_instrumentation__.Notify(15767)
	if !(frontierChanged || func() bool {
		__antithesis_instrumentation__.Notify(15770)
		return j.progressUpdatesSkipped == true
	}() == true) {
		__antithesis_instrumentation__.Notify(15771)
		return false
	} else {
		__antithesis_instrumentation__.Notify(15772)
	}
	__antithesis_instrumentation__.Notify(15768)

	minAdvance := changefeedbase.MinHighWaterMarkCheckpointAdvance.Get(&j.settings.SV)
	if j.checkpointDuration > 0 && func() bool {
		__antithesis_instrumentation__.Notify(15773)
		return j.ts.Now().Before(j.lastProgressUpdate.Add(j.checkpointDuration+minAdvance)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(15774)

		j.progressUpdatesSkipped = true
		return false
	} else {
		__antithesis_instrumentation__.Notify(15775)
	}
	__antithesis_instrumentation__.Notify(15769)

	return true
}

func (j *jobState) checkpointCompleted(ctx context.Context, checkpointDuration time.Duration) {
	__antithesis_instrumentation__.Notify(15776)
	minAdvance := changefeedbase.MinHighWaterMarkCheckpointAdvance.Get(&j.settings.SV)
	if j.progressUpdatesSkipped {
		__antithesis_instrumentation__.Notify(15778)

		warnThreshold := 2 * minAdvance
		if warnThreshold < 60*time.Second {
			__antithesis_instrumentation__.Notify(15780)
			warnThreshold = 60 * time.Second
		} else {
			__antithesis_instrumentation__.Notify(15781)
		}
		__antithesis_instrumentation__.Notify(15779)
		behind := j.ts.Now().Sub(j.lastProgressUpdate)
		if behind > warnThreshold {
			__antithesis_instrumentation__.Notify(15782)
			log.Warningf(ctx, "high water mark update delayed by %s; mean checkpoint duration %s",
				behind, j.checkpointDuration)
		} else {
			__antithesis_instrumentation__.Notify(15783)
		}
	} else {
		__antithesis_instrumentation__.Notify(15784)
	}
	__antithesis_instrumentation__.Notify(15777)

	j.metrics.CheckpointHistNanos.RecordValue(checkpointDuration.Nanoseconds())
	j.lastProgressUpdate = j.ts.Now()
	j.checkpointDuration = time.Duration(j.metrics.CheckpointHistNanos.Snapshot().Mean())
	j.progressUpdatesSkipped = false
}

var _ execinfra.Processor = &changeFrontier{}
var _ execinfra.RowSource = &changeFrontier{}

func newChangeFrontierProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.ChangeFrontierSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(15785)
	ctx := flowCtx.EvalCtx.Ctx()
	memMonitor := execinfra.NewMonitor(ctx, flowCtx.EvalCtx.Mon, "changefntr-mem")
	sf, err := makeSchemaChangeFrontier(hlc.Timestamp{}, spec.TrackedSpans...)
	if err != nil {
		__antithesis_instrumentation__.Notify(15790)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15791)
	}
	__antithesis_instrumentation__.Notify(15786)
	cf := &changeFrontier{
		flowCtx:       flowCtx,
		spec:          spec,
		memAcc:        memMonitor.MakeBoundAccount(),
		input:         input,
		frontier:      sf,
		slowLogEveryN: log.Every(slowSpanMaxFrequency),
	}
	if err := cf.Init(
		cf,
		post,
		input.OutputTypes(),
		flowCtx,
		processorID,
		output,
		memMonitor,
		execinfra.ProcStateOpts{
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(15792)
				cf.close()
				return nil
			},
			InputsToDrain: []execinfra.RowSource{cf.input},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(15793)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15794)
	}
	__antithesis_instrumentation__.Notify(15787)

	if r, ok := cf.spec.Feed.Opts[changefeedbase.OptResolvedTimestamps]; ok {
		__antithesis_instrumentation__.Notify(15795)
		var err error
		if r == `` {
			__antithesis_instrumentation__.Notify(15796)

			cf.freqEmitResolved = emitAllResolved
		} else {
			__antithesis_instrumentation__.Notify(15797)
			if cf.freqEmitResolved, err = time.ParseDuration(r); err != nil {
				__antithesis_instrumentation__.Notify(15798)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(15799)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(15800)
		cf.freqEmitResolved = emitNoResolved
	}
	__antithesis_instrumentation__.Notify(15788)

	if cf.encoder, err = getEncoder(spec.Feed.Opts, AllTargets(spec.Feed)); err != nil {
		__antithesis_instrumentation__.Notify(15801)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15802)
	}
	__antithesis_instrumentation__.Notify(15789)

	return cf, nil
}

func (cf *changeFrontier) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(15803)
	return true
}

func (cf *changeFrontier) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(15804)
	if cf.spec.JobID != 0 {
		__antithesis_instrumentation__.Notify(15810)
		ctx = logtags.AddTag(ctx, "job", cf.spec.JobID)
	} else {
		__antithesis_instrumentation__.Notify(15811)
	}
	__antithesis_instrumentation__.Notify(15805)

	ctx = cf.StartInternal(ctx, changeFrontierProcName)
	cf.input.Start(ctx)

	cf.metrics = cf.flowCtx.Cfg.JobRegistry.MetricsStruct().Changefeed.(*Metrics)

	var nilOracle timestampLowerBoundOracle
	var err error
	sli, err := cf.metrics.getSLIMetrics(cf.spec.Feed.Opts[changefeedbase.OptMetricsScope])
	if err != nil {
		__antithesis_instrumentation__.Notify(15812)
		cf.MoveToDraining(err)
		return
	} else {
		__antithesis_instrumentation__.Notify(15813)
	}
	__antithesis_instrumentation__.Notify(15806)
	cf.sliMetrics = sli
	cf.sink, err = getSink(ctx, cf.flowCtx.Cfg, cf.spec.Feed, nilOracle,
		cf.spec.User(), cf.spec.JobID, sli)

	if err != nil {
		__antithesis_instrumentation__.Notify(15814)
		err = changefeedbase.MarkRetryableError(err)
		cf.MoveToDraining(err)
		return
	} else {
		__antithesis_instrumentation__.Notify(15815)
	}
	__antithesis_instrumentation__.Notify(15807)

	if b, ok := cf.sink.(*bufferSink); ok {
		__antithesis_instrumentation__.Notify(15816)
		cf.resolvedBuf = &b.buf
	} else {
		__antithesis_instrumentation__.Notify(15817)
	}
	__antithesis_instrumentation__.Notify(15808)

	cf.sink = &errorWrapperSink{wrapped: cf.sink}

	cf.highWaterAtStart = cf.spec.Feed.StatementTime
	if cf.spec.JobID != 0 {
		__antithesis_instrumentation__.Notify(15818)
		job, err := cf.flowCtx.Cfg.JobRegistry.LoadClaimedJob(ctx, cf.spec.JobID)
		if err != nil {
			__antithesis_instrumentation__.Notify(15822)
			cf.MoveToDraining(err)
			return
		} else {
			__antithesis_instrumentation__.Notify(15823)
		}
		__antithesis_instrumentation__.Notify(15819)
		cf.js = newJobState(job, cf.flowCtx.Cfg.Settings, cf.metrics, timeutil.DefaultTimeSource{})

		if changefeedbase.FrontierCheckpointFrequency.Get(&cf.flowCtx.Cfg.Settings.SV) == 0 {
			__antithesis_instrumentation__.Notify(15824)
			log.Warning(ctx,
				"Frontier checkpointing disabled; set changefeed.frontier_checkpoint_frequency to non-zero value to re-enable")
		} else {
			__antithesis_instrumentation__.Notify(15825)
		}
		__antithesis_instrumentation__.Notify(15820)

		p := job.Progress()
		if ts := p.GetHighWater(); ts != nil {
			__antithesis_instrumentation__.Notify(15826)
			cf.highWaterAtStart.Forward(*ts)
			cf.frontier.initialHighWater = *ts
			for _, span := range cf.spec.TrackedSpans {
				__antithesis_instrumentation__.Notify(15827)
				if _, err := cf.frontier.Forward(span, *ts); err != nil {
					__antithesis_instrumentation__.Notify(15828)
					cf.MoveToDraining(err)
					return
				} else {
					__antithesis_instrumentation__.Notify(15829)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(15830)
		}
		__antithesis_instrumentation__.Notify(15821)

		if p.RunningStatus != "" {
			__antithesis_instrumentation__.Notify(15831)

			cf.js.lastRunStatusUpdate = timeutil.Now()
		} else {
			__antithesis_instrumentation__.Notify(15832)
		}
	} else {
		__antithesis_instrumentation__.Notify(15833)
	}
	__antithesis_instrumentation__.Notify(15809)

	cf.metrics.mu.Lock()
	cf.metricsID = cf.metrics.mu.id
	cf.metrics.mu.id++
	sli.RunningCount.Inc(1)
	cf.metrics.mu.Unlock()

	go func() {
		__antithesis_instrumentation__.Notify(15834)
		<-ctx.Done()
		cf.closeMetrics()
	}()
}

func (cf *changeFrontier) close() {
	__antithesis_instrumentation__.Notify(15835)
	if cf.InternalClose() {
		__antithesis_instrumentation__.Notify(15836)
		if cf.metrics != nil {
			__antithesis_instrumentation__.Notify(15839)
			cf.closeMetrics()
		} else {
			__antithesis_instrumentation__.Notify(15840)
		}
		__antithesis_instrumentation__.Notify(15837)
		if cf.sink != nil {
			__antithesis_instrumentation__.Notify(15841)
			if err := cf.sink.Close(); err != nil {
				__antithesis_instrumentation__.Notify(15842)
				log.Warningf(cf.Ctx, `error closing sink. goroutines may have leaked: %v`, err)
			} else {
				__antithesis_instrumentation__.Notify(15843)
			}
		} else {
			__antithesis_instrumentation__.Notify(15844)
		}
		__antithesis_instrumentation__.Notify(15838)
		cf.memAcc.Close(cf.Ctx)
		cf.MemMonitor.Stop(cf.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(15845)
	}
}

func (cf *changeFrontier) closeMetrics() {
	__antithesis_instrumentation__.Notify(15846)

	cf.metrics.mu.Lock()
	if cf.metricsID > 0 {
		__antithesis_instrumentation__.Notify(15848)
		cf.sliMetrics.RunningCount.Dec(1)
	} else {
		__antithesis_instrumentation__.Notify(15849)
	}
	__antithesis_instrumentation__.Notify(15847)
	delete(cf.metrics.mu.resolved, cf.metricsID)
	cf.metricsID = -1
	cf.metrics.mu.Unlock()
}

func (cf *changeFrontier) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(15850)
	for cf.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(15852)
		if !cf.passthroughBuf.IsEmpty() {
			__antithesis_instrumentation__.Notify(15858)
			return cf.ProcessRowHelper(cf.passthroughBuf.Pop()), nil
		} else {
			__antithesis_instrumentation__.Notify(15859)
			if !cf.resolvedBuf.IsEmpty() {
				__antithesis_instrumentation__.Notify(15860)
				return cf.ProcessRowHelper(cf.resolvedBuf.Pop()), nil
			} else {
				__antithesis_instrumentation__.Notify(15861)
			}
		}
		__antithesis_instrumentation__.Notify(15853)

		if cf.frontier.schemaChangeBoundaryReached() && func() bool {
			__antithesis_instrumentation__.Notify(15862)
			return (cf.frontier.boundaryType == jobspb.ResolvedSpan_EXIT || func() bool {
				__antithesis_instrumentation__.Notify(15863)
				return cf.frontier.boundaryType == jobspb.ResolvedSpan_RESTART == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(15864)
			var err error
			endTime := cf.spec.Feed.EndTime
			if endTime.IsEmpty() || func() bool {
				__antithesis_instrumentation__.Notify(15866)
				return endTime.Less(cf.frontier.boundaryTime.Next()) == true
			}() == true {
				__antithesis_instrumentation__.Notify(15867)
				err = pgerror.Newf(pgcode.SchemaChangeOccurred,
					"schema change occurred at %v", cf.frontier.boundaryTime.Next().AsOfSystemTime())

				if cf.frontier.boundaryType == jobspb.ResolvedSpan_RESTART {
					__antithesis_instrumentation__.Notify(15868)
					err = changefeedbase.MarkRetryableError(err)
				} else {
					__antithesis_instrumentation__.Notify(15869)
				}
			} else {
				__antithesis_instrumentation__.Notify(15870)
			}
			__antithesis_instrumentation__.Notify(15865)

			cf.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(15871)
		}
		__antithesis_instrumentation__.Notify(15854)

		row, meta := cf.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(15872)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(15874)
				cf.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(15875)
			}
			__antithesis_instrumentation__.Notify(15873)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(15876)
		}
		__antithesis_instrumentation__.Notify(15855)
		if row == nil {
			__antithesis_instrumentation__.Notify(15877)
			cf.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(15878)
		}
		__antithesis_instrumentation__.Notify(15856)

		if row[0].IsNull() {
			__antithesis_instrumentation__.Notify(15879)

			cf.passthroughBuf.Push(row)
			continue
		} else {
			__antithesis_instrumentation__.Notify(15880)
		}
		__antithesis_instrumentation__.Notify(15857)

		if err := cf.noteAggregatorProgress(row[0]); err != nil {
			__antithesis_instrumentation__.Notify(15881)
			cf.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(15882)
		}
	}
	__antithesis_instrumentation__.Notify(15851)
	return nil, cf.DrainHelper()
}

func (cf *changeFrontier) noteAggregatorProgress(d rowenc.EncDatum) error {
	__antithesis_instrumentation__.Notify(15883)
	if err := d.EnsureDecoded(changefeeddist.ChangefeedResultTypes[0], &cf.a); err != nil {
		__antithesis_instrumentation__.Notify(15888)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15889)
	}
	__antithesis_instrumentation__.Notify(15884)
	raw, ok := d.Datum.(*tree.DBytes)
	if !ok {
		__antithesis_instrumentation__.Notify(15890)
		return errors.AssertionFailedf(`unexpected datum type %T: %s`, d.Datum, d.Datum)
	} else {
		__antithesis_instrumentation__.Notify(15891)
	}
	__antithesis_instrumentation__.Notify(15885)

	var resolvedSpans jobspb.ResolvedSpans
	if cf.flowCtx.Cfg.Settings.Version.IsActive(cf.Ctx, clusterversion.ChangefeedIdleness) {
		__antithesis_instrumentation__.Notify(15892)
		if err := protoutil.Unmarshal([]byte(*raw), &resolvedSpans); err != nil {
			__antithesis_instrumentation__.Notify(15894)
			return errors.NewAssertionErrorWithWrappedErrf(err,
				`unmarshalling aggregator progress update: %x`, raw)
		} else {
			__antithesis_instrumentation__.Notify(15895)
		}
		__antithesis_instrumentation__.Notify(15893)

		cf.maybeMarkJobIdle(resolvedSpans.Stats.RecentKvCount)
	} else {
		__antithesis_instrumentation__.Notify(15896)

		var resolved jobspb.ResolvedSpan
		if err := protoutil.Unmarshal([]byte(*raw), &resolved); err != nil {
			__antithesis_instrumentation__.Notify(15898)
			return errors.NewAssertionErrorWithWrappedErrf(err,
				`unmarshalling resolved span: %x`, raw)
		} else {
			__antithesis_instrumentation__.Notify(15899)
		}
		__antithesis_instrumentation__.Notify(15897)
		resolvedSpans = jobspb.ResolvedSpans{
			ResolvedSpans: []jobspb.ResolvedSpan{resolved},
		}
	}
	__antithesis_instrumentation__.Notify(15886)

	for _, resolved := range resolvedSpans.ResolvedSpans {
		__antithesis_instrumentation__.Notify(15900)

		if !resolved.Timestamp.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(15902)
			return resolved.Timestamp.Less(cf.highWaterAtStart) == true
		}() == true {
			__antithesis_instrumentation__.Notify(15903)
			logcrash.ReportOrPanic(cf.Ctx, &cf.flowCtx.Cfg.Settings.SV,
				`got a span level timestamp %s for %s that is less than the initial high-water %s`,
				redact.Safe(resolved.Timestamp), resolved.Span, redact.Safe(cf.highWaterAtStart))
			continue
		} else {
			__antithesis_instrumentation__.Notify(15904)
		}
		__antithesis_instrumentation__.Notify(15901)
		if err := cf.forwardFrontier(resolved); err != nil {
			__antithesis_instrumentation__.Notify(15905)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15906)
		}
	}
	__antithesis_instrumentation__.Notify(15887)

	return nil
}

func (cf *changeFrontier) forwardFrontier(resolved jobspb.ResolvedSpan) error {
	__antithesis_instrumentation__.Notify(15907)
	frontierChanged, err := cf.frontier.ForwardResolvedSpan(resolved)
	if err != nil {
		__antithesis_instrumentation__.Notify(15911)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15912)
	}
	__antithesis_instrumentation__.Notify(15908)

	cf.maybeLogBehindSpan(frontierChanged)

	emitResolved := frontierChanged

	if cf.js != nil {
		__antithesis_instrumentation__.Notify(15913)
		checkpointed, err := cf.maybeCheckpointJob(resolved, frontierChanged)
		if err != nil {
			__antithesis_instrumentation__.Notify(15915)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15916)
		}
		__antithesis_instrumentation__.Notify(15914)

		emitResolved = checkpointed
	} else {
		__antithesis_instrumentation__.Notify(15917)
	}
	__antithesis_instrumentation__.Notify(15909)

	if emitResolved {
		__antithesis_instrumentation__.Notify(15918)

		newResolved := cf.frontier.Frontier()
		cf.metrics.mu.Lock()
		if cf.metricsID != -1 {
			__antithesis_instrumentation__.Notify(15920)
			cf.metrics.mu.resolved[cf.metricsID] = newResolved
		} else {
			__antithesis_instrumentation__.Notify(15921)
		}
		__antithesis_instrumentation__.Notify(15919)
		cf.metrics.mu.Unlock()

		return cf.maybeEmitResolved(newResolved)
	} else {
		__antithesis_instrumentation__.Notify(15922)
	}
	__antithesis_instrumentation__.Notify(15910)

	return nil
}

func (cf *changeFrontier) maybeMarkJobIdle(recentKVCount uint64) {
	__antithesis_instrumentation__.Notify(15923)
	if cf.spec.JobID == 0 {
		__antithesis_instrumentation__.Notify(15927)
		return
	} else {
		__antithesis_instrumentation__.Notify(15928)
	}
	__antithesis_instrumentation__.Notify(15924)

	if recentKVCount > 0 {
		__antithesis_instrumentation__.Notify(15929)
		cf.frontier.ForwardLatestKV(timeutil.Now())
	} else {
		__antithesis_instrumentation__.Notify(15930)
	}
	__antithesis_instrumentation__.Notify(15925)

	idleTimeout := changefeedbase.IdleTimeout.Get(&cf.flowCtx.Cfg.Settings.SV)
	if idleTimeout == 0 {
		__antithesis_instrumentation__.Notify(15931)
		return
	} else {
		__antithesis_instrumentation__.Notify(15932)
	}
	__antithesis_instrumentation__.Notify(15926)

	isIdle := timeutil.Since(cf.frontier.latestKV) > idleTimeout
	cf.js.job.MarkIdle(isIdle)
}

func (cf *changeFrontier) maybeCheckpointJob(
	resolvedSpan jobspb.ResolvedSpan, frontierChanged bool,
) (bool, error) {
	__antithesis_instrumentation__.Notify(15933)

	inBackfill := !frontierChanged && func() bool {
		__antithesis_instrumentation__.Notify(15936)
		return resolvedSpan.Timestamp.Equal(cf.frontier.BackfillTS()) == true
	}() == true

	updateCheckpoint := inBackfill && func() bool {
		__antithesis_instrumentation__.Notify(15937)
		return cf.js.canCheckpointBackfill() == true
	}() == true

	var checkpoint jobspb.ChangefeedProgress_Checkpoint
	if updateCheckpoint {
		__antithesis_instrumentation__.Notify(15938)
		maxBytes := changefeedbase.FrontierCheckpointMaxBytes.Get(&cf.flowCtx.Cfg.Settings.SV)
		checkpoint.Spans = cf.frontier.getCheckpointSpans(maxBytes)
	} else {
		__antithesis_instrumentation__.Notify(15939)
	}
	__antithesis_instrumentation__.Notify(15934)

	updateHighWater :=
		!inBackfill && func() bool {
			__antithesis_instrumentation__.Notify(15940)
			return (cf.frontier.schemaChangeBoundaryReached() || func() bool {
				__antithesis_instrumentation__.Notify(15941)
				return cf.js.canCheckpointHighWatermark(frontierChanged) == true
			}() == true) == true
		}() == true

	if updateCheckpoint || func() bool {
		__antithesis_instrumentation__.Notify(15942)
		return updateHighWater == true
	}() == true {
		__antithesis_instrumentation__.Notify(15943)
		checkpointStart := timeutil.Now()
		if err := cf.checkpointJobProgress(cf.frontier.Frontier(), checkpoint); err != nil {
			__antithesis_instrumentation__.Notify(15945)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(15946)
		}
		__antithesis_instrumentation__.Notify(15944)
		cf.js.checkpointCompleted(cf.Ctx, timeutil.Since(checkpointStart))
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(15947)
	}
	__antithesis_instrumentation__.Notify(15935)

	return false, nil
}

func (cf *changeFrontier) checkpointJobProgress(
	frontier hlc.Timestamp, checkpoint jobspb.ChangefeedProgress_Checkpoint,
) (err error) {
	__antithesis_instrumentation__.Notify(15948)
	updateRunStatus := timeutil.Since(cf.js.lastRunStatusUpdate) > runStatusUpdateFrequency
	if updateRunStatus {
		__antithesis_instrumentation__.Notify(15950)
		defer func() { __antithesis_instrumentation__.Notify(15951); cf.js.lastRunStatusUpdate = timeutil.Now() }()
	} else {
		__antithesis_instrumentation__.Notify(15952)
	}
	__antithesis_instrumentation__.Notify(15949)
	cf.metrics.FrontierUpdates.Inc(1)

	return cf.js.job.Update(cf.Ctx, nil, func(
		txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater,
	) error {
		__antithesis_instrumentation__.Notify(15953)
		if err := md.CheckRunningOrReverting(); err != nil {
			__antithesis_instrumentation__.Notify(15959)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15960)
		}
		__antithesis_instrumentation__.Notify(15954)

		progress := md.Progress
		progress.Progress = &jobspb.Progress_HighWater{
			HighWater: &frontier,
		}

		changefeedProgress := progress.Details.(*jobspb.Progress_Changefeed).Changefeed
		changefeedProgress.Checkpoint = &checkpoint

		timestampManager := cf.manageProtectedTimestamps

		if !changefeedbase.ActiveProtectedTimestampsEnabled.Get(&cf.flowCtx.Cfg.Settings.SV) {
			__antithesis_instrumentation__.Notify(15961)
			timestampManager = cf.deprecatedManageProtectedTimestamps
		} else {
			__antithesis_instrumentation__.Notify(15962)
		}
		__antithesis_instrumentation__.Notify(15955)
		if err := timestampManager(cf.Ctx, txn, changefeedProgress); err != nil {
			__antithesis_instrumentation__.Notify(15963)
			log.Warningf(cf.Ctx, "error managing protected timestamp record: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(15964)
		}
		__antithesis_instrumentation__.Notify(15956)

		if updateRunStatus {
			__antithesis_instrumentation__.Notify(15965)
			md.Progress.RunningStatus = fmt.Sprintf("running: resolved=%s", frontier)
		} else {
			__antithesis_instrumentation__.Notify(15966)
		}
		__antithesis_instrumentation__.Notify(15957)

		ju.UpdateProgress(progress)

		if md.RunStats != nil {
			__antithesis_instrumentation__.Notify(15967)
			ju.UpdateRunStats(1, md.RunStats.LastRun)
		} else {
			__antithesis_instrumentation__.Notify(15968)
		}
		__antithesis_instrumentation__.Notify(15958)

		return nil
	})
}

func (cf *changeFrontier) manageProtectedTimestamps(
	ctx context.Context, txn *kv.Txn, progress *jobspb.ChangefeedProgress,
) error {
	__antithesis_instrumentation__.Notify(15969)
	ptsUpdateInterval := changefeedbase.ProtectTimestampInterval.Get(&cf.flowCtx.Cfg.Settings.SV)
	if timeutil.Since(cf.lastProtectedTimestampUpdate) < ptsUpdateInterval {
		__antithesis_instrumentation__.Notify(15973)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(15974)
	}
	__antithesis_instrumentation__.Notify(15970)
	cf.lastProtectedTimestampUpdate = timeutil.Now()

	pts := cf.flowCtx.Cfg.ProtectedTimestampProvider

	highWater := cf.frontier.Frontier()
	if highWater.Less(cf.highWaterAtStart) {
		__antithesis_instrumentation__.Notify(15975)
		highWater = cf.highWaterAtStart
	} else {
		__antithesis_instrumentation__.Notify(15976)
	}
	__antithesis_instrumentation__.Notify(15971)

	recordID := progress.ProtectedTimestampRecord
	if recordID == uuid.Nil {
		__antithesis_instrumentation__.Notify(15977)
		ptr := createProtectedTimestampRecord(ctx, cf.flowCtx.Codec(), cf.spec.JobID, AllTargets(cf.spec.Feed), highWater, progress)
		if err := pts.Protect(ctx, txn, ptr); err != nil {
			__antithesis_instrumentation__.Notify(15978)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15979)
		}
	} else {
		__antithesis_instrumentation__.Notify(15980)
		log.VEventf(ctx, 2, "updating protected timestamp %v at %v", recordID, highWater)
		if err := pts.UpdateTimestamp(ctx, txn, recordID, highWater); err != nil {
			__antithesis_instrumentation__.Notify(15981)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15982)
		}
	}
	__antithesis_instrumentation__.Notify(15972)

	return nil
}

func (cf *changeFrontier) deprecatedManageProtectedTimestamps(
	ctx context.Context, txn *kv.Txn, progress *jobspb.ChangefeedProgress,
) error {
	__antithesis_instrumentation__.Notify(15983)
	pts := cf.flowCtx.Cfg.ProtectedTimestampProvider
	if err := cf.deprecatedMaybeReleaseProtectedTimestamp(ctx, progress, pts, txn); err != nil {
		__antithesis_instrumentation__.Notify(15986)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15987)
	}
	__antithesis_instrumentation__.Notify(15984)

	schemaChangePolicy := changefeedbase.SchemaChangePolicy(cf.spec.Feed.Opts[changefeedbase.OptSchemaChangePolicy])
	shouldProtectBoundaries := schemaChangePolicy == changefeedbase.OptSchemaChangePolicyBackfill
	if cf.frontier.schemaChangeBoundaryReached() && func() bool {
		__antithesis_instrumentation__.Notify(15988)
		return shouldProtectBoundaries == true
	}() == true {
		__antithesis_instrumentation__.Notify(15989)
		highWater := cf.frontier.Frontier()
		ptr := createProtectedTimestampRecord(ctx, cf.flowCtx.Codec(), cf.spec.JobID, AllTargets(cf.spec.Feed), highWater, progress)
		return pts.Protect(ctx, txn, ptr)
	} else {
		__antithesis_instrumentation__.Notify(15990)
	}
	__antithesis_instrumentation__.Notify(15985)
	return nil
}

func (cf *changeFrontier) deprecatedMaybeReleaseProtectedTimestamp(
	ctx context.Context, progress *jobspb.ChangefeedProgress, pts protectedts.Storage, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(15991)
	if progress.ProtectedTimestampRecord == uuid.Nil {
		__antithesis_instrumentation__.Notify(15995)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(15996)
	}
	__antithesis_instrumentation__.Notify(15992)
	if !cf.frontier.schemaChangeBoundaryReached() && func() bool {
		__antithesis_instrumentation__.Notify(15997)
		return cf.isBehind() == true
	}() == true {
		__antithesis_instrumentation__.Notify(15998)
		log.VEventf(ctx, 2, "not releasing protected timestamp because changefeed is behind")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(15999)
	}
	__antithesis_instrumentation__.Notify(15993)
	log.VEventf(ctx, 2, "releasing protected timestamp %v",
		progress.ProtectedTimestampRecord)
	if err := pts.Release(ctx, txn, progress.ProtectedTimestampRecord); err != nil {
		__antithesis_instrumentation__.Notify(16000)
		return err
	} else {
		__antithesis_instrumentation__.Notify(16001)
	}
	__antithesis_instrumentation__.Notify(15994)
	progress.ProtectedTimestampRecord = uuid.Nil
	return nil
}

func (cf *changeFrontier) maybeEmitResolved(newResolved hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(16002)
	if cf.freqEmitResolved == emitNoResolved || func() bool {
		__antithesis_instrumentation__.Notify(16006)
		return newResolved.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(16007)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(16008)
	}
	__antithesis_instrumentation__.Notify(16003)
	sinceEmitted := newResolved.GoTime().Sub(cf.lastEmitResolved)
	shouldEmit := sinceEmitted >= cf.freqEmitResolved || func() bool {
		__antithesis_instrumentation__.Notify(16009)
		return cf.frontier.schemaChangeBoundaryReached() == true
	}() == true
	if !shouldEmit {
		__antithesis_instrumentation__.Notify(16010)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(16011)
	}
	__antithesis_instrumentation__.Notify(16004)
	if err := emitResolvedTimestamp(cf.Ctx, cf.encoder, cf.sink, newResolved); err != nil {
		__antithesis_instrumentation__.Notify(16012)
		return err
	} else {
		__antithesis_instrumentation__.Notify(16013)
	}
	__antithesis_instrumentation__.Notify(16005)
	cf.lastEmitResolved = newResolved.GoTime()
	return nil
}

func (cf *changeFrontier) isBehind() bool {
	__antithesis_instrumentation__.Notify(16014)
	frontier := cf.frontier.Frontier()
	if frontier.IsEmpty() {
		__antithesis_instrumentation__.Notify(16016)

		return true
	} else {
		__antithesis_instrumentation__.Notify(16017)
	}
	__antithesis_instrumentation__.Notify(16015)

	return timeutil.Since(frontier.GoTime()) > cf.slownessThreshold()
}

func (cf *changeFrontier) maybeLogBehindSpan(frontierChanged bool) {
	__antithesis_instrumentation__.Notify(16018)
	if !cf.isBehind() {
		__antithesis_instrumentation__.Notify(16023)
		return
	} else {
		__antithesis_instrumentation__.Notify(16024)
	}
	__antithesis_instrumentation__.Notify(16019)

	frontier := cf.frontier.Frontier()
	if frontier.IsEmpty() {
		__antithesis_instrumentation__.Notify(16025)
		return
	} else {
		__antithesis_instrumentation__.Notify(16026)
	}
	__antithesis_instrumentation__.Notify(16020)

	now := timeutil.Now()
	resolvedBehind := now.Sub(frontier.GoTime())

	description := "sinkless feed"
	if !cf.isSinkless() {
		__antithesis_instrumentation__.Notify(16027)
		description = fmt.Sprintf("job %d", cf.spec.JobID)
	} else {
		__antithesis_instrumentation__.Notify(16028)
	}
	__antithesis_instrumentation__.Notify(16021)
	if frontierChanged {
		__antithesis_instrumentation__.Notify(16029)
		log.Infof(cf.Ctx, "%s new resolved timestamp %s is behind by %s",
			description, frontier, resolvedBehind)
	} else {
		__antithesis_instrumentation__.Notify(16030)
	}
	__antithesis_instrumentation__.Notify(16022)

	if cf.slowLogEveryN.ShouldProcess(now) {
		__antithesis_instrumentation__.Notify(16031)
		s := cf.frontier.PeekFrontierSpan()
		log.Infof(cf.Ctx, "%s span %s is behind by %s", description, s, resolvedBehind)
	} else {
		__antithesis_instrumentation__.Notify(16032)
	}
}

func (cf *changeFrontier) slownessThreshold() time.Duration {
	__antithesis_instrumentation__.Notify(16033)
	clusterThreshold := changefeedbase.SlowSpanLogThreshold.Get(&cf.flowCtx.Cfg.Settings.SV)
	if clusterThreshold > 0 {
		__antithesis_instrumentation__.Notify(16035)
		return clusterThreshold
	} else {
		__antithesis_instrumentation__.Notify(16036)
	}
	__antithesis_instrumentation__.Notify(16034)

	pollInterval := changefeedbase.TableDescriptorPollInterval.Get(&cf.flowCtx.Cfg.Settings.SV)
	closedtsInterval := closedts.TargetDuration.Get(&cf.flowCtx.Cfg.Settings.SV)
	return time.Second + 10*(pollInterval+closedtsInterval)
}

func (cf *changeFrontier) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(16037)

	cf.close()
}

func (cf *changeFrontier) isSinkless() bool {
	__antithesis_instrumentation__.Notify(16038)
	return cf.spec.JobID == 0
}

type spanFrontier struct {
	*span.Frontier
}

func (s *spanFrontier) frontierTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(16039)
	return s.Frontier.Frontier()
}

type schemaChangeFrontier struct {
	*spanFrontier

	boundaryTime hlc.Timestamp

	boundaryType jobspb.ResolvedSpan_BoundaryType

	latestTs hlc.Timestamp

	initialHighWater hlc.Timestamp

	latestKV time.Time
}

func makeSchemaChangeFrontier(
	initialHighWater hlc.Timestamp, spans ...roachpb.Span,
) (*schemaChangeFrontier, error) {
	__antithesis_instrumentation__.Notify(16040)
	sf, err := span.MakeFrontier(spans...)
	if err != nil {
		__antithesis_instrumentation__.Notify(16043)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16044)
	}
	__antithesis_instrumentation__.Notify(16041)
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(16045)
		if _, err := sf.Forward(span, initialHighWater); err != nil {
			__antithesis_instrumentation__.Notify(16046)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16047)
		}
	}
	__antithesis_instrumentation__.Notify(16042)
	return &schemaChangeFrontier{spanFrontier: &spanFrontier{sf}, initialHighWater: initialHighWater, latestTs: initialHighWater}, nil
}

func (f *schemaChangeFrontier) ForwardResolvedSpan(r jobspb.ResolvedSpan) (bool, error) {
	__antithesis_instrumentation__.Notify(16048)
	if r.BoundaryType != jobspb.ResolvedSpan_NONE {
		__antithesis_instrumentation__.Notify(16051)
		if !f.boundaryTime.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(16053)
			return r.Timestamp.Less(f.boundaryTime) == true
		}() == true {
			__antithesis_instrumentation__.Notify(16054)

			return false, errors.AssertionFailedf("received boundary timestamp %v < %v "+
				"of type %v before reaching existing boundary of type %v",
				r.Timestamp, f.boundaryTime, r.BoundaryType, f.boundaryType)
		} else {
			__antithesis_instrumentation__.Notify(16055)
		}
		__antithesis_instrumentation__.Notify(16052)
		f.boundaryTime = r.Timestamp
		f.boundaryType = r.BoundaryType
	} else {
		__antithesis_instrumentation__.Notify(16056)
	}
	__antithesis_instrumentation__.Notify(16049)

	if f.latestTs.Less(r.Timestamp) {
		__antithesis_instrumentation__.Notify(16057)
		f.latestTs = r.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(16058)
	}
	__antithesis_instrumentation__.Notify(16050)
	return f.Forward(r.Span, r.Timestamp)
}

func (f *schemaChangeFrontier) ForwardLatestKV(ts time.Time) {
	__antithesis_instrumentation__.Notify(16059)
	if f.latestKV.Before(ts) {
		__antithesis_instrumentation__.Notify(16060)
		f.latestKV = ts
	} else {
		__antithesis_instrumentation__.Notify(16061)
	}
}

func (f *schemaChangeFrontier) Frontier() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(16062)
	return f.frontierTimestamp()
}

func (f *schemaChangeFrontier) SpanFrontier() *span.Frontier {
	__antithesis_instrumentation__.Notify(16063)
	return f.spanFrontier.Frontier
}

func (f *schemaChangeFrontier) getCheckpointSpans(maxBytes int64) (checkpoint []roachpb.Span) {
	__antithesis_instrumentation__.Notify(16064)
	var used int64
	frontier := f.frontierTimestamp()
	f.Entries(func(s roachpb.Span, ts hlc.Timestamp) span.OpResult {
		__antithesis_instrumentation__.Notify(16066)
		if frontier.Less(ts) {
			__antithesis_instrumentation__.Notify(16068)
			used += int64(len(s.Key)) + int64(len(s.EndKey))
			if used > maxBytes {
				__antithesis_instrumentation__.Notify(16070)
				return span.StopMatch
			} else {
				__antithesis_instrumentation__.Notify(16071)
			}
			__antithesis_instrumentation__.Notify(16069)
			checkpoint = append(checkpoint, s)
		} else {
			__antithesis_instrumentation__.Notify(16072)
		}
		__antithesis_instrumentation__.Notify(16067)
		return span.ContinueMatch
	})
	__antithesis_instrumentation__.Notify(16065)
	return checkpoint
}

func (f *schemaChangeFrontier) BackfillTS() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(16073)
	frontier := f.Frontier()

	if frontier.IsEmpty() {
		__antithesis_instrumentation__.Notify(16076)
		return f.initialHighWater
	} else {
		__antithesis_instrumentation__.Notify(16077)
	}
	__antithesis_instrumentation__.Notify(16074)

	backfilling := f.boundaryType == jobspb.ResolvedSpan_BACKFILL && func() bool {
		__antithesis_instrumentation__.Notify(16078)
		return frontier.Equal(f.boundaryTime) == true
	}() == true

	restarted := frontier.Equal(f.initialHighWater)
	if backfilling || func() bool {
		__antithesis_instrumentation__.Notify(16079)
		return restarted == true
	}() == true {
		__antithesis_instrumentation__.Notify(16080)
		return frontier.Next()
	} else {
		__antithesis_instrumentation__.Notify(16081)
	}
	__antithesis_instrumentation__.Notify(16075)
	return hlc.Timestamp{}
}

func (f *schemaChangeFrontier) schemaChangeBoundaryReached() (r bool) {
	__antithesis_instrumentation__.Notify(16082)
	return f.boundaryTime.Equal(f.Frontier()) && func() bool {
		__antithesis_instrumentation__.Notify(16083)
		return f.latestTs.Equal(f.boundaryTime) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(16084)
		return f.boundaryType != jobspb.ResolvedSpan_NONE == true
	}() == true
}
