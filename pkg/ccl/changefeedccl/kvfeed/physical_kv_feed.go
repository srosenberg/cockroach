package kvfeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type physicalFeedFactory interface {
	Run(ctx context.Context, sink kvevent.Writer, cfg physicalConfig) error
}

type physicalConfig struct {
	Spans     []roachpb.Span
	Timestamp hlc.Timestamp
	WithDiff  bool
	Knobs     TestingKnobs
}

type rangefeedFactory func(
	ctx context.Context,
	spans []roachpb.Span,
	startFrom hlc.Timestamp,
	withDiff bool,
	eventC chan<- *roachpb.RangeFeedEvent,
) error

type rangefeed struct {
	memBuf kvevent.Writer
	cfg    physicalConfig
	eventC chan *roachpb.RangeFeedEvent
}

func (p rangefeedFactory) Run(ctx context.Context, sink kvevent.Writer, cfg physicalConfig) error {
	__antithesis_instrumentation__.Notify(17346)

	feed := rangefeed{
		memBuf: sink,
		cfg:    cfg,
		eventC: make(chan *roachpb.RangeFeedEvent, 128),
	}
	g := ctxgroup.WithContext(ctx)
	g.GoCtx(feed.addEventsToBuffer)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(17348)
		return p(ctx, cfg.Spans, cfg.Timestamp, cfg.WithDiff, feed.eventC)
	})
	__antithesis_instrumentation__.Notify(17347)
	return g.Wait()
}

func (p *rangefeed) addEventsToBuffer(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17349)
	var backfillTimestamp hlc.Timestamp
	for {
		__antithesis_instrumentation__.Notify(17350)
		select {
		case e := <-p.eventC:
			__antithesis_instrumentation__.Notify(17351)
			switch t := e.GetValue().(type) {
			case *roachpb.RangeFeedValue:
				__antithesis_instrumentation__.Notify(17353)
				kv := roachpb.KeyValue{Key: t.Key, Value: t.Value}
				if p.cfg.Knobs.OnRangeFeedValue != nil {
					__antithesis_instrumentation__.Notify(17360)
					if err := p.cfg.Knobs.OnRangeFeedValue(kv); err != nil {
						__antithesis_instrumentation__.Notify(17361)
						return err
					} else {
						__antithesis_instrumentation__.Notify(17362)
					}
				} else {
					__antithesis_instrumentation__.Notify(17363)
				}
				__antithesis_instrumentation__.Notify(17354)
				var prevVal roachpb.Value
				if p.cfg.WithDiff {
					__antithesis_instrumentation__.Notify(17364)
					prevVal = t.PrevValue
				} else {
					__antithesis_instrumentation__.Notify(17365)
				}
				__antithesis_instrumentation__.Notify(17355)
				if err := p.memBuf.Add(
					ctx,
					kvevent.MakeKVEvent(kv, prevVal, backfillTimestamp),
				); err != nil {
					__antithesis_instrumentation__.Notify(17366)
					return err
				} else {
					__antithesis_instrumentation__.Notify(17367)
				}
			case *roachpb.RangeFeedCheckpoint:
				__antithesis_instrumentation__.Notify(17356)
				if !t.ResolvedTS.IsEmpty() && func() bool {
					__antithesis_instrumentation__.Notify(17368)
					return t.ResolvedTS.Less(p.cfg.Timestamp) == true
				}() == true {
					__antithesis_instrumentation__.Notify(17369)

					continue
				} else {
					__antithesis_instrumentation__.Notify(17370)
				}
				__antithesis_instrumentation__.Notify(17357)
				if err := p.memBuf.Add(
					ctx,
					kvevent.MakeResolvedEvent(t.Span, t.ResolvedTS, jobspb.ResolvedSpan_NONE),
				); err != nil {
					__antithesis_instrumentation__.Notify(17371)
					return err
				} else {
					__antithesis_instrumentation__.Notify(17372)
				}
			case *roachpb.RangeFeedSSTable:
				__antithesis_instrumentation__.Notify(17358)

				return errors.Errorf("unexpected SST ingestion: %v", t)

			default:
				__antithesis_instrumentation__.Notify(17359)
				return errors.Errorf("unexpected RangeFeedEvent variant %v", t)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(17352)
			return ctx.Err()
		}
	}
}
