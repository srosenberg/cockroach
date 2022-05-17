package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowTransactions(n *tree.ShowTransactions) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465865)
	const query = `SELECT node_id, id AS txn_id, application_name, num_stmts, num_retries, num_auto_retries FROM `
	table := `"".crdb_internal.node_transactions`
	if n.Cluster {
		__antithesis_instrumentation__.Notify(465868)
		table = `"".crdb_internal.cluster_transactions`
	} else {
		__antithesis_instrumentation__.Notify(465869)
	}
	__antithesis_instrumentation__.Notify(465866)
	var filter string
	if !n.All {
		__antithesis_instrumentation__.Notify(465870)
		filter = " WHERE application_name NOT LIKE '" + catconstants.InternalAppNamePrefix + "%'"
	} else {
		__antithesis_instrumentation__.Notify(465871)
	}
	__antithesis_instrumentation__.Notify(465867)
	return parse(query + table + filter)
}
