// Package gcjobnotifier provides a mechanism to share a SystemConfigDeltaFilter
// among all gc jobs.
//
// It exists in a separate package to avoid import cycles between sql and gcjob.
package gcjobnotifier

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type Notifier struct {
	provider config.SystemConfigProvider
	prefix   roachpb.Key
	stopper  *stop.Stopper
	settings *cluster.Settings
	mu       struct {
		syncutil.Mutex
		started, stopped bool
		deltaFilter      *gossip.SystemConfigDeltaFilter
		notifyees        map[chan struct{}]struct{}
	}
}

func New(
	settings *cluster.Settings,
	provider config.SystemConfigProvider,
	codec keys.SQLCodec,
	stopper *stop.Stopper,
) *Notifier {
	__antithesis_instrumentation__.Notify(492421)
	n := &Notifier{
		provider: provider,
		prefix:   codec.IndexPrefix(keys.ZonesTableID, keys.ZonesTablePrimaryIndexID),
		stopper:  stopper,
		settings: settings,
	}
	n.mu.notifyees = make(map[chan struct{}]struct{})
	return n
}

func noopFunc() { __antithesis_instrumentation__.Notify(492422) }

func (n *Notifier) AddNotifyee(ctx context.Context) (onChange <-chan struct{}, cleanup func()) {
	__antithesis_instrumentation__.Notify(492423)
	n.mu.Lock()
	defer n.mu.Unlock()
	if !n.mu.started {
		__antithesis_instrumentation__.Notify(492427)
		logcrash.ReportOrPanic(ctx, &n.settings.SV,
			"adding a notifyee to a Notifier before starting")
	} else {
		__antithesis_instrumentation__.Notify(492428)
	}
	__antithesis_instrumentation__.Notify(492424)
	if n.mu.stopped {
		__antithesis_instrumentation__.Notify(492429)
		return nil, noopFunc
	} else {
		__antithesis_instrumentation__.Notify(492430)
	}
	__antithesis_instrumentation__.Notify(492425)
	if n.mu.deltaFilter == nil {
		__antithesis_instrumentation__.Notify(492431)
		zoneCfgFilter := gossip.MakeSystemConfigDeltaFilter(n.prefix)
		n.mu.deltaFilter = &zoneCfgFilter

		cfg := n.provider.GetSystemConfig()
		if cfg != nil {
			__antithesis_instrumentation__.Notify(492432)
			n.mu.deltaFilter.ForModified(cfg, func(kv roachpb.KeyValue) { __antithesis_instrumentation__.Notify(492433) })
		} else {
			__antithesis_instrumentation__.Notify(492434)
		}
	} else {
		__antithesis_instrumentation__.Notify(492435)
	}
	__antithesis_instrumentation__.Notify(492426)
	c := make(chan struct{}, 1)
	n.mu.notifyees[c] = struct{}{}
	return c, func() { __antithesis_instrumentation__.Notify(492436); n.cleanup(c) }
}

func (n *Notifier) cleanup(c chan struct{}) {
	__antithesis_instrumentation__.Notify(492437)
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.mu.notifyees, c)
	if len(n.mu.notifyees) == 0 {
		__antithesis_instrumentation__.Notify(492438)
		n.mu.deltaFilter = nil
	} else {
		__antithesis_instrumentation__.Notify(492439)
	}
}

func (n *Notifier) markStopped() {
	__antithesis_instrumentation__.Notify(492440)
	n.mu.Lock()
	defer n.mu.Unlock()
	n.mu.stopped = true
}

func (n *Notifier) markStarted() (alreadyStarted bool) {
	__antithesis_instrumentation__.Notify(492441)
	n.mu.Lock()
	defer n.mu.Unlock()
	alreadyStarted = n.mu.started
	n.mu.started = true
	return alreadyStarted
}

func (n *Notifier) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(492442)
	if alreadyStarted := n.markStarted(); alreadyStarted {
		__antithesis_instrumentation__.Notify(492444)
		logcrash.ReportOrPanic(ctx, &n.settings.SV, "started Notifier more than once")
		return
	} else {
		__antithesis_instrumentation__.Notify(492445)
	}
	__antithesis_instrumentation__.Notify(492443)
	if err := n.stopper.RunAsyncTask(ctx, "gcjob.Notifier", n.run); err != nil {
		__antithesis_instrumentation__.Notify(492446)
		n.markStopped()
	} else {
		__antithesis_instrumentation__.Notify(492447)
	}
}

func (n *Notifier) run(_ context.Context) {
	__antithesis_instrumentation__.Notify(492448)
	defer n.markStopped()
	systemConfigUpdateCh, _ := n.provider.RegisterSystemConfigChannel()
	for {
		__antithesis_instrumentation__.Notify(492449)
		select {
		case <-n.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(492450)
			return
		case <-systemConfigUpdateCh:
			__antithesis_instrumentation__.Notify(492451)
			n.maybeNotify()
		}
	}
}

func (n *Notifier) maybeNotify() {
	__antithesis_instrumentation__.Notify(492452)
	n.mu.Lock()
	defer n.mu.Unlock()

	if len(n.mu.notifyees) == 0 {
		__antithesis_instrumentation__.Notify(492456)
		return
	} else {
		__antithesis_instrumentation__.Notify(492457)
	}
	__antithesis_instrumentation__.Notify(492453)

	cfg := n.provider.GetSystemConfig()
	zoneConfigUpdated := false
	n.mu.deltaFilter.ForModified(cfg, func(kv roachpb.KeyValue) {
		__antithesis_instrumentation__.Notify(492458)
		zoneConfigUpdated = true
	})
	__antithesis_instrumentation__.Notify(492454)

	if !zoneConfigUpdated {
		__antithesis_instrumentation__.Notify(492459)
		return
	} else {
		__antithesis_instrumentation__.Notify(492460)
	}
	__antithesis_instrumentation__.Notify(492455)

	for c := range n.mu.notifyees {
		__antithesis_instrumentation__.Notify(492461)
		select {
		case c <- struct{}{}:
			__antithesis_instrumentation__.Notify(492462)
		default:
			__antithesis_instrumentation__.Notify(492463)
		}
	}
}
