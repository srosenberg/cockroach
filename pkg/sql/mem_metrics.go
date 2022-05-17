package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

type BaseMemoryMetrics struct {
	MaxBytesHist  *metric.Histogram
	CurBytesCount *metric.Gauge
}

type MemoryMetrics struct {
	BaseMemoryMetrics
	TxnMaxBytesHist      *metric.Histogram
	TxnCurBytesCount     *metric.Gauge
	SessionMaxBytesHist  *metric.Histogram
	SessionCurBytesCount *metric.Gauge
}

func (MemoryMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(501767) }

var _ metric.Struct = MemoryMetrics{}

const log10int64times1000 = 19 * 1000

func makeMemMetricMetadata(name, help string) metric.Metadata {
	__antithesis_instrumentation__.Notify(501768)
	return metric.Metadata{
		Name:        name,
		Help:        help,
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
}

func MakeBaseMemMetrics(endpoint string, histogramWindow time.Duration) BaseMemoryMetrics {
	__antithesis_instrumentation__.Notify(501769)
	prefix := "sql.mem." + endpoint
	MetaMemMaxBytes := makeMemMetricMetadata(prefix+".max", "Memory usage per sql statement for "+endpoint)
	MetaMemCurBytes := makeMemMetricMetadata(prefix+".current", "Current sql statement memory usage for "+endpoint)
	return BaseMemoryMetrics{
		MaxBytesHist:  metric.NewHistogram(MetaMemMaxBytes, histogramWindow, log10int64times1000, 3),
		CurBytesCount: metric.NewGauge(MetaMemCurBytes),
	}
}

func MakeMemMetrics(endpoint string, histogramWindow time.Duration) MemoryMetrics {
	__antithesis_instrumentation__.Notify(501770)
	base := MakeBaseMemMetrics(endpoint, histogramWindow)
	prefix := "sql.mem." + endpoint
	MetaMemMaxTxnBytes := makeMemMetricMetadata(prefix+".txn.max", "Memory usage per sql transaction for "+endpoint)
	MetaMemTxnCurBytes := makeMemMetricMetadata(prefix+".txn.current", "Current sql transaction memory usage for "+endpoint)
	MetaMemMaxSessionBytes := makeMemMetricMetadata(prefix+".session.max", "Memory usage per sql session for "+endpoint)
	MetaMemSessionCurBytes := makeMemMetricMetadata(prefix+".session.current", "Current sql session memory usage for "+endpoint)
	return MemoryMetrics{
		BaseMemoryMetrics:    base,
		TxnMaxBytesHist:      metric.NewHistogram(MetaMemMaxTxnBytes, histogramWindow, log10int64times1000, 3),
		TxnCurBytesCount:     metric.NewGauge(MetaMemTxnCurBytes),
		SessionMaxBytesHist:  metric.NewHistogram(MetaMemMaxSessionBytes, histogramWindow, log10int64times1000, 3),
		SessionCurBytesCount: metric.NewGauge(MetaMemSessionCurBytes),
	}

}
