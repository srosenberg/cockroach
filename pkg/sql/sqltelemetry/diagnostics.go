package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/sha256"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

var StatementDiagnosticsCollectedCounter = telemetry.GetCounterOnce("sql.diagnostics.collected")

func HashedFeatureCounter(feature string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625778)
	sum := sha256.Sum256([]byte(feature))
	return telemetry.GetCounter(fmt.Sprintf("sql.hashed.%x", sum))
}
