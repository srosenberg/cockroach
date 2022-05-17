package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/sql/vtable"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
	"golang.org/x/text/collate"
)

const (
	pgCatalogName = catconstants.PgCatalogName
)

var pgCatalogNameDString = tree.NewDString(pgCatalogName)

var informationSchema = virtualSchema{
	name: catconstants.InformationSchemaName,
	undefinedTables: buildStringSet(

		"_pg_foreign_data_wrappers",
		"_pg_foreign_servers",
		"_pg_foreign_table_columns",
		"_pg_foreign_tables",
		"_pg_user_mappings",
		"sql_languages",
		"sql_packages",
		"sql_sizing_profiles",
	),
	tableDefs: map[descpb.ID]virtualSchemaDef{
		catconstants.InformationSchemaAdministrableRoleAuthorizationsID:   informationSchemaAdministrableRoleAuthorizations,
		catconstants.InformationSchemaApplicableRolesID:                   informationSchemaApplicableRoles,
		catconstants.InformationSchemaAttributesTableID:                   informationSchemaAttributesTable,
		catconstants.InformationSchemaCharacterSets:                       informationSchemaCharacterSets,
		catconstants.InformationSchemaCheckConstraintRoutineUsageTableID:  informationSchemaCheckConstraintRoutineUsageTable,
		catconstants.InformationSchemaCheckConstraints:                    informationSchemaCheckConstraints,
		catconstants.InformationSchemaCollationCharacterSetApplicability:  informationSchemaCollationCharacterSetApplicability,
		catconstants.InformationSchemaCollations:                          informationSchemaCollations,
		catconstants.InformationSchemaColumnColumnUsageTableID:            informationSchemaColumnColumnUsageTable,
		catconstants.InformationSchemaColumnDomainUsageTableID:            informationSchemaColumnDomainUsageTable,
		catconstants.InformationSchemaColumnOptionsTableID:                informationSchemaColumnOptionsTable,
		catconstants.InformationSchemaColumnPrivilegesID:                  informationSchemaColumnPrivileges,
		catconstants.InformationSchemaColumnStatisticsTableID:             informationSchemaColumnStatisticsTable,
		catconstants.InformationSchemaColumnUDTUsageID:                    informationSchemaColumnUDTUsage,
		catconstants.InformationSchemaColumnsExtensionsTableID:            informationSchemaColumnsExtensionsTable,
		catconstants.InformationSchemaColumnsTableID:                      informationSchemaColumnsTable,
		catconstants.InformationSchemaConstraintColumnUsageTableID:        informationSchemaConstraintColumnUsageTable,
		catconstants.InformationSchemaConstraintTableUsageTableID:         informationSchemaConstraintTableUsageTable,
		catconstants.InformationSchemaDataTypePrivilegesTableID:           informationSchemaDataTypePrivilegesTable,
		catconstants.InformationSchemaDomainConstraintsTableID:            informationSchemaDomainConstraintsTable,
		catconstants.InformationSchemaDomainUdtUsageTableID:               informationSchemaDomainUdtUsageTable,
		catconstants.InformationSchemaDomainsTableID:                      informationSchemaDomainsTable,
		catconstants.InformationSchemaElementTypesTableID:                 informationSchemaElementTypesTable,
		catconstants.InformationSchemaEnabledRolesID:                      informationSchemaEnabledRoles,
		catconstants.InformationSchemaEnginesTableID:                      informationSchemaEnginesTable,
		catconstants.InformationSchemaEventsTableID:                       informationSchemaEventsTable,
		catconstants.InformationSchemaFilesTableID:                        informationSchemaFilesTable,
		catconstants.InformationSchemaForeignDataWrapperOptionsTableID:    informationSchemaForeignDataWrapperOptionsTable,
		catconstants.InformationSchemaForeignDataWrappersTableID:          informationSchemaForeignDataWrappersTable,
		catconstants.InformationSchemaForeignServerOptionsTableID:         informationSchemaForeignServerOptionsTable,
		catconstants.InformationSchemaForeignServersTableID:               informationSchemaForeignServersTable,
		catconstants.InformationSchemaForeignTableOptionsTableID:          informationSchemaForeignTableOptionsTable,
		catconstants.InformationSchemaForeignTablesTableID:                informationSchemaForeignTablesTable,
		catconstants.InformationSchemaInformationSchemaCatalogNameTableID: informationSchemaInformationSchemaCatalogNameTable,
		catconstants.InformationSchemaKeyColumnUsageTableID:               informationSchemaKeyColumnUsageTable,
		catconstants.InformationSchemaKeywordsTableID:                     informationSchemaKeywordsTable,
		catconstants.InformationSchemaOptimizerTraceTableID:               informationSchemaOptimizerTraceTable,
		catconstants.InformationSchemaParametersTableID:                   informationSchemaParametersTable,
		catconstants.InformationSchemaPartitionsTableID:                   informationSchemaPartitionsTable,
		catconstants.InformationSchemaPluginsTableID:                      informationSchemaPluginsTable,
		catconstants.InformationSchemaProcesslistTableID:                  informationSchemaProcesslistTable,
		catconstants.InformationSchemaProfilingTableID:                    informationSchemaProfilingTable,
		catconstants.InformationSchemaReferentialConstraintsTableID:       informationSchemaReferentialConstraintsTable,
		catconstants.InformationSchemaResourceGroupsTableID:               informationSchemaResourceGroupsTable,
		catconstants.InformationSchemaRoleColumnGrantsTableID:             informationSchemaRoleColumnGrantsTable,
		catconstants.InformationSchemaRoleRoutineGrantsTableID:            informationSchemaRoleRoutineGrantsTable,
		catconstants.InformationSchemaRoleTableGrantsID:                   informationSchemaRoleTableGrants,
		catconstants.InformationSchemaRoleUdtGrantsTableID:                informationSchemaRoleUdtGrantsTable,
		catconstants.InformationSchemaRoleUsageGrantsTableID:              informationSchemaRoleUsageGrantsTable,
		catconstants.InformationSchemaRoutinePrivilegesTableID:            informationSchemaRoutinePrivilegesTable,
		catconstants.InformationSchemaRoutineTableID:                      informationSchemaRoutineTable,
		catconstants.InformationSchemaSQLFeaturesTableID:                  informationSchemaSQLFeaturesTable,
		catconstants.InformationSchemaSQLImplementationInfoTableID:        informationSchemaSQLImplementationInfoTable,
		catconstants.InformationSchemaSQLPartsTableID:                     informationSchemaSQLPartsTable,
		catconstants.InformationSchemaSQLSizingTableID:                    informationSchemaSQLSizingTable,
		catconstants.InformationSchemaSchemataExtensionsTableID:           informationSchemaSchemataExtensionsTable,
		catconstants.InformationSchemaSchemataTableID:                     informationSchemaSchemataTable,
		catconstants.InformationSchemaSchemataTablePrivilegesID:           informationSchemaSchemataTablePrivileges,
		catconstants.InformationSchemaSequencesID:                         informationSchemaSequences,
		catconstants.InformationSchemaSessionVariables:                    informationSchemaSessionVariables,
		catconstants.InformationSchemaStGeometryColumnsTableID:            informationSchemaStGeometryColumnsTable,
		catconstants.InformationSchemaStSpatialReferenceSystemsTableID:    informationSchemaStSpatialReferenceSystemsTable,
		catconstants.InformationSchemaStUnitsOfMeasureTableID:             informationSchemaStUnitsOfMeasureTable,
		catconstants.InformationSchemaStatisticsTableID:                   informationSchemaStatisticsTable,
		catconstants.InformationSchemaTableConstraintTableID:              informationSchemaTableConstraintTable,
		catconstants.InformationSchemaTableConstraintsExtensionsTableID:   informationSchemaTableConstraintsExtensionsTable,
		catconstants.InformationSchemaTablePrivilegesID:                   informationSchemaTablePrivileges,
		catconstants.InformationSchemaTablesExtensionsTableID:             informationSchemaTablesExtensionsTable,
		catconstants.InformationSchemaTablesTableID:                       informationSchemaTablesTable,
		catconstants.InformationSchemaTablespacesExtensionsTableID:        informationSchemaTablespacesExtensionsTable,
		catconstants.InformationSchemaTablespacesTableID:                  informationSchemaTablespacesTable,
		catconstants.InformationSchemaTransformsTableID:                   informationSchemaTransformsTable,
		catconstants.InformationSchemaTriggeredUpdateColumnsTableID:       informationSchemaTriggeredUpdateColumnsTable,
		catconstants.InformationSchemaTriggersTableID:                     informationSchemaTriggersTable,
		catconstants.InformationSchemaTypePrivilegesID:                    informationSchemaTypePrivilegesTable,
		catconstants.InformationSchemaUdtPrivilegesTableID:                informationSchemaUdtPrivilegesTable,
		catconstants.InformationSchemaUsagePrivilegesTableID:              informationSchemaUsagePrivilegesTable,
		catconstants.InformationSchemaUserAttributesTableID:               informationSchemaUserAttributesTable,
		catconstants.InformationSchemaUserDefinedTypesTableID:             informationSchemaUserDefinedTypesTable,
		catconstants.InformationSchemaUserMappingOptionsTableID:           informationSchemaUserMappingOptionsTable,
		catconstants.InformationSchemaUserMappingsTableID:                 informationSchemaUserMappingsTable,
		catconstants.InformationSchemaUserPrivilegesID:                    informationSchemaUserPrivileges,
		catconstants.InformationSchemaViewColumnUsageTableID:              informationSchemaViewColumnUsageTable,
		catconstants.InformationSchemaViewRoutineUsageTableID:             informationSchemaViewRoutineUsageTable,
		catconstants.InformationSchemaViewTableUsageTableID:               informationSchemaViewTableUsageTable,
		catconstants.InformationSchemaViewsTableID:                        informationSchemaViewsTable,
	},
	tableValidator:             validateInformationSchemaTable,
	validWithNoDatabaseContext: true,
}

func buildStringSet(ss ...string) map[string]struct{} {
	__antithesis_instrumentation__.Notify(496707)
	m := map[string]struct{}{}
	for _, s := range ss {
		__antithesis_instrumentation__.Notify(496709)
		m[s] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(496708)
	return m
}

var (
	emptyString = tree.NewDString("")

	yesString = tree.NewDString("YES")
	noString  = tree.NewDString("NO")
)

func yesOrNoDatum(b bool) tree.Datum {
	__antithesis_instrumentation__.Notify(496710)
	if b {
		__antithesis_instrumentation__.Notify(496712)
		return yesString
	} else {
		__antithesis_instrumentation__.Notify(496713)
	}
	__antithesis_instrumentation__.Notify(496711)
	return noString
}

func dNameOrNull(s string) tree.Datum {
	__antithesis_instrumentation__.Notify(496714)
	if s == "" {
		__antithesis_instrumentation__.Notify(496716)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(496717)
	}
	__antithesis_instrumentation__.Notify(496715)
	return tree.NewDName(s)
}

func dIntFnOrNull(fn func() (int32, bool)) tree.Datum {
	__antithesis_instrumentation__.Notify(496718)
	if n, ok := fn(); ok {
		__antithesis_instrumentation__.Notify(496720)
		return tree.NewDInt(tree.DInt(n))
	} else {
		__antithesis_instrumentation__.Notify(496721)
	}
	__antithesis_instrumentation__.Notify(496719)
	return tree.DNull
}

func validateInformationSchemaTable(table *descpb.TableDescriptor) error {
	__antithesis_instrumentation__.Notify(496722)

	for i := range table.Columns {
		__antithesis_instrumentation__.Notify(496724)
		if table.Columns[i].Type.Family() == types.BoolFamily {
			__antithesis_instrumentation__.Notify(496725)
			return errors.Errorf("information_schema tables should never use BOOL columns. "+
				"See the comment about yesOrNoDatum. Found BOOL column in %s.", table.Name)
		} else {
			__antithesis_instrumentation__.Notify(496726)
		}
	}
	__antithesis_instrumentation__.Notify(496723)
	return nil
}

var informationSchemaAdministrableRoleAuthorizations = virtualSchemaTable{
	comment: `roles for which the current user has admin option
` + docs.URL("information-schema.html#administrable_role_authorizations") + `
https://www.postgresql.org/docs/9.5/infoschema-administrable-role-authorizations.html`,
	schema: vtable.InformationSchemaAdministrableRoleAuthorizations,
	populate: func(
		ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error,
	) error {
		__antithesis_instrumentation__.Notify(496727)
		return populateRoleHierarchy(ctx, p, addRow, true)
	},
}

var informationSchemaApplicableRoles = virtualSchemaTable{
	comment: `roles available to the current user
` + docs.URL("information-schema.html#applicable_roles") + `
https://www.postgresql.org/docs/9.5/infoschema-applicable-roles.html`,
	schema: vtable.InformationSchemaApplicableRoles,
	populate: func(
		ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error,
	) error {
		__antithesis_instrumentation__.Notify(496728)
		return populateRoleHierarchy(ctx, p, addRow, false)
	},
}

func populateRoleHierarchy(
	ctx context.Context, p *planner, addRow func(...tree.Datum) error, onlyIsAdmin bool,
) error {
	__antithesis_instrumentation__.Notify(496729)
	allRoles, err := p.MemberOfWithAdminOption(ctx, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(496731)
		return err
	} else {
		__antithesis_instrumentation__.Notify(496732)
	}
	__antithesis_instrumentation__.Notify(496730)
	return forEachRoleMembership(
		ctx, p.ExecCfg().InternalExecutor, p.Txn(),
		func(role, member security.SQLUsername, isAdmin bool) error {
			__antithesis_instrumentation__.Notify(496733)

			isRole := member == p.User()
			_, hasRole := allRoles[member]
			if (hasRole || func() bool {
				__antithesis_instrumentation__.Notify(496735)
				return isRole == true
			}() == true) && func() bool {
				__antithesis_instrumentation__.Notify(496736)
				return (!onlyIsAdmin || func() bool {
					__antithesis_instrumentation__.Notify(496737)
					return isAdmin == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(496738)
				if err := addRow(
					tree.NewDString(member.Normalized()),
					tree.NewDString(role.Normalized()),
					yesOrNoDatum(isAdmin),
				); err != nil {
					__antithesis_instrumentation__.Notify(496739)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496740)
				}
			} else {
				__antithesis_instrumentation__.Notify(496741)
			}
			__antithesis_instrumentation__.Notify(496734)
			return nil
		},
	)
}

var informationSchemaCharacterSets = virtualSchemaTable{
	comment: `character sets available in the current database
` + docs.URL("information-schema.html#character_sets") + `
https://www.postgresql.org/docs/9.5/infoschema-character-sets.html`,
	schema: vtable.InformationSchemaCharacterSets,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496742)
		return forEachDatabaseDesc(ctx, p, nil, true,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(496743)
				return addRow(
					tree.DNull,
					tree.DNull,
					tree.NewDString("UTF8"),
					tree.NewDString("UCS"),
					tree.NewDString("UTF8"),
					tree.NewDString(db.GetName()),
					tree.DNull,
					tree.DNull,
				)
			})
	},
}

