package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

func (d *delegator) delegateShowFullTableScans() (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465564)
	sqltelemetry.IncrementShowCounter(sqltelemetry.FullTableScans)
	const query = `
  SELECT 
    key AS query, count, rows_read_avg, bytes_read_avg, service_lat_avg, contention_time_avg, max_mem_usage_avg, network_bytes_avg, max_retries
  FROM crdb_internal.node_statement_statistics WHERE full_scan = TRUE ORDER BY count DESC`
	return parse(query)
}
