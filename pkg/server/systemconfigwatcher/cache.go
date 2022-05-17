package systemconfigwatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedbuffer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedcache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type Cache struct {
	w                   *rangefeedcache.Watcher
	defaultZoneConfig   *zonepb.ZoneConfig
	additionalKVsSource config.SystemConfigProvider
	mu                  struct {
		syncutil.RWMutex

		cfg       *config.SystemConfig
		timestamp hlc.Timestamp

		registry notificationRegistry

		additionalKVs []roachpb.KeyValue
	}
}

func New(
	codec keys.SQLCodec, clock *hlc.Clock, f *rangefeed.Factory, defaultZoneConfig *zonepb.ZoneConfig,
) *Cache {
	__antithesis_instrumentation__.Notify(238089)
	return NewWithAdditionalProvider(
		codec, clock, f, defaultZoneConfig, nil,
	)
}

func NewWithAdditionalProvider(
	codec keys.SQLCodec,
	clock *hlc.Clock,
	f *rangefeed.Factory,
	defaultZoneConfig *zonepb.ZoneConfig,
	additional config.SystemConfigProvider,
) *Cache {
	__antithesis_instrumentation__.Notify(238090)

	const bufferSize = 1 << 20
	const withPrevValue = false
	c := Cache{
		defaultZoneConfig: defaultZoneConfig,
	}
	c.mu.registry = notificationRegistry{}
	c.additionalKVsSource = additional

	span := roachpb.Span{
		Key:    append(codec.TenantPrefix(), keys.SystemConfigSplitKey...),
		EndKey: append(codec.TenantPrefix(), keys.SystemConfigTableDataMax...),
	}
	c.w = rangefeedcache.NewWatcher(
		"system-config-cache", clock, f,
		bufferSize,
		[]roachpb.Span{span},
		withPrevValue,
		passThroughTranslation,
		c.handleUpdate,
		nil)
	return &c
}

func (c *Cache) Start(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(238091)
	if err := rangefeedcache.Start(ctx, stopper, c.w, nil); err != nil {
		__antithesis_instrumentation__.Notify(238094)
		return err
	} else {
		__antithesis_instrumentation__.Notify(238095)
	}
	__antithesis_instrumentation__.Notify(238092)
	if c.additionalKVsSource != nil {
		__antithesis_instrumentation__.Notify(238096)
		setAdditionalKeys := func() {
			__antithesis_instrumentation__.Notify(238099)
			if cfg := c.additionalKVsSource.GetSystemConfig(); cfg != nil {
				__antithesis_instrumentation__.Notify(238100)
				c.setAdditionalKeys(cfg.Values)
			} else {
				__antithesis_instrumentation__.Notify(238101)
			}
		}
		__antithesis_instrumentation__.Notify(238097)
		ch, unregister := c.additionalKVsSource.RegisterSystemConfigChannel()

		select {
		case <-ch:
			__antithesis_instrumentation__.Notify(238102)
			setAdditionalKeys()
		default:
			__antithesis_instrumentation__.Notify(238103)
		}
		__antithesis_instrumentation__.Notify(238098)
		if err := stopper.RunAsyncTask(ctx, "systemconfigwatcher-additional", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(238104)
			for {
				__antithesis_instrumentation__.Notify(238105)
				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(238106)
					return
				case <-stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(238107)
					return
				case <-ch:
					__antithesis_instrumentation__.Notify(238108)
					setAdditionalKeys()
				}
			}
		}); err != nil {
			__antithesis_instrumentation__.Notify(238109)
			unregister()
			return err
		} else {
			__antithesis_instrumentation__.Notify(238110)
		}
	} else {
		__antithesis_instrumentation__.Notify(238111)
	}
	__antithesis_instrumentation__.Notify(238093)
	return nil
}

func (c *Cache) GetSystemConfig() *config.SystemConfig {
	__antithesis_instrumentation__.Notify(238112)
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mu.cfg
}

func (c *Cache) RegisterSystemConfigChannel() (_ <-chan struct{}, unregister func()) {
	__antithesis_instrumentation__.Notify(238113)
	ch := make(chan struct{}, 1)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mu.cfg != nil {
		__antithesis_instrumentation__.Notify(238115)
		ch <- struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(238116)
	}
	__antithesis_instrumentation__.Notify(238114)

	c.mu.registry[ch] = struct{}{}
	return ch, func() {
		__antithesis_instrumentation__.Notify(238117)
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.mu.registry, ch)
	}
}

func (c *Cache) LastUpdated() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(238118)
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mu.timestamp
}

