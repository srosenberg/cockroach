package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

const getVirtualSchemaEntry = "sql.schema.get_virtual_table.%s.%s"

var trackedSchemas = map[string]struct{}{
	"pg_catalog":         {},
	"information_schema": {},
}

func IncrementGetVirtualTableEntry(schema, tableName string) {
	__antithesis_instrumentation__.Notify(625848)
	if _, ok := trackedSchemas[schema]; ok {
		__antithesis_instrumentation__.Notify(625849)
		telemetry.Inc(telemetry.GetCounter(fmt.Sprintf(getVirtualSchemaEntry, schema, tableName)))
	} else {
		__antithesis_instrumentation__.Notify(625850)
	}
}
