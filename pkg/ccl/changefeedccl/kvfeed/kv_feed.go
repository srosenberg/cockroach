// Package kvfeed provides an abstraction to stream kvs to a buffer.
//
// The kvfeed coordinated performing logical backfills in the face of schema
// changes and then running rangefeeds.
package kvfeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/schemafeed"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/span"
	"github.com/cockroachdb/errors"
)

type Config struct {
	Settings                *cluster.Settings
	DB                      *kv.DB
	Codec                   keys.SQLCodec
	Clock                   *hlc.Clock
	Gossip                  gossip.OptionalGossip
	Spans                   []roachpb.Span
	BackfillCheckpoint      []roachpb.Span
	Targets                 []jobspb.ChangefeedTargetSpecification
	Writer                  kvevent.Writer
	Metrics                 *kvevent.Metrics
	OnBackfillCallback      func() func()
	OnBackfillRangeCallback func(int64) (func(), func())
	MM                      *mon.BytesMonitor
	WithDiff                bool
	SchemaChangeEvents      changefeedbase.SchemaChangeEventClass
	SchemaChangePolicy      changefeedbase.SchemaChangePolicy
	SchemaFeed              schemafeed.SchemaFeed

	NeedsInitialScan bool

	InitialHighWater hlc.Timestamp

	EndTime hlc.Timestamp

	Knobs TestingKnobs
}

func Run(ctx context.Context, cfg Config) error {
	__antithesis_instrumentation__.Notify(17155)

	var sc kvScanner
	{
		__antithesis_instrumentation__.Notify(17162)
		sc = &scanRequestScanner{
			settings:                cfg.Settings,
			gossip:                  cfg.Gossip,
			db:                      cfg.DB,
			onBackfillRangeCallback: cfg.OnBackfillRangeCallback,
		}
	}
	__antithesis_instrumentation__.Notify(17156)
	var pff physicalFeedFactory
	{
		__antithesis_instrumentation__.Notify(17163)
		sender := cfg.DB.NonTransactionalSender()
		distSender := sender.(*kv.CrossRangeTxnWrapperSender).Wrapped().(*kvcoord.DistSender)
		pff = rangefeedFactory(distSender.RangeFeed)
	}
	__antithesis_instrumentation__.Notify(17157)

	bf := func() kvevent.Buffer {
		__antithesis_instrumentation__.Notify(17164)
		return kvevent.NewErrorWrapperEventBuffer(
			kvevent.NewMemBuffer(cfg.MM.MakeBoundAccount(), &cfg.Settings.SV, cfg.Metrics))
	}
	__antithesis_instrumentation__.Notify(17158)

	f := newKVFeed(
		cfg.Writer, cfg.Spans, cfg.BackfillCheckpoint,
		cfg.SchemaChangeEvents, cfg.SchemaChangePolicy,
		cfg.NeedsInitialScan, cfg.WithDiff,
		cfg.InitialHighWater, cfg.EndTime,
		cfg.Codec,
		cfg.SchemaFeed,
		sc, pff, bf, cfg.Knobs)
	f.onBackfillCallback = cfg.OnBackfillCallback

	g := ctxgroup.WithContext(ctx)
	g.GoCtx(cfg.SchemaFeed.Run)
	g.GoCtx(f.run)
	err := g.Wait()

	var scErr schemaChangeDetectedError
	isChangefeedCompleted := errors.Is(err, errChangefeedCompleted)
	if !(isChangefeedCompleted || func() bool {
		__antithesis_instrumentation__.Notify(17165)
		return errors.As(err, &scErr) == true
	}() == true) {
		__antithesis_instrumentation__.Notify(17166)

		return errors.CombineErrors(err, f.writer.CloseWithReason(ctx, err))
	} else {
		__antithesis_instrumentation__.Notify(17167)
	}
	__antithesis_instrumentation__.Notify(17159)

	if isChangefeedCompleted {
		__antithesis_instrumentation__.Notify(17168)
		log.Info(ctx, "stopping kv feed: changefeed completed")
	} else {
		__antithesis_instrumentation__.Notify(17169)
		log.Infof(ctx, "stopping kv feed due to schema change at %v", scErr.ts)
	}
	__antithesis_instrumentation__.Notify(17160)

	err = errors.CombineErrors(f.writer.Drain(ctx), f.writer.CloseWithReason(ctx, kvevent.ErrNormalRestartReason))

	if err == nil {
		__antithesis_instrumentation__.Notify(17170)

		<-ctx.Done()
	} else {
		__antithesis_instrumentation__.Notify(17171)
	}
	__antithesis_instrumentation__.Notify(17161)

	return err
}

