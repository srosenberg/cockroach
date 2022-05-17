package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (d *delegator) delegateShowGrants(n *tree.ShowGrants) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465565)
	var params []string

	const dbPrivQuery = `
SELECT database_name,
       grantee,
       privilege_type,
			 is_grantable::boolean
  FROM "".crdb_internal.cluster_database_privileges`
	const schemaPrivQuery = `
SELECT table_catalog AS database_name,
       table_schema AS schema_name,
       grantee,
       privilege_type,
       is_grantable::boolean
  FROM "".information_schema.schema_privileges`
	const tablePrivQuery = `
SELECT table_catalog AS database_name,
       table_schema AS schema_name,
       table_name,
       grantee,
       privilege_type,
       is_grantable::boolean
FROM "".information_schema.table_privileges`
	const typePrivQuery = `
SELECT type_catalog AS database_name,
       type_schema AS schema_name,
       type_name,
       grantee,
       privilege_type,
       is_grantable::boolean
FROM "".information_schema.type_privileges`

	var source bytes.Buffer
	var cond bytes.Buffer
	var orderBy string

	if n.Targets != nil && func() bool {
		__antithesis_instrumentation__.Notify(465569)
		return len(n.Targets.Databases) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(465570)

		dbNames := n.Targets.Databases.ToStrings()

		for _, db := range dbNames {
			__antithesis_instrumentation__.Notify(465572)
			name := cat.SchemaName{
				CatalogName:     tree.Name(db),
				SchemaName:      tree.Name(tree.PublicSchema),
				ExplicitCatalog: true,
				ExplicitSchema:  true,
			}
			_, _, err := d.catalog.ResolveSchema(d.ctx, cat.Flags{AvoidDescriptorCaches: true}, &name)
			if err != nil {
				__antithesis_instrumentation__.Notify(465574)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(465575)
			}
			__antithesis_instrumentation__.Notify(465573)
			params = append(params, lexbase.EscapeSQLString(db))
		}
		__antithesis_instrumentation__.Notify(465571)

		fmt.Fprint(&source, dbPrivQuery)
		orderBy = "1,2,3"
		if len(params) == 0 {
			__antithesis_instrumentation__.Notify(465576)

			cond.WriteString(`WHERE false`)
		} else {
			__antithesis_instrumentation__.Notify(465577)
			fmt.Fprintf(&cond, `WHERE database_name IN (%s)`, strings.Join(params, ","))
		}
	} else {
		__antithesis_instrumentation__.Notify(465578)
		if n.Targets != nil && func() bool {
			__antithesis_instrumentation__.Notify(465579)
			return len(n.Targets.Schemas) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(465580)
			currDB := d.evalCtx.SessionData().Database

			for _, schema := range n.Targets.Schemas {
				__antithesis_instrumentation__.Notify(465582)
				_, _, err := d.catalog.ResolveSchema(d.ctx, cat.Flags{AvoidDescriptorCaches: true}, &schema)
				if err != nil {
					__antithesis_instrumentation__.Notify(465585)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(465586)
				}
				__antithesis_instrumentation__.Notify(465583)
				dbName := currDB
				if schema.ExplicitCatalog {
					__antithesis_instrumentation__.Notify(465587)
					dbName = schema.Catalog()
				} else {
					__antithesis_instrumentation__.Notify(465588)
				}
				__antithesis_instrumentation__.Notify(465584)
				params = append(params, fmt.Sprintf("(%s,%s)", lexbase.EscapeSQLString(dbName), lexbase.EscapeSQLString(schema.Schema())))
			}
			__antithesis_instrumentation__.Notify(465581)

			fmt.Fprint(&source, schemaPrivQuery)
			orderBy = "1,2,3,4"

			if len(params) != 0 {
				__antithesis_instrumentation__.Notify(465589)
				fmt.Fprintf(
					&cond,
					`WHERE (database_name, schema_name) IN (%s)`,
					strings.Join(params, ","),
				)
			} else {
				__antithesis_instrumentation__.Notify(465590)
			}
		} else {
			__antithesis_instrumentation__.Notify(465591)
			if n.Targets != nil && func() bool {
				__antithesis_instrumentation__.Notify(465592)
				return len(n.Targets.Types) > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(465593)
				for _, typName := range n.Targets.Types {
					__antithesis_instrumentation__.Notify(465595)
					t, err := d.catalog.ResolveType(d.ctx, typName)
					if err != nil {
						__antithesis_instrumentation__.Notify(465597)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(465598)
					}
					__antithesis_instrumentation__.Notify(465596)
					if t.UserDefined() {
						__antithesis_instrumentation__.Notify(465599)
						params = append(
							params,
							fmt.Sprintf(
								"(%s, %s, %s)",
								lexbase.EscapeSQLString(t.TypeMeta.Name.Catalog),
								lexbase.EscapeSQLString(t.TypeMeta.Name.Schema),
								lexbase.EscapeSQLString(t.TypeMeta.Name.Name),
							),
						)
					} else {
						__antithesis_instrumentation__.Notify(465600)
						params = append(
							params,
							fmt.Sprintf(
								"(%s, 'pg_catalog', %s)",
								lexbase.EscapeSQLString(t.TypeMeta.Name.Catalog),
								lexbase.EscapeSQLString(t.TypeMeta.Name.Name),
							),
						)
					}
				}
				__antithesis_instrumentation__.Notify(465594)

				fmt.Fprint(&source, typePrivQuery)
				orderBy = "1,2,3,4,5"
				if len(params) == 0 {
					__antithesis_instrumentation__.Notify(465601)
					dbNameClause := "true"

					if currDB := d.evalCtx.SessionData().Database; currDB != "" {
						__antithesis_instrumentation__.Notify(465603)
						dbNameClause = fmt.Sprintf("database_name = %s", lexbase.EscapeSQLString(currDB))
					} else {
						__antithesis_instrumentation__.Notify(465604)
					}
					__antithesis_instrumentation__.Notify(465602)
					cond.WriteString(fmt.Sprintf(`WHERE %s`, dbNameClause))
				} else {
					__antithesis_instrumentation__.Notify(465605)
					fmt.Fprintf(
						&cond,
						`WHERE (database_name, schema_name, type_name) IN (%s)`,
						strings.Join(params, ","),
					)
				}
			} else {
				__antithesis_instrumentation__.Notify(465606)
				orderBy = "1,2,3,4,5"

				if n.Targets != nil {
					__antithesis_instrumentation__.Notify(465607)
					fmt.Fprint(&source, tablePrivQuery)

					var allTables tree.TableNames

					for _, tableTarget := range n.Targets.Tables {
						__antithesis_instrumentation__.Notify(465610)
						tableGlob, err := tableTarget.NormalizeTablePattern()
						if err != nil {
							__antithesis_instrumentation__.Notify(465613)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(465614)
						}
						__antithesis_instrumentation__.Notify(465611)

						tables, _, err := cat.ExpandDataSourceGlob(
							d.ctx, d.catalog, cat.Flags{AvoidDescriptorCaches: true}, tableGlob,
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(465615)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(465616)
						}
						__antithesis_instrumentation__.Notify(465612)
						allTables = append(allTables, tables...)
					}
					__antithesis_instrumentation__.Notify(465608)

					for i := range allTables {
						__antithesis_instrumentation__.Notify(465617)
						params = append(params, fmt.Sprintf("(%s,%s,%s)",
							lexbase.EscapeSQLString(allTables[i].Catalog()),
							lexbase.EscapeSQLString(allTables[i].Schema()),
							lexbase.EscapeSQLString(allTables[i].Table())))
					}
					__antithesis_instrumentation__.Notify(465609)

					if len(params) == 0 {
						__antithesis_instrumentation__.Notify(465618)

						cond.WriteString(`WHERE false`)
					} else {
						__antithesis_instrumentation__.Notify(465619)
						fmt.Fprintf(&cond, `WHERE (database_name, schema_name, table_name) IN (%s)`, strings.Join(params, ","))
					}
				} else {
					__antithesis_instrumentation__.Notify(465620)

					source.WriteString(
						`SELECT database_name, schema_name, table_name AS relation_name, grantee, privilege_type FROM (`,
					)
					source.WriteString(tablePrivQuery)
					source.WriteByte(')')
					source.WriteString(` UNION ALL ` +
						`SELECT database_name, schema_name, NULL::STRING AS relation_name, grantee, privilege_type FROM (`)
					source.WriteString(schemaPrivQuery)
					source.WriteByte(')')
					source.WriteString(` UNION ALL ` +
						`SELECT database_name, NULL::STRING AS schema_name, NULL::STRING AS relation_name, grantee, privilege_type FROM (`)
					source.WriteString(dbPrivQuery)
					source.WriteByte(')')
					source.WriteString(` UNION ALL ` +
						`SELECT database_name, schema_name, type_name AS relation_name, grantee, privilege_type FROM (`)
					source.WriteString(typePrivQuery)
					source.WriteByte(')')

					if currDB := d.evalCtx.SessionData().Database; currDB != "" {
						__antithesis_instrumentation__.Notify(465621)
						fmt.Fprintf(&cond, ` WHERE database_name = %s`, lexbase.EscapeSQLString(currDB))
					} else {
						__antithesis_instrumentation__.Notify(465622)
						cond.WriteString(`WHERE true`)
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(465566)

	if n.Grantees != nil {
		__antithesis_instrumentation__.Notify(465623)
		params = params[:0]
		grantees, err := n.Grantees.ToSQLUsernames(d.evalCtx.SessionData(), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(465626)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(465627)
		}
		__antithesis_instrumentation__.Notify(465624)
		for _, grantee := range grantees {
			__antithesis_instrumentation__.Notify(465628)
			params = append(params, lexbase.EscapeSQLString(grantee.Normalized()))
		}
		__antithesis_instrumentation__.Notify(465625)
		fmt.Fprintf(&cond, ` AND grantee IN (%s)`, strings.Join(params, ","))
	} else {
		__antithesis_instrumentation__.Notify(465629)
	}
	__antithesis_instrumentation__.Notify(465567)
	query := fmt.Sprintf(`
		SELECT * FROM (%s) %s ORDER BY %s
	`, source.String(), cond.String(), orderBy)

	for _, p := range n.Grantees {
		__antithesis_instrumentation__.Notify(465630)

		user, err := p.ToSQLUsername(d.evalCtx.SessionData(), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(465633)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(465634)
		}
		__antithesis_instrumentation__.Notify(465631)
		userExists, err := d.catalog.RoleExists(d.ctx, user)
		if err != nil {
			__antithesis_instrumentation__.Notify(465635)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(465636)
		}
		__antithesis_instrumentation__.Notify(465632)
		if !userExists {
			__antithesis_instrumentation__.Notify(465637)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "role/user %q does not exist", user)
		} else {
			__antithesis_instrumentation__.Notify(465638)
		}
	}
	__antithesis_instrumentation__.Notify(465568)

	return parse(query)
}
