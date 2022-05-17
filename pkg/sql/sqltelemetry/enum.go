package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

type EnumTelemetryType int

const (
	_ EnumTelemetryType = iota

	EnumCreate

	EnumAlter

	EnumDrop

	EnumInTable
)

var enumTelemetryMap = map[EnumTelemetryType]string{
	EnumCreate:  "create_enum",
	EnumAlter:   "alter_enum",
	EnumDrop:    "drop_enum",
	EnumInTable: "enum_used_in_table",
}

func (e EnumTelemetryType) String() string {
	__antithesis_instrumentation__.Notify(625780)
	return enumTelemetryMap[e]
}

var enumTelemetryCounters map[EnumTelemetryType]telemetry.Counter

func init() {
	enumTelemetryCounters = make(map[EnumTelemetryType]telemetry.Counter)
	for ty, s := range enumTelemetryMap {
		enumTelemetryCounters[ty] = telemetry.GetCounterOnce(fmt.Sprintf("sql.udts.%s", s))
	}
}

func IncrementEnumCounter(enumType EnumTelemetryType) {
	__antithesis_instrumentation__.Notify(625781)
	telemetry.Inc(enumTelemetryCounters[enumType])
}
