package bulk

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

type Metrics struct {
	MaxBytesHist  *metric.Histogram
	CurBytesCount *metric.Gauge
}

func (Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(86325) }

var _ metric.Struct = Metrics{}

var (
	metaMemMaxBytes = metric.Metadata{
		Name:        "sql.mem.bulk.max",
		Help:        "Memory usage per sql statement for bulk operations",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
	metaMemCurBytes = metric.Metadata{
		Name:        "sql.mem.bulk.current",
		Help:        "Current sql statement memory usage for bulk operations",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
)

const log10int64times1000 = 19 * 1000

func MakeBulkMetrics(histogramWindow time.Duration) Metrics {
	__antithesis_instrumentation__.Notify(86326)
	return Metrics{
		MaxBytesHist:  metric.NewHistogram(metaMemMaxBytes, histogramWindow, log10int64times1000, 3),
		CurBytesCount: metric.NewGauge(metaMemCurBytes),
	}
}
