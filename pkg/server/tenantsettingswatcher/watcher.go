package tenantsettingswatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedbuffer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedcache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

type Watcher struct {
	clock   *hlc.Clock
	f       *rangefeed.Factory
	stopper *stop.Stopper
	st      *cluster.Settings
	dec     RowDecoder
	store   overridesStore

	startCh  chan struct{}
	startErr error
}

func New(
	clock *hlc.Clock, f *rangefeed.Factory, stopper *stop.Stopper, st *cluster.Settings,
) *Watcher {
	__antithesis_instrumentation__.Notify(238939)
	w := &Watcher{
		clock:   clock,
		f:       f,
		stopper: stopper,
		st:      st,
		dec:     MakeRowDecoder(),
	}
	w.store.Init()
	return w
}

func (w *Watcher) Start(ctx context.Context, sysTableResolver catalog.SystemTableIDResolver) error {
	__antithesis_instrumentation__.Notify(238940)
	w.startCh = make(chan struct{})
	if w.st.Version.IsActive(ctx, clusterversion.TenantSettingsTable) {
		__antithesis_instrumentation__.Notify(238944)

		w.startErr = w.startRangeFeed(ctx, sysTableResolver)
		close(w.startCh)
		return w.startErr
	} else {
		__antithesis_instrumentation__.Notify(238945)
	}
	__antithesis_instrumentation__.Notify(238941)

	versionOkCh := make(chan struct{})
	var once sync.Once
	w.st.Version.SetOnChange(func(ctx context.Context, newVersion clusterversion.ClusterVersion) {
		__antithesis_instrumentation__.Notify(238946)
		if newVersion.IsActive(clusterversion.TenantSettingsTable) {
			__antithesis_instrumentation__.Notify(238947)
			once.Do(func() {
				__antithesis_instrumentation__.Notify(238948)
				close(versionOkCh)
			})
		} else {
			__antithesis_instrumentation__.Notify(238949)
		}
	})
	__antithesis_instrumentation__.Notify(238942)

	if w.st.Version.IsActive(ctx, clusterversion.TenantSettingsTable) {
		__antithesis_instrumentation__.Notify(238950)
		w.startErr = w.startRangeFeed(ctx, sysTableResolver)
		close(w.startCh)
		return w.startErr
	} else {
		__antithesis_instrumentation__.Notify(238951)
	}
	__antithesis_instrumentation__.Notify(238943)
	return w.stopper.RunAsyncTask(ctx, "tenantsettingswatcher-start", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(238952)
		log.Infof(ctx, "tenantsettingswatcher waiting for the appropriate version")
		select {
		case <-versionOkCh:
			__antithesis_instrumentation__.Notify(238955)
		case <-w.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(238956)
			return
		}
		__antithesis_instrumentation__.Notify(238953)
		log.Infof(ctx, "tenantsettingswatcher can now start")
		w.startErr = w.startRangeFeed(ctx, sysTableResolver)
		if w.startErr != nil {
			__antithesis_instrumentation__.Notify(238957)

			log.Warningf(ctx, "error starting tenantsettingswatcher rangefeed: %v", w.startErr)
		} else {
			__antithesis_instrumentation__.Notify(238958)
		}
		__antithesis_instrumentation__.Notify(238954)
		close(w.startCh)
	})
}