var informationSchemaCheckConstraints = virtualSchemaTable{
	comment: `check constraints
` + docs.URL("information-schema.html#check_constraints") + `
https://www.postgresql.org/docs/9.5/infoschema-check-constraints.html`,
	schema: vtable.InformationSchemaCheckConstraints,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496744)
		h := makeOidHasher()
		return forEachTableDescWithTableLookup(ctx, p, dbContext, hideVirtual, func(
			db catalog.DatabaseDescriptor,
			scName string,
			table catalog.TableDescriptor,
			tableLookup tableLookupFn,
		) error {
			__antithesis_instrumentation__.Notify(496745)
			conInfo, err := table.GetConstraintInfoWithLookup(tableLookup.getTableByID)
			if err != nil {
				__antithesis_instrumentation__.Notify(496749)
				return err
			} else {
				__antithesis_instrumentation__.Notify(496750)
			}
			__antithesis_instrumentation__.Notify(496746)
			dbNameStr := tree.NewDString(db.GetName())
			scNameStr := tree.NewDString(scName)
			for conName, con := range conInfo {
				__antithesis_instrumentation__.Notify(496751)

				if con.Kind != descpb.ConstraintTypeCheck {
					__antithesis_instrumentation__.Notify(496753)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496754)
				}
				__antithesis_instrumentation__.Notify(496752)
				conNameStr := tree.NewDString(conName)

				chkExprStr := tree.NewDString(fmt.Sprintf("((%s))", con.Details))
				if err := addRow(
					dbNameStr,
					scNameStr,
					conNameStr,
					chkExprStr,
				); err != nil {
					__antithesis_instrumentation__.Notify(496755)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496756)
				}
			}
			__antithesis_instrumentation__.Notify(496747)

			for _, column := range table.PublicColumns() {
				__antithesis_instrumentation__.Notify(496757)

				if column.IsHidden() || func() bool {
					__antithesis_instrumentation__.Notify(496759)
					return column.IsNullable() == true
				}() == true {
					__antithesis_instrumentation__.Notify(496760)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496761)
				}
				__antithesis_instrumentation__.Notify(496758)

				conNameStr := tree.NewDString(fmt.Sprintf(
					"%s_%s_%d_not_null", h.NamespaceOid(db.GetID(), scName), tableOid(table.GetID()), column.Ordinal()+1,
				))
				chkExprStr := tree.NewDString(fmt.Sprintf(
					"%s IS NOT NULL", column.GetName(),
				))
				if err := addRow(
					dbNameStr,
					scNameStr,
					conNameStr,
					chkExprStr,
				); err != nil {
					__antithesis_instrumentation__.Notify(496762)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496763)
				}
			}
			__antithesis_instrumentation__.Notify(496748)
			return nil
		})
	},
}

var informationSchemaColumnPrivileges = virtualSchemaTable{
	comment: `column privilege grants (incomplete)
` + docs.URL("information-schema.html#column_privileges") + `
https://www.postgresql.org/docs/9.5/infoschema-column-privileges.html`,
	schema: vtable.InformationSchemaColumnPrivileges,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496764)
		return forEachTableDesc(ctx, p, dbContext, virtualMany, func(
			db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor,
		) error {
			__antithesis_instrumentation__.Notify(496765)
			dbNameStr := tree.NewDString(db.GetName())
			scNameStr := tree.NewDString(scName)
			columndata := privilege.List{privilege.SELECT, privilege.INSERT, privilege.UPDATE}
			for _, u := range table.GetPrivileges().Users {
				__antithesis_instrumentation__.Notify(496767)
				for _, priv := range columndata {
					__antithesis_instrumentation__.Notify(496768)
					if priv.Mask()&u.Privileges != 0 {
						__antithesis_instrumentation__.Notify(496769)
						for _, cd := range table.PublicColumns() {
							__antithesis_instrumentation__.Notify(496770)
							if err := addRow(
								tree.DNull,
								tree.NewDString(u.User().Normalized()),
								dbNameStr,
								scNameStr,
								tree.NewDString(table.GetName()),
								tree.NewDString(cd.GetName()),
								tree.NewDString(priv.String()),
								tree.DNull,
							); err != nil {
								__antithesis_instrumentation__.Notify(496771)
								return err
							} else {
								__antithesis_instrumentation__.Notify(496772)
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(496773)
					}
				}
			}
			__antithesis_instrumentation__.Notify(496766)
			return nil
		})
	},
}

var informationSchemaColumnsTable = virtualSchemaTable{
	comment: `table and view columns (incomplete)
` + docs.URL("information-schema.html#columns") + `
https://www.postgresql.org/docs/9.5/infoschema-columns.html`,
	schema: vtable.InformationSchemaColumns,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496774)

		comments, err := getComments(ctx, p)
		if err != nil {
			__antithesis_instrumentation__.Notify(496777)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496778)
		}
		__antithesis_instrumentation__.Notify(496775)

		commentMap := make(map[tree.DInt]map[tree.DInt]string)
		for _, comment := range comments {
			__antithesis_instrumentation__.Notify(496779)
			objID := tree.MustBeDInt(comment[0])
			objSubID := tree.MustBeDInt(comment[1])
			description := comment[2].String()
			commentType := tree.MustBeDInt(comment[3])
			if commentType == 2 {
				__antithesis_instrumentation__.Notify(496780)
				if commentMap[objID] == nil {
					__antithesis_instrumentation__.Notify(496782)
					commentMap[objID] = make(map[tree.DInt]string)
				} else {
					__antithesis_instrumentation__.Notify(496783)
				}
				__antithesis_instrumentation__.Notify(496781)
				commentMap[objID][objSubID] = description
			} else {
				__antithesis_instrumentation__.Notify(496784)
			}
		}
		__antithesis_instrumentation__.Notify(496776)

		return forEachTableDesc(ctx, p, dbContext, virtualMany, func(
			db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor,
		) error {
			__antithesis_instrumentation__.Notify(496785)
			dbNameStr := tree.NewDString(db.GetName())
			scNameStr := tree.NewDString(scName)
			for _, column := range table.AccessibleColumns() {
				__antithesis_instrumentation__.Notify(496787)
				collationCatalog := tree.DNull
				collationSchema := tree.DNull
				collationName := tree.DNull
				if locale := column.GetType().Locale(); locale != "" {
					__antithesis_instrumentation__.Notify(496793)
					collationCatalog = dbNameStr
					collationSchema = pgCatalogNameDString
					collationName = tree.NewDString(locale)
				} else {
					__antithesis_instrumentation__.Notify(496794)
				}
				__antithesis_instrumentation__.Notify(496788)
				colDefault := tree.DNull
				if column.HasDefault() {
					__antithesis_instrumentation__.Notify(496795)
					colExpr, err := schemaexpr.FormatExprForDisplay(
						ctx, table, column.GetDefaultExpr(), &p.semaCtx, p.SessionData(), tree.FmtParsable,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(496797)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496798)
					}
					__antithesis_instrumentation__.Notify(496796)
					colDefault = tree.NewDString(colExpr)
				} else {
					__antithesis_instrumentation__.Notify(496799)
				}
				__antithesis_instrumentation__.Notify(496789)
				colComputed := emptyString
				if column.IsComputed() {
					__antithesis_instrumentation__.Notify(496800)
					colExpr, err := schemaexpr.FormatExprForDisplay(
						ctx, table, column.GetComputeExpr(), &p.semaCtx, p.SessionData(), tree.FmtSimple,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(496802)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496803)
					}
					__antithesis_instrumentation__.Notify(496801)
					colComputed = tree.NewDString(colExpr)
				} else {
					__antithesis_instrumentation__.Notify(496804)
				}
				__antithesis_instrumentation__.Notify(496790)
				colGeneratedAsIdentity := emptyString
				if column.IsGeneratedAsIdentity() {
					__antithesis_instrumentation__.Notify(496805)
					if column.IsGeneratedAlwaysAsIdentity() {
						__antithesis_instrumentation__.Notify(496806)
						colGeneratedAsIdentity = tree.NewDString(
							"generated always as identity")
					} else {
						__antithesis_instrumentation__.Notify(496807)
						if column.IsGeneratedByDefaultAsIdentity() {
							__antithesis_instrumentation__.Notify(496808)
							colGeneratedAsIdentity = tree.NewDString(
								"generated by default as identity")
						} else {
							__antithesis_instrumentation__.Notify(496809)
							return errors.AssertionFailedf(
								"column %s is of wrong generated as identity type (neither ALWAYS nor BY DEFAULT)",
								column.GetName(),
							)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(496810)
				}
				__antithesis_instrumentation__.Notify(496791)

				tableID := tree.DInt(table.GetID())
				columnID := tree.DInt(column.GetID())
				description := commentMap[tableID][columnID]

				udtSchema := pgCatalogNameDString
				typeMetaName := column.GetType().TypeMeta.Name
				if typeMetaName != nil {
					__antithesis_instrumentation__.Notify(496811)
					udtSchema = tree.NewDString(typeMetaName.Schema)
				} else {
					__antithesis_instrumentation__.Notify(496812)
				}
				__antithesis_instrumentation__.Notify(496792)

				err := addRow(
					dbNameStr,
					scNameStr,
					tree.NewDString(table.GetName()),
					tree.NewDString(column.GetName()),
					tree.NewDString(description),
					tree.NewDInt(tree.DInt(column.GetPGAttributeNum())),
					colDefault,
					yesOrNoDatum(column.IsNullable()),
					tree.NewDString(column.GetType().InformationSchemaName()),
					characterMaximumLength(column.GetType()),
					characterOctetLength(column.GetType()),
					numericPrecision(column.GetType()),
					numericPrecisionRadix(column.GetType()),
					numericScale(column.GetType()),
					datetimePrecision(column.GetType()),
					tree.DNull,
					tree.DNull,
					tree.DNull,
					tree.DNull,
					tree.DNull,
					collationCatalog,
					collationSchema,
					collationName,
					tree.DNull,
					tree.DNull,
					tree.DNull,
					dbNameStr,
					udtSchema,
					tree.NewDString(column.GetType().PGName()),
					tree.DNull,
					tree.DNull,
					tree.DNull,
					tree.DNull,
					tree.DNull,
					tree.DNull,
					yesOrNoDatum(column.IsGeneratedAsIdentity()),
					colGeneratedAsIdentity,

					tree.DNull,
					tree.DNull,
					tree.DNull,
					tree.DNull,
					tree.DNull,
					yesOrNoDatum(column.IsComputed()),
					colComputed,
					yesOrNoDatum(table.IsTable() && func() bool {
						__antithesis_instrumentation__.Notify(496813)
						return !table.IsVirtualTable() == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(496814)
						return !column.IsComputed() == true
					}() == true,
					),
					yesOrNoDatum(column.IsHidden()),
					tree.NewDString(column.GetType().SQLString()),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(496815)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496816)
				}
			}
			__antithesis_instrumentation__.Notify(496786)
			return nil
		})
	},
}

