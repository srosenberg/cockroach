package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/server/telemetry"

func CreateDropOwnedByCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(625779)
	return telemetry.GetCounter("sql.drop_owned_by")
}
