package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

func (d *delegator) delegateShowRoles() (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465760)
	sqltelemetry.IncrementShowCounter(sqltelemetry.Roles)
	return parse(`
SELECT
	u.username,
	IFNULL(string_agg(o.option || COALESCE('=' || o.value, ''), ', ' ORDER BY o.option), '') AS options,
	ARRAY (SELECT role FROM system.role_members AS rm WHERE rm.member = u.username ORDER BY 1) AS member_of
FROM
	system.users AS u LEFT JOIN system.role_options AS o ON u.username = o.username
GROUP BY
	u.username
ORDER BY 1;
`)
}
