package contention

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/metric"

type Metrics struct {
	ResolverQueueSize *metric.Gauge
	ResolverRetries   *metric.Counter
	ResolverFailed    *metric.Counter
}

var _ metric.Struct = Metrics{}

func (Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(459117) }

func NewMetrics() Metrics {
	__antithesis_instrumentation__.Notify(459118)
	return Metrics{
		ResolverQueueSize: metric.NewGauge(metric.Metadata{
			Name:        "sql.contention.resolver.queue_size",
			Help:        "Length of queued unresolved contention events",
			Measurement: "Queue length",
			Unit:        metric.Unit_COUNT,
		}),
		ResolverRetries: metric.NewCounter(metric.Metadata{
			Name:        "sql.contention.resolver.retries",
			Help:        "Number of times transaction id resolution has been retried",
			Measurement: "Retry count",
			Unit:        metric.Unit_COUNT,
		}),
		ResolverFailed: metric.NewCounter(metric.Metadata{
			Name:        "sql.contention.resolver.failed_resolutions",
			Help:        "Number of failed transaction ID resolution attempts",
			Measurement: "Failed transaction ID resolution count",
			Unit:        metric.Unit_COUNT,
		}),
	}
}
