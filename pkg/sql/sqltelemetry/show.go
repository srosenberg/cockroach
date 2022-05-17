package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

type ShowTelemetryType int

const (
	_ ShowTelemetryType = iota

	Ranges

	Regions

	RegionsFromCluster

	RegionsFromAllDatabases

	RegionsFromDatabase

	SurvivalGoal

	Partitions

	Locality

	Create

	CreateSchedule

	RangeForRow

	Queries

	Indexes

	Constraints

	Jobs

	Roles

	Schedules

	FullTableScans

	SuperRegions
)

var showTelemetryNameMap = map[ShowTelemetryType]string{
	Ranges:                  "ranges",
	Partitions:              "partitions",
	Locality:                "locality",
	Create:                  "create",
	CreateSchedule:          "create_schedule",
	RangeForRow:             "rangeforrow",
	Regions:                 "regions",
	RegionsFromCluster:      "regions_from_cluster",
	RegionsFromDatabase:     "regions_from_database",
	RegionsFromAllDatabases: "regions_from_all_databases",
	SurvivalGoal:            "survival_goal",
	Queries:                 "queries",
	Indexes:                 "indexes",
	Constraints:             "constraints",
	Jobs:                    "jobs",
	Roles:                   "roles",
	Schedules:               "schedules",
	FullTableScans:          "full_table_scans",
	SuperRegions:            "super_regions",
}

func (s ShowTelemetryType) String() string {
	__antithesis_instrumentation__.Notify(625844)
	return showTelemetryNameMap[s]
}

var showTelemetryCounters map[ShowTelemetryType]telemetry.Counter

func init() {
	showTelemetryCounters = make(map[ShowTelemetryType]telemetry.Counter)
	for ty, s := range showTelemetryNameMap {
		showTelemetryCounters[ty] = telemetry.GetCounterOnce(fmt.Sprintf("sql.show.%s", s))
	}
}

func IncrementShowCounter(showType ShowTelemetryType) {
	__antithesis_instrumentation__.Notify(625845)
	telemetry.Inc(showTelemetryCounters[showType])
}
