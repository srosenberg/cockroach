package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowRoleGrants(n *tree.ShowRoleGrants) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465740)
	const selectQuery = `
SELECT role AS role_name,
       member,
       "isAdmin" AS is_admin
 FROM system.role_members`

	var query bytes.Buffer
	query.WriteString(selectQuery)

	if n.Roles != nil {
		__antithesis_instrumentation__.Notify(465743)
		var roles []string
		sqlUsernames, err := n.Roles.ToSQLUsernames(d.evalCtx.SessionData(), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(465746)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(465747)
		}
		__antithesis_instrumentation__.Notify(465744)
		for _, r := range sqlUsernames {
			__antithesis_instrumentation__.Notify(465748)
			roles = append(roles, lexbase.EscapeSQLString(r.Normalized()))
		}
		__antithesis_instrumentation__.Notify(465745)
		fmt.Fprintf(&query, ` WHERE "role" IN (%s)`, strings.Join(roles, ","))
	} else {
		__antithesis_instrumentation__.Notify(465749)
	}
	__antithesis_instrumentation__.Notify(465741)

	if n.Grantees != nil {
		__antithesis_instrumentation__.Notify(465750)
		if n.Roles == nil {
			__antithesis_instrumentation__.Notify(465754)

			query.WriteString(" WHERE ")
		} else {
			__antithesis_instrumentation__.Notify(465755)

			query.WriteString(" AND ")
		}
		__antithesis_instrumentation__.Notify(465751)

		var grantees []string
		granteeSQLUsernames, err := n.Grantees.ToSQLUsernames(d.evalCtx.SessionData(), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(465756)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(465757)
		}
		__antithesis_instrumentation__.Notify(465752)
		for _, g := range granteeSQLUsernames {
			__antithesis_instrumentation__.Notify(465758)
			grantees = append(grantees, lexbase.EscapeSQLString(g.Normalized()))
		}
		__antithesis_instrumentation__.Notify(465753)
		fmt.Fprintf(&query, ` member IN (%s)`, strings.Join(grantees, ","))

	} else {
		__antithesis_instrumentation__.Notify(465759)
	}
	__antithesis_instrumentation__.Notify(465742)

	return parse(query.String())
}
