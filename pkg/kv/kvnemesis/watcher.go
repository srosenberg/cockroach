package kvnemesis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"reflect"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/span"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type Watcher struct {
	env *Env
	mu  struct {
		syncutil.Mutex
		kvs             *Engine
		frontier        *span.Frontier
		frontierWaiters map[hlc.Timestamp][]chan error
	}
	cancel func()
	g      ctxgroup.Group
}

func Watch(ctx context.Context, env *Env, dbs []*kv.DB, dataSpan roachpb.Span) (*Watcher, error) {
	__antithesis_instrumentation__.Notify(93871)
	if len(dbs) < 1 {
		__antithesis_instrumentation__.Notify(93879)
		return nil, errors.New(`at least one db must be given`)
	} else {
		__antithesis_instrumentation__.Notify(93880)
	}
	__antithesis_instrumentation__.Notify(93872)
	firstDB := dbs[0]

	w := &Watcher{
		env: env,
	}
	var err error
	if w.mu.kvs, err = MakeEngine(); err != nil {
		__antithesis_instrumentation__.Notify(93881)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(93882)
	}
	__antithesis_instrumentation__.Notify(93873)
	w.mu.frontier, err = span.MakeFrontier(dataSpan)
	if err != nil {
		__antithesis_instrumentation__.Notify(93883)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(93884)
	}
	__antithesis_instrumentation__.Notify(93874)
	w.mu.frontierWaiters = make(map[hlc.Timestamp][]chan error)
	ctx, w.cancel = context.WithCancel(ctx)
	w.g = ctxgroup.WithContext(ctx)

	dss := make([]*kvcoord.DistSender, len(dbs))
	for i := range dbs {
		__antithesis_instrumentation__.Notify(93885)
		sender := dbs[i].NonTransactionalSender()
		dss[i] = sender.(*kv.CrossRangeTxnWrapperSender).Wrapped().(*kvcoord.DistSender)
	}
	__antithesis_instrumentation__.Notify(93875)

	startTs := firstDB.Clock().Now()
	eventC := make(chan *roachpb.RangeFeedEvent, 128)
	w.g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(93886)
		ts := startTs
		for i := 0; ; i = (i + 1) % len(dbs) {
			__antithesis_instrumentation__.Notify(93887)
			w.mu.Lock()
			ts.Forward(w.mu.frontier.Frontier())
			w.mu.Unlock()

			ds := dss[i]
			err := ds.RangeFeed(ctx, []roachpb.Span{dataSpan}, ts, true, eventC)
			if isRetryableRangeFeedErr(err) {
				__antithesis_instrumentation__.Notify(93889)
				log.Infof(ctx, "got retryable RangeFeed error: %+v", err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(93890)
			}
			__antithesis_instrumentation__.Notify(93888)
			return err
		}
	})
	__antithesis_instrumentation__.Notify(93876)
	w.g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(93891)
		return w.processEvents(ctx, eventC)
	})
	__antithesis_instrumentation__.Notify(93877)

	if err := w.WaitForFrontier(ctx, startTs); err != nil {
		__antithesis_instrumentation__.Notify(93892)
		_ = w.Finish()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(93893)
	}
	__antithesis_instrumentation__.Notify(93878)

	return w, nil
}

func isRetryableRangeFeedErr(err error) bool {
	__antithesis_instrumentation__.Notify(93894)
	switch {
	case errors.Is(err, context.Canceled):
		__antithesis_instrumentation__.Notify(93895)
		return false
	default:
		__antithesis_instrumentation__.Notify(93896)
		return true
	}
}

func (w *Watcher) Finish() *Engine {
	__antithesis_instrumentation__.Notify(93897)
	if w.cancel == nil {
		__antithesis_instrumentation__.Notify(93899)

		return w.mu.kvs
	} else {
		__antithesis_instrumentation__.Notify(93900)
	}
	__antithesis_instrumentation__.Notify(93898)
	w.cancel()
	w.cancel = nil

	_ = w.g.Wait()
	return w.mu.kvs
}

func (w *Watcher) WaitForFrontier(ctx context.Context, ts hlc.Timestamp) (retErr error) {
	__antithesis_instrumentation__.Notify(93901)
	log.Infof(ctx, `watcher waiting for %s`, ts)
	if err := w.env.SetClosedTimestampInterval(ctx, 1*time.Millisecond); err != nil {
		__antithesis_instrumentation__.Notify(93904)
		return err
	} else {
		__antithesis_instrumentation__.Notify(93905)
	}
	__antithesis_instrumentation__.Notify(93902)
	defer func() {
		__antithesis_instrumentation__.Notify(93906)
		if err := w.env.ResetClosedTimestampInterval(ctx); err != nil {
			__antithesis_instrumentation__.Notify(93907)
			retErr = errors.WithSecondaryError(retErr, err)
		} else {
			__antithesis_instrumentation__.Notify(93908)
		}
	}()
	__antithesis_instrumentation__.Notify(93903)
	resultCh := make(chan error, 1)
	w.mu.Lock()
	w.mu.frontierWaiters[ts] = append(w.mu.frontierWaiters[ts], resultCh)
	w.mu.Unlock()
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(93909)
		return ctx.Err()
	case err := <-resultCh:
		__antithesis_instrumentation__.Notify(93910)
		return err
	}
}

