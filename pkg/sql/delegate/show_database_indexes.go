package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowDatabaseIndexes(
	n *tree.ShowDatabaseIndexes,
) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465521)
	getAllIndexesQuery := `
SELECT
	table_name,
	index_name,
	non_unique::BOOL,
	seq_in_index,
	column_name,
	direction,
	storing::BOOL,
	implicit::BOOL`

	if n.WithComment {
		__antithesis_instrumentation__.Notify(465524)
		getAllIndexesQuery += `,
	obj_description(pg_class.oid) AS comment`
	} else {
		__antithesis_instrumentation__.Notify(465525)
	}
	__antithesis_instrumentation__.Notify(465522)

	getAllIndexesQuery += `
FROM
	%s.information_schema.statistics`

	if n.WithComment {
		__antithesis_instrumentation__.Notify(465526)
		getAllIndexesQuery += `
	LEFT JOIN pg_class ON
		statistics.index_name = pg_class.relname`
	} else {
		__antithesis_instrumentation__.Notify(465527)
	}
	__antithesis_instrumentation__.Notify(465523)

	return parse(fmt.Sprintf(getAllIndexesQuery, n.Database.String()))
}