var informationSchemaColumnUDTUsage = virtualSchemaTable{
	comment: `columns with user defined types
` + docs.URL("information-schema.html#column_udt_usage") + `
https://www.postgresql.org/docs/current/infoschema-column-udt-usage.html`,
	schema: vtable.InformationSchemaColumnUDTUsage,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496817)
		return forEachTableDesc(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(496818)
				dbNameStr := tree.NewDString(db.GetName())
				scNameStr := tree.NewDString(scName)
				tbNameStr := tree.NewDString(table.GetName())
				for _, col := range table.PublicColumns() {
					__antithesis_instrumentation__.Notify(496820)
					if !col.GetType().UserDefined() {
						__antithesis_instrumentation__.Notify(496822)
						continue
					} else {
						__antithesis_instrumentation__.Notify(496823)
					}
					__antithesis_instrumentation__.Notify(496821)
					if err := addRow(
						tree.NewDString(col.GetType().TypeMeta.Name.Catalog),
						tree.NewDString(col.GetType().TypeMeta.Name.Schema),
						tree.NewDString(col.GetType().TypeMeta.Name.Name),
						dbNameStr,
						scNameStr,
						tbNameStr,
						tree.NewDString(col.GetName()),
					); err != nil {
						__antithesis_instrumentation__.Notify(496824)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496825)
					}
				}
				__antithesis_instrumentation__.Notify(496819)
				return nil
			},
		)
	},
}

var informationSchemaEnabledRoles = virtualSchemaTable{
	comment: `roles for the current user
` + docs.URL("information-schema.html#enabled_roles") + `
https://www.postgresql.org/docs/9.5/infoschema-enabled-roles.html`,
	schema: vtable.InformationSchemaEnabledRoles,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496826)
		currentUser := p.SessionData().User()
		memberMap, err := p.MemberOfWithAdminOption(ctx, currentUser)
		if err != nil {
			__antithesis_instrumentation__.Notify(496830)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496831)
		}
		__antithesis_instrumentation__.Notify(496827)

		if err := addRow(
			tree.NewDString(currentUser.Normalized()),
		); err != nil {
			__antithesis_instrumentation__.Notify(496832)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496833)
		}
		__antithesis_instrumentation__.Notify(496828)

		for roleName := range memberMap {
			__antithesis_instrumentation__.Notify(496834)
			if err := addRow(
				tree.NewDString(roleName.Normalized()),
			); err != nil {
				__antithesis_instrumentation__.Notify(496835)
				return err
			} else {
				__antithesis_instrumentation__.Notify(496836)
			}
		}
		__antithesis_instrumentation__.Notify(496829)

		return nil
	},
}

func characterMaximumLength(colType *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(496837)
	return dIntFnOrNull(func() (int32, bool) {
		__antithesis_instrumentation__.Notify(496838)

		if colType.Oid() == oid.T_char {
			__antithesis_instrumentation__.Notify(496841)
			return 0, false
		} else {
			__antithesis_instrumentation__.Notify(496842)
		}
		__antithesis_instrumentation__.Notify(496839)
		switch colType.Family() {
		case types.StringFamily, types.CollatedStringFamily, types.BitFamily:
			__antithesis_instrumentation__.Notify(496843)
			if colType.Width() > 0 {
				__antithesis_instrumentation__.Notify(496845)
				return colType.Width(), true
			} else {
				__antithesis_instrumentation__.Notify(496846)
			}
		default:
			__antithesis_instrumentation__.Notify(496844)
		}
		__antithesis_instrumentation__.Notify(496840)
		return 0, false
	})
}

func characterOctetLength(colType *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(496847)
	return dIntFnOrNull(func() (int32, bool) {
		__antithesis_instrumentation__.Notify(496848)

		if colType.Oid() == oid.T_char {
			__antithesis_instrumentation__.Notify(496851)
			return 0, false
		} else {
			__antithesis_instrumentation__.Notify(496852)
		}
		__antithesis_instrumentation__.Notify(496849)
		switch colType.Family() {
		case types.StringFamily, types.CollatedStringFamily:
			__antithesis_instrumentation__.Notify(496853)
			if colType.Width() > 0 {
				__antithesis_instrumentation__.Notify(496855)
				return colType.Width() * utf8.UTFMax, true
			} else {
				__antithesis_instrumentation__.Notify(496856)
			}
		default:
			__antithesis_instrumentation__.Notify(496854)
		}
		__antithesis_instrumentation__.Notify(496850)
		return 0, false
	})
}

func numericPrecision(colType *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(496857)
	return dIntFnOrNull(func() (int32, bool) {
		__antithesis_instrumentation__.Notify(496858)
		switch colType.Family() {
		case types.IntFamily:
			__antithesis_instrumentation__.Notify(496860)
			return colType.Width(), true
		case types.FloatFamily:
			__antithesis_instrumentation__.Notify(496861)
			if colType.Width() == 32 {
				__antithesis_instrumentation__.Notify(496865)
				return 24, true
			} else {
				__antithesis_instrumentation__.Notify(496866)
			}
			__antithesis_instrumentation__.Notify(496862)
			return 53, true
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(496863)
			if colType.Precision() > 0 {
				__antithesis_instrumentation__.Notify(496867)
				return colType.Precision(), true
			} else {
				__antithesis_instrumentation__.Notify(496868)
			}
		default:
			__antithesis_instrumentation__.Notify(496864)
		}
		__antithesis_instrumentation__.Notify(496859)
		return 0, false
	})
}

func numericPrecisionRadix(colType *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(496869)
	return dIntFnOrNull(func() (int32, bool) {
		__antithesis_instrumentation__.Notify(496870)
		switch colType.Family() {
		case types.IntFamily:
			__antithesis_instrumentation__.Notify(496872)
			return 2, true
		case types.FloatFamily:
			__antithesis_instrumentation__.Notify(496873)
			return 2, true
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(496874)
			return 10, true
		default:
			__antithesis_instrumentation__.Notify(496875)
		}
		__antithesis_instrumentation__.Notify(496871)
		return 0, false
	})
}

func numericScale(colType *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(496876)
	return dIntFnOrNull(func() (int32, bool) {
		__antithesis_instrumentation__.Notify(496877)
		switch colType.Family() {
		case types.IntFamily:
			__antithesis_instrumentation__.Notify(496879)
			return 0, true
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(496880)
			if colType.Precision() > 0 {
				__antithesis_instrumentation__.Notify(496882)
				return colType.Width(), true
			} else {
				__antithesis_instrumentation__.Notify(496883)
			}
		default:
			__antithesis_instrumentation__.Notify(496881)
		}
		__antithesis_instrumentation__.Notify(496878)
		return 0, false
	})
}

func datetimePrecision(colType *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(496884)
	return dIntFnOrNull(func() (int32, bool) {
		__antithesis_instrumentation__.Notify(496885)
		switch colType.Family() {
		case types.TimeFamily, types.TimeTZFamily, types.TimestampFamily, types.TimestampTZFamily, types.IntervalFamily:
			__antithesis_instrumentation__.Notify(496887)
			return colType.Precision(), true
		default:
			__antithesis_instrumentation__.Notify(496888)
		}
		__antithesis_instrumentation__.Notify(496886)
		return 0, false
	})
}

var informationSchemaConstraintColumnUsageTable = virtualSchemaTable{
	comment: `columns usage by constraints
https://www.postgresql.org/docs/9.5/infoschema-constraint-column-usage.html`,
	schema: vtable.InformationSchemaConstraintColumnUsage,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496889)
		return forEachTableDescWithTableLookup(ctx, p, dbContext, hideVirtual, func(
			db catalog.DatabaseDescriptor,
			scName string,
			table catalog.TableDescriptor,
			tableLookup tableLookupFn,
		) error {
			__antithesis_instrumentation__.Notify(496890)
			conInfo, err := table.GetConstraintInfoWithLookup(tableLookup.getTableByID)
			if err != nil {
				__antithesis_instrumentation__.Notify(496893)
				return err
			} else {
				__antithesis_instrumentation__.Notify(496894)
			}
			__antithesis_instrumentation__.Notify(496891)
			scNameStr := tree.NewDString(scName)
			dbNameStr := tree.NewDString(db.GetName())

			for conName, con := range conInfo {
				__antithesis_instrumentation__.Notify(496895)
				conTable := table
				conCols := con.Columns
				conNameStr := tree.NewDString(conName)
				if con.Kind == descpb.ConstraintTypeFK {
					__antithesis_instrumentation__.Notify(496897)

					conTable = tabledesc.NewBuilder(con.ReferencedTable).BuildImmutableTable()
					conCols, err = conTable.NamesForColumnIDs(con.FK.ReferencedColumnIDs)
					if err != nil {
						__antithesis_instrumentation__.Notify(496898)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496899)
					}
				} else {
					__antithesis_instrumentation__.Notify(496900)
				}
				__antithesis_instrumentation__.Notify(496896)
				tableNameStr := tree.NewDString(conTable.GetName())
				for _, col := range conCols {
					__antithesis_instrumentation__.Notify(496901)
					if err := addRow(
						dbNameStr,
						scNameStr,
						tableNameStr,
						tree.NewDString(col),
						dbNameStr,
						scNameStr,
						conNameStr,
					); err != nil {
						__antithesis_instrumentation__.Notify(496902)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496903)
					}
				}
			}
			__antithesis_instrumentation__.Notify(496892)
			return nil
		})
	},
}

var informationSchemaKeyColumnUsageTable = virtualSchemaTable{
	comment: `column usage by indexes and key constraints
` + docs.URL("information-schema.html#key_column_usage") + `
https://www.postgresql.org/docs/9.5/infoschema-key-column-usage.html`,
	schema: vtable.InformationSchemaKeyColumnUsage,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496904)
		return forEachTableDescWithTableLookup(ctx, p, dbContext, hideVirtual, func(
			db catalog.DatabaseDescriptor,
			scName string,
			table catalog.TableDescriptor,
			tableLookup tableLookupFn,
		) error {
			__antithesis_instrumentation__.Notify(496905)
			conInfo, err := table.GetConstraintInfoWithLookup(tableLookup.getTableByID)
			if err != nil {
				__antithesis_instrumentation__.Notify(496908)
				return err
			} else {
				__antithesis_instrumentation__.Notify(496909)
			}
			__antithesis_instrumentation__.Notify(496906)
			dbNameStr := tree.NewDString(db.GetName())
			scNameStr := tree.NewDString(scName)
			tbNameStr := tree.NewDString(table.GetName())
			for conName, con := range conInfo {
				__antithesis_instrumentation__.Notify(496910)

				switch con.Kind {
				case descpb.ConstraintTypePK:
					__antithesis_instrumentation__.Notify(496912)
				case descpb.ConstraintTypeFK:
					__antithesis_instrumentation__.Notify(496913)
				case descpb.ConstraintTypeUnique:
					__antithesis_instrumentation__.Notify(496914)
				default:
					__antithesis_instrumentation__.Notify(496915)
					continue
				}
				__antithesis_instrumentation__.Notify(496911)

				cstNameStr := tree.NewDString(conName)

				for pos, col := range con.Columns {
					__antithesis_instrumentation__.Notify(496916)
					ordinalPos := tree.NewDInt(tree.DInt(pos + 1))
					uniquePos := tree.DNull
					if con.Kind == descpb.ConstraintTypeFK {
						__antithesis_instrumentation__.Notify(496918)
						uniquePos = ordinalPos
					} else {
						__antithesis_instrumentation__.Notify(496919)
					}
					__antithesis_instrumentation__.Notify(496917)
					if err := addRow(
						dbNameStr,
						scNameStr,
						cstNameStr,
						dbNameStr,
						scNameStr,
						tbNameStr,
						tree.NewDString(col),
						ordinalPos,
						uniquePos,
					); err != nil {
						__antithesis_instrumentation__.Notify(496920)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496921)
					}
				}
			}
			__antithesis_instrumentation__.Notify(496907)
			return nil
		})
	},
}

var informationSchemaParametersTable = virtualSchemaTable{
	comment: `built-in function parameters (empty - introspection not yet supported)
https://www.postgresql.org/docs/9.5/infoschema-parameters.html`,
	schema: vtable.InformationSchemaParameters,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496922)
		return nil
	},
	unimplemented: true,
}

var (
	matchOptionFull    = tree.NewDString("FULL")
	matchOptionPartial = tree.NewDString("PARTIAL")
	matchOptionNone    = tree.NewDString("NONE")

	matchOptionMap = map[descpb.ForeignKeyReference_Match]tree.Datum{
		descpb.ForeignKeyReference_SIMPLE:  matchOptionNone,
		descpb.ForeignKeyReference_FULL:    matchOptionFull,
		descpb.ForeignKeyReference_PARTIAL: matchOptionPartial,
	}

	refConstraintRuleNoAction   = tree.NewDString("NO ACTION")
	refConstraintRuleRestrict   = tree.NewDString("RESTRICT")
	refConstraintRuleSetNull    = tree.NewDString("SET NULL")
	refConstraintRuleSetDefault = tree.NewDString("SET DEFAULT")
	refConstraintRuleCascade    = tree.NewDString("CASCADE")
)

