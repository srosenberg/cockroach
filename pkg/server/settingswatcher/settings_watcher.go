// Package settingswatcher provides utilities to update cluster settings using
// a range feed.
package settingswatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedbuffer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedcache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type SettingsWatcher struct {
	clock    *hlc.Clock
	codec    keys.SQLCodec
	settings *cluster.Settings
	f        *rangefeed.Factory
	stopper  *stop.Stopper
	dec      RowDecoder
	storage  Storage

	overridesMonitor OverridesMonitor

	mu struct {
		syncutil.Mutex

		updater   settings.Updater
		values    map[string]settings.EncodedValue
		overrides map[string]settings.EncodedValue
	}

	testingWatcherKnobs *rangefeedcache.TestingKnobs
}

type Storage interface {
	SnapshotKVs(ctx context.Context, kvs []roachpb.KeyValue)
}

func New(
	clock *hlc.Clock,
	codec keys.SQLCodec,
	settingsToUpdate *cluster.Settings,
	f *rangefeed.Factory,
	stopper *stop.Stopper,
	storage Storage,
) *SettingsWatcher {
	__antithesis_instrumentation__.Notify(235124)
	return &SettingsWatcher{
		clock:    clock,
		codec:    codec,
		settings: settingsToUpdate,
		f:        f,
		stopper:  stopper,
		dec:      MakeRowDecoder(codec),
		storage:  storage,
	}
}

func NewWithOverrides(
	clock *hlc.Clock,
	codec keys.SQLCodec,
	settingsToUpdate *cluster.Settings,
	f *rangefeed.Factory,
	stopper *stop.Stopper,
	overridesMonitor OverridesMonitor,
	storage Storage,
) *SettingsWatcher {
	__antithesis_instrumentation__.Notify(235125)
	s := New(clock, codec, settingsToUpdate, f, stopper, storage)
	s.overridesMonitor = overridesMonitor
	settingsToUpdate.OverridesInformer = s
	return s
}

