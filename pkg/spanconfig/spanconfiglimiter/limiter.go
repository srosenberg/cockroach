// Package spanconfiglimiter is used to limit how many span configs are
// installed by tenants.
package spanconfiglimiter

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/errors"
)

var _ spanconfig.Limiter = &Limiter{}

var tenantLimitSetting = settings.RegisterIntSetting(
	settings.TenantReadOnly,
	"spanconfig.tenant_limit",
	"limit on the number of span configs a tenant is allowed to install",
	5000,
)

type Limiter struct {
	ie       sqlutil.InternalExecutor
	settings *cluster.Settings
	knobs    *spanconfig.TestingKnobs
}

func New(
	ie sqlutil.InternalExecutor, settings *cluster.Settings, knobs *spanconfig.TestingKnobs,
) *Limiter {
	__antithesis_instrumentation__.Notify(240683)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(240685)
		knobs = &spanconfig.TestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(240686)
	}
	__antithesis_instrumentation__.Notify(240684)
	return &Limiter{
		ie:       ie,
		settings: settings,
		knobs:    knobs,
	}
}

func (l *Limiter) ShouldLimit(ctx context.Context, txn *kv.Txn, delta int) (bool, error) {
	__antithesis_instrumentation__.Notify(240687)
	if !l.settings.Version.IsActive(ctx, clusterversion.PreSeedSpanCountTable) {
		__antithesis_instrumentation__.Notify(240694)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(240695)
	}
	__antithesis_instrumentation__.Notify(240688)

	if delta == 0 {
		__antithesis_instrumentation__.Notify(240696)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(240697)
	}
	__antithesis_instrumentation__.Notify(240689)

	limit := tenantLimitSetting.Get(&l.settings.SV)
	if overrideFn := l.knobs.LimiterLimitOverride; overrideFn != nil {
		__antithesis_instrumentation__.Notify(240698)
		limit = overrideFn()
	} else {
		__antithesis_instrumentation__.Notify(240699)
	}
	__antithesis_instrumentation__.Notify(240690)

	const updateSpanCountStmt = `
INSERT INTO system.span_count (span_count) VALUES ($1)
ON CONFLICT (singleton)
DO UPDATE SET span_count = system.span_count.span_count + $1
RETURNING span_count
`
	datums, err := l.ie.QueryRowEx(ctx, "update-span-count", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		updateSpanCountStmt, delta)
	if err != nil {
		__antithesis_instrumentation__.Notify(240700)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(240701)
	}
	__antithesis_instrumentation__.Notify(240691)
	if len(datums) != 1 {
		__antithesis_instrumentation__.Notify(240702)
		return false, errors.AssertionFailedf("expected to return 1 row, return %d", len(datums))
	} else {
		__antithesis_instrumentation__.Notify(240703)
	}
	__antithesis_instrumentation__.Notify(240692)

	if delta < 0 {
		__antithesis_instrumentation__.Notify(240704)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(240705)
	}
	__antithesis_instrumentation__.Notify(240693)
	spanCountWithDelta := int64(tree.MustBeDInt(datums[0]))
	return spanCountWithDelta > limit, nil
}
