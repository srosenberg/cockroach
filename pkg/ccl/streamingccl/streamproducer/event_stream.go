package streamproducer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streampb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/span"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type eventStream struct {
	streamID streaming.StreamID
	execCfg  *sql.ExecutorConfig
	spec     streampb.StreamPartitionSpec
	mon      *mon.BytesMonitor
	acc      mon.BoundAccount

	data tree.Datums

	rf          *rangefeed.RangeFeed
	streamGroup ctxgroup.Group
	eventsCh    chan roachpb.RangeFeedEvent
	errCh       chan error
	streamCh    chan tree.Datums
	sp          *tracing.Span
}

var _ tree.ValueGenerator = (*eventStream)(nil)

var eventStreamReturnType = types.MakeLabeledTuple(
	[]*types.T{types.Bytes},
	[]string{"stream_event"},
)

func (s *eventStream) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(26819)
	return eventStreamReturnType
}

func (s *eventStream) Start(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(26820)

	if s.errCh != nil {
		__antithesis_instrumentation__.Notify(26827)
		return errors.AssertionFailedf("expected to be started once")
	} else {
		__antithesis_instrumentation__.Notify(26828)
	}
	__antithesis_instrumentation__.Notify(26821)

	s.acc = s.mon.MakeBoundAccount()

	s.errCh = make(chan error)

	s.eventsCh = make(chan roachpb.RangeFeedEvent)

	s.streamCh = make(chan tree.Datums)

	opts := []rangefeed.Option{
		rangefeed.WithOnCheckpoint(s.onCheckpoint),

		rangefeed.WithOnInternalError(func(ctx context.Context, err error) {
			__antithesis_instrumentation__.Notify(26829)
			s.maybeSetError(err)
		}),

		rangefeed.WithMemoryMonitor(s.mon),
	}
	__antithesis_instrumentation__.Notify(26822)

	frontier, err := span.MakeFrontier(s.spec.Spans...)
	if err != nil {
		__antithesis_instrumentation__.Notify(26830)
		return err
	} else {
		__antithesis_instrumentation__.Notify(26831)
	}
	__antithesis_instrumentation__.Notify(26823)

	if s.spec.StartFrom.IsEmpty() {
		__antithesis_instrumentation__.Notify(26832)

		s.spec.StartFrom = s.execCfg.Clock.Now()

		opts = append(opts,
			rangefeed.WithInitialScan(func(ctx context.Context) { __antithesis_instrumentation__.Notify(26833) }),
			rangefeed.WithScanRetryBehavior(rangefeed.ScanRetryRemaining),

			rangefeed.WithOnInitialScanError(func(ctx context.Context, err error) (shouldFail bool) {
				__antithesis_instrumentation__.Notify(26834)

				return false
			}),

			rangefeed.WithInitialScanParallelismFn(func() int {
				__antithesis_instrumentation__.Notify(26835)
				return int(s.spec.Config.InitialScanParallelism)
			}),

			rangefeed.WithOnScanCompleted(s.onSpanCompleted),
		)
	} else {
		__antithesis_instrumentation__.Notify(26836)

		for _, sp := range s.spec.Spans {
			__antithesis_instrumentation__.Notify(26837)
			if _, err := frontier.Forward(sp, s.spec.StartFrom); err != nil {
				__antithesis_instrumentation__.Notify(26838)
				return err
			} else {
				__antithesis_instrumentation__.Notify(26839)
			}
		}
	}
	__antithesis_instrumentation__.Notify(26824)

	s.rf = s.execCfg.RangeFeedFactory.New(
		fmt.Sprintf("streamID=%d", s.streamID), s.spec.StartFrom, s.onEvent, opts...)
	if err := s.rf.Start(ctx, s.spec.Spans); err != nil {
		__antithesis_instrumentation__.Notify(26840)
		return err
	} else {
		__antithesis_instrumentation__.Notify(26841)
	}
	__antithesis_instrumentation__.Notify(26825)

	if err := s.acc.Grow(ctx, s.spec.Config.BatchByteSize); err != nil {
		__antithesis_instrumentation__.Notify(26842)
		return errors.Wrapf(err, "failed to allocated %d bytes from monitor", s.spec.Config.BatchByteSize)
	} else {
		__antithesis_instrumentation__.Notify(26843)
	}
	__antithesis_instrumentation__.Notify(26826)

	s.startStreamProcessor(ctx, frontier)
	return nil
}

