package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/server/telemetry"

func CreateExtensionCounter(ext string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625783)
	return telemetry.GetCounter("sql.extension.create." + ext)
}