func dStringForFKAction(action catpb.ForeignKeyAction) tree.Datum {
	__antithesis_instrumentation__.Notify(496923)
	switch action {
	case catpb.ForeignKeyAction_NO_ACTION:
		__antithesis_instrumentation__.Notify(496925)
		return refConstraintRuleNoAction
	case catpb.ForeignKeyAction_RESTRICT:
		__antithesis_instrumentation__.Notify(496926)
		return refConstraintRuleRestrict
	case catpb.ForeignKeyAction_SET_NULL:
		__antithesis_instrumentation__.Notify(496927)
		return refConstraintRuleSetNull
	case catpb.ForeignKeyAction_SET_DEFAULT:
		__antithesis_instrumentation__.Notify(496928)
		return refConstraintRuleSetDefault
	case catpb.ForeignKeyAction_CASCADE:
		__antithesis_instrumentation__.Notify(496929)
		return refConstraintRuleCascade
	default:
		__antithesis_instrumentation__.Notify(496930)
	}
	__antithesis_instrumentation__.Notify(496924)
	panic(errors.Errorf("unexpected ForeignKeyReference_Action: %v", action))
}

var informationSchemaReferentialConstraintsTable = virtualSchemaTable{
	comment: `foreign key constraints
` + docs.URL("information-schema.html#referential_constraints") + `
https://www.postgresql.org/docs/9.5/infoschema-referential-constraints.html`,
	schema: vtable.InformationSchemaReferentialConstraints,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496931)
		return forEachTableDescWithTableLookup(ctx, p, dbContext, hideVirtual, func(
			db catalog.DatabaseDescriptor,
			scName string,
			table catalog.TableDescriptor,
			tableLookup tableLookupFn,
		) error {
			__antithesis_instrumentation__.Notify(496932)
			dbNameStr := tree.NewDString(db.GetName())
			scNameStr := tree.NewDString(scName)
			tbNameStr := tree.NewDString(table.GetName())
			return table.ForeachOutboundFK(func(fk *descpb.ForeignKeyConstraint) error {
				__antithesis_instrumentation__.Notify(496933)
				refTable, err := tableLookup.getTableByID(fk.ReferencedTableID)
				if err != nil {
					__antithesis_instrumentation__.Notify(496937)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496938)
				}
				__antithesis_instrumentation__.Notify(496934)
				var matchType = tree.DNull
				if r, ok := matchOptionMap[fk.Match]; ok {
					__antithesis_instrumentation__.Notify(496939)
					matchType = r
				} else {
					__antithesis_instrumentation__.Notify(496940)
				}
				__antithesis_instrumentation__.Notify(496935)
				refConstraint, err := tabledesc.FindFKReferencedUniqueConstraint(
					refTable, fk.ReferencedColumnIDs,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(496941)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496942)
				}
				__antithesis_instrumentation__.Notify(496936)
				return addRow(
					dbNameStr,
					scNameStr,
					tree.NewDString(fk.Name),
					dbNameStr,
					scNameStr,
					tree.NewDString(refConstraint.GetName()),
					matchType,
					dStringForFKAction(fk.OnUpdate),
					dStringForFKAction(fk.OnDelete),
					tbNameStr,
					tree.NewDString(refTable.GetName()),
				)
			})
		})
	},
}

var informationSchemaRoleTableGrants = virtualSchemaTable{
	comment: `privileges granted on table or views (incomplete; see also information_schema.table_privileges; may contain excess users or roles)
` + docs.URL("information-schema.html#role_table_grants") + `
https://www.postgresql.org/docs/9.5/infoschema-role-table-grants.html`,
	schema: vtable.InformationSchemaRoleTableGrants,

	populate: populateTablePrivileges,
}

var informationSchemaRoutineTable = virtualSchemaTable{
	comment: `built-in functions (empty - introspection not yet supported)
https://www.postgresql.org/docs/9.5/infoschema-routines.html`,
	schema: vtable.InformationSchemaRoutines,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496943)
		return nil
	},
	unimplemented: true,
}

var informationSchemaSchemataTable = virtualSchemaTable{
	comment: `database schemas (may contain schemata without permission)
` + docs.URL("information-schema.html#schemata") + `
https://www.postgresql.org/docs/9.5/infoschema-schemata.html`,
	schema: vtable.InformationSchemaSchemata,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496944)
		return forEachDatabaseDesc(ctx, p, dbContext, true,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(496945)
				return forEachSchema(ctx, p, db, func(sc catalog.SchemaDescriptor) error {
					__antithesis_instrumentation__.Notify(496946)
					return addRow(
						tree.NewDString(db.GetName()),
						tree.NewDString(sc.GetName()),
						tree.DNull,
						tree.DNull,
						yesOrNoDatum(sc.SchemaKind() == catalog.SchemaUserDefined),
					)
				})
			})
	},
}

var builtinTypePrivileges = []struct {
	grantee *tree.DString
	kind    *tree.DString
}{
	{tree.NewDString(security.RootUser), tree.NewDString(privilege.ALL.String())},
	{tree.NewDString(security.AdminRole), tree.NewDString(privilege.ALL.String())},
	{tree.NewDString(security.PublicRole), tree.NewDString(privilege.USAGE.String())},
}

var informationSchemaTypePrivilegesTable = virtualSchemaTable{
	comment: `type privileges (incomplete; may contain excess users or roles)
` + docs.URL("information-schema.html#type_privileges"),
	schema: vtable.InformationSchemaTypePrivileges,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496947)
		return forEachDatabaseDesc(ctx, p, dbContext, true,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(496948)
				dbNameStr := tree.NewDString(db.GetName())
				pgCatalogStr := tree.NewDString("pg_catalog")
				populateGrantOption := p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.ValidateGrantOption)
				var isGrantable tree.Datum
				if populateGrantOption {
					__antithesis_instrumentation__.Notify(496951)
					isGrantable = noString
				} else {
					__antithesis_instrumentation__.Notify(496952)
					isGrantable = tree.DNull
				}
				__antithesis_instrumentation__.Notify(496949)

				for _, typ := range types.OidToType {
					__antithesis_instrumentation__.Notify(496953)
					typeNameStr := tree.NewDString(typ.Name())
					for _, it := range builtinTypePrivileges {
						__antithesis_instrumentation__.Notify(496954)
						if err := addRow(
							it.grantee,
							dbNameStr,
							pgCatalogStr,
							typeNameStr,
							it.kind,
							isGrantable,
						); err != nil {
							__antithesis_instrumentation__.Notify(496955)
							return err
						} else {
							__antithesis_instrumentation__.Notify(496956)
						}
					}
				}
				__antithesis_instrumentation__.Notify(496950)

				return forEachTypeDesc(ctx, p, db, func(db catalog.DatabaseDescriptor, sc string, typeDesc catalog.TypeDescriptor) error {
					__antithesis_instrumentation__.Notify(496957)
					scNameStr := tree.NewDString(sc)
					typeNameStr := tree.NewDString(typeDesc.GetName())

					privs := typeDesc.GetPrivileges().Show(privilege.Type)
					for _, u := range privs {
						__antithesis_instrumentation__.Notify(496959)
						userNameStr := tree.NewDString(u.User.Normalized())
						for _, priv := range u.Privileges {
							__antithesis_instrumentation__.Notify(496960)
							var isGrantable tree.Datum
							if populateGrantOption {
								__antithesis_instrumentation__.Notify(496962)
								isGrantable = yesOrNoDatum(priv.GrantOption)
							} else {
								__antithesis_instrumentation__.Notify(496963)
								isGrantable = tree.DNull
							}
							__antithesis_instrumentation__.Notify(496961)
							if err := addRow(
								userNameStr,
								dbNameStr,
								scNameStr,
								typeNameStr,
								tree.NewDString(priv.Kind.String()),
								isGrantable,
							); err != nil {
								__antithesis_instrumentation__.Notify(496964)
								return err
							} else {
								__antithesis_instrumentation__.Notify(496965)
							}
						}
					}
					__antithesis_instrumentation__.Notify(496958)
					return nil
				})
			})
	},
}

var informationSchemaSchemataTablePrivileges = virtualSchemaTable{
	comment: `schema privileges (incomplete; may contain excess users or roles)
` + docs.URL("information-schema.html#schema_privileges"),
	schema: vtable.InformationSchemaSchemaPrivileges,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496966)
		return forEachDatabaseDesc(ctx, p, dbContext, true,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(496967)
				return forEachSchema(ctx, p, db, func(sc catalog.SchemaDescriptor) error {
					__antithesis_instrumentation__.Notify(496968)
					privs := sc.GetPrivileges().Show(privilege.Schema)
					dbNameStr := tree.NewDString(db.GetName())
					scNameStr := tree.NewDString(sc.GetName())

					populateGrantOption := p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.ValidateGrantOption)
					for _, u := range privs {
						__antithesis_instrumentation__.Notify(496970)
						userNameStr := tree.NewDString(u.User.Normalized())
						for _, priv := range u.Privileges {
							__antithesis_instrumentation__.Notify(496971)
							var isGrantable tree.Datum
							if populateGrantOption {
								__antithesis_instrumentation__.Notify(496973)
								isGrantable = yesOrNoDatum(priv.GrantOption)
							} else {
								__antithesis_instrumentation__.Notify(496974)
								isGrantable = tree.DNull
							}
							__antithesis_instrumentation__.Notify(496972)
							if err := addRow(
								userNameStr,
								dbNameStr,
								scNameStr,
								tree.NewDString(priv.Kind.String()),
								isGrantable,
							); err != nil {
								__antithesis_instrumentation__.Notify(496975)
								return err
							} else {
								__antithesis_instrumentation__.Notify(496976)
							}
						}
					}
					__antithesis_instrumentation__.Notify(496969)
					return nil
				})
			})
	},
}

var (
	indexDirectionNA   = tree.NewDString("N/A")
	indexDirectionAsc  = tree.NewDString(descpb.IndexDescriptor_ASC.String())
	indexDirectionDesc = tree.NewDString(descpb.IndexDescriptor_DESC.String())
)

func dStringForIndexDirection(dir descpb.IndexDescriptor_Direction) tree.Datum {
	__antithesis_instrumentation__.Notify(496977)
	switch dir {
	case descpb.IndexDescriptor_ASC:
		__antithesis_instrumentation__.Notify(496979)
		return indexDirectionAsc
	case descpb.IndexDescriptor_DESC:
		__antithesis_instrumentation__.Notify(496980)
		return indexDirectionDesc
	default:
		__antithesis_instrumentation__.Notify(496981)
	}
	__antithesis_instrumentation__.Notify(496978)
	panic("unreachable")
}

var informationSchemaSequences = virtualSchemaTable{
	comment: `sequences
` + docs.URL("information-schema.html#sequences") + `
https://www.postgresql.org/docs/9.5/infoschema-sequences.html`,
	schema: vtable.InformationSchemaSequences,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496982)
		return forEachTableDesc(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(496983)
				if !table.IsSequence() {
					__antithesis_instrumentation__.Notify(496985)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(496986)
				}
				__antithesis_instrumentation__.Notify(496984)
				return addRow(
					tree.NewDString(db.GetName()),
					tree.NewDString(scName),
					tree.NewDString(table.GetName()),
					tree.NewDString("bigint"),
					tree.NewDInt(64),
					tree.NewDInt(2),
					tree.NewDInt(0),
					tree.NewDString(strconv.FormatInt(table.GetSequenceOpts().Start, 10)),
					tree.NewDString(strconv.FormatInt(table.GetSequenceOpts().MinValue, 10)),
					tree.NewDString(strconv.FormatInt(table.GetSequenceOpts().MaxValue, 10)),
					tree.NewDString(strconv.FormatInt(table.GetSequenceOpts().Increment, 10)),
					noString,
				)
			})
	},
}