func (s *eventStream) maybeSetError(err error) {
	__antithesis_instrumentation__.Notify(26844)
	select {
	case s.errCh <- err:
		__antithesis_instrumentation__.Notify(26845)
	default:
		__antithesis_instrumentation__.Notify(26846)
	}
}

func (s *eventStream) startStreamProcessor(ctx context.Context, frontier *span.Frontier) {
	__antithesis_instrumentation__.Notify(26847)
	type ctxGroupFn = func(ctx context.Context) error

	withErrCapture := func(fn ctxGroupFn) ctxGroupFn {
		__antithesis_instrumentation__.Notify(26849)
		return func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(26850)
			err := fn(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(26852)

				log.Errorf(ctx, "event stream %d terminating with error %v", s.streamID, err)
				s.maybeSetError(err)
			} else {
				__antithesis_instrumentation__.Notify(26853)
			}
			__antithesis_instrumentation__.Notify(26851)
			return err
		}
	}
	__antithesis_instrumentation__.Notify(26848)

	streamCtx, sp := tracing.ChildSpan(ctx, "event stream")
	s.sp = sp
	s.streamGroup = ctxgroup.WithContext(streamCtx)
	s.streamGroup.GoCtx(withErrCapture(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(26854)
		return s.streamLoop(ctx, frontier)
	}))

}

func (s *eventStream) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(26855)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(26856)
		return false, ctx.Err()
	case err := <-s.errCh:
		__antithesis_instrumentation__.Notify(26857)
		return false, err
	case s.data = <-s.streamCh:
		__antithesis_instrumentation__.Notify(26858)
		return true, nil
	}
}

func (s *eventStream) Values() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(26859)
	return s.data, nil
}

func (s *eventStream) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(26860)
	s.rf.Close()
	s.acc.Close(ctx)

	if err := s.streamGroup.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(26862)

		log.Errorf(ctx, "partition stream %d terminated with error %v", s.streamID, err)
	} else {
		__antithesis_instrumentation__.Notify(26863)
	}
	__antithesis_instrumentation__.Notify(26861)

	s.sp.Finish()
}

func (s *eventStream) onEvent(ctx context.Context, value *roachpb.RangeFeedValue) {
	__antithesis_instrumentation__.Notify(26864)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(26865)
	case s.eventsCh <- roachpb.RangeFeedEvent{Val: value}:
		__antithesis_instrumentation__.Notify(26866)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(26867)
			log.Infof(ctx, "onEvent: %s@%s", value.Key, value.Value.Timestamp)
		} else {
			__antithesis_instrumentation__.Notify(26868)
		}
	}
}

func (s *eventStream) onCheckpoint(ctx context.Context, checkpoint *roachpb.RangeFeedCheckpoint) {
	__antithesis_instrumentation__.Notify(26869)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(26870)
	case s.eventsCh <- roachpb.RangeFeedEvent{Checkpoint: checkpoint}:
		__antithesis_instrumentation__.Notify(26871)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(26872)
			log.Infof(ctx, "onCheckpoint: %s@%s", checkpoint.Span, checkpoint.ResolvedTS)
		} else {
			__antithesis_instrumentation__.Notify(26873)
		}
	}
}

func (s *eventStream) onSpanCompleted(ctx context.Context, sp roachpb.Span) error {
	__antithesis_instrumentation__.Notify(26874)
	checkpoint := roachpb.RangeFeedCheckpoint{
		Span:       sp,
		ResolvedTS: s.spec.StartFrom,
	}
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(26875)
		return ctx.Err()
	case s.eventsCh <- roachpb.RangeFeedEvent{Checkpoint: &checkpoint}:
		__antithesis_instrumentation__.Notify(26876)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(26878)
			log.Infof(ctx, "onSpanCompleted: %s@%s", checkpoint.Span, checkpoint.ResolvedTS)
		} else {
			__antithesis_instrumentation__.Notify(26879)
		}
		__antithesis_instrumentation__.Notify(26877)
		return nil
	}
}

func makeCheckpoint(f *span.Frontier) (checkpoint streampb.StreamEvent_StreamCheckpoint) {
	__antithesis_instrumentation__.Notify(26880)
	f.Entries(func(sp roachpb.Span, ts hlc.Timestamp) (done span.OpResult) {
		__antithesis_instrumentation__.Notify(26882)
		checkpoint.Spans = append(checkpoint.Spans, streampb.StreamEvent_SpanCheckpoint{
			Span:      sp,
			Timestamp: ts,
		})
		return span.ContinueMatch
	})
	__antithesis_instrumentation__.Notify(26881)
	return
}

