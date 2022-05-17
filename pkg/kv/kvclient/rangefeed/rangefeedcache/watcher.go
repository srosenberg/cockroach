package rangefeedcache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedbuffer"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type Watcher struct {
	name             redact.SafeString
	clock            *hlc.Clock
	rangefeedFactory *rangefeed.Factory
	spans            []roachpb.Span
	bufferSize       int
	withPrevValue    bool

	started int32

	translateEvent TranslateEventFunc
	onUpdate       OnUpdateFunc

	lastFrontierTS hlc.Timestamp

	knobs TestingKnobs
}

type UpdateType bool

const (
	CompleteUpdate UpdateType = true

	IncrementalUpdate UpdateType = false
)

type TranslateEventFunc func(
	context.Context, *roachpb.RangeFeedValue,
) rangefeedbuffer.Event

type OnUpdateFunc func(context.Context, Update)

type Update struct {
	Type      UpdateType
	Timestamp hlc.Timestamp
	Events    []rangefeedbuffer.Event
}

type TestingKnobs struct {
	PostRangeFeedStart func()

	PreExit func()

	OnTimestampAdvance func(timestamp hlc.Timestamp)

	ErrorInjectionCh chan error
}

func (k *TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(89902) }

var _ base.ModuleTestingKnobs = (*TestingKnobs)(nil)

func NewWatcher(
	name redact.SafeString,
	clock *hlc.Clock,
	rangeFeedFactory *rangefeed.Factory,
	bufferSize int,
	spans []roachpb.Span,
	withPrevValue bool,
	translateEvent TranslateEventFunc,
	onUpdate OnUpdateFunc,
	knobs *TestingKnobs,
) *Watcher {
	__antithesis_instrumentation__.Notify(89903)
	w := &Watcher{
		name:             name,
		clock:            clock,
		rangefeedFactory: rangeFeedFactory,
		spans:            spans,
		bufferSize:       bufferSize,
		withPrevValue:    withPrevValue,
		translateEvent:   translateEvent,
		onUpdate:         onUpdate,
	}
	if knobs != nil {
		__antithesis_instrumentation__.Notify(89905)
		w.knobs = *knobs
	} else {
		__antithesis_instrumentation__.Notify(89906)
	}
	__antithesis_instrumentation__.Notify(89904)
	return w
}

func Start(ctx context.Context, stopper *stop.Stopper, c *Watcher, onError func(error)) error {
	__antithesis_instrumentation__.Notify(89907)
	return stopper.RunAsyncTask(ctx, string(c.name), func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(89908)
		ctx, cancel := stopper.WithCancelOnQuiesce(ctx)
		defer cancel()

		const aWhile = 5 * time.Minute
		for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
			__antithesis_instrumentation__.Notify(89909)
			started := timeutil.Now()
			if err := c.Run(ctx); err != nil {
				__antithesis_instrumentation__.Notify(89911)
				if errors.Is(err, context.Canceled) {
					__antithesis_instrumentation__.Notify(89915)
					return
				} else {
					__antithesis_instrumentation__.Notify(89916)
				}
				__antithesis_instrumentation__.Notify(89912)
				if onError != nil {
					__antithesis_instrumentation__.Notify(89917)
					onError(err)
				} else {
					__antithesis_instrumentation__.Notify(89918)
				}
				__antithesis_instrumentation__.Notify(89913)

				if timeutil.Since(started) > aWhile {
					__antithesis_instrumentation__.Notify(89919)
					r.Reset()
				} else {
					__antithesis_instrumentation__.Notify(89920)
				}
				__antithesis_instrumentation__.Notify(89914)

				log.Warningf(ctx, "%s: failed with %v, retrying...", c.name, err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(89921)
			}
			__antithesis_instrumentation__.Notify(89910)

			return
		}
	})
}

