package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

var ValidVars = make(map[string]struct{})

func (d *delegator) delegateShowVar(n *tree.ShowVar) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465874)
	origName := n.Name
	name := strings.ToLower(n.Name)

	if name == "locality" {
		__antithesis_instrumentation__.Notify(465878)
		sqltelemetry.IncrementShowCounter(sqltelemetry.Locality)
	} else {
		__antithesis_instrumentation__.Notify(465879)
	}
	__antithesis_instrumentation__.Notify(465875)

	if name == "all" {
		__antithesis_instrumentation__.Notify(465880)
		return parse(
			"SELECT variable, value FROM crdb_internal.session_variables WHERE hidden = FALSE",
		)
	} else {
		__antithesis_instrumentation__.Notify(465881)
	}
	__antithesis_instrumentation__.Notify(465876)

	if _, ok := ValidVars[name]; !ok {
		__antithesis_instrumentation__.Notify(465882)

		if strings.Contains(name, ".") {
			__antithesis_instrumentation__.Notify(465884)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(465885)
		}
		__antithesis_instrumentation__.Notify(465883)
		return nil, pgerror.Newf(pgcode.UndefinedObject,
			"unrecognized configuration parameter %q", origName)
	} else {
		__antithesis_instrumentation__.Notify(465886)
	}
	__antithesis_instrumentation__.Notify(465877)

	varName := lexbase.EscapeSQLString(name)
	nm := tree.Name(name)
	return parse(fmt.Sprintf(
		`SELECT value AS %[1]s FROM crdb_internal.session_variables WHERE variable = %[2]s`,
		nm.String(), varName,
	))
}