type schemaChangeDetectedError struct {
	ts hlc.Timestamp
}

func (e schemaChangeDetectedError) Error() string {
	__antithesis_instrumentation__.Notify(17172)
	return fmt.Sprintf("schema change detected at %v", e.ts)
}

type kvFeed struct {
	spans               []roachpb.Span
	checkpoint          []roachpb.Span
	withDiff            bool
	withInitialBackfill bool
	initialHighWater    hlc.Timestamp
	endTime             hlc.Timestamp
	writer              kvevent.Writer
	codec               keys.SQLCodec

	onBackfillCallback func() func()
	schemaChangeEvents changefeedbase.SchemaChangeEventClass
	schemaChangePolicy changefeedbase.SchemaChangePolicy

	bufferFactory func() kvevent.Buffer
	tableFeed     schemafeed.SchemaFeed
	scanner       kvScanner
	physicalFeed  physicalFeedFactory
	knobs         TestingKnobs
}

func newKVFeed(
	writer kvevent.Writer,
	spans []roachpb.Span,
	checkpoint []roachpb.Span,
	schemaChangeEvents changefeedbase.SchemaChangeEventClass,
	schemaChangePolicy changefeedbase.SchemaChangePolicy,
	withInitialBackfill, withDiff bool,
	initialHighWater hlc.Timestamp,
	endTime hlc.Timestamp,
	codec keys.SQLCodec,
	tf schemafeed.SchemaFeed,
	sc kvScanner,
	pff physicalFeedFactory,
	bf func() kvevent.Buffer,
	knobs TestingKnobs,
) *kvFeed {
	__antithesis_instrumentation__.Notify(17173)
	return &kvFeed{
		writer:              writer,
		spans:               spans,
		checkpoint:          checkpoint,
		withInitialBackfill: withInitialBackfill,
		withDiff:            withDiff,
		initialHighWater:    initialHighWater,
		endTime:             endTime,
		schemaChangeEvents:  schemaChangeEvents,
		schemaChangePolicy:  schemaChangePolicy,
		codec:               codec,
		tableFeed:           tf,
		scanner:             sc,
		physicalFeed:        pff,
		bufferFactory:       bf,
		knobs:               knobs,
	}
}

var errChangefeedCompleted = errors.New("changefeed completed")

