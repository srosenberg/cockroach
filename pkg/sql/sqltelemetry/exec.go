package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

var DistSQLExecCounter = telemetry.GetCounterOnce("sql.exec.query.is-distributed")

var VecExecCounter = telemetry.GetCounterOnce("sql.exec.query.is-vectorized")

func VecModeCounter(mode string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625782)
	return telemetry.GetCounter(fmt.Sprintf("sql.exec.vectorized-setting.%s", mode))
}

var CascadesLimitReached = telemetry.GetCounterOnce("sql.exec.cascade-limit-reached")

var HashAggregationDiskSpillingDisabled = telemetry.GetCounterOnce("sql.exec.hash-agg-spilling-disabled")

var DistSQLFlowsScheduled = telemetry.GetCounterOnce("sql.distsql.flows.scheduled")

var DistSQLFlowsQueued = telemetry.GetCounterOnce("sql.distsql.flows.queued")