func (w *Watcher) processEvents(ctx context.Context, eventC chan *roachpb.RangeFeedEvent) error {
	__antithesis_instrumentation__.Notify(93911)
	for {
		__antithesis_instrumentation__.Notify(93912)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(93913)
			return nil
		case event := <-eventC:
			__antithesis_instrumentation__.Notify(93914)
			switch e := event.GetValue().(type) {
			case *roachpb.RangeFeedError:
				__antithesis_instrumentation__.Notify(93915)
				return e.Error.GoError()
			case *roachpb.RangeFeedValue:
				__antithesis_instrumentation__.Notify(93916)
				log.Infof(ctx, `rangefeed Put %s %s -> %s (prev %s)`,
					e.Key, e.Value.Timestamp, e.Value.PrettyPrint(), e.PrevValue.PrettyPrint())
				w.mu.Lock()

				if len(e.Value.RawBytes) == 0 {
					__antithesis_instrumentation__.Notify(93923)
					w.mu.kvs.Delete(storage.MVCCKey{Key: e.Key, Timestamp: e.Value.Timestamp})
				} else {
					__antithesis_instrumentation__.Notify(93924)
					w.mu.kvs.Put(storage.MVCCKey{Key: e.Key, Timestamp: e.Value.Timestamp}, e.Value.RawBytes)
				}
				__antithesis_instrumentation__.Notify(93917)
				prevTs := e.Value.Timestamp.Prev()
				prevValue := w.mu.kvs.Get(e.Key, prevTs)

				prevValue.Timestamp = hlc.Timestamp{}

				if len(e.PrevValue.RawBytes) == 0 {
					__antithesis_instrumentation__.Notify(93925)
					e.PrevValue.RawBytes = nil
				} else {
					__antithesis_instrumentation__.Notify(93926)
				}
				__antithesis_instrumentation__.Notify(93918)
				prevValueMismatch := !reflect.DeepEqual(prevValue, e.PrevValue)
				var engineContents string
				if prevValueMismatch {
					__antithesis_instrumentation__.Notify(93927)
					engineContents = w.mu.kvs.DebugPrint("  ")
				} else {
					__antithesis_instrumentation__.Notify(93928)
				}
				__antithesis_instrumentation__.Notify(93919)
				w.mu.Unlock()

				if prevValueMismatch {
					__antithesis_instrumentation__.Notify(93929)
					log.Infof(ctx, "rangefeed mismatch\n%s", engineContents)
					panic(errors.Errorf(
						`expected (%s, %s) previous value %s got: %s`, e.Key, prevTs, prevValue, e.PrevValue))
				} else {
					__antithesis_instrumentation__.Notify(93930)
				}
			case *roachpb.RangeFeedCheckpoint:
				__antithesis_instrumentation__.Notify(93920)
				w.mu.Lock()
				frontierAdvanced, err := w.mu.frontier.Forward(e.Span, e.ResolvedTS)
				if err != nil {
					__antithesis_instrumentation__.Notify(93931)
					panic(errors.Wrapf(err, "unexpected frontier error advancing to %s@%s", e.Span, e.ResolvedTS))
				} else {
					__antithesis_instrumentation__.Notify(93932)
				}
				__antithesis_instrumentation__.Notify(93921)
				if frontierAdvanced {
					__antithesis_instrumentation__.Notify(93933)
					frontier := w.mu.frontier.Frontier()
					log.Infof(ctx, `watcher reached frontier %s lagging by %s`,
						frontier, timeutil.Since(frontier.GoTime()))
					for ts, chs := range w.mu.frontierWaiters {
						__antithesis_instrumentation__.Notify(93934)
						if frontier.Less(ts) {
							__antithesis_instrumentation__.Notify(93936)
							continue
						} else {
							__antithesis_instrumentation__.Notify(93937)
						}
						__antithesis_instrumentation__.Notify(93935)
						log.Infof(ctx, `watcher notifying %s`, ts)
						delete(w.mu.frontierWaiters, ts)
						for _, ch := range chs {
							__antithesis_instrumentation__.Notify(93938)
							ch <- nil
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(93939)
				}
				__antithesis_instrumentation__.Notify(93922)
				w.mu.Unlock()
			}
		}
	}
}
