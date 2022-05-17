package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "sort"

var internalTSMetricsNames []string

func init() {
	internalTSMetricsNames = allInternalTSMetricsNames()
}

func AllInternalTimeseriesMetricNames() []string {
	__antithesis_instrumentation__.Notify(647941)
	return internalTSMetricsNames
}

var _ = `
./cockroach demo --empty -e \
    "select name from crdb_internal.node_metrics where name like '%-p50'" | \
    sed -E 's/^(.*)-p50$/"\1": {},/'
`

var histogramMetricsNames = map[string]struct{}{
	"sql.txn.latency.internal":                  {},
	"sql.conn.latency":                          {},
	"sql.mem.sql.session.max":                   {},
	"sql.stats.flush.duration":                  {},
	"changefeed.checkpoint_hist_nanos":          {},
	"admission.wait_durations.sql-sql-response": {},
	"admission.wait_durations.sql-kv-response":  {},
	"sql.exec.latency":                          {},
	"sql.stats.mem.max.internal":                {},
	"admission.wait_durations.sql-leaf-start":   {},
	"sql.disk.distsql.max":                      {},
	"txn.restarts":                              {},
	"sql.stats.flush.duration.internal":         {},
	"sql.distsql.exec.latency":                  {},
	"sql.mem.internal.txn.max":                  {},
	"changefeed.emit_hist_nanos":                {},
	"changefeed.flush_hist_nanos":               {},
	"sql.service.latency":                       {},
	"round-trip-latency":                        {},
	"admission.wait_durations.kv":               {},
	"sql.mem.distsql.max":                       {},
	"kv.prober.write.latency":                   {},
	"exec.latency":                              {},
	"admission.wait_durations.sql-root-start":   {},
	"sql.mem.bulk.max":                          {},
	"sql.distsql.flows.queue_wait":              {},
	"sql.txn.latency":                           {},
	"sql.mem.root.max":                          {},
	"admission.wait_durations.kv-stores":        {},
	"sql.stats.mem.max":                         {},
	"sql.distsql.service.latency.internal":      {},
	"sql.stats.reported.mem.max.internal":       {},
	"sql.stats.reported.mem.max":                {},
	"sql.exec.latency.internal":                 {},
	"sql.mem.internal.session.max":              {},
	"sql.distsql.exec.latency.internal":         {},
	"kv.prober.read.latency":                    {},
	"sql.distsql.service.latency":               {},
	"sql.service.latency.internal":              {},
	"sql.mem.sql.txn.max":                       {},
	"liveness.heartbeatlatency":                 {},
	"txn.durations":                             {},
	"raft.process.handleready.latency":          {},
	"raft.process.commandcommit.latency":        {},
	"raft.process.logcommit.latency":            {},
	"raft.scheduler.latency":                    {},
	"txnwaitqueue.pusher.wait_time":             {},
	"txnwaitqueue.query.wait_time":              {},
	"raft.process.applycommitted.latency":       {},
	"sql.stats.txn_stats_collection.duration":   {},
}

func allInternalTSMetricsNames() []string {
	__antithesis_instrumentation__.Notify(647942)
	m := map[string]struct{}{}
	for _, section := range charts {
		__antithesis_instrumentation__.Notify(647945)
		for _, chart := range section.Charts {
			__antithesis_instrumentation__.Notify(647946)
			for _, metric := range chart.Metrics {
				__antithesis_instrumentation__.Notify(647947)

				_, isHist := histogramMetricsNames[metric]
				if !isHist {
					__antithesis_instrumentation__.Notify(647949)
					m[metric] = struct{}{}
					continue
				} else {
					__antithesis_instrumentation__.Notify(647950)
				}
				__antithesis_instrumentation__.Notify(647948)
				for _, p := range []string{
					"p50",
					"p75",
					"p90",
					"p99",
					"p99.9",
					"p99.99",
					"p99.999",
				} {
					__antithesis_instrumentation__.Notify(647951)
					m[metric+"-"+p] = struct{}{}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(647943)

	names := make([]string, 0, 2*len(m))
	for name := range m {
		__antithesis_instrumentation__.Notify(647952)
		names = append(names, "cr.store."+name, "cr.node."+name)
	}
	__antithesis_instrumentation__.Notify(647944)
	sort.Strings(names)
	return names
}
