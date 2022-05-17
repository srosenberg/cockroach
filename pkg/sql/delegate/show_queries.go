package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

func (d *delegator) delegateShowQueries(n *tree.ShowQueries) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465683)
	sqltelemetry.IncrementShowCounter(sqltelemetry.Queries)
	const query = `SELECT query_id, node_id, session_id, user_name, start, query, client_address, application_name, distributed, phase FROM crdb_internal.`
	table := `node_queries`
	if n.Cluster {
		__antithesis_instrumentation__.Notify(465686)
		table = `cluster_queries`
	} else {
		__antithesis_instrumentation__.Notify(465687)
	}
	__antithesis_instrumentation__.Notify(465684)
	var filter string
	if !n.All {
		__antithesis_instrumentation__.Notify(465688)
		filter = " WHERE application_name NOT LIKE '" + catconstants.InternalAppNamePrefix + "%'"
	} else {
		__antithesis_instrumentation__.Notify(465689)
	}
	__antithesis_instrumentation__.Notify(465685)
	return parse(query + table + filter)
}