var informationSchemaStatisticsTable = virtualSchemaTable{
	comment: `index metadata and statistics (incomplete)
` + docs.URL("information-schema.html#statistics"),
	schema: vtable.InformationSchemaStatistics,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(496987)
		return forEachTableDesc(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(496988)
				dbNameStr := tree.NewDString(db.GetName())
				scNameStr := tree.NewDString(scName)
				tbNameStr := tree.NewDString(table.GetName())

				appendRow := func(index catalog.Index, colName string, sequence int,
					direction tree.Datum, isStored, isImplicit bool,
				) error {
					__antithesis_instrumentation__.Notify(496990)
					return addRow(
						dbNameStr,
						scNameStr,
						tbNameStr,
						yesOrNoDatum(!index.IsUnique()),
						scNameStr,
						tree.NewDString(index.GetName()),
						tree.NewDInt(tree.DInt(sequence)),
						tree.NewDString(colName),
						tree.DNull,
						tree.DNull,
						direction,
						yesOrNoDatum(isStored),
						yesOrNoDatum(isImplicit),
					)
				}
				__antithesis_instrumentation__.Notify(496989)

				return catalog.ForEachIndex(table, catalog.IndexOpts{}, func(index catalog.Index) error {
					__antithesis_instrumentation__.Notify(496991)

					var implicitCols map[string]struct{}
					var hasImplicitCols bool
					if index.HasOldStoredColumns() {
						__antithesis_instrumentation__.Notify(496997)

						hasImplicitCols = index.NumKeySuffixColumns() > index.NumSecondaryStoredColumns()
					} else {
						__antithesis_instrumentation__.Notify(496998)

						hasImplicitCols = index.NumKeySuffixColumns() > 0
					}
					__antithesis_instrumentation__.Notify(496992)
					if hasImplicitCols {
						__antithesis_instrumentation__.Notify(496999)
						implicitCols = make(map[string]struct{})
						for i := 0; i < table.GetPrimaryIndex().NumKeyColumns(); i++ {
							__antithesis_instrumentation__.Notify(497000)
							col := table.GetPrimaryIndex().GetKeyColumnName(i)
							implicitCols[col] = struct{}{}
						}
					} else {
						__antithesis_instrumentation__.Notify(497001)
					}
					__antithesis_instrumentation__.Notify(496993)

					sequence := 1
					for i := 0; i < index.NumKeyColumns(); i++ {
						__antithesis_instrumentation__.Notify(497002)
						col := index.GetKeyColumnName(i)

						dir := dStringForIndexDirection(index.GetKeyColumnDirection(i))
						if err := appendRow(
							index,
							col,
							sequence,
							dir,
							false,
							i < index.ExplicitColumnStartIdx(),
						); err != nil {
							__antithesis_instrumentation__.Notify(497004)
							return err
						} else {
							__antithesis_instrumentation__.Notify(497005)
						}
						__antithesis_instrumentation__.Notify(497003)
						sequence++
						delete(implicitCols, col)
					}
					__antithesis_instrumentation__.Notify(496994)
					for i := 0; i < index.NumPrimaryStoredColumns()+index.NumSecondaryStoredColumns(); i++ {
						__antithesis_instrumentation__.Notify(497006)
						col := index.GetStoredColumnName(i)

						if err := appendRow(index, col, sequence,
							indexDirectionNA, true, false); err != nil {
							__antithesis_instrumentation__.Notify(497008)
							return err
						} else {
							__antithesis_instrumentation__.Notify(497009)
						}
						__antithesis_instrumentation__.Notify(497007)
						sequence++
						delete(implicitCols, col)
					}
					__antithesis_instrumentation__.Notify(496995)
					if len(implicitCols) > 0 {
						__antithesis_instrumentation__.Notify(497010)

						for i := 0; i < table.GetPrimaryIndex().NumKeyColumns(); i++ {
							__antithesis_instrumentation__.Notify(497011)
							col := table.GetPrimaryIndex().GetKeyColumnName(i)
							if _, isImplicit := implicitCols[col]; isImplicit {
								__antithesis_instrumentation__.Notify(497012)

								if err := appendRow(index, col, sequence,
									indexDirectionAsc, index.IsUnique(), true); err != nil {
									__antithesis_instrumentation__.Notify(497014)
									return err
								} else {
									__antithesis_instrumentation__.Notify(497015)
								}
								__antithesis_instrumentation__.Notify(497013)
								sequence++
							} else {
								__antithesis_instrumentation__.Notify(497016)
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(497017)
					}
					__antithesis_instrumentation__.Notify(496996)
					return nil
				})
			})
	},
}

var informationSchemaTableConstraintTable = virtualSchemaTable{
	comment: `table constraints
` + docs.URL("information-schema.html#table_constraints") + `
https://www.postgresql.org/docs/9.5/infoschema-table-constraints.html`,
	schema: vtable.InformationSchemaTableConstraint,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497018)
		h := makeOidHasher()
		return forEachTableDescWithTableLookup(ctx, p, dbContext, hideVirtual,
			func(
				db catalog.DatabaseDescriptor,
				scName string,
				table catalog.TableDescriptor,
				tableLookup tableLookupFn,
			) error {
				__antithesis_instrumentation__.Notify(497019)
				conInfo, err := table.GetConstraintInfoWithLookup(tableLookup.getTableByID)
				if err != nil {
					__antithesis_instrumentation__.Notify(497023)
					return err
				} else {
					__antithesis_instrumentation__.Notify(497024)
				}
				__antithesis_instrumentation__.Notify(497020)

				dbNameStr := tree.NewDString(db.GetName())
				scNameStr := tree.NewDString(scName)
				tbNameStr := tree.NewDString(table.GetName())

				for conName, c := range conInfo {
					__antithesis_instrumentation__.Notify(497025)
					if err := addRow(
						dbNameStr,
						scNameStr,
						tree.NewDString(conName),
						dbNameStr,
						scNameStr,
						tbNameStr,
						tree.NewDString(string(c.Kind)),
						yesOrNoDatum(false),
						yesOrNoDatum(false),
					); err != nil {
						__antithesis_instrumentation__.Notify(497026)
						return err
					} else {
						__antithesis_instrumentation__.Notify(497027)
					}
				}
				__antithesis_instrumentation__.Notify(497021)

				for _, col := range table.PublicColumns() {
					__antithesis_instrumentation__.Notify(497028)
					if col.IsNullable() {
						__antithesis_instrumentation__.Notify(497030)
						continue
					} else {
						__antithesis_instrumentation__.Notify(497031)
					}
					__antithesis_instrumentation__.Notify(497029)

					conNameStr := tree.NewDString(fmt.Sprintf(
						"%s_%s_%d_not_null", h.NamespaceOid(db.GetID(), scName), tableOid(table.GetID()), col.Ordinal()+1,
					))
					if err := addRow(
						dbNameStr,
						scNameStr,
						conNameStr,
						dbNameStr,
						scNameStr,
						tbNameStr,
						tree.NewDString("CHECK"),
						yesOrNoDatum(false),
						yesOrNoDatum(false),
					); err != nil {
						__antithesis_instrumentation__.Notify(497032)
						return err
					} else {
						__antithesis_instrumentation__.Notify(497033)
					}
				}
				__antithesis_instrumentation__.Notify(497022)
				return nil
			})
	},
}

var informationSchemaUserPrivileges = virtualSchemaTable{
	comment: `grantable privileges (incomplete)`,
	schema:  vtable.InformationSchemaUserPrivileges,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497034)
		return forEachDatabaseDesc(ctx, p, dbContext, true,
			func(dbDesc catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(497035)
				dbNameStr := tree.NewDString(dbDesc.GetName())
				for _, u := range []string{security.RootUser, security.AdminRole} {
					__antithesis_instrumentation__.Notify(497037)
					grantee := tree.NewDString(u)
					for _, p := range privilege.GetValidPrivilegesForObject(privilege.Table).SortedNames() {
						__antithesis_instrumentation__.Notify(497038)
						if err := addRow(
							grantee,
							dbNameStr,
							tree.NewDString(p),
							tree.DNull,
						); err != nil {
							__antithesis_instrumentation__.Notify(497039)
							return err
						} else {
							__antithesis_instrumentation__.Notify(497040)
						}
					}
				}
				__antithesis_instrumentation__.Notify(497036)
				return nil
			})
	},
}

var informationSchemaTablePrivileges = virtualSchemaTable{
	comment: `privileges granted on table or views (incomplete; may contain excess users or roles)
` + docs.URL("information-schema.html#table_privileges") + `
https://www.postgresql.org/docs/9.5/infoschema-table-privileges.html`,
	schema:   vtable.InformationSchemaTablePrivileges,
	populate: populateTablePrivileges,
}

func populateTablePrivileges(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	addRow func(...tree.Datum) error,
) error {
	__antithesis_instrumentation__.Notify(497041)
	return forEachTableDesc(ctx, p, dbContext, virtualMany,
		func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
			__antithesis_instrumentation__.Notify(497042)
			dbNameStr := tree.NewDString(db.GetName())
			scNameStr := tree.NewDString(scName)
			tbNameStr := tree.NewDString(table.GetName())

			populateGrantOption := p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.ValidateGrantOption)
			for _, u := range table.GetPrivileges().Show(privilege.Table) {
				__antithesis_instrumentation__.Notify(497044)
				granteeNameStr := tree.NewDString(u.User.Normalized())
				for _, priv := range u.Privileges {
					__antithesis_instrumentation__.Notify(497045)
					var isGrantable tree.Datum
					if populateGrantOption {
						__antithesis_instrumentation__.Notify(497047)
						isGrantable = yesOrNoDatum(priv.GrantOption)
					} else {
						__antithesis_instrumentation__.Notify(497048)
						isGrantable = tree.DNull
					}
					__antithesis_instrumentation__.Notify(497046)
					if err := addRow(
						tree.DNull,
						granteeNameStr,
						dbNameStr,
						scNameStr,
						tbNameStr,
						tree.NewDString(priv.Kind.String()),
						isGrantable,
						yesOrNoDatum(priv.Kind == privilege.SELECT),
					); err != nil {
						__antithesis_instrumentation__.Notify(497049)
						return err
					} else {
						__antithesis_instrumentation__.Notify(497050)
					}
				}
			}
			__antithesis_instrumentation__.Notify(497043)
			return nil
		})
}

var (
	tableTypeSystemView = tree.NewDString("SYSTEM VIEW")
	tableTypeBaseTable  = tree.NewDString("BASE TABLE")
	tableTypeView       = tree.NewDString("VIEW")
	tableTypeTemporary  = tree.NewDString("LOCAL TEMPORARY")
)

var informationSchemaTablesTable = virtualSchemaTable{
	comment: `tables and views
` + docs.URL("information-schema.html#tables") + `
https://www.postgresql.org/docs/9.5/infoschema-tables.html`,
	schema: vtable.InformationSchemaTables,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497051)
		return forEachTableDesc(ctx, p, dbContext, virtualMany, addTablesTableRow(addRow))
	},
}

func addTablesTableRow(
	addRow func(...tree.Datum) error,
) func(
	db catalog.DatabaseDescriptor,
	scName string,
	table catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(497052)
	return func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
		__antithesis_instrumentation__.Notify(497053)
		if table.IsSequence() {
			__antithesis_instrumentation__.Notify(497056)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(497057)
		}
		__antithesis_instrumentation__.Notify(497054)
		tableType := tableTypeBaseTable
		insertable := yesString
		if table.IsVirtualTable() {
			__antithesis_instrumentation__.Notify(497058)
			tableType = tableTypeSystemView
			insertable = noString
		} else {
			__antithesis_instrumentation__.Notify(497059)
			if table.IsView() {
				__antithesis_instrumentation__.Notify(497060)
				tableType = tableTypeView
				insertable = noString
			} else {
				__antithesis_instrumentation__.Notify(497061)
				if table.IsTemporary() {
					__antithesis_instrumentation__.Notify(497062)
					tableType = tableTypeTemporary
				} else {
					__antithesis_instrumentation__.Notify(497063)
				}
			}
		}
		__antithesis_instrumentation__.Notify(497055)
		dbNameStr := tree.NewDString(db.GetName())
		scNameStr := tree.NewDString(scName)
		tbNameStr := tree.NewDString(table.GetName())
		return addRow(
			dbNameStr,
			scNameStr,
			tbNameStr,
			tableType,
			insertable,
			tree.NewDInt(tree.DInt(table.GetVersion())),
		)
	}
}

var informationSchemaViewsTable = virtualSchemaTable{
	comment: `views (incomplete)
` + docs.URL("information-schema.html#views") + `
https://www.postgresql.org/docs/9.5/infoschema-views.html`,
	schema: vtable.InformationSchemaViews,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497064)
		return forEachTableDesc(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(497065)
				if !table.IsView() {
					__antithesis_instrumentation__.Notify(497067)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(497068)
				}
				__antithesis_instrumentation__.Notify(497066)

				return addRow(
					tree.NewDString(db.GetName()),
					tree.NewDString(scName),
					tree.NewDString(table.GetName()),
					tree.NewDString(table.GetViewQuery()),
					tree.DNull,
					noString,
					noString,
					noString,
					noString,
					noString,
				)
			})
	},
}

var informationSchemaCollations = virtualSchemaTable{
	comment: `shows the collations available in the current database
https://www.postgresql.org/docs/current/infoschema-collations.html`,
	schema: vtable.InformationSchemaCollations,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497069)
		dbNameStr := tree.NewDString(p.CurrentDatabase())
		add := func(collName string) error {
			__antithesis_instrumentation__.Notify(497073)
			return addRow(
				dbNameStr,
				pgCatalogNameDString,
				tree.NewDString(collName),

				tree.NewDString("NO PAD"),
			)
		}
		__antithesis_instrumentation__.Notify(497070)
		if err := add(tree.DefaultCollationTag); err != nil {
			__antithesis_instrumentation__.Notify(497074)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497075)
		}
		__antithesis_instrumentation__.Notify(497071)
		for _, tag := range collate.Supported() {
			__antithesis_instrumentation__.Notify(497076)
			collName := tag.String()
			if err := add(collName); err != nil {
				__antithesis_instrumentation__.Notify(497077)
				return err
			} else {
				__antithesis_instrumentation__.Notify(497078)
			}
		}
		__antithesis_instrumentation__.Notify(497072)
		return nil
	},
}

