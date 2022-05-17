package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
)

var CancelRequestCounter = telemetry.GetCounterOnce("pgwire.cancel_request")

func UnimplementedClientStatusParameterCounter(key string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625806)
	return telemetry.GetCounter(fmt.Sprintf("unimplemented.pgwire.parameter.%s", key))
}

var BinaryDecimalInfinityCounter = telemetry.GetCounterOnce("pgwire.#32489.binary_decimal_infinity")

var UncategorizedErrorCounter = telemetry.GetCounterOnce("othererror." + pgcode.Uncategorized.String())

var InterleavedPortalRequestCounter = telemetry.GetCounterOnce("pgwire.#40195.interleaved_portal")

var PortalWithLimitRequestCounter = telemetry.GetCounterOnce("pgwire.portal_with_limit_request")

var ParseRequestCounter = telemetry.GetCounterOnce("pgwire.command.parse")

var BindRequestCounter = telemetry.GetCounterOnce("pgwire.command.bind")

var DescribeRequestCounter = telemetry.GetCounterOnce("pgwire.command.describe")

var ExecuteRequestCounter = telemetry.GetCounterOnce("pgwire.command.execute")

var CloseRequestCounter = telemetry.GetCounterOnce("pgwire.command.close")

var FlushRequestCounter = telemetry.GetCounterOnce("pgwire.command.flush")
