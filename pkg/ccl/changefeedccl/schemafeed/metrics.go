package schemafeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

var metaChangefeedTableMetadataNanos = metric.Metadata{
	Name:        "changefeed.table_metadata_nanos",
	Help:        "Time blocked while verifying table metadata histories",
	Measurement: "Nanoseconds",
	Unit:        metric.Unit_NANOSECONDS,
}

type Metrics struct {
	TableMetadataNanos *metric.Counter
}

func (Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(17710) }

func MakeMetrics(histogramWindow time.Duration) Metrics {
	__antithesis_instrumentation__.Notify(17711)
	return Metrics{
		TableMetadataNanos: metric.NewCounter(metaChangefeedTableMetadataNanos),
	}
}

var _ (metric.Struct) = (*Metrics)(nil)
