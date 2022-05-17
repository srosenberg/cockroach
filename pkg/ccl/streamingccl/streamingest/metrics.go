package streamingest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

var (
	metaStreamingEventsIngested = metric.Metadata{
		Name:        "streaming.events_ingested",
		Help:        "Events ingested by all ingestion jobs",
		Measurement: "Events",
		Unit:        metric.Unit_COUNT,
	}
	metaStreamingResolvedEventsIngested = metric.Metadata{
		Name:        "streaming.resolved_events_ingested",
		Help:        "Resolved events ingested by all ingestion jobs",
		Measurement: "Events",
		Unit:        metric.Unit_COUNT,
	}
	metaStreamingIngestedBytes = metric.Metadata{
		Name:        "streaming.ingested_bytes",
		Help:        "Bytes ingested by all ingestion jobs",
		Measurement: "Bytes",
		Unit:        metric.Unit_BYTES,
	}
	metaStreamingFlushes = metric.Metadata{
		Name:        "streaming.flushes",
		Help:        "Total flushes across all ingestion jobs",
		Measurement: "Flushes",
		Unit:        metric.Unit_COUNT,
	}
)

type Metrics struct {
	IngestedEvents *metric.Counter
	IngestedBytes  *metric.Counter
	Flushes        *metric.Counter
	ResolvedEvents *metric.Counter
}

func (*Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(25182) }

func MakeMetrics(histogramWindow time.Duration) metric.Struct {
	__antithesis_instrumentation__.Notify(25183)
	m := &Metrics{
		IngestedEvents: metric.NewCounter(metaStreamingEventsIngested),
		IngestedBytes:  metric.NewCounter(metaStreamingIngestedBytes),
		Flushes:        metric.NewCounter(metaStreamingFlushes),
		ResolvedEvents: metric.NewCounter(metaStreamingResolvedEventsIngested),
	}
	return m
}

func init() {
	jobs.MakeStreamIngestMetricsHook = MakeMetrics
}
