package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

func (d *delegator) delegateShowSurvivalGoal(n *tree.ShowSurvivalGoal) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465799)
	sqltelemetry.IncrementShowCounter(sqltelemetry.SurvivalGoal)
	dbName := string(n.DatabaseName)
	if dbName == "" {
		__antithesis_instrumentation__.Notify(465801)
		dbName = d.evalCtx.SessionData().Database
	} else {
		__antithesis_instrumentation__.Notify(465802)
	}
	__antithesis_instrumentation__.Notify(465800)
	query := fmt.Sprintf(
		`SELECT
	name AS "database",
	survival_goal
FROM crdb_internal.databases
WHERE name = %s`,
		lexbase.EscapeSQLString(dbName),
	)
	return parse(query)
}
