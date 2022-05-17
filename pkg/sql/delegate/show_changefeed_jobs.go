package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

func (d *delegator) delegateShowChangefeedJobs(n *tree.ShowChangefeedJobs) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465477)
	sqltelemetry.IncrementShowCounter(sqltelemetry.Jobs)

	const (
		selectClause = `
WITH payload AS (
  SELECT 
    id, 
    crdb_internal.pb_to_json(
      'cockroach.sql.jobs.jobspb.Payload', 
      payload, false, true
    )->'changefeed' AS changefeed_details 
  FROM 
    system.jobs
) 
SELECT 
  job_id, 
  description, 
  user_name, 
  status, 
  running_status, 
  created, 
  started, 
  finished, 
  modified, 
  high_water_timestamp, 
  error, 
  replace(
    changefeed_details->>'sink_uri', 
    '\u0026', '&'
  ) AS sink_uri, 
  ARRAY (
    SELECT 
      concat(
        database_name, '.', schema_name, '.', 
        name
      ) 
    FROM 
      crdb_internal.tables 
    WHERE 
      table_id = ANY (descriptor_ids)
  ) AS full_table_names, 
  changefeed_details->'opts'->>'topics' AS topics,
  changefeed_details->'opts'->>'format' AS format 
FROM 
  crdb_internal.jobs 
  INNER JOIN payload ON id = job_id`
	)

	var whereClause, orderbyClause string
	typePredicate := fmt.Sprintf("job_type = '%s'", jobspb.TypeChangefeed)

	if n.Jobs == nil {
		__antithesis_instrumentation__.Notify(465479)

		whereClause = fmt.Sprintf(
			`WHERE %s AND (finished IS NULL OR finished > now() - '12h':::interval)`, typePredicate)

		orderbyClause = `ORDER BY COALESCE(finished, now()) DESC, started DESC`
	} else {
		__antithesis_instrumentation__.Notify(465480)

		whereClause = fmt.Sprintf(`WHERE %s AND job_id in (%s)`, typePredicate, n.Jobs.String())
	}
	__antithesis_instrumentation__.Notify(465478)

	sqlStmt := fmt.Sprintf("%s %s %s", selectClause, whereClause, orderbyClause)

	return parse(sqlStmt)
}