var informationSchemaCollationCharacterSetApplicability = virtualSchemaTable{
	comment: `identifies which character set the available collations are 
applicable to. As UTF-8 is the only available encoding this table does not
provide much useful information.
https://www.postgresql.org/docs/current/infoschema-collation-character-set-applicab.html`,
	schema: vtable.InformationSchemaCollationCharacterSetApplicability,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497079)
		dbNameStr := tree.NewDString(p.CurrentDatabase())
		add := func(collName string) error {
			__antithesis_instrumentation__.Notify(497083)
			return addRow(
				dbNameStr,
				pgCatalogNameDString,
				tree.NewDString(collName),
				tree.DNull,
				tree.DNull,
				tree.NewDString("UTF8"),
			)
		}
		__antithesis_instrumentation__.Notify(497080)
		if err := add(tree.DefaultCollationTag); err != nil {
			__antithesis_instrumentation__.Notify(497084)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497085)
		}
		__antithesis_instrumentation__.Notify(497081)
		for _, tag := range collate.Supported() {
			__antithesis_instrumentation__.Notify(497086)
			collName := tag.String()
			if err := add(collName); err != nil {
				__antithesis_instrumentation__.Notify(497087)
				return err
			} else {
				__antithesis_instrumentation__.Notify(497088)
			}
		}
		__antithesis_instrumentation__.Notify(497082)
		return nil
	},
}

var informationSchemaSessionVariables = virtualSchemaTable{
	comment: `exposes the session variables.`,
	schema:  vtable.InformationSchemaSessionVariables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497089)
		for _, vName := range varNames {
			__antithesis_instrumentation__.Notify(497091)
			gen := varGen[vName]
			value, err := gen.Get(&p.extendedEvalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(497093)
				return err
			} else {
				__antithesis_instrumentation__.Notify(497094)
			}
			__antithesis_instrumentation__.Notify(497092)
			if err := addRow(
				tree.NewDString(vName),
				tree.NewDString(value),
			); err != nil {
				__antithesis_instrumentation__.Notify(497095)
				return err
			} else {
				__antithesis_instrumentation__.Notify(497096)
			}
		}
		__antithesis_instrumentation__.Notify(497090)
		return nil
	},
}

var informationSchemaRoutinePrivilegesTable = virtualSchemaTable{
	comment: "routine_privileges was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaRoutinePrivileges,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497097)
		return nil
	},
	unimplemented: true,
}

var informationSchemaRoleRoutineGrantsTable = virtualSchemaTable{
	comment: "role_routine_grants was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaRoleRoutineGrants,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497098)
		return nil
	},
	unimplemented: true,
}

var informationSchemaElementTypesTable = virtualSchemaTable{
	comment: "element_types was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaElementTypes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497099)
		return nil
	},
	unimplemented: true,
}

var informationSchemaRoleUdtGrantsTable = virtualSchemaTable{
	comment: "role_udt_grants was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaRoleUdtGrants,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497100)
		return nil
	},
	unimplemented: true,
}

var informationSchemaColumnOptionsTable = virtualSchemaTable{
	comment: "column_options was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaColumnOptions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497101)
		return nil
	},
	unimplemented: true,
}

var informationSchemaForeignDataWrapperOptionsTable = virtualSchemaTable{
	comment: "foreign_data_wrapper_options was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaForeignDataWrapperOptions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497102)
		return nil
	},
	unimplemented: true,
}

var informationSchemaTransformsTable = virtualSchemaTable{
	comment: "transforms was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaTransforms,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497103)
		return nil
	},
	unimplemented: true,
}

var informationSchemaViewColumnUsageTable = virtualSchemaTable{
	comment: "view_column_usage was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaViewColumnUsage,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497104)
		return nil
	},
	unimplemented: true,
}

var informationSchemaInformationSchemaCatalogNameTable = virtualSchemaTable{
	comment: "information_schema_catalog_name was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaInformationSchemaCatalogName,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497105)
		return nil
	},
	unimplemented: true,
}

var informationSchemaForeignTablesTable = virtualSchemaTable{
	comment: "foreign_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaForeignTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497106)
		return nil
	},
	unimplemented: true,
}

var informationSchemaViewRoutineUsageTable = virtualSchemaTable{
	comment: "view_routine_usage was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaViewRoutineUsage,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497107)
		return nil
	},
	unimplemented: true,
}

var informationSchemaRoleColumnGrantsTable = virtualSchemaTable{
	comment: "role_column_grants was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaRoleColumnGrants,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497108)
		return nil
	},
	unimplemented: true,
}

var informationSchemaAttributesTable = virtualSchemaTable{
	comment: "attributes was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaAttributes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497109)
		return nil
	},
	unimplemented: true,
}

var informationSchemaDomainConstraintsTable = virtualSchemaTable{
	comment: "domain_constraints was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaDomainConstraints,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497110)
		return nil
	},
	unimplemented: true,
}

var informationSchemaUserMappingsTable = virtualSchemaTable{
	comment: "user_mappings was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaUserMappings,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497111)
		return nil
	},
	unimplemented: true,
}

var informationSchemaCheckConstraintRoutineUsageTable = virtualSchemaTable{
	comment: "check_constraint_routine_usage was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaCheckConstraintRoutineUsage,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497112)
		return nil
	},
	unimplemented: true,
}

var informationSchemaColumnDomainUsageTable = virtualSchemaTable{
	comment: "column_domain_usage was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaColumnDomainUsage,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497113)
		return nil
	},
	unimplemented: true,
}

var informationSchemaForeignDataWrappersTable = virtualSchemaTable{
	comment: "foreign_data_wrappers was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaForeignDataWrappers,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497114)
		return nil
	},
	unimplemented: true,
}

var informationSchemaColumnColumnUsageTable = virtualSchemaTable{
	comment: "column_column_usage was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaColumnColumnUsage,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497115)
		return nil
	},
	unimplemented: true,
}

var informationSchemaSQLSizingTable = virtualSchemaTable{
	comment: "sql_sizing was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaSQLSizing,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497116)
		return nil
	},
	unimplemented: true,
}

var informationSchemaUsagePrivilegesTable = virtualSchemaTable{
	comment: "usage_privileges was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaUsagePrivileges,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497117)
		return nil
	},
	unimplemented: true,
}

var informationSchemaDomainsTable = virtualSchemaTable{
	comment: "domains was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaDomains,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497118)
		return nil
	},
	unimplemented: true,
}

var informationSchemaSQLImplementationInfoTable = virtualSchemaTable{
	comment: "sql_implementation_info was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaSQLImplementationInfo,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497119)
		return nil
	},
	unimplemented: true,
}

var informationSchemaUdtPrivilegesTable = virtualSchemaTable{
	comment: "udt_privileges was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaUdtPrivileges,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497120)
		return nil
	},
	unimplemented: true,
}

var informationSchemaPartitionsTable = virtualSchemaTable{
	comment: "partitions was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaPartitions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497121)
		return nil
	},
	unimplemented: true,
}

var informationSchemaTablespacesExtensionsTable = virtualSchemaTable{
	comment: "tablespaces_extensions was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaTablespacesExtensions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497122)
		return nil
	},
	unimplemented: true,
}

var informationSchemaResourceGroupsTable = virtualSchemaTable{
	comment: "resource_groups was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaResourceGroups,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497123)
		return nil
	},
	unimplemented: true,
}

var informationSchemaForeignServerOptionsTable = virtualSchemaTable{
	comment: "foreign_server_options was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaForeignServerOptions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497124)
		return nil
	},
	unimplemented: true,
}

var informationSchemaStUnitsOfMeasureTable = virtualSchemaTable{
	comment: "st_units_of_measure was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaStUnitsOfMeasure,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497125)
		return nil
	},
	unimplemented: true,
}

var informationSchemaSchemataExtensionsTable = virtualSchemaTable{
	comment: "schemata_extensions was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaSchemataExtensions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497126)
		return nil
	},
	unimplemented: true,
}

var informationSchemaColumnStatisticsTable = virtualSchemaTable{
	comment: "column_statistics was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaColumnStatistics,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497127)
		return nil
	},
	unimplemented: true,
}

var informationSchemaConstraintTableUsageTable = virtualSchemaTable{
	comment: "constraint_table_usage was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaConstraintTableUsage,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497128)
		return nil
	},
	unimplemented: true,
}

var informationSchemaDataTypePrivilegesTable = virtualSchemaTable{
	comment: "data_type_privileges was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaDataTypePrivileges,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497129)
		return nil
	},
	unimplemented: true,
}

var informationSchemaRoleUsageGrantsTable = virtualSchemaTable{
	comment: "role_usage_grants was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaRoleUsageGrants,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497130)
		return nil
	},
	unimplemented: true,
}

var informationSchemaFilesTable = virtualSchemaTable{
	comment: "files was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaFiles,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497131)
		return nil
	},
	unimplemented: true,
}

var informationSchemaEnginesTable = virtualSchemaTable{
	comment: "engines was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaEngines,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497132)
		return nil
	},
	unimplemented: true,
}

var informationSchemaForeignTableOptionsTable = virtualSchemaTable{
	comment: "foreign_table_options was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaForeignTableOptions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497133)
		return nil
	},
	unimplemented: true,
}

var informationSchemaEventsTable = virtualSchemaTable{
	comment: "events was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaEvents,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497134)
		return nil
	},
	unimplemented: true,
}

var informationSchemaDomainUdtUsageTable = virtualSchemaTable{
	comment: "domain_udt_usage was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaDomainUdtUsage,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497135)
		return nil
	},
	unimplemented: true,
}

var informationSchemaUserAttributesTable = virtualSchemaTable{
	comment: "user_attributes was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaUserAttributes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497136)
		return nil
	},
	unimplemented: true,
}

var informationSchemaKeywordsTable = virtualSchemaTable{
	comment: "keywords was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaKeywords,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497137)
		return nil
	},
	unimplemented: true,
}

var informationSchemaUserMappingOptionsTable = virtualSchemaTable{
	comment: "user_mapping_options was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaUserMappingOptions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497138)
		return nil
	},
	unimplemented: true,
}

var informationSchemaOptimizerTraceTable = virtualSchemaTable{
	comment: "optimizer_trace was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaOptimizerTrace,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497139)
		return nil
	},
	unimplemented: true,
}

var informationSchemaTableConstraintsExtensionsTable = virtualSchemaTable{
	comment: "table_constraints_extensions was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaTableConstraintsExtensions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497140)
		return nil
	},
	unimplemented: true,
}

var informationSchemaColumnsExtensionsTable = virtualSchemaTable{
	comment: "columns_extensions was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaColumnsExtensions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497141)
		return nil
	},
	unimplemented: true,
}

var informationSchemaUserDefinedTypesTable = virtualSchemaTable{
	comment: "user_defined_types was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaUserDefinedTypes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497142)
		return nil
	},
	unimplemented: true,
}

var informationSchemaSQLFeaturesTable = virtualSchemaTable{
	comment: "sql_features was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaSQLFeatures,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497143)
		return nil
	},
	unimplemented: true,
}

var informationSchemaStGeometryColumnsTable = virtualSchemaTable{
	comment: "st_geometry_columns was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaStGeometryColumns,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497144)
		return nil
	},
	unimplemented: true,
}

var informationSchemaSQLPartsTable = virtualSchemaTable{
	comment: "sql_parts was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaSQLParts,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497145)
		return nil
	},
	unimplemented: true,
}

var informationSchemaPluginsTable = virtualSchemaTable{
	comment: "plugins was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaPlugins,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497146)
		return nil
	},
	unimplemented: true,
}

var informationSchemaStSpatialReferenceSystemsTable = virtualSchemaTable{
	comment: "st_spatial_reference_systems was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaStSpatialReferenceSystems,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497147)
		return nil
	},
	unimplemented: true,
}

var informationSchemaProcesslistTable = virtualSchemaTable{
	comment: "processlist was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaProcesslist,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497148)
		return nil
	},
	unimplemented: true,
}

var informationSchemaForeignServersTable = virtualSchemaTable{
	comment: "foreign_servers was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaForeignServers,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497149)
		return nil
	},
	unimplemented: true,
}

var informationSchemaTriggeredUpdateColumnsTable = virtualSchemaTable{
	comment: "triggered_update_columns was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaTriggeredUpdateColumns,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497150)
		return nil
	},
	unimplemented: true,
}

var informationSchemaTriggersTable = virtualSchemaTable{
	comment: "triggers was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaTriggers,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497151)
		return nil
	},
	unimplemented: true,
}

var informationSchemaTablesExtensionsTable = virtualSchemaTable{
	comment: "tables_extensions was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaTablesExtensions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497152)
		return nil
	},
	unimplemented: true,
}

var informationSchemaProfilingTable = virtualSchemaTable{
	comment: "profiling was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaProfiling,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497153)
		return nil
	},
	unimplemented: true,
}

var informationSchemaTablespacesTable = virtualSchemaTable{
	comment: "tablespaces was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaTablespaces,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497154)
		return nil
	},
	unimplemented: true,
}