func (c *Cache) setAdditionalKeys(kvs []roachpb.KeyValue) {
	__antithesis_instrumentation__.Notify(238119)
	c.mu.Lock()
	defer c.mu.Unlock()

	sort.Sort(keyValues(kvs))
	if c.mu.cfg == nil {
		__antithesis_instrumentation__.Notify(238121)
		c.mu.additionalKVs = kvs
		return
	} else {
		__antithesis_instrumentation__.Notify(238122)
	}
	__antithesis_instrumentation__.Notify(238120)

	cloned := append([]roachpb.KeyValue(nil), c.mu.cfg.Values...)
	trimmed := append(trimOldKVs(cloned, c.mu.additionalKVs), kvs...)
	sort.Sort(keyValues(trimmed))
	c.mu.cfg = config.NewSystemConfig(c.defaultZoneConfig)
	c.mu.cfg.Values = trimmed
	c.mu.additionalKVs = kvs
	c.mu.registry.notify()
}

func trimOldKVs(cloned, prev []roachpb.KeyValue) []roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(238123)
	trimmed := cloned[:0]
	shouldSkip := func(clonedOrd int) (shouldSkip bool) {
		__antithesis_instrumentation__.Notify(238126)
		for len(prev) > 0 {
			__antithesis_instrumentation__.Notify(238128)
			if cmp := prev[0].Key.Compare(cloned[clonedOrd].Key); cmp >= 0 {
				__antithesis_instrumentation__.Notify(238130)
				return cmp == 0
			} else {
				__antithesis_instrumentation__.Notify(238131)
			}
			__antithesis_instrumentation__.Notify(238129)
			prev = prev[1:]
		}
		__antithesis_instrumentation__.Notify(238127)
		return false
	}
	__antithesis_instrumentation__.Notify(238124)
	for i := range cloned {
		__antithesis_instrumentation__.Notify(238132)
		if !shouldSkip(i) {
			__antithesis_instrumentation__.Notify(238133)
			trimmed = append(trimmed, cloned[i])
		} else {
			__antithesis_instrumentation__.Notify(238134)
		}
	}
	__antithesis_instrumentation__.Notify(238125)
	return trimmed
}

type keyValues []roachpb.KeyValue

func (k keyValues) Len() int { __antithesis_instrumentation__.Notify(238135); return len(k) }
func (k keyValues) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(238136)
	k[i], k[j] = k[j], k[i]
}
func (k keyValues) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(238137)
	return k[i].Key.Compare(k[j].Key) < 0
}

var _ sort.Interface = (keyValues)(nil)

func (c *Cache) handleUpdate(_ context.Context, update rangefeedcache.Update) {
	__antithesis_instrumentation__.Notify(238138)
	updateKVs := rangefeedbuffer.EventsToKVs(update.Events,
		rangefeedbuffer.RangeFeedValueEventToKV)
	c.mu.Lock()
	defer c.mu.Unlock()
	var updatedData []roachpb.KeyValue
	switch update.Type {
	case rangefeedcache.CompleteUpdate:
		__antithesis_instrumentation__.Notify(238140)
		updatedData = rangefeedbuffer.MergeKVs(c.mu.additionalKVs, updateKVs)
	case rangefeedcache.IncrementalUpdate:
		__antithesis_instrumentation__.Notify(238141)

		prev := c.mu.cfg

		if len(updateKVs) == 0 {
			__antithesis_instrumentation__.Notify(238144)
			c.setUpdatedConfigLocked(prev, update.Timestamp)
			return
		} else {
			__antithesis_instrumentation__.Notify(238145)
		}
		__antithesis_instrumentation__.Notify(238142)
		updatedData = rangefeedbuffer.MergeKVs(prev.Values, updateKVs)
	default:
		__antithesis_instrumentation__.Notify(238143)
	}
	__antithesis_instrumentation__.Notify(238139)

	updatedCfg := config.NewSystemConfig(c.defaultZoneConfig)
	updatedCfg.Values = updatedData
	c.setUpdatedConfigLocked(updatedCfg, update.Timestamp)
}

func (c *Cache) setUpdatedConfigLocked(updated *config.SystemConfig, ts hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(238146)
	changed := c.mu.cfg != updated
	c.mu.cfg = updated
	c.mu.timestamp = ts
	if changed {
		__antithesis_instrumentation__.Notify(238147)
		c.mu.registry.notify()
	} else {
		__antithesis_instrumentation__.Notify(238148)
	}
}

func passThroughTranslation(
	ctx context.Context, value *roachpb.RangeFeedValue,
) rangefeedbuffer.Event {
	__antithesis_instrumentation__.Notify(238149)
	return value
}

var _ config.SystemConfigProvider = (*Cache)(nil)

type notificationRegistry map[chan<- struct{}]struct{}

func (nr notificationRegistry) notify() {
	__antithesis_instrumentation__.Notify(238150)
	for ch := range nr {
		__antithesis_instrumentation__.Notify(238151)
		select {
		case ch <- struct{}{}:
			__antithesis_instrumentation__.Notify(238152)
		default:
			__antithesis_instrumentation__.Notify(238153)
		}
	}
}
