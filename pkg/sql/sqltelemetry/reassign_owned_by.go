package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/server/telemetry"

func CreateReassignOwnedByCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(625812)
	return telemetry.GetCounter("sql.reassign_owned_by")
}