var informationSchemaViewTableUsageTable = virtualSchemaTable{
	comment: "view_table_usage was created for compatibility and is currently unimplemented",
	schema:  vtable.InformationSchemaViewTableUsage,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(497155)
		return nil
	},
	unimplemented: true,
}

func forEachSchema(
	ctx context.Context,
	p *planner,
	db catalog.DatabaseDescriptor,
	fn func(sc catalog.SchemaDescriptor) error,
) error {
	__antithesis_instrumentation__.Notify(497156)
	schemaNames, err := getSchemaNames(ctx, p, db)
	if err != nil {
		__antithesis_instrumentation__.Notify(497164)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497165)
	}
	__antithesis_instrumentation__.Notify(497157)

	vtableEntries := p.getVirtualTabler().getSchemas()
	schemas := make([]catalog.SchemaDescriptor, 0, len(schemaNames)+len(vtableEntries))
	var userDefinedSchemaIDs []descpb.ID
	for id, name := range schemaNames {
		__antithesis_instrumentation__.Notify(497166)
		switch {
		case strings.HasPrefix(name, catconstants.PgTempSchemaName):
			__antithesis_instrumentation__.Notify(497167)
			schemas = append(schemas, schemadesc.NewTemporarySchema(name, id, db.GetID()))
		case name == tree.PublicSchema:
			__antithesis_instrumentation__.Notify(497168)

			if id == keys.PublicSchemaID {
				__antithesis_instrumentation__.Notify(497170)
				schemas = append(schemas, schemadesc.GetPublicSchema())
			} else {
				__antithesis_instrumentation__.Notify(497171)

				userDefinedSchemaIDs = append(userDefinedSchemaIDs, id)
			}
		default:
			__antithesis_instrumentation__.Notify(497169)

			userDefinedSchemaIDs = append(userDefinedSchemaIDs, id)
		}
	}
	__antithesis_instrumentation__.Notify(497158)

	userDefinedSchemas, err := p.Descriptors().Direct().GetSchemaDescriptorsFromIDs(ctx, p.txn, userDefinedSchemaIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(497172)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497173)
	}
	__antithesis_instrumentation__.Notify(497159)
	for i := range userDefinedSchemas {
		__antithesis_instrumentation__.Notify(497174)
		desc := userDefinedSchemas[i]
		canSeeDescriptor, err := userCanSeeDescriptor(ctx, p, desc, db, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(497177)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497178)
		}
		__antithesis_instrumentation__.Notify(497175)
		if !canSeeDescriptor {
			__antithesis_instrumentation__.Notify(497179)
			continue
		} else {
			__antithesis_instrumentation__.Notify(497180)
		}
		__antithesis_instrumentation__.Notify(497176)
		schemas = append(schemas, desc)
	}
	__antithesis_instrumentation__.Notify(497160)

	for _, schema := range vtableEntries {
		__antithesis_instrumentation__.Notify(497181)
		schemas = append(schemas, schema.Desc())
	}
	__antithesis_instrumentation__.Notify(497161)

	sort.Slice(schemas, func(i int, j int) bool {
		__antithesis_instrumentation__.Notify(497182)
		return schemas[i].GetName() < schemas[j].GetName()
	})
	__antithesis_instrumentation__.Notify(497162)

	for _, sc := range schemas {
		__antithesis_instrumentation__.Notify(497183)
		if err := fn(sc); err != nil {
			__antithesis_instrumentation__.Notify(497184)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497185)
		}
	}
	__antithesis_instrumentation__.Notify(497163)

	return nil
}

func forEachDatabaseDesc(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	requiresPrivileges bool,
	fn func(descriptor catalog.DatabaseDescriptor) error,
) error {
	__antithesis_instrumentation__.Notify(497186)
	var dbDescs []catalog.DatabaseDescriptor
	if dbContext == nil {
		__antithesis_instrumentation__.Notify(497189)
		allDbDescs, err := p.Descriptors().GetAllDatabaseDescriptors(ctx, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(497191)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497192)
		}
		__antithesis_instrumentation__.Notify(497190)
		dbDescs = allDbDescs
	} else {
		__antithesis_instrumentation__.Notify(497193)
		dbDescs = append(dbDescs, dbContext)
	}
	__antithesis_instrumentation__.Notify(497187)

	for _, dbDesc := range dbDescs {
		__antithesis_instrumentation__.Notify(497194)
		canSeeDescriptor := !requiresPrivileges
		if requiresPrivileges {
			__antithesis_instrumentation__.Notify(497196)
			var err error
			canSeeDescriptor, err = userCanSeeDescriptor(ctx, p, dbDesc, nil, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(497197)
				return err
			} else {
				__antithesis_instrumentation__.Notify(497198)
			}
		} else {
			__antithesis_instrumentation__.Notify(497199)
		}
		__antithesis_instrumentation__.Notify(497195)
		if canSeeDescriptor {
			__antithesis_instrumentation__.Notify(497200)
			if err := fn(dbDesc); err != nil {
				__antithesis_instrumentation__.Notify(497201)
				return err
			} else {
				__antithesis_instrumentation__.Notify(497202)
			}
		} else {
			__antithesis_instrumentation__.Notify(497203)
		}
	}
	__antithesis_instrumentation__.Notify(497188)

	return nil
}

func forEachTypeDesc(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	fn func(db catalog.DatabaseDescriptor, sc string, typ catalog.TypeDescriptor) error,
) error {
	__antithesis_instrumentation__.Notify(497204)
	all, err := p.Descriptors().GetAllDescriptors(ctx, p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(497207)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497208)
	}
	__antithesis_instrumentation__.Notify(497205)
	lCtx := newInternalLookupCtx(all.OrderedDescriptors(), dbContext)
	for _, id := range lCtx.typIDs {
		__antithesis_instrumentation__.Notify(497209)
		typ := lCtx.typDescs[id]
		dbDesc, err := lCtx.getDatabaseByID(typ.GetParentID())
		if err != nil {
			__antithesis_instrumentation__.Notify(497214)
			continue
		} else {
			__antithesis_instrumentation__.Notify(497215)
		}
		__antithesis_instrumentation__.Notify(497210)
		scName, err := lCtx.getSchemaNameByID(typ.GetParentSchemaID())
		if err != nil {
			__antithesis_instrumentation__.Notify(497216)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497217)
		}
		__antithesis_instrumentation__.Notify(497211)
		canSeeDescriptor, err := userCanSeeDescriptor(ctx, p, typ, dbDesc, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(497218)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497219)
		}
		__antithesis_instrumentation__.Notify(497212)
		if !canSeeDescriptor {
			__antithesis_instrumentation__.Notify(497220)
			continue
		} else {
			__antithesis_instrumentation__.Notify(497221)
		}
		__antithesis_instrumentation__.Notify(497213)
		if err := fn(dbDesc, scName, typ); err != nil {
			__antithesis_instrumentation__.Notify(497222)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497223)
		}
	}
	__antithesis_instrumentation__.Notify(497206)
	return nil
}

func forEachTableDesc(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	virtualOpts virtualOpts,
	fn func(catalog.DatabaseDescriptor, string, catalog.TableDescriptor) error,
) error {
	__antithesis_instrumentation__.Notify(497224)
	return forEachTableDescWithTableLookup(ctx, p, dbContext, virtualOpts, func(
		db catalog.DatabaseDescriptor,
		scName string,
		table catalog.TableDescriptor,
		_ tableLookupFn,
	) error {
		__antithesis_instrumentation__.Notify(497225)
		return fn(db, scName, table)
	})
}

type virtualOpts int

const (
	virtualMany virtualOpts = iota

	virtualCurrentDB

	hideVirtual
)

func forEachTableDescAll(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	virtualOpts virtualOpts,
	fn func(catalog.DatabaseDescriptor, string, catalog.TableDescriptor) error,
) error {
	__antithesis_instrumentation__.Notify(497226)
	return forEachTableDescAllWithTableLookup(ctx, p, dbContext, virtualOpts, func(
		db catalog.DatabaseDescriptor,
		scName string,
		table catalog.TableDescriptor,
		_ tableLookupFn,
	) error {
		__antithesis_instrumentation__.Notify(497227)
		return fn(db, scName, table)
	})
}

func forEachTableDescAllWithTableLookup(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	virtualOpts virtualOpts,
	fn func(catalog.DatabaseDescriptor, string, catalog.TableDescriptor, tableLookupFn) error,
) error {
	__antithesis_instrumentation__.Notify(497228)
	return forEachTableDescWithTableLookupInternal(
		ctx, p, dbContext, virtualOpts, true, fn,
	)
}

func forEachTableDescWithTableLookup(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	virtualOpts virtualOpts,
	fn func(catalog.DatabaseDescriptor, string, catalog.TableDescriptor, tableLookupFn) error,
) error {
	__antithesis_instrumentation__.Notify(497229)
	return forEachTableDescWithTableLookupInternal(
		ctx, p, dbContext, virtualOpts, false, fn,
	)
}

func getSchemaNames(
	ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor,
) (map[descpb.ID]string, error) {
	__antithesis_instrumentation__.Notify(497230)
	if dbContext != nil {
		__antithesis_instrumentation__.Notify(497234)
		return p.Descriptors().GetSchemasForDatabase(ctx, p.txn, dbContext)
	} else {
		__antithesis_instrumentation__.Notify(497235)
	}
	__antithesis_instrumentation__.Notify(497231)
	ret := make(map[descpb.ID]string)
	allDbDescs, err := p.Descriptors().GetAllDatabaseDescriptors(ctx, p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(497236)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(497237)
	}
	__antithesis_instrumentation__.Notify(497232)
	for _, db := range allDbDescs {
		__antithesis_instrumentation__.Notify(497238)
		if db == nil {
			__antithesis_instrumentation__.Notify(497241)
			return nil, catalog.ErrDescriptorNotFound
		} else {
			__antithesis_instrumentation__.Notify(497242)
		}
		__antithesis_instrumentation__.Notify(497239)
		schemas, err := p.Descriptors().GetSchemasForDatabase(ctx, p.txn, db)
		if err != nil {
			__antithesis_instrumentation__.Notify(497243)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(497244)
		}
		__antithesis_instrumentation__.Notify(497240)
		for id, name := range schemas {
			__antithesis_instrumentation__.Notify(497245)
			ret[id] = name
		}
	}
	__antithesis_instrumentation__.Notify(497233)
	return ret, nil
}

func forEachTableDescWithTableLookupInternal(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	virtualOpts virtualOpts,
	allowAdding bool,
	fn func(catalog.DatabaseDescriptor, string, catalog.TableDescriptor, tableLookupFn) error,
) error {
	__antithesis_instrumentation__.Notify(497246)
	all, err := p.Descriptors().GetAllDescriptors(ctx, p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(497248)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497249)
	}
	__antithesis_instrumentation__.Notify(497247)
	return forEachTableDescWithTableLookupInternalFromDescriptors(
		ctx, p, dbContext, virtualOpts, allowAdding, all, fn)
}

func forEachTypeDescWithTableLookupInternalFromDescriptors(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	allowAdding bool,
	c nstree.Catalog,
	fn func(catalog.DatabaseDescriptor, string, catalog.TypeDescriptor, tableLookupFn) error,
) error {
	__antithesis_instrumentation__.Notify(497250)
	lCtx := newInternalLookupCtx(c.OrderedDescriptors(), dbContext)

	for _, typID := range lCtx.typIDs {
		__antithesis_instrumentation__.Notify(497252)
		typDesc := lCtx.typDescs[typID]
		if typDesc.Dropped() {
			__antithesis_instrumentation__.Notify(497258)
			continue
		} else {
			__antithesis_instrumentation__.Notify(497259)
		}
		__antithesis_instrumentation__.Notify(497253)
		dbDesc, err := lCtx.getDatabaseByID(typDesc.GetParentID())
		if err != nil {
			__antithesis_instrumentation__.Notify(497260)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497261)
		}
		__antithesis_instrumentation__.Notify(497254)
		canSeeDescriptor, err := userCanSeeDescriptor(ctx, p, typDesc, dbDesc, allowAdding)
		if err != nil {
			__antithesis_instrumentation__.Notify(497262)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497263)
		}
		__antithesis_instrumentation__.Notify(497255)
		if !canSeeDescriptor {
			__antithesis_instrumentation__.Notify(497264)
			continue
		} else {
			__antithesis_instrumentation__.Notify(497265)
		}
		__antithesis_instrumentation__.Notify(497256)
		scName, err := lCtx.getSchemaNameByID(typDesc.GetParentSchemaID())
		if err != nil {
			__antithesis_instrumentation__.Notify(497266)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497267)
		}
		__antithesis_instrumentation__.Notify(497257)
		if err := fn(dbDesc, scName, typDesc, lCtx); err != nil {
			__antithesis_instrumentation__.Notify(497268)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497269)
		}
	}
	__antithesis_instrumentation__.Notify(497251)
	return nil
}

