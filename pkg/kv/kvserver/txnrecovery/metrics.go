package txnrecovery

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/metric"

type Metrics struct {
	AttemptsPending      *metric.Gauge
	Attempts             *metric.Counter
	SuccessesAsCommitted *metric.Counter
	SuccessesAsAborted   *metric.Counter
	SuccessesAsPending   *metric.Counter
	Failures             *metric.Counter
}

func makeMetrics() Metrics {
	__antithesis_instrumentation__.Notify(127286)
	return Metrics{
		AttemptsPending: metric.NewGauge(
			metric.Metadata{
				Name:        "txnrecovery.attempts.pending",
				Help:        "Number of transaction recovery attempts currently in-flight",
				Measurement: "Recovery Attempts",
				Unit:        metric.Unit_COUNT,
			},
		),

		Attempts: metric.NewCounter(
			metric.Metadata{
				Name:        "txnrecovery.attempts.total",
				Help:        "Number of transaction recovery attempts executed",
				Measurement: "Recovery Attempts",
				Unit:        metric.Unit_COUNT,
			},
		),

		SuccessesAsCommitted: metric.NewCounter(
			metric.Metadata{
				Name:        "txnrecovery.successes.committed",
				Help:        "Number of transaction recovery attempts that committed a transaction",
				Measurement: "Recovery Attempts",
				Unit:        metric.Unit_COUNT,
			},
		),

		SuccessesAsAborted: metric.NewCounter(
			metric.Metadata{
				Name:        "txnrecovery.successes.aborted",
				Help:        "Number of transaction recovery attempts that aborted a transaction",
				Measurement: "Recovery Attempts",
				Unit:        metric.Unit_COUNT,
			},
		),

		SuccessesAsPending: metric.NewCounter(
			metric.Metadata{
				Name:        "txnrecovery.successes.pending",
				Help:        "Number of transaction recovery attempts that left a transaction pending",
				Measurement: "Recovery Attempts",
				Unit:        metric.Unit_COUNT,
			},
		),

		Failures: metric.NewCounter(
			metric.Metadata{
				Name:        "txnrecovery.failures",
				Help:        "Number of transaction recovery attempts that failed",
				Measurement: "Recovery Attempts",
				Unit:        metric.Unit_COUNT,
			},
		),
	}
}
