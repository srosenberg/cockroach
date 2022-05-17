package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

type PartitioningTelemetryType int

const (
	_ PartitioningTelemetryType = iota

	AlterAllPartitions

	PartitionConstrainedScan
)

var partitioningTelemetryMap = map[PartitioningTelemetryType]string{
	AlterAllPartitions:       "alter-all-partitions",
	PartitionConstrainedScan: "partition-constrained-scan",
}

func (p PartitioningTelemetryType) String() string {
	__antithesis_instrumentation__.Notify(625804)
	return partitioningTelemetryMap[p]
}

var partitioningTelemetryCounters map[PartitioningTelemetryType]telemetry.Counter

func init() {
	partitioningTelemetryCounters = make(map[PartitioningTelemetryType]telemetry.Counter)
	for ty, s := range partitioningTelemetryMap {
		partitioningTelemetryCounters[ty] = telemetry.GetCounterOnce(fmt.Sprintf("sql.partitioning.%s", s))
	}
}

func IncrementPartitioningCounter(partitioningType PartitioningTelemetryType) {
	__antithesis_instrumentation__.Notify(625805)
	telemetry.Inc(partitioningTelemetryCounters[partitioningType])
}
