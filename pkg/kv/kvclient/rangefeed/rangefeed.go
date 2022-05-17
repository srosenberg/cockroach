package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/span"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
)

type DB interface {
	RangeFeed(
		ctx context.Context,
		spans []roachpb.Span,
		startFrom hlc.Timestamp,
		withDiff bool,
		eventC chan<- *roachpb.RangeFeedEvent,
	) error

	Scan(
		ctx context.Context,
		spans []roachpb.Span,
		asOf hlc.Timestamp,
		rowFn func(value roachpb.KeyValue),
		cfg scanConfig,
	) error
}

type Factory struct {
	stopper *stop.Stopper
	client  DB
	knobs   *TestingKnobs
}

type TestingKnobs struct {
	OnRangefeedRestart func()
}

func (t TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(89764) }

var _ base.ModuleTestingKnobs = (*TestingKnobs)(nil)

func NewFactory(
	stopper *stop.Stopper, db *kv.DB, st *cluster.Settings, knobs *TestingKnobs,
) (*Factory, error) {
	__antithesis_instrumentation__.Notify(89765)
	kvDB, err := newDBAdapter(db, st)
	if err != nil {
		__antithesis_instrumentation__.Notify(89767)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(89768)
	}
	__antithesis_instrumentation__.Notify(89766)
	return newFactory(stopper, kvDB, knobs), nil
}

func newFactory(stopper *stop.Stopper, client DB, knobs *TestingKnobs) *Factory {
	__antithesis_instrumentation__.Notify(89769)
	return &Factory{
		stopper: stopper,
		client:  client,
		knobs:   knobs,
	}
}

func (f *Factory) RangeFeed(
	ctx context.Context,
	name string,
	spans []roachpb.Span,
	initialTimestamp hlc.Timestamp,
	onValue OnValue,
	options ...Option,
) (_ *RangeFeed, err error) {
	__antithesis_instrumentation__.Notify(89770)
	r := f.New(name, initialTimestamp, onValue, options...)
	if err := r.Start(ctx, spans); err != nil {
		__antithesis_instrumentation__.Notify(89772)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(89773)
	}
	__antithesis_instrumentation__.Notify(89771)
	return r, nil
}

func (f *Factory) New(
	name string, initialTimestamp hlc.Timestamp, onValue OnValue, options ...Option,
) *RangeFeed {
	__antithesis_instrumentation__.Notify(89774)
	r := RangeFeed{
		client:  f.client,
		stopper: f.stopper,
		knobs:   f.knobs,

		initialTimestamp: initialTimestamp,
		name:             name,
		onValue:          onValue,

		stopped: make(chan struct{}),
	}
	initConfig(&r.config, options)
	return &r
}

type OnValue func(ctx context.Context, value *roachpb.RangeFeedValue)

type RangeFeed struct {
	config
	name    string
	client  DB
	stopper *stop.Stopper
	knobs   *TestingKnobs

	initialTimestamp hlc.Timestamp
	spans            []roachpb.Span
	spansDebugStr    string

	onValue OnValue

	closeOnce sync.Once
	cancel    context.CancelFunc
	stopped   chan struct{}

	started int32
}

