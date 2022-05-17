package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowDatabases(stmt *tree.ShowDatabases) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465528)
	query := `SELECT
	name AS database_name, owner, primary_region, regions, survival_goal`

	if stmt.WithComment {
		__antithesis_instrumentation__.Notify(465531)
		query += `, comment`
	} else {
		__antithesis_instrumentation__.Notify(465532)
	}
	__antithesis_instrumentation__.Notify(465529)

	query += `
FROM
  "".crdb_internal.databases d
`
	if stmt.WithComment {
		__antithesis_instrumentation__.Notify(465533)
		query += fmt.Sprintf(`
LEFT JOIN
	(
		SELECT 
			object_id, type, comment
		FROM
			system.comments
		WHERE
			type = %d
	) c
ON
	c.object_id = d.id`, keys.DatabaseCommentType)
	} else {
		__antithesis_instrumentation__.Notify(465534)
	}
	__antithesis_instrumentation__.Notify(465530)

	query += `
ORDER BY
	database_name`

	return parse(query)
}
