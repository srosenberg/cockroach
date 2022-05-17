package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

func (d *delegator) delegateShowSchemas(n *tree.ShowSchemas) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465776)
	name, err := d.getSpecifiedOrCurrentDatabase(n.Database)
	if err != nil {
		__antithesis_instrumentation__.Notify(465778)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465779)
	}
	__antithesis_instrumentation__.Notify(465777)
	getSchemasQuery := fmt.Sprintf(`
      SELECT nspname AS schema_name, rolname AS owner
      FROM %[1]s.information_schema.schemata i
      INNER JOIN %[1]s.pg_catalog.pg_namespace n ON (n.nspname = i.schema_name)
      LEFT JOIN %[1]s.pg_catalog.pg_roles r ON (n.nspowner = r.oid)
			WHERE catalog_name = %[2]s
			ORDER BY schema_name`,
		name.String(),
		lexbase.EscapeSQLString(string(name)),
	)

	return parse(getSchemasQuery)
}

func (d *delegator) delegateShowCreateAllSchemas() (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465780)
	sqltelemetry.IncrementShowCounter(sqltelemetry.Create)

	const showCreateAllSchemasQuery = `
	SELECT crdb_internal.show_create_all_schemas(%[1]s) AS create_statement;
`
	databaseLiteral := d.evalCtx.SessionData().Database

	query := fmt.Sprintf(showCreateAllSchemasQuery,
		lexbase.EscapeSQLString(databaseLiteral),
	)

	return parse(query)
}

func (d *delegator) getSpecifiedOrCurrentDatabase(specifiedDB tree.Name) (tree.Name, error) {
	__antithesis_instrumentation__.Notify(465781)
	var name cat.SchemaName
	if specifiedDB != "" {
		__antithesis_instrumentation__.Notify(465784)

		name.SchemaName = specifiedDB
		name.ExplicitSchema = true
	} else {
		__antithesis_instrumentation__.Notify(465785)
	}
	__antithesis_instrumentation__.Notify(465782)

	flags := cat.Flags{AvoidDescriptorCaches: true}
	_, resName, err := d.catalog.ResolveSchema(d.ctx, flags, &name)
	if err != nil {
		__antithesis_instrumentation__.Notify(465786)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(465787)
	}
	__antithesis_instrumentation__.Notify(465783)
	return resName.CatalogName, nil
}