func (s *SettingsWatcher) Start(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(235126)
	settingsTablePrefix := s.codec.TablePrefix(keys.SettingsTableID)
	settingsTableSpan := roachpb.Span{
		Key:    settingsTablePrefix,
		EndKey: settingsTablePrefix.PrefixEnd(),
	}
	s.resetUpdater()
	var initialScan = struct {
		ch   chan struct{}
		done bool
		err  error
	}{
		ch: make(chan struct{}),
	}
	noteUpdate := func(update rangefeedcache.Update) {
		__antithesis_instrumentation__.Notify(235133)
		if update.Type != rangefeedcache.CompleteUpdate {
			__antithesis_instrumentation__.Notify(235135)
			return
		} else {
			__antithesis_instrumentation__.Notify(235136)
		}
		__antithesis_instrumentation__.Notify(235134)
		s.mu.Lock()
		defer s.mu.Unlock()
		s.mu.updater.ResetRemaining(ctx)
		if !initialScan.done {
			__antithesis_instrumentation__.Notify(235137)
			initialScan.done = true
			close(initialScan.ch)
		} else {
			__antithesis_instrumentation__.Notify(235138)
		}
	}
	__antithesis_instrumentation__.Notify(235127)

	s.mu.values = make(map[string]settings.EncodedValue)

	if s.overridesMonitor != nil {
		__antithesis_instrumentation__.Notify(235139)
		s.mu.overrides = make(map[string]settings.EncodedValue)

		s.updateOverrides(ctx)

		if err := s.stopper.RunAsyncTask(ctx, "setting-overrides", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(235140)
			overridesCh := s.overridesMonitor.RegisterOverridesChannel()
			for {
				__antithesis_instrumentation__.Notify(235141)
				select {
				case <-overridesCh:
					__antithesis_instrumentation__.Notify(235142)
					s.updateOverrides(ctx)

				case <-s.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(235143)
					return
				}
			}
		}); err != nil {
			__antithesis_instrumentation__.Notify(235144)

			return err
		} else {
			__antithesis_instrumentation__.Notify(235145)
		}
	} else {
		__antithesis_instrumentation__.Notify(235146)
	}
	__antithesis_instrumentation__.Notify(235128)

	var bufferSize int
	if s.storage != nil {
		__antithesis_instrumentation__.Notify(235147)
		bufferSize = settings.MaxSettings * 3
	} else {
		__antithesis_instrumentation__.Notify(235148)
	}
	__antithesis_instrumentation__.Notify(235129)
	var snapshot []roachpb.KeyValue
	maybeUpdateSnapshot := func(update rangefeedcache.Update) {
		__antithesis_instrumentation__.Notify(235149)

		if s.storage == nil || func() bool {
			__antithesis_instrumentation__.Notify(235152)
			return (update.Type == rangefeedcache.IncrementalUpdate && func() bool {
				__antithesis_instrumentation__.Notify(235153)
				return len(update.Events) == 0 == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(235154)
			return
		} else {
			__antithesis_instrumentation__.Notify(235155)
		}
		__antithesis_instrumentation__.Notify(235150)
		eventKVs := rangefeedbuffer.EventsToKVs(update.Events,
			rangefeedbuffer.RangeFeedValueEventToKV)
		switch update.Type {
		case rangefeedcache.CompleteUpdate:
			__antithesis_instrumentation__.Notify(235156)
			snapshot = eventKVs
		case rangefeedcache.IncrementalUpdate:
			__antithesis_instrumentation__.Notify(235157)
			snapshot = rangefeedbuffer.MergeKVs(snapshot, eventKVs)
		default:
			__antithesis_instrumentation__.Notify(235158)
		}
		__antithesis_instrumentation__.Notify(235151)
		s.storage.SnapshotKVs(ctx, snapshot)
	}
	__antithesis_instrumentation__.Notify(235130)
	c := rangefeedcache.NewWatcher(
		"settings-watcher",
		s.clock, s.f,
		bufferSize,
		[]roachpb.Span{settingsTableSpan},
		false,
		func(ctx context.Context, kv *roachpb.RangeFeedValue) rangefeedbuffer.Event {
			__antithesis_instrumentation__.Notify(235159)
			return s.handleKV(ctx, kv)
		},
		func(ctx context.Context, update rangefeedcache.Update) {
			__antithesis_instrumentation__.Notify(235160)
			noteUpdate(update)
			maybeUpdateSnapshot(update)
		},
		s.testingWatcherKnobs,
	)
	__antithesis_instrumentation__.Notify(235131)

	if err := rangefeedcache.Start(ctx, s.stopper, c, func(err error) {
		__antithesis_instrumentation__.Notify(235161)
		if !initialScan.done {
			__antithesis_instrumentation__.Notify(235162)
			initialScan.err = err
			initialScan.done = true
			close(initialScan.ch)
		} else {
			__antithesis_instrumentation__.Notify(235163)
			s.resetUpdater()
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(235164)
		return err
	} else {
		__antithesis_instrumentation__.Notify(235165)
	}
	__antithesis_instrumentation__.Notify(235132)

	select {
	case <-initialScan.ch:
		__antithesis_instrumentation__.Notify(235166)
		return initialScan.err

	case <-s.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(235167)
		return errors.Wrap(stop.ErrUnavailable, "failed to retrieve initial cluster settings")

	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(235168)
		return errors.Wrap(ctx.Err(), "failed to retrieve initial cluster settings")
	}
}

func (s *SettingsWatcher) handleKV(
	ctx context.Context, kv *roachpb.RangeFeedValue,
) rangefeedbuffer.Event {
	__antithesis_instrumentation__.Notify(235169)
	name, val, tombstone, err := s.dec.DecodeRow(roachpb.KeyValue{
		Key:   kv.Key,
		Value: kv.Value,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(235174)
		log.Warningf(ctx, "failed to decode settings row %v: %v", kv.Key, err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(235175)
	}
	__antithesis_instrumentation__.Notify(235170)

	if !s.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(235176)
		setting, ok := settings.Lookup(name, settings.LookupForLocalAccess, s.codec.ForSystemTenant())
		if !ok {
			__antithesis_instrumentation__.Notify(235178)
			log.Warningf(ctx, "unknown setting %s, skipping update", redact.Safe(name))
			return nil
		} else {
			__antithesis_instrumentation__.Notify(235179)
		}
		__antithesis_instrumentation__.Notify(235177)
		if setting.Class() != settings.TenantWritable {
			__antithesis_instrumentation__.Notify(235180)
			log.Warningf(ctx, "ignoring read-only setting %s", redact.Safe(name))
			return nil
		} else {
			__antithesis_instrumentation__.Notify(235181)
		}
	} else {
		__antithesis_instrumentation__.Notify(235182)
	}
	__antithesis_instrumentation__.Notify(235171)

	s.mu.Lock()
	defer s.mu.Unlock()
	_, hasOverride := s.mu.overrides[name]
	if tombstone {
		__antithesis_instrumentation__.Notify(235183)

		delete(s.mu.values, name)
		if !hasOverride {
			__antithesis_instrumentation__.Notify(235184)
			s.setDefaultLocked(ctx, name)
		} else {
			__antithesis_instrumentation__.Notify(235185)
		}
	} else {
		__antithesis_instrumentation__.Notify(235186)
		s.mu.values[name] = val
		if !hasOverride {
			__antithesis_instrumentation__.Notify(235187)
			s.setLocked(ctx, name, val)
		} else {
			__antithesis_instrumentation__.Notify(235188)
		}
	}
	__antithesis_instrumentation__.Notify(235172)
	if s.storage != nil {
		__antithesis_instrumentation__.Notify(235189)
		return kv
	} else {
		__antithesis_instrumentation__.Notify(235190)
	}
	__antithesis_instrumentation__.Notify(235173)
	return nil
}

const versionSettingKey = "version"

func (s *SettingsWatcher) setLocked(ctx context.Context, key string, val settings.EncodedValue) {
	__antithesis_instrumentation__.Notify(235191)

	if key == versionSettingKey && func() bool {
		__antithesis_instrumentation__.Notify(235193)
		return !s.codec.ForSystemTenant() == true
	}() == true {
		__antithesis_instrumentation__.Notify(235194)
		var v clusterversion.ClusterVersion
		if err := protoutil.Unmarshal([]byte(val.Value), &v); err != nil {
			__antithesis_instrumentation__.Notify(235196)
			log.Warningf(ctx, "failed to set cluster version: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(235197)
			if err := s.settings.Version.SetActiveVersion(ctx, v); err != nil {
				__antithesis_instrumentation__.Notify(235198)
				log.Warningf(ctx, "failed to set cluster version: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(235199)
				log.Infof(ctx, "set cluster version to: %v", v)
			}
		}
		__antithesis_instrumentation__.Notify(235195)
		return
	} else {
		__antithesis_instrumentation__.Notify(235200)
	}
	__antithesis_instrumentation__.Notify(235192)

	if err := s.mu.updater.Set(ctx, key, val); err != nil {
		__antithesis_instrumentation__.Notify(235201)
		log.Warningf(ctx, "failed to set setting %s to %s: %v", redact.Safe(key), val.Value, err)
	} else {
		__antithesis_instrumentation__.Notify(235202)
	}
}

func (s *SettingsWatcher) setDefaultLocked(ctx context.Context, key string) {
	__antithesis_instrumentation__.Notify(235203)
	setting, ok := settings.Lookup(key, settings.LookupForLocalAccess, s.codec.ForSystemTenant())
	if !ok {
		__antithesis_instrumentation__.Notify(235206)
		log.Warningf(ctx, "failed to find setting %s, skipping update", redact.Safe(key))
		return
	} else {
		__antithesis_instrumentation__.Notify(235207)
	}
	__antithesis_instrumentation__.Notify(235204)
	ws, ok := setting.(settings.NonMaskedSetting)
	if !ok {
		__antithesis_instrumentation__.Notify(235208)
		log.Fatalf(ctx, "expected non-masked setting, got %T", s)
	} else {
		__antithesis_instrumentation__.Notify(235209)
	}
	__antithesis_instrumentation__.Notify(235205)
	val := settings.EncodedValue{
		Value: ws.EncodedDefault(),
		Type:  ws.Typ(),
	}
	s.setLocked(ctx, key, val)
}

func (s *SettingsWatcher) updateOverrides(ctx context.Context) {
	__antithesis_instrumentation__.Notify(235210)
	newOverrides := s.overridesMonitor.Overrides()

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, val := range newOverrides {
		__antithesis_instrumentation__.Notify(235212)
		if key == versionSettingKey {
			__antithesis_instrumentation__.Notify(235215)
			log.Warningf(ctx, "ignoring attempt to override %s", key)
			continue
		} else {
			__antithesis_instrumentation__.Notify(235216)
		}
		__antithesis_instrumentation__.Notify(235213)
		if oldVal, hasExisting := s.mu.overrides[key]; hasExisting && func() bool {
			__antithesis_instrumentation__.Notify(235217)
			return oldVal == val == true
		}() == true {
			__antithesis_instrumentation__.Notify(235218)

			continue
		} else {
			__antithesis_instrumentation__.Notify(235219)
		}
		__antithesis_instrumentation__.Notify(235214)

		s.mu.overrides[key] = val
		s.setLocked(ctx, key, val)
	}
	__antithesis_instrumentation__.Notify(235211)

	for key := range s.mu.overrides {
		__antithesis_instrumentation__.Notify(235220)
		if _, ok := newOverrides[key]; !ok {
			__antithesis_instrumentation__.Notify(235221)
			delete(s.mu.overrides, key)

			if val, ok := s.mu.values[key]; ok {
				__antithesis_instrumentation__.Notify(235222)
				s.setLocked(ctx, key, val)
			} else {
				__antithesis_instrumentation__.Notify(235223)
				s.setDefaultLocked(ctx, key)
			}
		} else {
			__antithesis_instrumentation__.Notify(235224)
		}
	}
}

func (s *SettingsWatcher) resetUpdater() {
	__antithesis_instrumentation__.Notify(235225)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.updater = s.settings.MakeUpdater()
}

func (s *SettingsWatcher) SetTestingKnobs(knobs *rangefeedcache.TestingKnobs) {
	__antithesis_instrumentation__.Notify(235226)
	s.testingWatcherKnobs = knobs
}

func (s *SettingsWatcher) IsOverridden(settingName string) bool {
	__antithesis_instrumentation__.Notify(235227)
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.mu.overrides[settingName]
	return exists
}