func (f *kvFeed) run(ctx context.Context) (err error) {
	__antithesis_instrumentation__.Notify(17174)
	emitResolved := func(ts hlc.Timestamp, boundary jobspb.ResolvedSpan_BoundaryType) error {
		__antithesis_instrumentation__.Notify(17176)
		for _, sp := range f.spans {
			__antithesis_instrumentation__.Notify(17178)
			if err := f.writer.Add(ctx, kvevent.MakeResolvedEvent(sp, ts, boundary)); err != nil {
				__antithesis_instrumentation__.Notify(17179)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17180)
			}
		}
		__antithesis_instrumentation__.Notify(17177)
		return nil
	}
	__antithesis_instrumentation__.Notify(17175)

	highWater := f.initialHighWater
	for i := 0; ; i++ {
		__antithesis_instrumentation__.Notify(17181)
		initialScan := i == 0
		if err = f.scanIfShould(ctx, initialScan, highWater); err != nil {
			__antithesis_instrumentation__.Notify(17187)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17188)
		}
		__antithesis_instrumentation__.Notify(17182)

		initialScanOnly := f.endTime.EqOrdering(f.initialHighWater)
		if initialScanOnly {
			__antithesis_instrumentation__.Notify(17189)
			if err := emitResolved(f.initialHighWater, jobspb.ResolvedSpan_EXIT); err != nil {
				__antithesis_instrumentation__.Notify(17191)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17192)
			}
			__antithesis_instrumentation__.Notify(17190)
			return errChangefeedCompleted
		} else {
			__antithesis_instrumentation__.Notify(17193)
		}
		__antithesis_instrumentation__.Notify(17183)

		highWater, err = f.runUntilTableEvent(ctx, highWater)
		if err != nil {
			__antithesis_instrumentation__.Notify(17194)
			if tErr := (*errEndTimeReached)(nil); errors.As(err, &tErr) {
				__antithesis_instrumentation__.Notify(17196)
				if err := emitResolved(highWater, jobspb.ResolvedSpan_EXIT); err != nil {
					__antithesis_instrumentation__.Notify(17198)
					return err
				} else {
					__antithesis_instrumentation__.Notify(17199)
				}
				__antithesis_instrumentation__.Notify(17197)
				return errChangefeedCompleted
			} else {
				__antithesis_instrumentation__.Notify(17200)
			}
			__antithesis_instrumentation__.Notify(17195)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17201)
		}
		__antithesis_instrumentation__.Notify(17184)

		boundaryType := jobspb.ResolvedSpan_BACKFILL
		if f.schemaChangePolicy == changefeedbase.OptSchemaChangePolicyStop {
			__antithesis_instrumentation__.Notify(17202)
			boundaryType = jobspb.ResolvedSpan_EXIT
		} else {
			__antithesis_instrumentation__.Notify(17203)
			if events, err := f.tableFeed.Peek(ctx, highWater.Next()); err == nil && func() bool {
				__antithesis_instrumentation__.Notify(17204)
				return isPrimaryKeyChange(events) == true
			}() == true {
				__antithesis_instrumentation__.Notify(17205)
				boundaryType = jobspb.ResolvedSpan_RESTART
			} else {
				__antithesis_instrumentation__.Notify(17206)
				if err != nil {
					__antithesis_instrumentation__.Notify(17207)
					return err
				} else {
					__antithesis_instrumentation__.Notify(17208)
				}
			}
		}
		__antithesis_instrumentation__.Notify(17185)

		if f.schemaChangePolicy != changefeedbase.OptSchemaChangePolicyNoBackfill || func() bool {
			__antithesis_instrumentation__.Notify(17209)
			return boundaryType == jobspb.ResolvedSpan_RESTART == true
		}() == true {
			__antithesis_instrumentation__.Notify(17210)
			if err := emitResolved(highWater, boundaryType); err != nil {
				__antithesis_instrumentation__.Notify(17211)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17212)
			}
		} else {
			__antithesis_instrumentation__.Notify(17213)
		}
		__antithesis_instrumentation__.Notify(17186)

		if boundaryType == jobspb.ResolvedSpan_RESTART || func() bool {
			__antithesis_instrumentation__.Notify(17214)
			return boundaryType == jobspb.ResolvedSpan_EXIT == true
		}() == true {
			__antithesis_instrumentation__.Notify(17215)
			return schemaChangeDetectedError{highWater.Next()}
		} else {
			__antithesis_instrumentation__.Notify(17216)
		}
	}
}

func isPrimaryKeyChange(events []schemafeed.TableEvent) bool {
	__antithesis_instrumentation__.Notify(17217)
	for _, ev := range events {
		__antithesis_instrumentation__.Notify(17219)
		if schemafeed.IsPrimaryIndexChange(ev) {
			__antithesis_instrumentation__.Notify(17220)
			return true
		} else {
			__antithesis_instrumentation__.Notify(17221)
		}
	}
	__antithesis_instrumentation__.Notify(17218)
	return false
}

func filterCheckpointSpans(spans []roachpb.Span, completed []roachpb.Span) []roachpb.Span {
	__antithesis_instrumentation__.Notify(17222)
	var sg roachpb.SpanGroup
	sg.Add(spans...)
	sg.Sub(completed...)
	return sg.Slice()
}

