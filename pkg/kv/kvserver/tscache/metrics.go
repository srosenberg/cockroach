package tscache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/metric"

type Metrics struct {
	Skl sklMetrics
}

type sklMetrics struct {
	Pages         *metric.Gauge
	PageRotations *metric.Counter
}

func (sklMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(127104) }

var _ metric.Struct = sklMetrics{}

var (
	metaSklPages = metric.Metadata{
		Name:        "tscache.skl.pages",
		Help:        "Number of pages in the timestamp cache",
		Measurement: "Pages",
		Unit:        metric.Unit_COUNT,
	}
	metaSklRotations = metric.Metadata{
		Name:        "tscache.skl.rotations",
		Help:        "Number of page rotations in the timestamp cache",
		Measurement: "Page Rotations",
		Unit:        metric.Unit_COUNT,
	}
)

func makeMetrics() Metrics {
	__antithesis_instrumentation__.Notify(127105)
	return Metrics{
		Skl: sklMetrics{
			Pages:         metric.NewGauge(metaSklPages),
			PageRotations: metric.NewCounter(metaSklRotations),
		},
	}
}
