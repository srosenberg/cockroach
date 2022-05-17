package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowDefaultPrivileges(
	n *tree.ShowDefaultPrivileges,
) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465535)
	currentDatabase, err := d.getSpecifiedOrCurrentDatabase("")
	if err != nil {
		__antithesis_instrumentation__.Notify(465539)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465540)
	}
	__antithesis_instrumentation__.Notify(465536)

	schemaClause := " AND schema_name IS NULL"
	if n.Schema != "" {
		__antithesis_instrumentation__.Notify(465541)
		schemaClause = fmt.Sprintf(" AND schema_name = %s", lexbase.EscapeSQLString(n.Schema.String()))
	} else {
		__antithesis_instrumentation__.Notify(465542)
	}
	__antithesis_instrumentation__.Notify(465537)

	query := fmt.Sprintf(
		"SELECT role, for_all_roles, object_type, grantee, privilege_type FROM crdb_internal.default_privileges WHERE database_name = %s%s",
		lexbase.EscapeSQLString(currentDatabase.Normalize()),
		schemaClause,
	)

	if n.ForAllRoles {
		__antithesis_instrumentation__.Notify(465543)
		query += " AND for_all_roles=true"
	} else {
		__antithesis_instrumentation__.Notify(465544)
		if len(n.Roles) > 0 {
			__antithesis_instrumentation__.Notify(465545)
			targetRoles, err := n.Roles.ToSQLUsernames(d.evalCtx.SessionData(), security.UsernameValidation)
			if err != nil {
				__antithesis_instrumentation__.Notify(465548)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(465549)
			}
			__antithesis_instrumentation__.Notify(465546)

			query = fmt.Sprintf("%s AND for_all_roles=false AND role IN (", query)
			for i, role := range targetRoles {
				__antithesis_instrumentation__.Notify(465550)
				if i != 0 {
					__antithesis_instrumentation__.Notify(465551)
					query += fmt.Sprintf(", '%s'", role.Normalized())
				} else {
					__antithesis_instrumentation__.Notify(465552)
					query += fmt.Sprintf("'%s'", role.Normalized())
				}
			}
			__antithesis_instrumentation__.Notify(465547)

			query += ")"
		} else {
			__antithesis_instrumentation__.Notify(465553)
			query = fmt.Sprintf("%s AND for_all_roles=false AND role = '%s'",
				query, d.evalCtx.SessionData().User())
		}
	}
	__antithesis_instrumentation__.Notify(465538)
	query += " ORDER BY 1,2,3,4,5"
	return parse(query)
}
