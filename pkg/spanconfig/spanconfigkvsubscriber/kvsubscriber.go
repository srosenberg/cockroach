package spanconfigkvsubscriber

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedbuffer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedcache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigstore"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type KVSubscriber struct {
	fallback roachpb.SpanConfig
	knobs    *spanconfig.TestingKnobs

	rfc *rangefeedcache.Watcher

	mu struct {
		syncutil.RWMutex
		lastUpdated hlc.Timestamp

		internal spanconfig.Store
		handlers []handler
	}
}

var _ spanconfig.KVSubscriber = &KVSubscriber{}

const spanConfigurationsTableRowSize = 5 << 10

func New(
	clock *hlc.Clock,
	rangeFeedFactory *rangefeed.Factory,
	spanConfigurationsTableID uint32,
	bufferMemLimit int64,
	fallback roachpb.SpanConfig,
	knobs *spanconfig.TestingKnobs,
) *KVSubscriber {
	__antithesis_instrumentation__.Notify(240593)
	spanConfigTableStart := keys.SystemSQLCodec.IndexPrefix(
		spanConfigurationsTableID,
		keys.SpanConfigurationsTablePrimaryKeyIndexID,
	)
	spanConfigTableSpan := roachpb.Span{
		Key:    spanConfigTableStart,
		EndKey: spanConfigTableStart.PrefixEnd(),
	}
	spanConfigStore := spanconfigstore.New(fallback)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(240596)
		knobs = &spanconfig.TestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(240597)
	}
	__antithesis_instrumentation__.Notify(240594)
	s := &KVSubscriber{
		fallback: fallback,
		knobs:    knobs,
	}
	var rfCacheKnobs *rangefeedcache.TestingKnobs
	if knobs != nil {
		__antithesis_instrumentation__.Notify(240598)
		rfCacheKnobs, _ = knobs.KVSubscriberRangeFeedKnobs.(*rangefeedcache.TestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(240599)
	}
	__antithesis_instrumentation__.Notify(240595)
	s.rfc = rangefeedcache.NewWatcher(
		"spanconfig-subscriber",
		clock, rangeFeedFactory,
		int(bufferMemLimit/spanConfigurationsTableRowSize),
		[]roachpb.Span{spanConfigTableSpan},
		true,
		newSpanConfigDecoder().translateEvent,
		s.handleUpdate,
		rfCacheKnobs,
	)
	s.mu.internal = spanConfigStore
	return s
}

func (s *KVSubscriber) Start(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(240600)
	return rangefeedcache.Start(ctx, stopper, s.rfc, nil)
}

func (s *KVSubscriber) Subscribe(fn func(context.Context, roachpb.Span)) {
	__antithesis_instrumentation__.Notify(240601)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mu.handlers = append(s.mu.handlers, handler{fn: fn})
}

func (s *KVSubscriber) LastUpdated() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(240602)
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.mu.lastUpdated
}

func (s *KVSubscriber) NeedsSplit(ctx context.Context, start, end roachpb.RKey) bool {
	__antithesis_instrumentation__.Notify(240603)
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.mu.internal.NeedsSplit(ctx, start, end)
}

func (s *KVSubscriber) ComputeSplitKey(ctx context.Context, start, end roachpb.RKey) roachpb.RKey {
	__antithesis_instrumentation__.Notify(240604)
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.mu.internal.ComputeSplitKey(ctx, start, end)
}

func (s *KVSubscriber) GetSpanConfigForKey(
	ctx context.Context, key roachpb.RKey,
) (roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(240605)
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.mu.internal.GetSpanConfigForKey(ctx, key)
}

