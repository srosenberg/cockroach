package tenantcostserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type instance struct {
	db         *kv.DB
	executor   *sql.InternalExecutor
	metrics    Metrics
	timeSource timeutil.TimeSource
	settings   *cluster.Settings
}

var instanceInactivity = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"tenant_usage_instance_inactivity",
	"instances that have not reported consumption for longer than this value are cleaned up; "+
		"should be at least four times higher than the tenant_cost_control_period of any tenant",
	1*time.Minute, settings.PositiveDuration,
)

func newInstance(
	settings *cluster.Settings,
	db *kv.DB,
	executor *sql.InternalExecutor,
	timeSource timeutil.TimeSource,
) *instance {
	__antithesis_instrumentation__.Notify(20224)
	res := &instance{
		db:         db,
		executor:   executor,
		timeSource: timeSource,
		settings:   settings,
	}
	res.metrics.init()
	return res
}

func (s *instance) Metrics() metric.Struct {
	__antithesis_instrumentation__.Notify(20225)
	return &s.metrics
}

var _ multitenant.TenantUsageServer = (*instance)(nil)

func init() {
	server.NewTenantUsageServer = func(
		settings *cluster.Settings,
		db *kv.DB,
		executor *sql.InternalExecutor,
	) multitenant.TenantUsageServer {
		return newInstance(settings, db, executor, timeutil.DefaultTimeSource{})
	}
}