func (s *Watcher) Run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(89922)
	if !atomic.CompareAndSwapInt32(&s.started, 0, 1) {
		__antithesis_instrumentation__.Notify(89932)
		log.Fatal(ctx, "currently started: only allowed once at any point in time")
	} else {
		__antithesis_instrumentation__.Notify(89933)
	}
	__antithesis_instrumentation__.Notify(89923)
	if fn := s.knobs.PreExit; fn != nil {
		__antithesis_instrumentation__.Notify(89934)
		defer fn()
	} else {
		__antithesis_instrumentation__.Notify(89935)
	}
	__antithesis_instrumentation__.Notify(89924)
	defer func() { __antithesis_instrumentation__.Notify(89936); atomic.StoreInt32(&s.started, 0) }()
	__antithesis_instrumentation__.Notify(89925)

	buffer := rangefeedbuffer.New(math.MaxInt)
	frontierBumpedCh, initialScanDoneCh, errCh := make(chan struct{}), make(chan struct{}), make(chan error)
	mu := struct {
		syncutil.Mutex
		frontierTS hlc.Timestamp
	}{}

	defer func() {
		__antithesis_instrumentation__.Notify(89937)
		mu.Lock()
		s.lastFrontierTS.Forward(mu.frontierTS)
		mu.Unlock()
	}()
	__antithesis_instrumentation__.Notify(89926)

	onValue := func(ctx context.Context, ev *roachpb.RangeFeedValue) {
		__antithesis_instrumentation__.Notify(89938)
		bEv := s.translateEvent(ctx, ev)
		if bEv == nil {
			__antithesis_instrumentation__.Notify(89940)
			return
		} else {
			__antithesis_instrumentation__.Notify(89941)
		}
		__antithesis_instrumentation__.Notify(89939)

		if err := buffer.Add(bEv); err != nil {
			__antithesis_instrumentation__.Notify(89942)
			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(89943)

			case errCh <- err:
				__antithesis_instrumentation__.Notify(89944)
			}
		} else {
			__antithesis_instrumentation__.Notify(89945)
		}
	}
	__antithesis_instrumentation__.Notify(89927)

	initialScanTS := s.clock.Now()
	if initialScanTS.Less(s.lastFrontierTS) {
		__antithesis_instrumentation__.Notify(89946)
		log.Fatalf(ctx, "%s: initial scan timestamp (%s) regressed from last recorded frontier (%s)", s.name, initialScanTS, s.lastFrontierTS)
	} else {
		__antithesis_instrumentation__.Notify(89947)
	}
	__antithesis_instrumentation__.Notify(89928)

	rangeFeed := s.rangefeedFactory.New(string(s.name), initialScanTS,
		onValue,
		rangefeed.WithInitialScan(func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(89948)
			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(89949)

			case initialScanDoneCh <- struct{}{}:
				__antithesis_instrumentation__.Notify(89950)
			}
		}),
		rangefeed.WithOnFrontierAdvance(func(ctx context.Context, frontierTS hlc.Timestamp) {
			__antithesis_instrumentation__.Notify(89951)
			mu.Lock()
			mu.frontierTS = frontierTS
			mu.Unlock()

			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(89952)
			case frontierBumpedCh <- struct{}{}:
				__antithesis_instrumentation__.Notify(89953)
			}
		}),
		rangefeed.WithDiff(s.withPrevValue),
		rangefeed.WithRowTimestampInInitialScan(true),
		rangefeed.WithOnInitialScanError(func(ctx context.Context, err error) (shouldFail bool) {
			__antithesis_instrumentation__.Notify(89954)

			if grpcutil.IsAuthError(err) || func() bool {
				__antithesis_instrumentation__.Notify(89956)
				return strings.Contains(err.Error(), "rpc error: code = Unauthenticated") == true
			}() == true {
				__antithesis_instrumentation__.Notify(89957)
				return true
			} else {
				__antithesis_instrumentation__.Notify(89958)
			}
			__antithesis_instrumentation__.Notify(89955)
			return false
		}),
	)
	__antithesis_instrumentation__.Notify(89929)
	if err := rangeFeed.Start(ctx, s.spans); err != nil {
		__antithesis_instrumentation__.Notify(89959)
		return err
	} else {
		__antithesis_instrumentation__.Notify(89960)
	}
	__antithesis_instrumentation__.Notify(89930)
	defer rangeFeed.Close()
	if fn := s.knobs.PostRangeFeedStart; fn != nil {
		__antithesis_instrumentation__.Notify(89961)
		fn()
	} else {
		__antithesis_instrumentation__.Notify(89962)
	}
	__antithesis_instrumentation__.Notify(89931)

	log.Infof(ctx, "%s: established range feed cache", s.name)

	for {
		__antithesis_instrumentation__.Notify(89963)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(89964)
			return ctx.Err()

		case <-frontierBumpedCh:
			__antithesis_instrumentation__.Notify(89965)
			mu.Lock()
			frontierTS := mu.frontierTS
			mu.Unlock()
			s.handleUpdate(ctx, buffer, frontierTS, IncrementalUpdate)

		case <-initialScanDoneCh:
			__antithesis_instrumentation__.Notify(89966)
			s.handleUpdate(ctx, buffer, initialScanTS, CompleteUpdate)

			buffer.SetLimit(s.bufferSize)

		case err := <-errCh:
			__antithesis_instrumentation__.Notify(89967)
			return err
		case err := <-s.knobs.ErrorInjectionCh:
			__antithesis_instrumentation__.Notify(89968)
			return err
		}
	}
}

func (s *Watcher) handleUpdate(
	ctx context.Context, buffer *rangefeedbuffer.Buffer, ts hlc.Timestamp, updateType UpdateType,
) {
	__antithesis_instrumentation__.Notify(89969)
	s.onUpdate(ctx, Update{
		Type:      updateType,
		Timestamp: ts,
		Events:    buffer.Flush(ctx, ts),
	})
	if fn := s.knobs.OnTimestampAdvance; fn != nil {
		__antithesis_instrumentation__.Notify(89970)
		fn(ts)
	} else {
		__antithesis_instrumentation__.Notify(89971)
	}
}