func (s *eventStream) flushEvent(ctx context.Context, event *streampb.StreamEvent) error {
	__antithesis_instrumentation__.Notify(26883)
	data, err := protoutil.Marshal(event)
	if err != nil {
		__antithesis_instrumentation__.Notify(26885)
		return err
	} else {
		__antithesis_instrumentation__.Notify(26886)
	}
	__antithesis_instrumentation__.Notify(26884)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(26887)
		return ctx.Err()
	case s.streamCh <- tree.Datums{tree.NewDBytes(tree.DBytes(data))}:
		__antithesis_instrumentation__.Notify(26888)
		return nil
	}
}

type checkpointPacer struct {
	pace    time.Duration
	next    time.Time
	skipped bool
}

func makeCheckpointPacer(frequency time.Duration) checkpointPacer {
	__antithesis_instrumentation__.Notify(26889)
	return checkpointPacer{
		pace:    frequency,
		next:    timeutil.Now().Add(frequency),
		skipped: false,
	}
}

func (p *checkpointPacer) shouldCheckpoint(
	currentFrontier hlc.Timestamp, frontierAdvanced bool,
) bool {
	__antithesis_instrumentation__.Notify(26890)
	now := timeutil.Now()
	enoughTimeElapsed := p.next.Before(now)

	if p.skipped {
		__antithesis_instrumentation__.Notify(26893)
		if enoughTimeElapsed {
			__antithesis_instrumentation__.Notify(26895)
			p.skipped = false
			p.next = now.Add(p.pace)
			return true
		} else {
			__antithesis_instrumentation__.Notify(26896)
		}
		__antithesis_instrumentation__.Notify(26894)
		return false
	} else {
		__antithesis_instrumentation__.Notify(26897)
	}
	__antithesis_instrumentation__.Notify(26891)

	isInitialScanCheckpoint := currentFrontier.IsEmpty()

	if frontierAdvanced || func() bool {
		__antithesis_instrumentation__.Notify(26898)
		return isInitialScanCheckpoint == true
	}() == true {
		__antithesis_instrumentation__.Notify(26899)
		if enoughTimeElapsed {
			__antithesis_instrumentation__.Notify(26901)
			p.next = now.Add(p.pace)
			return true
		} else {
			__antithesis_instrumentation__.Notify(26902)
		}
		__antithesis_instrumentation__.Notify(26900)
		p.skipped = true
		return false
	} else {
		__antithesis_instrumentation__.Notify(26903)
	}
	__antithesis_instrumentation__.Notify(26892)
	return false
}

func (s *eventStream) streamLoop(ctx context.Context, frontier *span.Frontier) error {
	__antithesis_instrumentation__.Notify(26904)
	pacer := makeCheckpointPacer(s.spec.Config.MinCheckpointFrequency)

	var batch streampb.StreamEvent_Batch
	batchSize := 0
	addValue := func(v *roachpb.RangeFeedValue) {
		__antithesis_instrumentation__.Notify(26907)
		keyValue := roachpb.KeyValue{
			Key:   v.Key,
			Value: v.Value,
		}
		batch.KeyValues = append(batch.KeyValues, keyValue)
		batchSize += keyValue.Size()
	}
	__antithesis_instrumentation__.Notify(26905)

	maybeFlushBatch := func(force bool) error {
		__antithesis_instrumentation__.Notify(26908)
		if (force && func() bool {
			__antithesis_instrumentation__.Notify(26910)
			return batchSize > 0 == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(26911)
			return batchSize > int(s.spec.Config.BatchByteSize) == true
		}() == true {
			__antithesis_instrumentation__.Notify(26912)
			defer func() {
				__antithesis_instrumentation__.Notify(26914)
				batchSize = 0
				batch.KeyValues = batch.KeyValues[:0]
			}()
			__antithesis_instrumentation__.Notify(26913)
			return s.flushEvent(ctx, &streampb.StreamEvent{Batch: &batch})
		} else {
			__antithesis_instrumentation__.Notify(26915)
		}
		__antithesis_instrumentation__.Notify(26909)
		return nil
	}
	__antithesis_instrumentation__.Notify(26906)

	const forceFlush = true
	const flushIfNeeded = false

	for {
		__antithesis_instrumentation__.Notify(26916)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(26917)
			return ctx.Err()
		case ev := <-s.eventsCh:
			__antithesis_instrumentation__.Notify(26918)
			switch {
			case ev.Val != nil:
				__antithesis_instrumentation__.Notify(26919)
				addValue(ev.Val)
				if err := maybeFlushBatch(flushIfNeeded); err != nil {
					__antithesis_instrumentation__.Notify(26923)
					return err
				} else {
					__antithesis_instrumentation__.Notify(26924)
				}
			case ev.Checkpoint != nil:
				__antithesis_instrumentation__.Notify(26920)
				advanced, err := frontier.Forward(ev.Checkpoint.Span, ev.Checkpoint.ResolvedTS)
				if err != nil {
					__antithesis_instrumentation__.Notify(26925)
					return err
				} else {
					__antithesis_instrumentation__.Notify(26926)
				}
				__antithesis_instrumentation__.Notify(26921)

				if pacer.shouldCheckpoint(frontier.Frontier(), advanced) {
					__antithesis_instrumentation__.Notify(26927)
					if err := maybeFlushBatch(forceFlush); err != nil {
						__antithesis_instrumentation__.Notify(26929)
						return err
					} else {
						__antithesis_instrumentation__.Notify(26930)
					}
					__antithesis_instrumentation__.Notify(26928)
					checkpoint := makeCheckpoint(frontier)
					if err := s.flushEvent(ctx, &streampb.StreamEvent{Checkpoint: &checkpoint}); err != nil {
						__antithesis_instrumentation__.Notify(26931)
						return err
					} else {
						__antithesis_instrumentation__.Notify(26932)
					}
				} else {
					__antithesis_instrumentation__.Notify(26933)
				}
			default:
				__antithesis_instrumentation__.Notify(26922)

				return errors.AssertionFailedf("unexpected event")
			}
		}
	}
}