func (s *KVSubscriber) GetProtectionTimestamps(
	ctx context.Context, sp roachpb.Span,
) (protectionTimestamps []hlc.Timestamp, asOf hlc.Timestamp, _ error) {
	__antithesis_instrumentation__.Notify(240606)
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.mu.internal.ForEachOverlappingSpanConfig(ctx, sp,
		func(sp roachpb.Span, config roachpb.SpanConfig) error {
			__antithesis_instrumentation__.Notify(240608)
			for _, protection := range config.GCPolicy.ProtectionPolicies {
				__antithesis_instrumentation__.Notify(240610)

				if config.ExcludeDataFromBackup && func() bool {
					__antithesis_instrumentation__.Notify(240612)
					return protection.IgnoreIfExcludedFromBackup == true
				}() == true {
					__antithesis_instrumentation__.Notify(240613)
					continue
				} else {
					__antithesis_instrumentation__.Notify(240614)
				}
				__antithesis_instrumentation__.Notify(240611)
				protectionTimestamps = append(protectionTimestamps, protection.ProtectedTimestamp)
			}
			__antithesis_instrumentation__.Notify(240609)
			return nil
		}); err != nil {
		__antithesis_instrumentation__.Notify(240615)
		return nil, hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(240616)
	}
	__antithesis_instrumentation__.Notify(240607)

	return protectionTimestamps, s.mu.lastUpdated, nil
}

func (s *KVSubscriber) handleUpdate(ctx context.Context, u rangefeedcache.Update) {
	__antithesis_instrumentation__.Notify(240617)
	switch u.Type {
	case rangefeedcache.CompleteUpdate:
		__antithesis_instrumentation__.Notify(240618)
		s.handleCompleteUpdate(ctx, u.Timestamp, u.Events)
	case rangefeedcache.IncrementalUpdate:
		__antithesis_instrumentation__.Notify(240619)
		s.handlePartialUpdate(ctx, u.Timestamp, u.Events)
	default:
		__antithesis_instrumentation__.Notify(240620)
	}
}

func (s *KVSubscriber) handleCompleteUpdate(
	ctx context.Context, ts hlc.Timestamp, events []rangefeedbuffer.Event,
) {
	__antithesis_instrumentation__.Notify(240621)
	freshStore := spanconfigstore.New(s.fallback)
	for _, ev := range events {
		__antithesis_instrumentation__.Notify(240623)
		freshStore.Apply(ctx, false, ev.(*bufferEvent).Update)
	}
	__antithesis_instrumentation__.Notify(240622)
	s.mu.Lock()
	s.mu.internal = freshStore
	s.mu.lastUpdated = ts
	handlers := s.mu.handlers
	s.mu.Unlock()
	for i := range handlers {
		__antithesis_instrumentation__.Notify(240624)
		handler := &handlers[i]
		handler.invoke(ctx, keys.EverythingSpan)
	}
}

func (s *KVSubscriber) handlePartialUpdate(
	ctx context.Context, ts hlc.Timestamp, events []rangefeedbuffer.Event,
) {
	__antithesis_instrumentation__.Notify(240625)
	s.mu.Lock()
	for _, ev := range events {
		__antithesis_instrumentation__.Notify(240627)

		s.mu.internal.Apply(ctx, false, ev.(*bufferEvent).Update)
	}
	__antithesis_instrumentation__.Notify(240626)
	s.mu.lastUpdated = ts
	handlers := s.mu.handlers
	s.mu.Unlock()

	for i := range handlers {
		__antithesis_instrumentation__.Notify(240628)
		handler := &handlers[i]
		for _, ev := range events {
			__antithesis_instrumentation__.Notify(240629)
			target := ev.(*bufferEvent).Update.GetTarget()
			handler.invoke(ctx, target.KeyspaceTargeted())
		}
	}
}

type handler struct {
	initialized bool
	fn          func(ctx context.Context, update roachpb.Span)
}

func (h *handler) invoke(ctx context.Context, update roachpb.Span) {
	__antithesis_instrumentation__.Notify(240630)
	if !h.initialized {
		__antithesis_instrumentation__.Notify(240632)
		h.fn(ctx, keys.EverythingSpan)
		h.initialized = true

		if update.Equal(keys.EverythingSpan) {
			__antithesis_instrumentation__.Notify(240633)
			return
		} else {
			__antithesis_instrumentation__.Notify(240634)
		}
	} else {
		__antithesis_instrumentation__.Notify(240635)
	}
	__antithesis_instrumentation__.Notify(240631)

	h.fn(ctx, update)
}

type bufferEvent struct {
	spanconfig.Update
	ts hlc.Timestamp
}

func (w *bufferEvent) Timestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(240636)
	return w.ts
}

var _ rangefeedbuffer.Event = &bufferEvent{}