func (f *kvFeed) scanIfShould(
	ctx context.Context, initialScan bool, highWater hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(17223)
	scanTime := highWater.Next()

	events, err := f.tableFeed.Peek(ctx, scanTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(17230)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17231)
	}
	__antithesis_instrumentation__.Notify(17224)

	isInitialScan := initialScan && func() bool {
		__antithesis_instrumentation__.Notify(17232)
		return f.withInitialBackfill == true
	}() == true
	var spansToBackfill []roachpb.Span
	if isInitialScan {
		__antithesis_instrumentation__.Notify(17233)
		scanTime = highWater
		spansToBackfill = f.spans
	} else {
		__antithesis_instrumentation__.Notify(17234)
		if len(events) > 0 {
			__antithesis_instrumentation__.Notify(17235)

			for _, ev := range events {
				__antithesis_instrumentation__.Notify(17236)

				if schemafeed.IsOnlyPrimaryIndexChange(ev) {
					__antithesis_instrumentation__.Notify(17239)
					continue
				} else {
					__antithesis_instrumentation__.Notify(17240)
				}
				__antithesis_instrumentation__.Notify(17237)
				tablePrefix := f.codec.TablePrefix(uint32(ev.After.GetID()))
				tableSpan := roachpb.Span{Key: tablePrefix, EndKey: tablePrefix.PrefixEnd()}
				for _, sp := range f.spans {
					__antithesis_instrumentation__.Notify(17241)
					if tableSpan.Overlaps(sp) {
						__antithesis_instrumentation__.Notify(17242)
						spansToBackfill = append(spansToBackfill, sp)
					} else {
						__antithesis_instrumentation__.Notify(17243)
					}
				}
				__antithesis_instrumentation__.Notify(17238)
				if !scanTime.Equal(ev.After.GetModificationTime()) {
					__antithesis_instrumentation__.Notify(17244)
					return errors.AssertionFailedf("found event in shouldScan which did not occur at the scan time %v: %v",
						scanTime, ev)
				} else {
					__antithesis_instrumentation__.Notify(17245)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(17246)
			return nil
		}
	}
	__antithesis_instrumentation__.Notify(17225)

	if _, err := f.tableFeed.Pop(ctx, scanTime); err != nil {
		__antithesis_instrumentation__.Notify(17247)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17248)
	}
	__antithesis_instrumentation__.Notify(17226)

	spansToBackfill = filterCheckpointSpans(spansToBackfill, f.checkpoint)

	if (!isInitialScan && func() bool {
		__antithesis_instrumentation__.Notify(17249)
		return f.schemaChangePolicy == changefeedbase.OptSchemaChangePolicyNoBackfill == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(17250)
		return len(spansToBackfill) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(17251)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(17252)
	}
	__antithesis_instrumentation__.Notify(17227)

	if f.onBackfillCallback != nil {
		__antithesis_instrumentation__.Notify(17253)
		defer f.onBackfillCallback()()
	} else {
		__antithesis_instrumentation__.Notify(17254)
	}
	__antithesis_instrumentation__.Notify(17228)

	if err := f.scanner.Scan(ctx, f.writer, physicalConfig{
		Spans:     spansToBackfill,
		Timestamp: scanTime,
		WithDiff: !isInitialScan && func() bool {
			__antithesis_instrumentation__.Notify(17255)
			return f.withDiff == true
		}() == true,
		Knobs: f.knobs,
	}); err != nil {
		__antithesis_instrumentation__.Notify(17256)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17257)
	}
	__antithesis_instrumentation__.Notify(17229)

	f.checkpoint = nil

	return nil
}