func (w *Watcher) startRangeFeed(
	ctx context.Context, sysTableResolver catalog.SystemTableIDResolver,
) error {
	__antithesis_instrumentation__.Notify(238959)
	tableID, err := sysTableResolver.LookupSystemTableID(ctx, systemschema.TenantSettingsTable.GetName())
	if err != nil {
		__antithesis_instrumentation__.Notify(238965)
		return err
	} else {
		__antithesis_instrumentation__.Notify(238966)
	}
	__antithesis_instrumentation__.Notify(238960)
	tenantSettingsTablePrefix := keys.SystemSQLCodec.TablePrefix(uint32(tableID))
	tenantSettingsTableSpan := roachpb.Span{
		Key:    tenantSettingsTablePrefix,
		EndKey: tenantSettingsTablePrefix.PrefixEnd(),
	}

	var initialScan = struct {
		ch   chan struct{}
		done bool
		err  error
	}{
		ch: make(chan struct{}),
	}

	allOverrides := make(map[roachpb.TenantID][]roachpb.TenantSetting)

	translateEvent := func(ctx context.Context, kv *roachpb.RangeFeedValue) rangefeedbuffer.Event {
		__antithesis_instrumentation__.Notify(238967)
		tenantID, setting, tombstone, err := w.dec.DecodeRow(roachpb.KeyValue{
			Key:   kv.Key,
			Value: kv.Value,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(238970)
			log.Warningf(ctx, "failed to decode settings row %v: %v", kv.Key, err)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(238971)
		}
		__antithesis_instrumentation__.Notify(238968)
		if allOverrides != nil {
			__antithesis_instrumentation__.Notify(238972)

			if tombstone {
				__antithesis_instrumentation__.Notify(238974)
				log.Warning(ctx, "unexpected empty value during rangefeed scan")
				return nil
			} else {
				__antithesis_instrumentation__.Notify(238975)
			}
			__antithesis_instrumentation__.Notify(238973)
			allOverrides[tenantID] = append(allOverrides[tenantID], setting)
		} else {
			__antithesis_instrumentation__.Notify(238976)

			w.store.SetTenantOverride(tenantID, setting)
		}
		__antithesis_instrumentation__.Notify(238969)
		return nil
	}
	__antithesis_instrumentation__.Notify(238961)

	onUpdate := func(ctx context.Context, update rangefeedcache.Update) {
		__antithesis_instrumentation__.Notify(238977)
		if update.Type == rangefeedcache.CompleteUpdate {
			__antithesis_instrumentation__.Notify(238978)

			w.store.SetAll(allOverrides)
			allOverrides = nil

			if !initialScan.done {
				__antithesis_instrumentation__.Notify(238979)
				initialScan.done = true
				close(initialScan.ch)
			} else {
				__antithesis_instrumentation__.Notify(238980)
			}
		} else {
			__antithesis_instrumentation__.Notify(238981)
		}
	}
	__antithesis_instrumentation__.Notify(238962)

	onError := func(err error) {
		__antithesis_instrumentation__.Notify(238982)
		if !initialScan.done {
			__antithesis_instrumentation__.Notify(238983)
			initialScan.err = err
			initialScan.done = true
			close(initialScan.ch)
		} else {
			__antithesis_instrumentation__.Notify(238984)

			allOverrides = make(map[roachpb.TenantID][]roachpb.TenantSetting)
		}
	}
	__antithesis_instrumentation__.Notify(238963)

	c := rangefeedcache.NewWatcher(
		"tenant-settings-watcher",
		w.clock, w.f,
		0,
		[]roachpb.Span{tenantSettingsTableSpan},
		false,
		translateEvent,
		onUpdate,
		nil,
	)

	if err := rangefeedcache.Start(ctx, w.stopper, c, onError); err != nil {
		__antithesis_instrumentation__.Notify(238985)
		return err
	} else {
		__antithesis_instrumentation__.Notify(238986)
	}
	__antithesis_instrumentation__.Notify(238964)

	select {
	case <-initialScan.ch:
		__antithesis_instrumentation__.Notify(238987)
		return initialScan.err

	case <-w.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(238988)
		return errors.Wrap(stop.ErrUnavailable, "failed to retrieve initial tenant settings")

	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(238989)
		return errors.Wrap(ctx.Err(), "failed to retrieve initial tenant settings")
	}
}

func (w *Watcher) WaitForStart(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(238990)

	select {
	case <-w.startCh:
		__antithesis_instrumentation__.Notify(238994)
		return w.startErr
	default:
		__antithesis_instrumentation__.Notify(238995)
	}
	__antithesis_instrumentation__.Notify(238991)
	if w.startCh == nil {
		__antithesis_instrumentation__.Notify(238996)
		return errors.AssertionFailedf("Start() was not yet called")
	} else {
		__antithesis_instrumentation__.Notify(238997)
	}
	__antithesis_instrumentation__.Notify(238992)
	if !w.st.Version.IsActive(ctx, clusterversion.TenantSettingsTable) {
		__antithesis_instrumentation__.Notify(238998)

		log.Warningf(ctx, "tenant requested settings before host cluster version upgrade")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(238999)
	}
	__antithesis_instrumentation__.Notify(238993)
	select {
	case <-w.startCh:
		__antithesis_instrumentation__.Notify(239000)
		return w.startErr
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(239001)
		return ctx.Err()
	}
}

func (w *Watcher) GetTenantOverrides(
	tenantID roachpb.TenantID,
) (overrides []roachpb.TenantSetting, changeCh <-chan struct{}) {
	__antithesis_instrumentation__.Notify(239002)
	o := w.store.GetTenantOverrides(tenantID)
	return o.overrides, o.changeCh
}

func (w *Watcher) GetAllTenantOverrides() (
	overrides []roachpb.TenantSetting,
	changeCh <-chan struct{},
) {
	__antithesis_instrumentation__.Notify(239003)
	return w.GetTenantOverrides(allTenantOverridesID)
}
