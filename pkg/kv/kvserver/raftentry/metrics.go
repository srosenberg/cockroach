package raftentry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/metric"

var (
	metaEntryCacheSize = metric.Metadata{
		Name:        "raft.entrycache.size",
		Help:        "Number of Raft entries in the Raft entry cache",
		Measurement: "Entry Count",
		Unit:        metric.Unit_COUNT,
	}
	metaEntryCacheBytes = metric.Metadata{
		Name:        "raft.entrycache.bytes",
		Help:        "Aggregate size of all Raft entries in the Raft entry cache",
		Measurement: "Entry Bytes",
		Unit:        metric.Unit_BYTES,
	}
	metaEntryCacheAccesses = metric.Metadata{
		Name:        "raft.entrycache.accesses",
		Help:        "Number of cache lookups in the Raft entry cache",
		Measurement: "Accesses",
		Unit:        metric.Unit_COUNT,
	}
	metaEntryCacheHits = metric.Metadata{
		Name:        "raft.entrycache.hits",
		Help:        "Number of successful cache lookups in the Raft entry cache",
		Measurement: "Hits",
		Unit:        metric.Unit_COUNT,
	}
)

type Metrics struct {
	Size     *metric.Gauge
	Bytes    *metric.Gauge
	Accesses *metric.Counter
	Hits     *metric.Counter
}

func makeMetrics() Metrics {
	__antithesis_instrumentation__.Notify(113343)
	return Metrics{
		Size:     metric.NewGauge(metaEntryCacheSize),
		Bytes:    metric.NewGauge(metaEntryCacheBytes),
		Accesses: metric.NewCounter(metaEntryCacheAccesses),
		Hits:     metric.NewCounter(metaEntryCacheHits),
	}
}
