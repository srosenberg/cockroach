package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/metric"

var (
	metaWriteSamples = metric.Metadata{
		Name:        "timeseries.write.samples",
		Help:        "Total number of metric samples written to disk",
		Measurement: "Metric Samples",
		Unit:        metric.Unit_COUNT,
	}
	metaWriteBytes = metric.Metadata{
		Name:        "timeseries.write.bytes",
		Help:        "Total size in bytes of metric samples written to disk",
		Measurement: "Storage",
		Unit:        metric.Unit_BYTES,
	}
	metaWriteErrors = metric.Metadata{
		Name:        "timeseries.write.errors",
		Help:        "Total errors encountered while attempting to write metrics to disk",
		Measurement: "Errors",
		Unit:        metric.Unit_COUNT,
	}
)

type TimeSeriesMetrics struct {
	WriteSamples *metric.Counter
	WriteBytes   *metric.Counter
	WriteErrors  *metric.Counter
}

func NewTimeSeriesMetrics() *TimeSeriesMetrics {
	__antithesis_instrumentation__.Notify(648096)
	return &TimeSeriesMetrics{
		WriteSamples: metric.NewCounter(metaWriteSamples),
		WriteBytes:   metric.NewCounter(metaWriteBytes),
		WriteErrors:  metric.NewCounter(metaWriteErrors),
	}
}