func (f *kvFeed) runUntilTableEvent(
	ctx context.Context, startFrom hlc.Timestamp,
) (resolvedUpTo hlc.Timestamp, err error) {
	__antithesis_instrumentation__.Notify(17258)

	if _, err := f.tableFeed.Peek(ctx, startFrom); err != nil {
		__antithesis_instrumentation__.Notify(17263)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(17264)
	}
	__antithesis_instrumentation__.Notify(17259)

	memBuf := f.bufferFactory()
	defer func() {
		__antithesis_instrumentation__.Notify(17265)
		err = errors.CombineErrors(err, memBuf.CloseWithReason(ctx, err))
	}()
	__antithesis_instrumentation__.Notify(17260)

	g := ctxgroup.WithContext(ctx)
	physicalCfg := physicalConfig{
		Spans:     f.spans,
		Timestamp: startFrom,
		WithDiff:  f.withDiff,
		Knobs:     f.knobs,
	}
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(17266)
		return copyFromSourceToDestUntilTableEvent(ctx, f.writer, memBuf, physicalCfg, f.tableFeed, f.endTime)
	})
	__antithesis_instrumentation__.Notify(17261)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(17267)
		return f.physicalFeed.Run(ctx, memBuf, physicalCfg)
	})
	__antithesis_instrumentation__.Notify(17262)

	err = g.Wait()
	if err == nil {
		__antithesis_instrumentation__.Notify(17268)
		return hlc.Timestamp{},
			errors.AssertionFailedf("feed exited with no error and no scan boundary")
	} else {
		__antithesis_instrumentation__.Notify(17269)
		if tErr := (*errTableEventReached)(nil); errors.As(err, &tErr) {
			__antithesis_instrumentation__.Notify(17270)

			return tErr.Timestamp().Prev(), nil
		} else {
			__antithesis_instrumentation__.Notify(17271)
			if tErr := (*errEndTimeReached)(nil); errors.As(err, &tErr) {
				__antithesis_instrumentation__.Notify(17272)
				return tErr.endTime.Prev(), err
			} else {
				__antithesis_instrumentation__.Notify(17273)
				if kvcoord.IsSendError(err) {
					__antithesis_instrumentation__.Notify(17274)

					err = changefeedbase.MarkRetryableError(err)
					return hlc.Timestamp{}, err
				} else {
					__antithesis_instrumentation__.Notify(17275)
					return hlc.Timestamp{}, err
				}
			}
		}
	}
}

type errBoundaryReached interface {
	error
	Timestamp() hlc.Timestamp
}

type errTableEventReached struct {
	schemafeed.TableEvent
}

func (e *errTableEventReached) Error() string {
	__antithesis_instrumentation__.Notify(17276)
	return "scan boundary reached: " + e.String()
}

type errEndTimeReached struct {
	endTime hlc.Timestamp
}

func (e *errEndTimeReached) Error() string {
	__antithesis_instrumentation__.Notify(17277)
	return "end time reached: " + e.endTime.String()
}

func (e *errEndTimeReached) Timestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(17278)
	return e.endTime
}

type errUnknownEvent struct {
	kvevent.Event
}

var _ errBoundaryReached = (*errTableEventReached)(nil)
var _ errBoundaryReached = (*errEndTimeReached)(nil)

func (e *errUnknownEvent) Error() string {
	__antithesis_instrumentation__.Notify(17279)
	return "unknown event type"
}

