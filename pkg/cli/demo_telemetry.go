package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

type demoTelemetry int

const (
	_ demoTelemetry = iota

	demo

	nodes

	demoLocality

	withLoad

	geoPartitionedReplicas
)

var demoTelemetryMap = map[demoTelemetry]string{
	demo:                   "demo",
	nodes:                  "nodes",
	demoLocality:           "demo-locality",
	withLoad:               "withload",
	geoPartitionedReplicas: "geo-partitioned-replicas",
}

var demoTelemetryCounters map[demoTelemetry]telemetry.Counter

func init() {
	demoTelemetryCounters = make(map[demoTelemetry]telemetry.Counter)
	for ty, s := range demoTelemetryMap {
		demoTelemetryCounters[ty] = telemetry.GetCounterOnce(fmt.Sprintf("cli.demo.%s", s))
	}
}

func incrementDemoCounter(d demoTelemetry) {
	__antithesis_instrumentation__.Notify(31940)
	telemetry.Inc(demoTelemetryCounters[d])
}
