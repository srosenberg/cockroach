package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

type UserDefinedSchemaTelemetryType int

const (
	_ UserDefinedSchemaTelemetryType = iota

	UserDefinedSchemaCreate

	UserDefinedSchemaAlter

	UserDefinedSchemaDrop

	UserDefinedSchemaReparentDatabase

	UserDefinedSchemaUsedByObject
)

var userDefinedSchemaTelemetryMap = map[UserDefinedSchemaTelemetryType]string{
	UserDefinedSchemaCreate:           "create_schema",
	UserDefinedSchemaAlter:            "alter_schema",
	UserDefinedSchemaDrop:             "drop_schema",
	UserDefinedSchemaReparentDatabase: "reparent_database",
	UserDefinedSchemaUsedByObject:     "schema_used_by_object",
}

func (u UserDefinedSchemaTelemetryType) String() string {
	__antithesis_instrumentation__.Notify(625846)
	return userDefinedSchemaTelemetryMap[u]
}

var userDefinedSchemaTelemetryCounters map[UserDefinedSchemaTelemetryType]telemetry.Counter

func init() {
	userDefinedSchemaTelemetryCounters = make(map[UserDefinedSchemaTelemetryType]telemetry.Counter)
	for ty, s := range userDefinedSchemaTelemetryMap {
		userDefinedSchemaTelemetryCounters[ty] = telemetry.GetCounterOnce(fmt.Sprintf("sql.uds.%s", s))
	}
}

func IncrementUserDefinedSchemaCounter(userDefinedSchemaType UserDefinedSchemaTelemetryType) {
	__antithesis_instrumentation__.Notify(625847)
	telemetry.Inc(userDefinedSchemaTelemetryCounters[userDefinedSchemaType])
}
