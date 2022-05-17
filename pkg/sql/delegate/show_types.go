package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

func (d *delegator) delegateShowTypes() (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465872)

	return parse(`
SELECT
  schema, name, owner
FROM
  [SHOW ENUMS]
ORDER BY
  (schema, name)`)
}

func (d *delegator) delegateShowCreateAllTypes() (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465873)
	sqltelemetry.IncrementShowCounter(sqltelemetry.Create)

	const showCreateAllTypesQuery = `
	SELECT crdb_internal.show_create_all_types(%[1]s) AS create_statement;
`
	databaseLiteral := d.evalCtx.SessionData().Database

	query := fmt.Sprintf(showCreateAllTypesQuery,
		lexbase.EscapeSQLString(databaseLiteral),
	)

	return parse(query)
}