func (f *RangeFeed) Start(ctx context.Context, spans []roachpb.Span) error {
	__antithesis_instrumentation__.Notify(89775)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(89784)
		return errors.AssertionFailedf("expected at least 1 span, got none")
	} else {
		__antithesis_instrumentation__.Notify(89785)
	}
	__antithesis_instrumentation__.Notify(89776)

	if !atomic.CompareAndSwapInt32(&f.started, 0, 1) {
		__antithesis_instrumentation__.Notify(89786)
		return errors.AssertionFailedf("rangefeed already started")
	} else {
		__antithesis_instrumentation__.Notify(89787)
	}
	__antithesis_instrumentation__.Notify(89777)

	frontier, err := span.MakeFrontier(spans...)
	if err != nil {
		__antithesis_instrumentation__.Notify(89788)
		return err
	} else {
		__antithesis_instrumentation__.Notify(89789)
	}
	__antithesis_instrumentation__.Notify(89778)

	for _, sp := range spans {
		__antithesis_instrumentation__.Notify(89790)
		if _, err := frontier.Forward(sp, f.initialTimestamp); err != nil {
			__antithesis_instrumentation__.Notify(89791)
			return err
		} else {
			__antithesis_instrumentation__.Notify(89792)
		}
	}
	__antithesis_instrumentation__.Notify(89779)

	frontier.Entries(func(sp roachpb.Span, _ hlc.Timestamp) (done span.OpResult) {
		__antithesis_instrumentation__.Notify(89793)
		f.spans = append(f.spans, sp)
		return span.ContinueMatch
	})
	__antithesis_instrumentation__.Notify(89780)

	runWithFrontier := func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(89794)

		defer pprof.SetGoroutineLabels(ctx)
		ctx = pprof.WithLabels(ctx, pprof.Labels(append(f.extraPProfLabels, "rangefeed", f.name)...))
		pprof.SetGoroutineLabels(ctx)
		f.run(ctx, frontier)
	}
	__antithesis_instrumentation__.Notify(89781)

	f.spansDebugStr = func() string {
		__antithesis_instrumentation__.Notify(89795)
		n := len(spans)
		if n == 1 {
			__antithesis_instrumentation__.Notify(89797)
			return spans[0].String()
		} else {
			__antithesis_instrumentation__.Notify(89798)
		}
		__antithesis_instrumentation__.Notify(89796)

		return fmt.Sprintf("{%s}", frontier.String())
	}()
	__antithesis_instrumentation__.Notify(89782)

	ctx = logtags.AddTag(ctx, "rangefeed", f.name)
	ctx, f.cancel = f.stopper.WithCancelOnQuiesce(ctx)
	if err := f.stopper.RunAsyncTask(ctx, "rangefeed", runWithFrontier); err != nil {
		__antithesis_instrumentation__.Notify(89799)
		f.cancel()
		return err
	} else {
		__antithesis_instrumentation__.Notify(89800)
	}
	__antithesis_instrumentation__.Notify(89783)
	return nil
}

func (f *RangeFeed) Close() {
	__antithesis_instrumentation__.Notify(89801)
	f.closeOnce.Do(func() {
		__antithesis_instrumentation__.Notify(89802)
		f.cancel()
		<-f.stopped
	})
}

const resetThreshold = 30 * time.Second

func (f *RangeFeed) run(ctx context.Context, frontier *span.Frontier) {
	__antithesis_instrumentation__.Notify(89803)
	defer close(f.stopped)
	r := retry.StartWithCtx(ctx, f.retryOptions)
	restartLogEvery := log.Every(10 * time.Second)

	if f.withInitialScan {
		__antithesis_instrumentation__.Notify(89806)
		if done := f.runInitialScan(ctx, &restartLogEvery, &r); done {
			__antithesis_instrumentation__.Notify(89807)
			return
		} else {
			__antithesis_instrumentation__.Notify(89808)
		}
	} else {
		__antithesis_instrumentation__.Notify(89809)
	}
	__antithesis_instrumentation__.Notify(89804)

	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(89810)
		return
	} else {
		__antithesis_instrumentation__.Notify(89811)
	}
	__antithesis_instrumentation__.Notify(89805)

	eventCh := make(chan *roachpb.RangeFeedEvent)

	for i := 0; r.Next(); i++ {
		__antithesis_instrumentation__.Notify(89812)
		ts := frontier.Frontier()
		if log.ExpensiveLogEnabled(ctx, 1) {
			__antithesis_instrumentation__.Notify(89820)
			log.Eventf(ctx, "starting rangefeed from %v on %v", ts, f.spansDebugStr)
		} else {
			__antithesis_instrumentation__.Notify(89821)
		}
		__antithesis_instrumentation__.Notify(89813)

		start := timeutil.Now()

		rangeFeedTask := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(89822)
			return f.client.RangeFeed(ctx, f.spans, ts, f.withDiff, eventCh)
		}
		__antithesis_instrumentation__.Notify(89814)
		processEventsTask := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(89823)
			return f.processEvents(ctx, frontier, eventCh)
		}
		__antithesis_instrumentation__.Notify(89815)

		err := ctxgroup.GoAndWait(ctx, rangeFeedTask, processEventsTask)
		if errors.HasType(err, &roachpb.BatchTimestampBeforeGCError{}) || func() bool {
			__antithesis_instrumentation__.Notify(89824)
			return errors.HasType(err, &roachpb.MVCCHistoryMutationError{}) == true
		}() == true {
			__antithesis_instrumentation__.Notify(89825)
			if errCallback := f.onUnrecoverableError; errCallback != nil {
				__antithesis_instrumentation__.Notify(89827)
				errCallback(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(89828)
			}
			__antithesis_instrumentation__.Notify(89826)

			log.VEventf(ctx, 1, "exiting rangefeed due to internal error: %v", err)
			return
		} else {
			__antithesis_instrumentation__.Notify(89829)
		}
		__antithesis_instrumentation__.Notify(89816)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(89830)
			return ctx.Err() == nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(89831)
			return restartLogEvery.ShouldLog() == true
		}() == true {
			__antithesis_instrumentation__.Notify(89832)
			log.Warningf(ctx, "rangefeed failed %d times, restarting: %v",
				redact.Safe(i), err)
		} else {
			__antithesis_instrumentation__.Notify(89833)
		}
		__antithesis_instrumentation__.Notify(89817)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(89834)
			log.VEventf(ctx, 1, "exiting rangefeed")
			return
		} else {
			__antithesis_instrumentation__.Notify(89835)
		}
		__antithesis_instrumentation__.Notify(89818)

		ranFor := timeutil.Since(start)
		log.VEventf(ctx, 1, "restarting rangefeed for %v after %v",
			f.spansDebugStr, ranFor)
		if f.knobs != nil && func() bool {
			__antithesis_instrumentation__.Notify(89836)
			return f.knobs.OnRangefeedRestart != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(89837)
			f.knobs.OnRangefeedRestart()
		} else {
			__antithesis_instrumentation__.Notify(89838)
		}
		__antithesis_instrumentation__.Notify(89819)

		if ranFor > resetThreshold {
			__antithesis_instrumentation__.Notify(89839)
			i = 1
			r.Reset()
		} else {
			__antithesis_instrumentation__.Notify(89840)
		}
	}
}

