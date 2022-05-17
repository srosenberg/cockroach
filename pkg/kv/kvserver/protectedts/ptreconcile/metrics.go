package ptreconcile

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

type Metrics struct {
	ReconcilationRuns    *metric.Counter
	RecordsProcessed     *metric.Counter
	RecordsRemoved       *metric.Counter
	ReconciliationErrors *metric.Counter
}

func makeMetrics() Metrics {
	__antithesis_instrumentation__.Notify(111748)
	return Metrics{
		ReconcilationRuns:    metric.NewCounter(metaReconciliationRuns),
		RecordsProcessed:     metric.NewCounter(metaRecordsProcessed),
		RecordsRemoved:       metric.NewCounter(metaRecordsRemoved),
		ReconciliationErrors: metric.NewCounter(metaReconciliationErrors),
	}
}

var _ metric.Struct = (*Metrics)(nil)

func (m *Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(111749) }

var (
	metaReconciliationRuns = metric.Metadata{
		Name:        "kv.protectedts.reconciliation.num_runs",
		Help:        "number of successful reconciliation runs on this node",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_COUNTER,
	}
	metaRecordsProcessed = metric.Metadata{
		Name:        "kv.protectedts.reconciliation.records_processed",
		Help:        "number of records processed without error during reconciliation on this node",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_COUNTER,
	}
	metaRecordsRemoved = metric.Metadata{
		Name:        "kv.protectedts.reconciliation.records_removed",
		Help:        "number of records removed during reconciliation runs on this node",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_COUNTER,
	}
	metaReconciliationErrors = metric.Metadata{
		Name:        "kv.protectedts.reconciliation.errors",
		Help:        "number of errors encountered during reconciliation runs on this node",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_COUNTER,
	}
)