func copyFromSourceToDestUntilTableEvent(
	ctx context.Context,
	dest kvevent.Writer,
	source kvevent.Reader,
	cfg physicalConfig,
	tables schemafeed.SchemaFeed,
	endTime hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(17280)

	frontier, err := span.MakeFrontier(cfg.Spans...)
	if err != nil {
		__antithesis_instrumentation__.Notify(17284)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17285)
	}
	__antithesis_instrumentation__.Notify(17281)
	for _, span := range cfg.Spans {
		__antithesis_instrumentation__.Notify(17286)
		if _, err := frontier.Forward(span, cfg.Timestamp); err != nil {
			__antithesis_instrumentation__.Notify(17287)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17288)
		}
	}
	__antithesis_instrumentation__.Notify(17282)
	var (
		scanBoundary         errBoundaryReached
		checkForScanBoundary = func(ts hlc.Timestamp) error {
			__antithesis_instrumentation__.Notify(17289)

			_, isEndTimeBoundary := scanBoundary.(*errEndTimeReached)
			if scanBoundary != nil && func() bool {
				__antithesis_instrumentation__.Notify(17293)
				return !isEndTimeBoundary == true
			}() == true {
				__antithesis_instrumentation__.Notify(17294)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(17295)
			}
			__antithesis_instrumentation__.Notify(17290)
			nextEvents, err := tables.Peek(ctx, ts)
			if err != nil {
				__antithesis_instrumentation__.Notify(17296)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17297)
			}
			__antithesis_instrumentation__.Notify(17291)

			if len(nextEvents) > 0 {
				__antithesis_instrumentation__.Notify(17298)
				scanBoundary = &errTableEventReached{nextEvents[0]}
			} else {
				__antithesis_instrumentation__.Notify(17299)
				if !endTime.IsEmpty() && func() bool {
					__antithesis_instrumentation__.Notify(17300)
					return scanBoundary == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(17301)
					scanBoundary = &errEndTimeReached{
						endTime: endTime,
					}
				} else {
					__antithesis_instrumentation__.Notify(17302)
				}
			}
			__antithesis_instrumentation__.Notify(17292)
			return nil
		}
		applyScanBoundary = func(e kvevent.Event) (skipEvent, reachedBoundary bool, err error) {
			__antithesis_instrumentation__.Notify(17303)
			if scanBoundary == nil {
				__antithesis_instrumentation__.Notify(17306)
				return false, false, nil
			} else {
				__antithesis_instrumentation__.Notify(17307)
			}
			__antithesis_instrumentation__.Notify(17304)
			if e.Timestamp().Less(scanBoundary.Timestamp()) {
				__antithesis_instrumentation__.Notify(17308)
				return false, false, nil
			} else {
				__antithesis_instrumentation__.Notify(17309)
			}
			__antithesis_instrumentation__.Notify(17305)
			switch e.Type() {
			case kvevent.TypeKV:
				__antithesis_instrumentation__.Notify(17310)
				return true, false, nil
			case kvevent.TypeResolved:
				__antithesis_instrumentation__.Notify(17311)
				boundaryResolvedTimestamp := scanBoundary.Timestamp().Prev()
				resolved := e.Resolved()
				if resolved.Timestamp.LessEq(boundaryResolvedTimestamp) {
					__antithesis_instrumentation__.Notify(17316)
					return false, false, nil
				} else {
					__antithesis_instrumentation__.Notify(17317)
				}
				__antithesis_instrumentation__.Notify(17312)
				if _, err := frontier.Forward(resolved.Span, boundaryResolvedTimestamp); err != nil {
					__antithesis_instrumentation__.Notify(17318)
					return false, false, err
				} else {
					__antithesis_instrumentation__.Notify(17319)
				}
				__antithesis_instrumentation__.Notify(17313)
				return true, frontier.Frontier().EqOrdering(boundaryResolvedTimestamp), nil
			case kvevent.TypeFlush:
				__antithesis_instrumentation__.Notify(17314)

				return false, false, nil

			default:
				__antithesis_instrumentation__.Notify(17315)
				return false, false, &errUnknownEvent{e}
			}
		}
		addEntry = func(e kvevent.Event) error {
			__antithesis_instrumentation__.Notify(17320)
			switch e.Type() {
			case kvevent.TypeKV, kvevent.TypeFlush:
				__antithesis_instrumentation__.Notify(17321)
				return dest.Add(ctx, e)
			case kvevent.TypeResolved:
				__antithesis_instrumentation__.Notify(17322)

				resolved := e.Resolved()
				if _, err := frontier.Forward(resolved.Span, resolved.Timestamp); err != nil {
					__antithesis_instrumentation__.Notify(17325)
					return err
				} else {
					__antithesis_instrumentation__.Notify(17326)
				}
				__antithesis_instrumentation__.Notify(17323)
				return dest.Add(ctx, e)
			default:
				__antithesis_instrumentation__.Notify(17324)
				return &errUnknownEvent{e}
			}
		}
		copyEvent = func(e kvevent.Event) error {
			__antithesis_instrumentation__.Notify(17327)
			if err := checkForScanBoundary(e.Timestamp()); err != nil {
				__antithesis_instrumentation__.Notify(17332)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17333)
			}
			__antithesis_instrumentation__.Notify(17328)
			skipEntry, scanBoundaryReached, err := applyScanBoundary(e)
			if err != nil {
				__antithesis_instrumentation__.Notify(17334)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17335)
			}
			__antithesis_instrumentation__.Notify(17329)
			if scanBoundaryReached {
				__antithesis_instrumentation__.Notify(17336)

				return scanBoundary
			} else {
				__antithesis_instrumentation__.Notify(17337)
			}
			__antithesis_instrumentation__.Notify(17330)
			if skipEntry {
				__antithesis_instrumentation__.Notify(17338)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(17339)
			}
			__antithesis_instrumentation__.Notify(17331)
			return addEntry(e)
		}
	)
	__antithesis_instrumentation__.Notify(17283)

	for {
		__antithesis_instrumentation__.Notify(17340)
		e, err := source.Get(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(17342)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17343)
		}
		__antithesis_instrumentation__.Notify(17341)
		if err := copyEvent(e); err != nil {
			__antithesis_instrumentation__.Notify(17344)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17345)
		}
	}
}