func forEachTableDescWithTableLookupInternalFromDescriptors(
	ctx context.Context,
	p *planner,
	dbContext catalog.DatabaseDescriptor,
	virtualOpts virtualOpts,
	allowAdding bool,
	c nstree.Catalog,
	fn func(catalog.DatabaseDescriptor, string, catalog.TableDescriptor, tableLookupFn) error,
) error {
	__antithesis_instrumentation__.Notify(497270)
	lCtx := newInternalLookupCtx(c.OrderedDescriptors(), dbContext)

	if virtualOpts == virtualMany || func() bool {
		__antithesis_instrumentation__.Notify(497273)
		return virtualOpts == virtualCurrentDB == true
	}() == true {
		__antithesis_instrumentation__.Notify(497274)

		vt := p.getVirtualTabler()
		vEntries := vt.getSchemas()
		vSchemaNames := vt.getSchemaNames()
		iterate := func(dbDesc catalog.DatabaseDescriptor) error {
			__antithesis_instrumentation__.Notify(497276)
			for _, virtSchemaName := range vSchemaNames {
				__antithesis_instrumentation__.Notify(497278)
				e := vEntries[virtSchemaName]
				for _, tName := range e.orderedDefNames {
					__antithesis_instrumentation__.Notify(497279)
					te := e.defs[tName]
					if err := fn(dbDesc, virtSchemaName, te.desc, lCtx); err != nil {
						__antithesis_instrumentation__.Notify(497280)
						return err
					} else {
						__antithesis_instrumentation__.Notify(497281)
					}
				}
			}
			__antithesis_instrumentation__.Notify(497277)
			return nil
		}
		__antithesis_instrumentation__.Notify(497275)

		switch virtualOpts {
		case virtualCurrentDB:
			__antithesis_instrumentation__.Notify(497282)
			if err := iterate(dbContext); err != nil {
				__antithesis_instrumentation__.Notify(497285)
				return err
			} else {
				__antithesis_instrumentation__.Notify(497286)
			}
		case virtualMany:
			__antithesis_instrumentation__.Notify(497283)
			for _, dbID := range lCtx.dbIDs {
				__antithesis_instrumentation__.Notify(497287)
				dbDesc := lCtx.dbDescs[dbID]
				if err := iterate(dbDesc); err != nil {
					__antithesis_instrumentation__.Notify(497288)
					return err
				} else {
					__antithesis_instrumentation__.Notify(497289)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(497284)
		}
	} else {
		__antithesis_instrumentation__.Notify(497290)
	}
	__antithesis_instrumentation__.Notify(497271)

	for _, tbID := range lCtx.tbIDs {
		__antithesis_instrumentation__.Notify(497291)
		table := lCtx.tbDescs[tbID]
		dbDesc, parentExists := lCtx.dbDescs[table.GetParentID()]
		canSeeDescriptor, err := userCanSeeDescriptor(ctx, p, table, dbDesc, allowAdding)
		if err != nil {
			__antithesis_instrumentation__.Notify(497295)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497296)
		}
		__antithesis_instrumentation__.Notify(497292)
		if table.Dropped() || func() bool {
			__antithesis_instrumentation__.Notify(497297)
			return !canSeeDescriptor == true
		}() == true {
			__antithesis_instrumentation__.Notify(497298)
			continue
		} else {
			__antithesis_instrumentation__.Notify(497299)
		}
		__antithesis_instrumentation__.Notify(497293)
		var scName string
		if parentExists {
			__antithesis_instrumentation__.Notify(497300)
			var ok bool
			scName, ok, err = lCtx.GetSchemaName(
				ctx, table.GetParentSchemaID(), table.GetParentID(), p.ExecCfg().Settings.Version,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(497303)
				return err
			} else {
				__antithesis_instrumentation__.Notify(497304)
			}
			__antithesis_instrumentation__.Notify(497301)

			if !ok && func() bool {
				__antithesis_instrumentation__.Notify(497305)
				return !table.IsTemporary() == true
			}() == true {
				__antithesis_instrumentation__.Notify(497306)
				return errors.AssertionFailedf("schema id %d not found", table.GetParentSchemaID())
			} else {
				__antithesis_instrumentation__.Notify(497307)
			}
			__antithesis_instrumentation__.Notify(497302)
			if !ok {
				__antithesis_instrumentation__.Notify(497308)
				namesForSchema, err := getSchemaNames(ctx, p, dbDesc)
				if err != nil {
					__antithesis_instrumentation__.Notify(497310)
					return errors.Wrapf(err, "failed to look up schema id %d",
						table.GetParentSchemaID())
				} else {
					__antithesis_instrumentation__.Notify(497311)
				}
				__antithesis_instrumentation__.Notify(497309)
				for id, n := range namesForSchema {
					__antithesis_instrumentation__.Notify(497312)
					_, exists, err := lCtx.GetSchemaName(ctx, id, dbDesc.GetID(), p.ExecCfg().Settings.Version)
					if err != nil {
						__antithesis_instrumentation__.Notify(497316)
						return err
					} else {
						__antithesis_instrumentation__.Notify(497317)
					}
					__antithesis_instrumentation__.Notify(497313)
					if exists {
						__antithesis_instrumentation__.Notify(497318)
						continue
					} else {
						__antithesis_instrumentation__.Notify(497319)
					}
					__antithesis_instrumentation__.Notify(497314)
					lCtx.schemaNames[id] = n
					var found bool
					scName, found, err = lCtx.GetSchemaName(ctx, id, dbDesc.GetID(), p.ExecCfg().Settings.Version)
					if err != nil {
						__antithesis_instrumentation__.Notify(497320)
						return err
					} else {
						__antithesis_instrumentation__.Notify(497321)
					}
					__antithesis_instrumentation__.Notify(497315)
					if !found {
						__antithesis_instrumentation__.Notify(497322)
						return errors.AssertionFailedf("schema id %d not found", id)
					} else {
						__antithesis_instrumentation__.Notify(497323)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(497324)
			}
		} else {
			__antithesis_instrumentation__.Notify(497325)
		}
		__antithesis_instrumentation__.Notify(497294)
		if err := fn(dbDesc, scName, table, lCtx); err != nil {
			__antithesis_instrumentation__.Notify(497326)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497327)
		}
	}
	__antithesis_instrumentation__.Notify(497272)
	return nil
}

type roleOptions struct {
	*tree.DJSON
}

func (r roleOptions) noLogin() (tree.DBool, error) {
	__antithesis_instrumentation__.Notify(497328)
	nologin, err := r.Exists("NOLOGIN")
	return tree.DBool(nologin), err
}

func (r roleOptions) validUntil(p *planner) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(497329)
	const validUntilKey = "VALID UNTIL"
	jsonValue, err := r.FetchValKey(validUntilKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(497335)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(497336)
	}
	__antithesis_instrumentation__.Notify(497330)
	if jsonValue == nil {
		__antithesis_instrumentation__.Notify(497337)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(497338)
	}
	__antithesis_instrumentation__.Notify(497331)
	validUntilText, err := jsonValue.AsText()
	if err != nil {
		__antithesis_instrumentation__.Notify(497339)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(497340)
	}
	__antithesis_instrumentation__.Notify(497332)
	if validUntilText == nil {
		__antithesis_instrumentation__.Notify(497341)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(497342)
	}
	__antithesis_instrumentation__.Notify(497333)
	validUntil, _, err := pgdate.ParseTimestamp(
		p.EvalContext().GetRelativeParseTime(),
		pgdate.DefaultDateStyle(),
		*validUntilText,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(497343)
		return nil, errors.Errorf("rolValidUntil string %s could not be parsed with datestyle %s", *validUntilText, p.EvalContext().GetDateStyle())
	} else {
		__antithesis_instrumentation__.Notify(497344)
	}
	__antithesis_instrumentation__.Notify(497334)
	return tree.MakeDTimestampTZ(validUntil, time.Second)
}

func (r roleOptions) createDB() (tree.DBool, error) {
	__antithesis_instrumentation__.Notify(497345)
	createDB, err := r.Exists("CREATEDB")
	return tree.DBool(createDB), err
}

func (r roleOptions) createRole() (tree.DBool, error) {
	__antithesis_instrumentation__.Notify(497346)
	createRole, err := r.Exists("CREATEROLE")
	return tree.DBool(createRole), err
}

func forEachRoleQuery(ctx context.Context, p *planner) string {
	__antithesis_instrumentation__.Notify(497347)
	return `
SELECT
	u.username,
	"isRole",
  drs.settings,
	json_object_agg(COALESCE(ro.option, 'null'), ro.value)
FROM
	system.users AS u
	LEFT JOIN system.role_options AS ro ON
			ro.username = u.username
  LEFT JOIN system.database_role_settings AS drs ON 
			drs.role_name = u.username AND drs.database_id = 0
GROUP BY
	u.username, "isRole", drs.settings;
`
}

func forEachRole(
	ctx context.Context,
	p *planner,
	fn func(username security.SQLUsername, isRole bool, options roleOptions, settings tree.Datum) error,
) error {
	__antithesis_instrumentation__.Notify(497348)
	query := forEachRoleQuery(ctx, p)

	rows, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryBuffered(
		ctx, "read-roles", p.txn, query,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(497351)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497352)
	}
	__antithesis_instrumentation__.Notify(497349)

	for _, row := range rows {
		__antithesis_instrumentation__.Notify(497353)
		usernameS := tree.MustBeDString(row[0])
		isRole, ok := row[1].(*tree.DBool)
		if !ok {
			__antithesis_instrumentation__.Notify(497356)
			return errors.Errorf("isRole should be a boolean value, found %s instead", row[1].ResolvedType())
		} else {
			__antithesis_instrumentation__.Notify(497357)
		}
		__antithesis_instrumentation__.Notify(497354)

		defaultSettings := row[2]
		roleOptionsJSON, ok := row[3].(*tree.DJSON)
		if !ok {
			__antithesis_instrumentation__.Notify(497358)
			return errors.Errorf("roleOptionJson should be a JSON value, found %s instead", row[3].ResolvedType())
		} else {
			__antithesis_instrumentation__.Notify(497359)
		}
		__antithesis_instrumentation__.Notify(497355)
		options := roleOptions{roleOptionsJSON}

		username := security.MakeSQLUsernameFromPreNormalizedString(string(usernameS))
		if err := fn(username, bool(*isRole), options, defaultSettings); err != nil {
			__antithesis_instrumentation__.Notify(497360)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497361)
		}
	}
	__antithesis_instrumentation__.Notify(497350)

	return nil
}

func forEachRoleMembership(
	ctx context.Context,
	ie sqlutil.InternalExecutor,
	txn *kv.Txn,
	fn func(role, member security.SQLUsername, isAdmin bool) error,
) (retErr error) {
	__antithesis_instrumentation__.Notify(497362)
	const query = `SELECT "role", "member", "isAdmin" FROM system.role_members`
	it, err := ie.QueryIterator(ctx, "read-members", txn, query)
	if err != nil {
		__antithesis_instrumentation__.Notify(497366)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497367)
	}
	__antithesis_instrumentation__.Notify(497363)

	defer func() {
		__antithesis_instrumentation__.Notify(497368)
		retErr = errors.CombineErrors(retErr, it.Close())
	}()
	__antithesis_instrumentation__.Notify(497364)

	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(497369)
		row := it.Cur()
		roleName := tree.MustBeDString(row[0])
		memberName := tree.MustBeDString(row[1])
		isAdmin := row[2].(*tree.DBool)

		if err := fn(
			security.MakeSQLUsernameFromPreNormalizedString(string(roleName)),
			security.MakeSQLUsernameFromPreNormalizedString(string(memberName)),
			bool(*isAdmin)); err != nil {
			__antithesis_instrumentation__.Notify(497370)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497371)
		}
	}
	__antithesis_instrumentation__.Notify(497365)
	return err
}

func userCanSeeDescriptor(
	ctx context.Context, p *planner, desc, parentDBDesc catalog.Descriptor, allowAdding bool,
) (bool, error) {
	__antithesis_instrumentation__.Notify(497372)
	if !descriptorIsVisible(desc, allowAdding) {
		__antithesis_instrumentation__.Notify(497375)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(497376)
	}
	__antithesis_instrumentation__.Notify(497373)

	canSeeDescriptor := p.CheckAnyPrivilege(ctx, desc) == nil

	if parentDBDesc != nil {
		__antithesis_instrumentation__.Notify(497377)
		canSeeDescriptor = canSeeDescriptor || func() bool {
			__antithesis_instrumentation__.Notify(497378)
			return p.CheckPrivilege(ctx, parentDBDesc, privilege.CONNECT) == nil == true
		}() == true
	} else {
		__antithesis_instrumentation__.Notify(497379)
	}
	__antithesis_instrumentation__.Notify(497374)
	return canSeeDescriptor, nil
}

func descriptorIsVisible(desc catalog.Descriptor, allowAdding bool) bool {
	__antithesis_instrumentation__.Notify(497380)
	return desc.Public() || func() bool {
		__antithesis_instrumentation__.Notify(497381)
		return (allowAdding && func() bool {
			__antithesis_instrumentation__.Notify(497382)
			return desc.Adding() == true
		}() == true) == true
	}() == true
}
