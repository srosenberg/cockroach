package tenantrate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type TestingKnobs struct {
	TimeSource timeutil.TimeSource
}

type LimiterFactory struct {
	knobs         TestingKnobs
	metrics       Metrics
	systemLimiter systemLimiter
	mu            struct {
		syncutil.RWMutex
		config  Config
		tenants map[roachpb.TenantID]*refCountedLimiter
	}
}

type refCountedLimiter struct {
	refCount int
	lim      limiter
}

func NewLimiterFactory(sv *settings.Values, knobs *TestingKnobs) *LimiterFactory {
	__antithesis_instrumentation__.Notify(126644)
	rl := &LimiterFactory{
		metrics: makeMetrics(),
	}
	if knobs != nil {
		__antithesis_instrumentation__.Notify(126648)
		rl.knobs = *knobs
	} else {
		__antithesis_instrumentation__.Notify(126649)
	}
	__antithesis_instrumentation__.Notify(126645)
	rl.mu.tenants = make(map[roachpb.TenantID]*refCountedLimiter)
	rl.mu.config = ConfigFromSettings(sv)
	rl.systemLimiter = systemLimiter{
		tenantMetrics: rl.metrics.tenantMetrics(roachpb.SystemTenantID),
	}
	updateFn := func(_ context.Context) {
		__antithesis_instrumentation__.Notify(126650)
		config := ConfigFromSettings(sv)
		rl.UpdateConfig(config)
	}
	__antithesis_instrumentation__.Notify(126646)
	for _, setting := range configSettings {
		__antithesis_instrumentation__.Notify(126651)
		setting.SetOnChange(sv, updateFn)
	}
	__antithesis_instrumentation__.Notify(126647)
	return rl
}

func (rl *LimiterFactory) GetTenant(
	ctx context.Context, tenantID roachpb.TenantID, closer <-chan struct{},
) Limiter {
	__antithesis_instrumentation__.Notify(126652)

	if tenantID == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(126655)
		return &rl.systemLimiter
	} else {
		__antithesis_instrumentation__.Notify(126656)
	}
	__antithesis_instrumentation__.Notify(126653)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rcLim, ok := rl.mu.tenants[tenantID]
	if !ok {
		__antithesis_instrumentation__.Notify(126657)
		var options []quotapool.Option
		if rl.knobs.TimeSource != nil {
			__antithesis_instrumentation__.Notify(126660)
			options = append(options, quotapool.WithTimeSource(rl.knobs.TimeSource))
		} else {
			__antithesis_instrumentation__.Notify(126661)
		}
		__antithesis_instrumentation__.Notify(126658)
		if closer != nil {
			__antithesis_instrumentation__.Notify(126662)
			options = append(options, quotapool.WithCloser(closer))
		} else {
			__antithesis_instrumentation__.Notify(126663)
		}
		__antithesis_instrumentation__.Notify(126659)
		rcLim = new(refCountedLimiter)
		rcLim.lim.init(rl, tenantID, rl.mu.config, rl.metrics.tenantMetrics(tenantID), options...)
		rl.mu.tenants[tenantID] = rcLim
		log.Infof(
			ctx, "tenant %s rate limiter initialized (rate: %g RU/s; burst: %g RU)",
			tenantID, rl.mu.config.Rate, rl.mu.config.Burst,
		)
	} else {
		__antithesis_instrumentation__.Notify(126664)
	}
	__antithesis_instrumentation__.Notify(126654)
	rcLim.refCount++
	return &rcLim.lim
}

func (rl *LimiterFactory) Release(lim Limiter) {
	__antithesis_instrumentation__.Notify(126665)
	if _, isSystem := lim.(*systemLimiter); isSystem {
		__antithesis_instrumentation__.Notify(126669)
		return
	} else {
		__antithesis_instrumentation__.Notify(126670)
	}
	__antithesis_instrumentation__.Notify(126666)
	l := lim.(*limiter)
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rcLim, ok := rl.mu.tenants[l.tenantID]
	if !ok {
		__antithesis_instrumentation__.Notify(126671)
		panic(errors.AssertionFailedf("expected to find entry for tenant %v", l.tenantID))
	} else {
		__antithesis_instrumentation__.Notify(126672)
	}
	__antithesis_instrumentation__.Notify(126667)
	if &rcLim.lim != lim {
		__antithesis_instrumentation__.Notify(126673)
		panic(errors.AssertionFailedf("two limiters exist for tenant %v", l.tenantID))
	} else {
		__antithesis_instrumentation__.Notify(126674)
	}
	__antithesis_instrumentation__.Notify(126668)
	if rcLim.refCount--; rcLim.refCount == 0 {
		__antithesis_instrumentation__.Notify(126675)
		l.metrics.destroy()
		delete(rl.mu.tenants, l.tenantID)
	} else {
		__antithesis_instrumentation__.Notify(126676)
	}
}

func (rl *LimiterFactory) UpdateConfig(config Config) {
	__antithesis_instrumentation__.Notify(126677)
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.mu.config = config
	for _, rcLim := range rl.mu.tenants {
		__antithesis_instrumentation__.Notify(126678)
		rcLim.lim.updateConfig(rl.mu.config)
	}
}

func (rl *LimiterFactory) Metrics() *Metrics {
	__antithesis_instrumentation__.Notify(126679)
	return &rl.metrics
}
