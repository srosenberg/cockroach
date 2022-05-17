package txnidcache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/metric"

type Metrics struct {
	CacheMissCounter *metric.Counter
	CacheReadCounter *metric.Counter
}

var _ metric.Struct = Metrics{}

func (Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(459379) }

func NewMetrics() Metrics {
	__antithesis_instrumentation__.Notify(459380)
	return Metrics{
		CacheMissCounter: metric.NewCounter(metric.Metadata{
			Name:        "sql.contention.txn_id_cache.miss",
			Help:        "Number of cache misses",
			Measurement: "Cache miss",
			Unit:        metric.Unit_COUNT,
		}),
		CacheReadCounter: metric.NewCounter(metric.Metadata{
			Name:        "sql.contention.txn_id_cache.read",
			Help:        "Number of cache read",
			Measurement: "Cache read",
			Unit:        metric.Unit_COUNT,
		}),
	}
}
