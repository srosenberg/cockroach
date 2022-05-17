package intentresolver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/metric"

var (
	metaIntentResolverAsyncThrottled = metric.Metadata{
		Name:        "intentresolver.async.throttled",
		Help:        "Number of intent resolution attempts not run asynchronously due to throttling",
		Measurement: "Intent Resolutions",
		Unit:        metric.Unit_COUNT,
	}
	metaFinalizedTxnCleanupFailed = metric.Metadata{
		Name: "intentresolver.finalized_txns.failed",
		Help: "Number of finalized transaction cleanup failures. Transaction " +
			"cleanup refers to the process of resolving all of a transactions intents " +
			"and then garbage collecting its transaction record.",
		Measurement: "Intent Resolutions",
		Unit:        metric.Unit_COUNT,
	}
	metaIntentCleanupFailed = metric.Metadata{
		Name: "intentresolver.intents.failed",
		Help: "Number of intent resolution failures. The unit of measurement " +
			"is a single intent, so if a batch of intent resolution requests fails, " +
			"the metric will be incremented for each request in the batch.",
		Measurement: "Intent Resolutions",
		Unit:        metric.Unit_COUNT,
	}
)

type Metrics struct {
	IntentResolverAsyncThrottled *metric.Counter

	FinalizedTxnCleanupFailed *metric.Counter

	IntentResolutionFailed *metric.Counter
}

func (*Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(101788) }

func makeMetrics() Metrics {
	__antithesis_instrumentation__.Notify(101789)
	return Metrics{
		IntentResolverAsyncThrottled: metric.NewCounter(metaIntentResolverAsyncThrottled),
		FinalizedTxnCleanupFailed:    metric.NewCounter(metaFinalizedTxnCleanupFailed),
		IntentResolutionFailed:       metric.NewCounter(metaIntentCleanupFailed),
	}
}