func setConfigDefaults(cfg *streampb.StreamPartitionSpec_ExecutionConfig) {
	__antithesis_instrumentation__.Notify(26934)
	const defaultInitialScanParallelism = 16
	const defaultMinCheckpointFrequency = 10 * time.Second
	const defaultBatchSize = 1 << 20

	if cfg.InitialScanParallelism <= 0 {
		__antithesis_instrumentation__.Notify(26937)
		cfg.InitialScanParallelism = defaultInitialScanParallelism
	} else {
		__antithesis_instrumentation__.Notify(26938)
	}
	__antithesis_instrumentation__.Notify(26935)

	if cfg.MinCheckpointFrequency <= 0 {
		__antithesis_instrumentation__.Notify(26939)
		cfg.MinCheckpointFrequency = defaultMinCheckpointFrequency
	} else {
		__antithesis_instrumentation__.Notify(26940)
	}
	__antithesis_instrumentation__.Notify(26936)

	if cfg.BatchByteSize <= 0 {
		__antithesis_instrumentation__.Notify(26941)
		cfg.BatchByteSize = defaultBatchSize
	} else {
		__antithesis_instrumentation__.Notify(26942)
	}
}

func streamPartition(
	evalCtx *tree.EvalContext, streamID streaming.StreamID, opaqueSpec []byte,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(26943)
	if !evalCtx.SessionData().AvoidBuffering {
		__antithesis_instrumentation__.Notify(26947)
		return nil, errors.New("partition streaming requires 'SET avoid_buffering = true' option")
	} else {
		__antithesis_instrumentation__.Notify(26948)
	}
	__antithesis_instrumentation__.Notify(26944)

	var spec streampb.StreamPartitionSpec
	if err := protoutil.Unmarshal(opaqueSpec, &spec); err != nil {
		__antithesis_instrumentation__.Notify(26949)
		return nil, errors.Wrapf(err, "invalid partition spec for stream %d", streamID)
	} else {
		__antithesis_instrumentation__.Notify(26950)
	}
	__antithesis_instrumentation__.Notify(26945)

	if len(spec.Spans) == 0 {
		__antithesis_instrumentation__.Notify(26951)
		return nil, errors.AssertionFailedf("expected at least one span, got none")
	} else {
		__antithesis_instrumentation__.Notify(26952)
	}
	__antithesis_instrumentation__.Notify(26946)

	setConfigDefaults(&spec.Config)

	execCfg := evalCtx.Planner.ExecutorConfig().(*sql.ExecutorConfig)

	return &eventStream{
		streamID: streamID,
		spec:     spec,
		execCfg:  execCfg,
		mon:      evalCtx.Mon,
	}, nil
}
