package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

var showEstimatedRowCountClusterSetting = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.show_tables.estimated_row_count.enabled",
	"whether the estimated_row_count is shown on SHOW TABLES. Turning this off "+
		"will improve SHOW TABLES performance.",
	true,
)

func (d *delegator) delegateShowTables(n *tree.ShowTables) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465841)
	flags := cat.Flags{AvoidDescriptorCaches: true}
	_, name, err := d.catalog.ResolveSchema(d.ctx, flags, &n.ObjectNamePrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(465847)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465848)
	}
	__antithesis_instrumentation__.Notify(465842)

	if name.ExplicitSchema && func() bool {
		__antithesis_instrumentation__.Notify(465849)
		return name.ExplicitCatalog == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(465850)
		return name.SchemaName == tree.PublicSchemaName == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(465851)
		return n.ExplicitSchema == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(465852)
		return !n.ExplicitCatalog == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(465853)
		return n.SchemaName == name.CatalogName == true
	}() == true {
		__antithesis_instrumentation__.Notify(465854)
		name.SchemaName, name.ExplicitSchema = "", false
	} else {
		__antithesis_instrumentation__.Notify(465855)
	}
	__antithesis_instrumentation__.Notify(465843)
	var schemaClause string
	if name.ExplicitSchema {
		__antithesis_instrumentation__.Notify(465856)
		schema := lexbase.EscapeSQLString(name.Schema())
		if name.Schema() == catconstants.PgTempSchemaName {
			__antithesis_instrumentation__.Notify(465858)
			schema = lexbase.EscapeSQLString(d.evalCtx.SessionData().SearchPath.GetTemporarySchemaName())
		} else {
			__antithesis_instrumentation__.Notify(465859)
		}
		__antithesis_instrumentation__.Notify(465857)
		schemaClause = fmt.Sprintf("AND ns.nspname = %s", schema)
	} else {
		__antithesis_instrumentation__.Notify(465860)

		schemaClause = "AND ns.nspname NOT IN ('information_schema', 'pg_catalog', 'crdb_internal', 'pg_extension')"
	}
	__antithesis_instrumentation__.Notify(465844)

	const getTablesQuery = `
SELECT ns.nspname AS schema_name,
       pc.relname AS table_name,
       CASE
       WHEN pc.relkind = 'v' THEN 'view'
       WHEN pc.relkind = 'm' THEN 'materialized view'
       WHEN pc.relkind = 'S' THEN 'sequence'
       ELSE 'table'
       END AS type,
       rl.rolname AS owner,
			 %[5]s
       ct.locality AS locality
       %[3]s
FROM %[1]s.pg_catalog.pg_class AS pc
LEFT JOIN %[1]s.pg_catalog.pg_roles AS rl on (pc.relowner = rl.oid)
JOIN %[1]s.pg_catalog.pg_namespace AS ns ON (ns.oid = pc.relnamespace)
%[4]s
%[6]s
LEFT JOIN crdb_internal.tables AS ct ON (pc.oid::int8 = ct.table_id)
WHERE pc.relkind IN ('r', 'v', 'S', 'm') %[2]s
ORDER BY schema_name, table_name
`
	var estimatedRowCount string
	var estimatedRowCountJoin string
	if showEstimatedRowCountClusterSetting.Get(&d.evalCtx.Settings.SV) {
		__antithesis_instrumentation__.Notify(465861)
		estimatedRowCount = "s.estimated_row_count AS estimated_row_count, "
		estimatedRowCountJoin = fmt.Sprintf(
			`LEFT JOIN %[1]s.crdb_internal.table_row_statistics AS s on (s.table_id = pc.oid::INT8)`,
			&name.CatalogName,
		)
	} else {
		__antithesis_instrumentation__.Notify(465862)
	}
	__antithesis_instrumentation__.Notify(465845)
	var descJoin string
	var comment string
	if n.WithComment {
		__antithesis_instrumentation__.Notify(465863)
		descJoin = fmt.Sprintf(
			`LEFT JOIN %s.pg_catalog.pg_description AS pd ON (pc.oid = pd.objoid AND pd.objsubid = 0)`,
			&name.CatalogName,
		)
		comment = `, COALESCE(pd.description, '') AS comment`
	} else {
		__antithesis_instrumentation__.Notify(465864)
	}
	__antithesis_instrumentation__.Notify(465846)
	query := fmt.Sprintf(
		getTablesQuery,
		&name.CatalogName,
		schemaClause,
		comment,
		descJoin,
		estimatedRowCount,
		estimatedRowCountJoin,
	)
	return parse(query)
}
