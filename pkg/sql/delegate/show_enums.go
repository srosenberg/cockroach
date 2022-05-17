package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowEnums(n *tree.ShowEnums) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465554)
	flags := cat.Flags{AvoidDescriptorCaches: true}
	_, name, err := d.catalog.ResolveSchema(d.ctx, flags, &n.ObjectNamePrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(465557)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465558)
	}
	__antithesis_instrumentation__.Notify(465555)

	schemaClause := ""
	if n.ExplicitSchema {
		__antithesis_instrumentation__.Notify(465559)
		schema := lexbase.EscapeSQLString(name.Schema())
		if name.Schema() == catconstants.PgTempSchemaName {
			__antithesis_instrumentation__.Notify(465561)
			schema = lexbase.EscapeSQLString(d.evalCtx.SessionData().SearchPath.GetTemporarySchemaName())
		} else {
			__antithesis_instrumentation__.Notify(465562)
		}
		__antithesis_instrumentation__.Notify(465560)
		schemaClause = fmt.Sprintf("AND nsp.nspname = %s", schema)
	} else {
		__antithesis_instrumentation__.Notify(465563)
	}
	__antithesis_instrumentation__.Notify(465556)

	query := fmt.Sprintf(`
WITH enums(enumtypid, values) AS (
	SELECT
		enums.enumtypid AS enumtypid,
		array_agg(enums.enumlabel) WITHIN GROUP (ORDER BY (enumsortorder)) AS values
	FROM %[1]s.pg_catalog.pg_enum AS enums
	GROUP BY enumtypid
)
SELECT
	nsp.nspname AS schema,
	types.typname AS name,
	values,
	rl.rolname AS owner
FROM
	%[1]s.pg_catalog.pg_type AS types
	LEFT JOIN enums ON (types.oid = enums.enumtypid)
	LEFT JOIN %[1]s.pg_catalog.pg_roles AS rl on (types.typowner = rl.oid)
	JOIN %[1]s.pg_catalog.pg_namespace AS nsp ON (types.typnamespace = nsp.oid)
WHERE types.typtype = 'e' %[2]s
ORDER BY (nsp.nspname, types.typname)
`, &name.CatalogName, schemaClause)
	return parse(query)
}
