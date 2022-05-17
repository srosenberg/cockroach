package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowSessions(n *tree.ShowSessions) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465792)
	const query = `SELECT node_id, session_id, user_name, client_address, application_name, active_queries, last_active_query, session_start, oldest_query_start FROM crdb_internal.`
	table := `node_sessions`
	if n.Cluster {
		__antithesis_instrumentation__.Notify(465795)
		table = `cluster_sessions`
	} else {
		__antithesis_instrumentation__.Notify(465796)
	}
	__antithesis_instrumentation__.Notify(465793)
	var filter string
	if !n.All {
		__antithesis_instrumentation__.Notify(465797)
		filter = " WHERE application_name NOT LIKE '" + catconstants.InternalAppNamePrefix + "%'"
	} else {
		__antithesis_instrumentation__.Notify(465798)
	}
	__antithesis_instrumentation__.Notify(465794)
	return parse(query + table + filter)
}
