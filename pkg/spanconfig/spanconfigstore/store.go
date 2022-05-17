package spanconfigstore

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

var EnabledSetting = settings.RegisterBoolSetting(
	settings.SystemOnly,
	"spanconfig.store.enabled",
	`use the span config infrastructure in KV instead of the system config span`,
	true,
)

type Store struct {
	mu struct {
		syncutil.RWMutex
		spanConfigStore       *spanConfigStore
		systemSpanConfigStore *systemSpanConfigStore
	}

	fallback roachpb.SpanConfig
}

var _ spanconfig.Store = &Store{}

func New(fallback roachpb.SpanConfig) *Store {
	__antithesis_instrumentation__.Notify(241587)
	s := &Store{fallback: fallback}
	s.mu.spanConfigStore = newSpanConfigStore()
	s.mu.systemSpanConfigStore = newSystemSpanConfigStore()
	return s
}

func (s *Store) NeedsSplit(ctx context.Context, start, end roachpb.RKey) bool {
	__antithesis_instrumentation__.Notify(241588)
	return len(s.ComputeSplitKey(ctx, start, end)) > 0
}

func (s *Store) ComputeSplitKey(_ context.Context, start, end roachpb.RKey) roachpb.RKey {
	__antithesis_instrumentation__.Notify(241589)
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.mu.spanConfigStore.computeSplitKey(start, end)
}

func (s *Store) GetSpanConfigForKey(
	ctx context.Context, key roachpb.RKey,
) (roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(241590)
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getSpanConfigForKeyRLocked(ctx, key)
}

func (s *Store) getSpanConfigForKeyRLocked(
	ctx context.Context, key roachpb.RKey,
) (roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(241591)
	conf, found, err := s.mu.spanConfigStore.getSpanConfigForKey(ctx, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(241594)
		return roachpb.SpanConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(241595)
	}
	__antithesis_instrumentation__.Notify(241592)
	if !found {
		__antithesis_instrumentation__.Notify(241596)
		conf = s.fallback
	} else {
		__antithesis_instrumentation__.Notify(241597)
	}
	__antithesis_instrumentation__.Notify(241593)
	return s.mu.systemSpanConfigStore.combine(key, conf)
}

func (s *Store) Apply(
	ctx context.Context, dryrun bool, updates ...spanconfig.Update,
) (deleted []spanconfig.Target, added []spanconfig.Record) {
	__antithesis_instrumentation__.Notify(241598)
	deleted, added, err := s.applyInternal(dryrun, updates...)
	if err != nil {
		__antithesis_instrumentation__.Notify(241600)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(241601)
	}
	__antithesis_instrumentation__.Notify(241599)
	return deleted, added
}

func (s *Store) ForEachOverlappingSpanConfig(
	ctx context.Context, span roachpb.Span, f func(roachpb.Span, roachpb.SpanConfig) error,
) error {
	__antithesis_instrumentation__.Notify(241602)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mu.spanConfigStore.forEachOverlapping(span, func(entry spanConfigEntry) error {
		__antithesis_instrumentation__.Notify(241603)
		config, err := s.getSpanConfigForKeyRLocked(ctx, roachpb.RKey(entry.span.Key))
		if err != nil {
			__antithesis_instrumentation__.Notify(241605)
			return err
		} else {
			__antithesis_instrumentation__.Notify(241606)
		}
		__antithesis_instrumentation__.Notify(241604)
		return f(entry.span, config)
	})
}

func (s *Store) Copy(ctx context.Context) *Store {
	__antithesis_instrumentation__.Notify(241607)
	s.mu.Lock()
	defer s.mu.Unlock()

	clone := New(s.fallback)
	clone.mu.spanConfigStore = s.mu.spanConfigStore.copy(ctx)
	clone.mu.systemSpanConfigStore = s.mu.systemSpanConfigStore.copy()
	return clone
}

func (s *Store) applyInternal(
	dryrun bool, updates ...spanconfig.Update,
) (deleted []spanconfig.Target, added []spanconfig.Record, err error) {
	__antithesis_instrumentation__.Notify(241608)
	s.mu.Lock()
	defer s.mu.Unlock()

	spanStoreUpdates := make([]spanconfig.Update, 0, len(updates))
	systemSpanConfigStoreUpdates := make([]spanconfig.Update, 0, len(updates))
	for _, update := range updates {
		__antithesis_instrumentation__.Notify(241615)
		switch {
		case update.GetTarget().IsSpanTarget():
			__antithesis_instrumentation__.Notify(241616)
			spanStoreUpdates = append(spanStoreUpdates, update)
		case update.GetTarget().IsSystemTarget():
			__antithesis_instrumentation__.Notify(241617)
			systemSpanConfigStoreUpdates = append(systemSpanConfigStoreUpdates, update)
		default:
			__antithesis_instrumentation__.Notify(241618)
			return nil, nil, errors.AssertionFailedf("unknown target type")
		}
	}
	__antithesis_instrumentation__.Notify(241609)
	deletedSpans, addedEntries, err := s.mu.spanConfigStore.apply(dryrun, spanStoreUpdates...)
	if err != nil {
		__antithesis_instrumentation__.Notify(241619)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(241620)
	}
	__antithesis_instrumentation__.Notify(241610)

	for _, sp := range deletedSpans {
		__antithesis_instrumentation__.Notify(241621)
		deleted = append(deleted, spanconfig.MakeTargetFromSpan(sp))
	}
	__antithesis_instrumentation__.Notify(241611)

	for _, entry := range addedEntries {
		__antithesis_instrumentation__.Notify(241622)
		record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(entry.span),
			entry.config)
		if err != nil {
			__antithesis_instrumentation__.Notify(241624)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(241625)
		}
		__antithesis_instrumentation__.Notify(241623)
		added = append(added, record)
	}
	__antithesis_instrumentation__.Notify(241612)

	deletedSystemTargets, addedSystemSpanConfigRecords, err := s.mu.systemSpanConfigStore.apply(
		systemSpanConfigStoreUpdates...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(241626)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(241627)
	}
	__antithesis_instrumentation__.Notify(241613)
	for _, systemTarget := range deletedSystemTargets {
		__antithesis_instrumentation__.Notify(241628)
		deleted = append(deleted, spanconfig.MakeTargetFromSystemTarget(systemTarget))
	}
	__antithesis_instrumentation__.Notify(241614)
	added = append(added, addedSystemSpanConfigRecords...)

	return deleted, added, nil
}

func (s *Store) Iterate(f func(spanconfig.Record) error) error {
	__antithesis_instrumentation__.Notify(241629)
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.mu.systemSpanConfigStore.iterate(f); err != nil {
		__antithesis_instrumentation__.Notify(241631)
		return err
	} else {
		__antithesis_instrumentation__.Notify(241632)
	}
	__antithesis_instrumentation__.Notify(241630)
	return s.mu.spanConfigStore.forEachOverlapping(
		keys.EverythingSpan,
		func(s spanConfigEntry) error {
			__antithesis_instrumentation__.Notify(241633)
			record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(s.span), s.config)
			if err != nil {
				__antithesis_instrumentation__.Notify(241635)
				return err
			} else {
				__antithesis_instrumentation__.Notify(241636)
			}
			__antithesis_instrumentation__.Notify(241634)
			return f(record)
		})
}
