package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

var DefaultIntSize4Counter = telemetry.GetCounterOnce("sql.default_int_size.4")

var ForceSavepointRestartCounter = telemetry.GetCounterOnce("sql.force_savepoint_restart")

var CockroachShellCounter = telemetry.GetCounterOnce("sql.connection.cockroach_cli")

func UnimplementedSessionVarValueCounter(varName, val string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625842)
	return telemetry.GetCounter(fmt.Sprintf("unimplemented.sql.session_var.%s.%s", varName, val))
}

func DummySessionVarValueCounter(varName string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625843)
	return telemetry.GetCounter(fmt.Sprintf("sql.session_var.dummy.%s", varName))
}