func (f *RangeFeed) processEvents(
	ctx context.Context, frontier *span.Frontier, eventCh <-chan *roachpb.RangeFeedEvent,
) error {
	__antithesis_instrumentation__.Notify(89841)
	for {
		__antithesis_instrumentation__.Notify(89842)
		select {
		case ev := <-eventCh:
			__antithesis_instrumentation__.Notify(89843)
			switch {
			case ev.Val != nil:
				__antithesis_instrumentation__.Notify(89845)
				f.onValue(ctx, ev.Val)
			case ev.Checkpoint != nil:
				__antithesis_instrumentation__.Notify(89846)
				advanced, err := frontier.Forward(ev.Checkpoint.Span, ev.Checkpoint.ResolvedTS)
				if err != nil {
					__antithesis_instrumentation__.Notify(89853)
					return err
				} else {
					__antithesis_instrumentation__.Notify(89854)
				}
				__antithesis_instrumentation__.Notify(89847)
				if f.onCheckpoint != nil {
					__antithesis_instrumentation__.Notify(89855)
					f.onCheckpoint(ctx, ev.Checkpoint)
				} else {
					__antithesis_instrumentation__.Notify(89856)
				}
				__antithesis_instrumentation__.Notify(89848)
				if advanced && func() bool {
					__antithesis_instrumentation__.Notify(89857)
					return f.onFrontierAdvance != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(89858)
					f.onFrontierAdvance(ctx, frontier.Frontier())
				} else {
					__antithesis_instrumentation__.Notify(89859)
				}
			case ev.SST != nil:
				__antithesis_instrumentation__.Notify(89849)
				if f.onSSTable == nil {
					__antithesis_instrumentation__.Notify(89860)
					return errors.AssertionFailedf(
						"received unexpected rangefeed SST event with no OnSSTable handler")
				} else {
					__antithesis_instrumentation__.Notify(89861)
				}
				__antithesis_instrumentation__.Notify(89850)
				f.onSSTable(ctx, ev.SST)
			case ev.Error != nil:
				__antithesis_instrumentation__.Notify(89851)
			default:
				__antithesis_instrumentation__.Notify(89852)

			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(89844)
			return ctx.Err()
		}
	}
}
