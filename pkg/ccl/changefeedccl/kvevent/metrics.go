package kvevent

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

var (
	metaChangefeedBufferEntriesIn = metric.Metadata{
		Name:        "changefeed.buffer_entries.in",
		Help:        "Total entries entering the buffer between raft and changefeed sinks",
		Measurement: "Entries",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedBufferEntriesOut = metric.Metadata{
		Name:        "changefeed.buffer_entries.out",
		Help:        "Total entries leaving the buffer between raft and changefeed sinks",
		Measurement: "Entries",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedBufferEntriesReleased = metric.Metadata{
		Name:        "changefeed.buffer_entries.released",
		Help:        "Total entries processed, emitted and acknowledged by the sinks",
		Measurement: "Entries",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedBufferMemAcquired = metric.Metadata{
		Name:        "changefeed.buffer_entries_mem.acquired",
		Help:        "Total amount of memory acquired for entries as they enter the system",
		Measurement: "Entries",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedBufferMemReleased = metric.Metadata{
		Name:        "changefeed.buffer_entries_mem.released",
		Help:        "Total amount of memory released by the entries after they have been emitted",
		Measurement: "Entries",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedBufferPushbackNanos = metric.Metadata{
		Name:        "changefeed.buffer_pushback_nanos",
		Help:        "Total time spent waiting while the buffer was full",
		Measurement: "Nanoseconds",
		Unit:        metric.Unit_NANOSECONDS,
	}
)

type Metrics struct {
	BufferEntriesIn          *metric.Counter
	BufferEntriesOut         *metric.Counter
	BufferEntriesReleased    *metric.Counter
	BufferPushbackNanos      *metric.Counter
	BufferEntriesMemAcquired *metric.Counter
	BufferEntriesMemReleased *metric.Counter
}

func MakeMetrics(histogramWindow time.Duration) Metrics {
	__antithesis_instrumentation__.Notify(17145)
	return Metrics{
		BufferEntriesIn:          metric.NewCounter(metaChangefeedBufferEntriesIn),
		BufferEntriesOut:         metric.NewCounter(metaChangefeedBufferEntriesOut),
		BufferEntriesReleased:    metric.NewCounter(metaChangefeedBufferEntriesReleased),
		BufferEntriesMemAcquired: metric.NewCounter(metaChangefeedBufferMemAcquired),
		BufferEntriesMemReleased: metric.NewCounter(metaChangefeedBufferMemReleased),
		BufferPushbackNanos:      metric.NewCounter(metaChangefeedBufferPushbackNanos),
	}
}

var _ (metric.Struct) = (*Metrics)(nil)

func (m Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(17146) }
