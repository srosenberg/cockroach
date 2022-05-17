package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"strings"
	"time"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catformat"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/sql/vtable"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
	"golang.org/x/text/collate"
)

var (
	oidZero   = tree.NewDOid(0)
	zeroVal   = tree.DZero
	negOneVal = tree.NewDInt(-1)

	passwdStarString = tree.NewDString("********")
)

const (
	indexTypeForwardIndex  = "prefix"
	indexTypeInvertedIndex = "inverted"
)

const (
	indoptionDesc = 0x01

	indoptionNullsFirst = 0x02
)

type RewriteEvTypes string

const evTypeSelect RewriteEvTypes = "1"

type PGShDependType string

const (
	sharedDependencyOwner PGShDependType = "o"

	sharedDependencyACL PGShDependType = "a"

	sharedDependencyPin PGShDependType = "p"
)

var forwardIndexOid = stringOid(indexTypeForwardIndex)
var invertedIndexOid = stringOid(indexTypeInvertedIndex)

var pgCatalog = virtualSchema{
	name: pgCatalogName,
	undefinedTables: buildStringSet(

		"pg_pltemplate",
	),
	tableDefs: map[descpb.ID]virtualSchemaDef{
		catconstants.PgCatalogAggregateTableID:                  pgCatalogAggregateTable,
		catconstants.PgCatalogAmTableID:                         pgCatalogAmTable,
		catconstants.PgCatalogAmopTableID:                       pgCatalogAmopTable,
		catconstants.PgCatalogAmprocTableID:                     pgCatalogAmprocTable,
		catconstants.PgCatalogAttrDefTableID:                    pgCatalogAttrDefTable,
		catconstants.PgCatalogAttributeTableID:                  pgCatalogAttributeTable,
		catconstants.PgCatalogAuthIDTableID:                     pgCatalogAuthIDTable,
		catconstants.PgCatalogAuthMembersTableID:                pgCatalogAuthMembersTable,
		catconstants.PgCatalogAvailableExtensionVersionsTableID: pgCatalogAvailableExtensionVersionsTable,
		catconstants.PgCatalogAvailableExtensionsTableID:        pgCatalogAvailableExtensionsTable,
		catconstants.PgCatalogCastTableID:                       pgCatalogCastTable,
		catconstants.PgCatalogClassTableID:                      pgCatalogClassTable,
		catconstants.PgCatalogCollationTableID:                  pgCatalogCollationTable,
		catconstants.PgCatalogConfigTableID:                     pgCatalogConfigTable,
		catconstants.PgCatalogConstraintTableID:                 pgCatalogConstraintTable,
		catconstants.PgCatalogConversionTableID:                 pgCatalogConversionTable,
		catconstants.PgCatalogCursorsTableID:                    pgCatalogCursorsTable,
		catconstants.PgCatalogDatabaseTableID:                   pgCatalogDatabaseTable,
		catconstants.PgCatalogDbRoleSettingTableID:              pgCatalogDbRoleSettingTable,
		catconstants.PgCatalogDefaultACLTableID:                 pgCatalogDefaultACLTable,
		catconstants.PgCatalogDependTableID:                     pgCatalogDependTable,
		catconstants.PgCatalogDescriptionTableID:                pgCatalogDescriptionTable,
		catconstants.PgCatalogEnumTableID:                       pgCatalogEnumTable,
		catconstants.PgCatalogEventTriggerTableID:               pgCatalogEventTriggerTable,
		catconstants.PgCatalogExtensionTableID:                  pgCatalogExtensionTable,
		catconstants.PgCatalogFileSettingsTableID:               pgCatalogFileSettingsTable,
		catconstants.PgCatalogForeignDataWrapperTableID:         pgCatalogForeignDataWrapperTable,
		catconstants.PgCatalogForeignServerTableID:              pgCatalogForeignServerTable,
		catconstants.PgCatalogForeignTableTableID:               pgCatalogForeignTableTable,
		catconstants.PgCatalogGroupTableID:                      pgCatalogGroupTable,
		catconstants.PgCatalogHbaFileRulesTableID:               pgCatalogHbaFileRulesTable,
		catconstants.PgCatalogIndexTableID:                      pgCatalogIndexTable,
		catconstants.PgCatalogIndexesTableID:                    pgCatalogIndexesTable,
		catconstants.PgCatalogInheritsTableID:                   pgCatalogInheritsTable,
		catconstants.PgCatalogInitPrivsTableID:                  pgCatalogInitPrivsTable,
		catconstants.PgCatalogLanguageTableID:                   pgCatalogLanguageTable,
		catconstants.PgCatalogLargeobjectMetadataTableID:        pgCatalogLargeobjectMetadataTable,
		catconstants.PgCatalogLargeobjectTableID:                pgCatalogLargeobjectTable,
		catconstants.PgCatalogLocksTableID:                      pgCatalogLocksTable,
		catconstants.PgCatalogMatViewsTableID:                   pgCatalogMatViewsTable,
		catconstants.PgCatalogNamespaceTableID:                  pgCatalogNamespaceTable,
		catconstants.PgCatalogOpclassTableID:                    pgCatalogOpclassTable,
		catconstants.PgCatalogOperatorTableID:                   pgCatalogOperatorTable,
		catconstants.PgCatalogOpfamilyTableID:                   pgCatalogOpfamilyTable,
		catconstants.PgCatalogPartitionedTableTableID:           pgCatalogPartitionedTableTable,
		catconstants.PgCatalogPoliciesTableID:                   pgCatalogPoliciesTable,
		catconstants.PgCatalogPolicyTableID:                     pgCatalogPolicyTable,
		catconstants.PgCatalogPreparedStatementsTableID:         pgCatalogPreparedStatementsTable,
		catconstants.PgCatalogPreparedXactsTableID:              pgCatalogPreparedXactsTable,
		catconstants.PgCatalogProcTableID:                       pgCatalogProcTable,
		catconstants.PgCatalogPublicationRelTableID:             pgCatalogPublicationRelTable,
		catconstants.PgCatalogPublicationTableID:                pgCatalogPublicationTable,
		catconstants.PgCatalogPublicationTablesTableID:          pgCatalogPublicationTablesTable,
		catconstants.PgCatalogRangeTableID:                      pgCatalogRangeTable,
		catconstants.PgCatalogReplicationOriginStatusTableID:    pgCatalogReplicationOriginStatusTable,
		catconstants.PgCatalogReplicationOriginTableID:          pgCatalogReplicationOriginTable,
		catconstants.PgCatalogReplicationSlotsTableID:           pgCatalogReplicationSlotsTable,
		catconstants.PgCatalogRewriteTableID:                    pgCatalogRewriteTable,
		catconstants.PgCatalogRolesTableID:                      pgCatalogRolesTable,
		catconstants.PgCatalogRulesTableID:                      pgCatalogRulesTable,
		catconstants.PgCatalogSecLabelsTableID:                  pgCatalogSecLabelsTable,
		catconstants.PgCatalogSecurityLabelTableID:              pgCatalogSecurityLabelTable,
		catconstants.PgCatalogSequenceTableID:                   pgCatalogSequenceTable,
		catconstants.PgCatalogSequencesTableID:                  pgCatalogSequencesTable,
		catconstants.PgCatalogSettingsTableID:                   pgCatalogSettingsTable,
		catconstants.PgCatalogShadowTableID:                     pgCatalogShadowTable,
		catconstants.PgCatalogSharedDescriptionTableID:          pgCatalogSharedDescriptionTable,
		catconstants.PgCatalogSharedSecurityLabelTableID:        pgCatalogSharedSecurityLabelTable,
		catconstants.PgCatalogShdependTableID:                   pgCatalogShdependTable,
		catconstants.PgCatalogShmemAllocationsTableID:           pgCatalogShmemAllocationsTable,
		catconstants.PgCatalogStatActivityTableID:               pgCatalogStatActivityTable,
		catconstants.PgCatalogStatAllIndexesTableID:             pgCatalogStatAllIndexesTable,
		catconstants.PgCatalogStatAllTablesTableID:              pgCatalogStatAllTablesTable,
		catconstants.PgCatalogStatArchiverTableID:               pgCatalogStatArchiverTable,
		catconstants.PgCatalogStatBgwriterTableID:               pgCatalogStatBgwriterTable,
		catconstants.PgCatalogStatDatabaseConflictsTableID:      pgCatalogStatDatabaseConflictsTable,
		catconstants.PgCatalogStatDatabaseTableID:               pgCatalogStatDatabaseTable,
		catconstants.PgCatalogStatGssapiTableID:                 pgCatalogStatGssapiTable,
		catconstants.PgCatalogStatProgressAnalyzeTableID:        pgCatalogStatProgressAnalyzeTable,
		catconstants.PgCatalogStatProgressBasebackupTableID:     pgCatalogStatProgressBasebackupTable,
		catconstants.PgCatalogStatProgressClusterTableID:        pgCatalogStatProgressClusterTable,
		catconstants.PgCatalogStatProgressCreateIndexTableID:    pgCatalogStatProgressCreateIndexTable,
		catconstants.PgCatalogStatProgressVacuumTableID:         pgCatalogStatProgressVacuumTable,
		catconstants.PgCatalogStatReplicationTableID:            pgCatalogStatReplicationTable,
		catconstants.PgCatalogStatSlruTableID:                   pgCatalogStatSlruTable,
		catconstants.PgCatalogStatSslTableID:                    pgCatalogStatSslTable,
		catconstants.PgCatalogStatSubscriptionTableID:           pgCatalogStatSubscriptionTable,
		catconstants.PgCatalogStatSysIndexesTableID:             pgCatalogStatSysIndexesTable,
		catconstants.PgCatalogStatSysTablesTableID:              pgCatalogStatSysTablesTable,
		catconstants.PgCatalogStatUserFunctionsTableID:          pgCatalogStatUserFunctionsTable,
		catconstants.PgCatalogStatUserIndexesTableID:            pgCatalogStatUserIndexesTable,
		catconstants.PgCatalogStatUserTablesTableID:             pgCatalogStatUserTablesTable,
		catconstants.PgCatalogStatWalReceiverTableID:            pgCatalogStatWalReceiverTable,
		catconstants.PgCatalogStatXactAllTablesTableID:          pgCatalogStatXactAllTablesTable,
		catconstants.PgCatalogStatXactSysTablesTableID:          pgCatalogStatXactSysTablesTable,
		catconstants.PgCatalogStatXactUserFunctionsTableID:      pgCatalogStatXactUserFunctionsTable,
		catconstants.PgCatalogStatXactUserTablesTableID:         pgCatalogStatXactUserTablesTable,
		catconstants.PgCatalogStatioAllIndexesTableID:           pgCatalogStatioAllIndexesTable,
		catconstants.PgCatalogStatioAllSequencesTableID:         pgCatalogStatioAllSequencesTable,
		catconstants.PgCatalogStatioAllTablesTableID:            pgCatalogStatioAllTablesTable,
		catconstants.PgCatalogStatioSysIndexesTableID:           pgCatalogStatioSysIndexesTable,
		catconstants.PgCatalogStatioSysSequencesTableID:         pgCatalogStatioSysSequencesTable,
		catconstants.PgCatalogStatioSysTablesTableID:            pgCatalogStatioSysTablesTable,
		catconstants.PgCatalogStatioUserIndexesTableID:          pgCatalogStatioUserIndexesTable,
		catconstants.PgCatalogStatioUserSequencesTableID:        pgCatalogStatioUserSequencesTable,
		catconstants.PgCatalogStatioUserTablesTableID:           pgCatalogStatioUserTablesTable,
		catconstants.PgCatalogStatisticExtDataTableID:           pgCatalogStatisticExtDataTable,
		catconstants.PgCatalogStatisticExtTableID:               pgCatalogStatisticExtTable,
		catconstants.PgCatalogStatisticTableID:                  pgCatalogStatisticTable,
		catconstants.PgCatalogStatsExtTableID:                   pgCatalogStatsExtTable,
		catconstants.PgCatalogStatsTableID:                      pgCatalogStatsTable,
		catconstants.PgCatalogSubscriptionRelTableID:            pgCatalogSubscriptionRelTable,
		catconstants.PgCatalogSubscriptionTableID:               pgCatalogSubscriptionTable,
		catconstants.PgCatalogTablesTableID:                     pgCatalogTablesTable,
		catconstants.PgCatalogTablespaceTableID:                 pgCatalogTablespaceTable,
		catconstants.PgCatalogTimezoneAbbrevsTableID:            pgCatalogTimezoneAbbrevsTable,
		catconstants.PgCatalogTimezoneNamesTableID:              pgCatalogTimezoneNamesTable,
		catconstants.PgCatalogTransformTableID:                  pgCatalogTransformTable,
		catconstants.PgCatalogTriggerTableID:                    pgCatalogTriggerTable,
		catconstants.PgCatalogTsConfigMapTableID:                pgCatalogTsConfigMapTable,
		catconstants.PgCatalogTsConfigTableID:                   pgCatalogTsConfigTable,
		catconstants.PgCatalogTsDictTableID:                     pgCatalogTsDictTable,
		catconstants.PgCatalogTsParserTableID:                   pgCatalogTsParserTable,
		catconstants.PgCatalogTsTemplateTableID:                 pgCatalogTsTemplateTable,
		catconstants.PgCatalogTypeTableID:                       pgCatalogTypeTable,
		catconstants.PgCatalogUserMappingTableID:                pgCatalogUserMappingTable,
		catconstants.PgCatalogUserMappingsTableID:               pgCatalogUserMappingsTable,
		catconstants.PgCatalogUserTableID:                       pgCatalogUserTable,
		catconstants.PgCatalogViewsTableID:                      pgCatalogViewsTable,
	},

	validWithNoDatabaseContext: false,
	containsTypes:              true,
}

var pgCatalogAmTable = virtualSchemaTable{
	comment: `index access methods (incomplete)
https://www.postgresql.org/docs/9.5/catalog-pg-am.html`,
	schema: vtable.PGCatalogAm,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557742)

		if err := addRow(
			forwardIndexOid,
			tree.NewDName(indexTypeForwardIndex),
			zeroVal,
			zeroVal,
			tree.DBoolTrue,
			tree.DBoolFalse,
			tree.DBoolTrue,
			tree.DBoolTrue,
			tree.DBoolTrue,
			tree.DBoolTrue,
			tree.DBoolTrue,
			tree.DBoolTrue,
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.DBoolFalse,
			oidZero,
			tree.DNull,
			tree.DNull,
			oidZero,
			oidZero,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.NewDString("i"),
		); err != nil {
			__antithesis_instrumentation__.Notify(557745)
			return err
		} else {
			__antithesis_instrumentation__.Notify(557746)
		}
		__antithesis_instrumentation__.Notify(557743)

		if err := addRow(
			invertedIndexOid,
			tree.NewDName(indexTypeInvertedIndex),
			zeroVal,
			zeroVal,
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.DBoolTrue,
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.DBoolFalse,
			oidZero,
			tree.DNull,
			tree.DNull,
			oidZero,
			oidZero,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.NewDString("i"),
		); err != nil {
			__antithesis_instrumentation__.Notify(557747)
			return err
		} else {
			__antithesis_instrumentation__.Notify(557748)
		}
		__antithesis_instrumentation__.Notify(557744)
		return nil
	},
}

var pgCatalogAttrDefTable = makeAllRelationsVirtualTableWithDescriptorIDIndex(
	`column default values
https://www.postgresql.org/docs/9.5/catalog-pg-attrdef.html`,
	vtable.PGCatalogAttrDef,
	virtualMany, false,
	func(ctx context.Context, p *planner, h oidHasher, db catalog.DatabaseDescriptor, scName string,
		table catalog.TableDescriptor,
		lookup simpleSchemaResolver,
		addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557749)
		for _, column := range table.PublicColumns() {
			__antithesis_instrumentation__.Notify(557751)
			if !column.HasDefault() {
				__antithesis_instrumentation__.Notify(557754)

				continue
			} else {
				__antithesis_instrumentation__.Notify(557755)
			}
			__antithesis_instrumentation__.Notify(557752)
			displayExpr, err := schemaexpr.FormatExprForDisplay(
				ctx, table, column.GetDefaultExpr(), &p.semaCtx, p.SessionData(), tree.FmtPGCatalog,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(557756)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557757)
			}
			__antithesis_instrumentation__.Notify(557753)
			defSrc := tree.NewDString(displayExpr)
			if err := addRow(
				h.ColumnOid(table.GetID(), column.GetID()),
				tableOid(table.GetID()),
				tree.NewDInt(tree.DInt(column.GetPGAttributeNum())),
				defSrc,
				defSrc,
			); err != nil {
				__antithesis_instrumentation__.Notify(557758)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557759)
			}
		}
		__antithesis_instrumentation__.Notify(557750)
		return nil
	})

var pgCatalogAttributeTable = makeAllRelationsVirtualTableWithDescriptorIDIndex(
	`table columns (incomplete - see also information_schema.columns)
https://www.postgresql.org/docs/12/catalog-pg-attribute.html`,
	vtable.PGCatalogAttribute,
	virtualMany, true,
	func(ctx context.Context, p *planner, h oidHasher, db catalog.DatabaseDescriptor, scName string,
		table catalog.TableDescriptor,
		lookup simpleSchemaResolver,
		addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557760)

		addColumn := func(column catalog.Column, attRelID tree.Datum, attNum uint32) error {
			__antithesis_instrumentation__.Notify(557763)
			colTyp := column.GetType()

			var isColumnComputed string
			if column.IsComputed() && func() bool {
				__antithesis_instrumentation__.Notify(557766)
				return !column.IsVirtual() == true
			}() == true {
				__antithesis_instrumentation__.Notify(557767)
				isColumnComputed = "s"
			} else {
				__antithesis_instrumentation__.Notify(557768)
				if column.IsComputed() {
					__antithesis_instrumentation__.Notify(557769)
					isColumnComputed = "v"
				} else {
					__antithesis_instrumentation__.Notify(557770)
					isColumnComputed = ""
				}
			}
			__antithesis_instrumentation__.Notify(557764)

			var generatedAsIdentityType string
			if column.IsGeneratedAsIdentity() {
				__antithesis_instrumentation__.Notify(557771)
				if column.IsGeneratedAlwaysAsIdentity() {
					__antithesis_instrumentation__.Notify(557772)
					generatedAsIdentityType = "a"
				} else {
					__antithesis_instrumentation__.Notify(557773)
					if column.IsGeneratedByDefaultAsIdentity() {
						__antithesis_instrumentation__.Notify(557774)
						generatedAsIdentityType = "b"
					} else {
						__antithesis_instrumentation__.Notify(557775)
						return errors.AssertionFailedf(
							"column %s is of wrong generated as identity type (neither ALWAYS nor BY DEFAULT)",
							column.GetName(),
						)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(557776)
				generatedAsIdentityType = ""
			}
			__antithesis_instrumentation__.Notify(557765)

			return addRow(
				attRelID,
				tree.NewDName(column.GetName()),
				typOid(colTyp),
				zeroVal,
				typLen(colTyp),
				tree.NewDInt(tree.DInt(attNum)),
				zeroVal,
				negOneVal,
				tree.NewDInt(tree.DInt(colTyp.TypeModifier())),
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.MakeDBool(tree.DBool(!column.IsNullable())),
				tree.MakeDBool(tree.DBool(column.HasDefault())),
				tree.NewDString(generatedAsIdentityType),
				tree.NewDString(isColumnComputed),
				tree.DBoolFalse,
				tree.DBoolTrue,
				zeroVal,
				typColl(colTyp, h),
				tree.DNull,
				tree.DNull,
				tree.DNull,

				tree.DNull,

				tree.DNull,
			)
		}
		__antithesis_instrumentation__.Notify(557761)

		for _, column := range table.AccessibleColumns() {
			__antithesis_instrumentation__.Notify(557777)
			tableID := tableOid(table.GetID())
			if err := addColumn(column, tableID, column.GetPGAttributeNum()); err != nil {
				__antithesis_instrumentation__.Notify(557778)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557779)
			}
		}
		__antithesis_instrumentation__.Notify(557762)

		columnIdxMap := catalog.ColumnIDToOrdinalMap(table.PublicColumns())
		return catalog.ForEachIndex(table, catalog.IndexOpts{}, func(index catalog.Index) error {
			__antithesis_instrumentation__.Notify(557780)
			for i := 0; i < index.NumKeyColumns(); i++ {
				__antithesis_instrumentation__.Notify(557782)
				colID := index.GetKeyColumnID(i)
				idxID := h.IndexOid(table.GetID(), index.GetID())
				column := table.PublicColumns()[columnIdxMap.GetDefault(colID)]
				if err := addColumn(column, idxID, column.GetPGAttributeNum()); err != nil {
					__antithesis_instrumentation__.Notify(557783)
					return err
				} else {
					__antithesis_instrumentation__.Notify(557784)
				}
			}
			__antithesis_instrumentation__.Notify(557781)
			return nil
		})
	})

var pgCatalogCastTable = virtualSchemaTable{
	comment: `casts (empty - needs filling out)
https://www.postgresql.org/docs/9.6/catalog-pg-cast.html`,
	schema: vtable.PGCatalogCast,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557785)

		return nil
	},
	unimplemented: true,
}

func userIsSuper(
	ctx context.Context, p *planner, username security.SQLUsername,
) (tree.DBool, error) {
	__antithesis_instrumentation__.Notify(557786)
	isSuper, err := p.UserHasAdminRole(ctx, username)
	return tree.DBool(isSuper), err
}

var pgCatalogAuthIDTable = virtualSchemaTable{
	comment: `authorization identifiers - differs from postgres as we do not display passwords, 
and thus do not require admin privileges for access. 
https://www.postgresql.org/docs/9.5/catalog-pg-authid.html`,
	schema: vtable.PGCatalogAuthID,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557787)
		h := makeOidHasher()
		return forEachRole(ctx, p, func(username security.SQLUsername, isRole bool, options roleOptions, _ tree.Datum) error {
			__antithesis_instrumentation__.Notify(557788)
			isRoot := tree.DBool(username.IsRootUser() || func() bool {
				__antithesis_instrumentation__.Notify(557794)
				return username.IsAdminRole() == true
			}() == true)

			roleInherits := tree.DBool(true)
			noLogin, err := options.noLogin()
			if err != nil {
				__antithesis_instrumentation__.Notify(557795)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557796)
			}
			__antithesis_instrumentation__.Notify(557789)
			roleCanLogin := !noLogin
			createDB, err := options.createDB()
			if err != nil {
				__antithesis_instrumentation__.Notify(557797)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557798)
			}
			__antithesis_instrumentation__.Notify(557790)
			rolValidUntil, err := options.validUntil(p)
			if err != nil {
				__antithesis_instrumentation__.Notify(557799)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557800)
			}
			__antithesis_instrumentation__.Notify(557791)
			createRole, err := options.createRole()
			if err != nil {
				__antithesis_instrumentation__.Notify(557801)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557802)
			}
			__antithesis_instrumentation__.Notify(557792)

			isSuper, err := userIsSuper(ctx, p, username)
			if err != nil {
				__antithesis_instrumentation__.Notify(557803)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557804)
			}
			__antithesis_instrumentation__.Notify(557793)

			return addRow(
				h.UserOid(username),
				tree.NewDName(username.Normalized()),
				tree.MakeDBool(isRoot || func() bool {
					__antithesis_instrumentation__.Notify(557805)
					return isSuper == true
				}() == true),
				tree.MakeDBool(roleInherits),
				tree.MakeDBool(isRoot || func() bool {
					__antithesis_instrumentation__.Notify(557806)
					return createRole == true
				}() == true),
				tree.MakeDBool(isRoot || func() bool {
					__antithesis_instrumentation__.Notify(557807)
					return createDB == true
				}() == true),
				tree.MakeDBool(roleCanLogin),
				tree.DBoolFalse,
				tree.DBoolFalse,
				negOneVal,
				passwdStarString,
				rolValidUntil,
			)
		})
	},
}

var pgCatalogAuthMembersTable = virtualSchemaTable{
	comment: `role membership
https://www.postgresql.org/docs/9.5/catalog-pg-auth-members.html`,
	schema: vtable.PGCatalogAuthMembers,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557808)
		h := makeOidHasher()
		return forEachRoleMembership(ctx, p.ExecCfg().InternalExecutor, p.Txn(),
			func(roleName, memberName security.SQLUsername, isAdmin bool) error {
				__antithesis_instrumentation__.Notify(557809)
				return addRow(
					h.UserOid(roleName),
					h.UserOid(memberName),
					tree.DNull,
					tree.MakeDBool(tree.DBool(isAdmin)),
				)
			})
	},
}

var pgCatalogAvailableExtensionsTable = virtualSchemaTable{
	comment: `available extensions
https://www.postgresql.org/docs/9.6/view-pg-available-extensions.html`,
	schema: vtable.PGCatalogAvailableExtensions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557810)

		return nil
	},
	unimplemented: true,
}

func getOwnerOID(desc catalog.Descriptor) tree.Datum {
	__antithesis_instrumentation__.Notify(557811)
	owner := getOwnerOfDesc(desc)
	h := makeOidHasher()
	return h.UserOid(owner)
}

func getOwnerName(desc catalog.Descriptor) tree.Datum {
	__antithesis_instrumentation__.Notify(557812)
	owner := getOwnerOfDesc(desc)
	return tree.NewDName(owner.Normalized())
}

var (
	relKindTable            = tree.NewDString("r")
	relKindIndex            = tree.NewDString("i")
	relKindView             = tree.NewDString("v")
	relKindMaterializedView = tree.NewDString("m")
	relKindSequence         = tree.NewDString("S")

	relPersistencePermanent = tree.NewDString("p")
	relPersistenceTemporary = tree.NewDString("t")
)

var pgCatalogClassTable = makeAllRelationsVirtualTableWithDescriptorIDIndex(
	`tables and relation-like objects (incomplete - see also information_schema.tables/sequences/views)
https://www.postgresql.org/docs/9.5/catalog-pg-class.html`,
	vtable.PGCatalogClass,
	virtualMany, true,
	func(ctx context.Context, p *planner, h oidHasher, db catalog.DatabaseDescriptor, scName string,
		table catalog.TableDescriptor, _ simpleSchemaResolver, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557813)

		relKind := relKindTable
		relAm := forwardIndexOid
		if table.IsView() {
			__antithesis_instrumentation__.Notify(557819)
			relKind = relKindView
			if table.MaterializedView() {
				__antithesis_instrumentation__.Notify(557821)
				relKind = relKindMaterializedView
			} else {
				__antithesis_instrumentation__.Notify(557822)
			}
			__antithesis_instrumentation__.Notify(557820)
			relAm = oidZero
		} else {
			__antithesis_instrumentation__.Notify(557823)
			if table.IsSequence() {
				__antithesis_instrumentation__.Notify(557824)
				relKind = relKindSequence
				relAm = oidZero
			} else {
				__antithesis_instrumentation__.Notify(557825)
			}
		}
		__antithesis_instrumentation__.Notify(557814)
		relPersistence := relPersistencePermanent
		if table.IsTemporary() {
			__antithesis_instrumentation__.Notify(557826)
			relPersistence = relPersistenceTemporary
		} else {
			__antithesis_instrumentation__.Notify(557827)
		}
		__antithesis_instrumentation__.Notify(557815)
		var relOptions tree.Datum = tree.DNull
		if storageParams := table.GetStorageParams(false); len(storageParams) > 0 {
			__antithesis_instrumentation__.Notify(557828)
			relOptionsArr := tree.NewDArray(types.String)
			for _, storageParam := range storageParams {
				__antithesis_instrumentation__.Notify(557830)
				if err := relOptionsArr.Append(tree.NewDString(storageParam)); err != nil {
					__antithesis_instrumentation__.Notify(557831)
					return err
				} else {
					__antithesis_instrumentation__.Notify(557832)
				}
			}
			__antithesis_instrumentation__.Notify(557829)
			relOptions = relOptionsArr
		} else {
			__antithesis_instrumentation__.Notify(557833)
		}
		__antithesis_instrumentation__.Notify(557816)
		namespaceOid := h.NamespaceOid(db.GetID(), scName)
		if err := addRow(
			tableOid(table.GetID()),
			tree.NewDName(table.GetName()),
			namespaceOid,
			tableIDToTypeOID(table),
			oidZero,
			getOwnerOID(table),
			relAm,
			oidZero,
			oidZero,
			tree.DNull,
			tree.DNull,
			zeroVal,
			oidZero,
			tree.MakeDBool(tree.DBool(table.IsPhysicalTable())),
			tree.DBoolFalse,
			relPersistence,
			tree.MakeDBool(tree.DBool(table.IsTemporary())),
			relKind,
			tree.NewDInt(tree.DInt(len(table.AccessibleColumns()))),
			tree.NewDInt(tree.DInt(len(table.GetChecks()))),
			tree.DBoolFalse,
			tree.MakeDBool(tree.DBool(table.IsPhysicalTable())),
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.DBoolFalse,
			zeroVal,
			tree.DNull,
			relOptions,

			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,

			tree.DNull,
		); err != nil {
			__antithesis_instrumentation__.Notify(557834)
			return err
		} else {
			__antithesis_instrumentation__.Notify(557835)
		}
		__antithesis_instrumentation__.Notify(557817)

		if table.IsSequence() {
			__antithesis_instrumentation__.Notify(557836)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(557837)
		}
		__antithesis_instrumentation__.Notify(557818)

		return catalog.ForEachIndex(table, catalog.IndexOpts{}, func(index catalog.Index) error {
			__antithesis_instrumentation__.Notify(557838)
			indexType := forwardIndexOid
			if index.GetType() == descpb.IndexDescriptor_INVERTED {
				__antithesis_instrumentation__.Notify(557840)
				indexType = invertedIndexOid
			} else {
				__antithesis_instrumentation__.Notify(557841)
			}
			__antithesis_instrumentation__.Notify(557839)
			return addRow(
				h.IndexOid(table.GetID(), index.GetID()),
				tree.NewDName(index.GetName()),
				namespaceOid,
				oidZero,
				oidZero,
				getOwnerOID(table),
				indexType,
				oidZero,
				oidZero,
				tree.DNull,
				tree.DNull,
				zeroVal,
				oidZero,
				tree.DBoolFalse,
				tree.DBoolFalse,
				relPersistencePermanent,
				tree.DBoolFalse,
				relKindIndex,
				tree.NewDInt(tree.DInt(index.NumKeyColumns())),
				zeroVal,
				tree.DBoolFalse,
				tree.DBoolFalse,
				tree.DBoolFalse,
				tree.DBoolFalse,
				tree.DBoolFalse,
				zeroVal,
				tree.DNull,
				tree.DNull,

				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,

				tree.DNull,
			)
		})
	})

var pgCatalogCollationTable = virtualSchemaTable{
	comment: `available collations (incomplete)
https://www.postgresql.org/docs/9.5/catalog-pg-collation.html`,
	schema: vtable.PGCatalogCollation,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557842)
		h := makeOidHasher()
		return forEachDatabaseDesc(ctx, p, dbContext, false, func(db catalog.DatabaseDescriptor) error {
			__antithesis_instrumentation__.Notify(557843)
			namespaceOid := h.NamespaceOid(db.GetID(), pgCatalogName)
			add := func(collName string) error {
				__antithesis_instrumentation__.Notify(557847)
				return addRow(
					h.CollationOid(collName),
					tree.NewDString(collName),
					namespaceOid,
					tree.DNull,
					builtins.DatEncodingUTFId,

					tree.DNull,
					tree.DNull,

					tree.DNull,
					tree.DNull,
					tree.DNull,
				)
			}
			__antithesis_instrumentation__.Notify(557844)
			if err := add(tree.DefaultCollationTag); err != nil {
				__antithesis_instrumentation__.Notify(557848)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557849)
			}
			__antithesis_instrumentation__.Notify(557845)
			for _, tag := range collate.Supported() {
				__antithesis_instrumentation__.Notify(557850)
				collName := tag.String()
				if err := add(collName); err != nil {
					__antithesis_instrumentation__.Notify(557851)
					return err
				} else {
					__antithesis_instrumentation__.Notify(557852)
				}
			}
			__antithesis_instrumentation__.Notify(557846)
			return nil
		})
	},
}

var (
	conTypeCheck     = tree.NewDString("c")
	conTypeFK        = tree.NewDString("f")
	conTypePKey      = tree.NewDString("p")
	conTypeUnique    = tree.NewDString("u")
	conTypeTrigger   = tree.NewDString("t")
	conTypeExclusion = tree.NewDString("x")

	_ = conTypeTrigger
	_ = conTypeExclusion

	fkActionNone       = tree.NewDString("a")
	fkActionRestrict   = tree.NewDString("r")
	fkActionCascade    = tree.NewDString("c")
	fkActionSetNull    = tree.NewDString("n")
	fkActionSetDefault = tree.NewDString("d")

	fkActionMap = map[catpb.ForeignKeyAction]tree.Datum{
		catpb.ForeignKeyAction_NO_ACTION:   fkActionNone,
		catpb.ForeignKeyAction_RESTRICT:    fkActionRestrict,
		catpb.ForeignKeyAction_CASCADE:     fkActionCascade,
		catpb.ForeignKeyAction_SET_NULL:    fkActionSetNull,
		catpb.ForeignKeyAction_SET_DEFAULT: fkActionSetDefault,
	}

	fkMatchTypeFull    = tree.NewDString("f")
	fkMatchTypePartial = tree.NewDString("p")
	fkMatchTypeSimple  = tree.NewDString("s")

	fkMatchMap = map[descpb.ForeignKeyReference_Match]tree.Datum{
		descpb.ForeignKeyReference_SIMPLE:  fkMatchTypeSimple,
		descpb.ForeignKeyReference_FULL:    fkMatchTypeFull,
		descpb.ForeignKeyReference_PARTIAL: fkMatchTypePartial,
	}
)

func populateTableConstraints(
	ctx context.Context,
	p *planner,
	h oidHasher,
	db catalog.DatabaseDescriptor,
	scName string,
	table catalog.TableDescriptor,
	tableLookup simpleSchemaResolver,
	addRow func(...tree.Datum) error,
) error {
	__antithesis_instrumentation__.Notify(557853)
	conInfo, err := table.GetConstraintInfoWithLookup(tableLookup.getTableByID)
	if err != nil {
		__antithesis_instrumentation__.Notify(557856)
		return err
	} else {
		__antithesis_instrumentation__.Notify(557857)
	}
	__antithesis_instrumentation__.Notify(557854)
	namespaceOid := h.NamespaceOid(db.GetID(), scName)
	tblOid := tableOid(table.GetID())
	for conName, con := range conInfo {
		__antithesis_instrumentation__.Notify(557858)
		oid := tree.DNull
		contype := tree.DNull
		conindid := oidZero
		confrelid := oidZero
		confupdtype := tree.DNull
		confdeltype := tree.DNull
		confmatchtype := tree.DNull
		conkey := tree.DNull
		confkey := tree.DNull
		consrc := tree.DNull
		conbin := tree.DNull
		condef := tree.DNull

		var err error
		switch con.Kind {
		case descpb.ConstraintTypePK:
			__antithesis_instrumentation__.Notify(557860)
			oid = h.PrimaryKeyConstraintOid(db.GetID(), scName, table.GetID(), con.Index)
			contype = conTypePKey
			conindid = h.IndexOid(table.GetID(), con.Index.ID)

			var err error
			if conkey, err = colIDArrayToDatum(con.Index.KeyColumnIDs); err != nil {
				__antithesis_instrumentation__.Notify(557878)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557879)
			}
			__antithesis_instrumentation__.Notify(557861)
			condef = tree.NewDString(tabledesc.PrimaryKeyString(table))

		case descpb.ConstraintTypeFK:
			__antithesis_instrumentation__.Notify(557862)
			oid = h.ForeignKeyConstraintOid(db.GetID(), scName, table.GetID(), con.FK)
			contype = conTypeFK

			referencedTable, err := tableLookup.getTableByID(con.FK.ReferencedTableID)
			if err != nil {
				__antithesis_instrumentation__.Notify(557880)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557881)
			}
			__antithesis_instrumentation__.Notify(557863)
			if refConstraint, err := tabledesc.FindFKReferencedUniqueConstraint(
				referencedTable, con.FK.ReferencedColumnIDs,
			); err != nil {
				__antithesis_instrumentation__.Notify(557882)

				log.Warningf(ctx, "broken fk reference: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(557883)
				if idx, ok := refConstraint.(*descpb.IndexDescriptor); ok {
					__antithesis_instrumentation__.Notify(557884)
					conindid = h.IndexOid(con.ReferencedTable.ID, idx.ID)
				} else {
					__antithesis_instrumentation__.Notify(557885)
				}
			}
			__antithesis_instrumentation__.Notify(557864)
			confrelid = tableOid(con.ReferencedTable.ID)
			if r, ok := fkActionMap[con.FK.OnUpdate]; ok {
				__antithesis_instrumentation__.Notify(557886)
				confupdtype = r
			} else {
				__antithesis_instrumentation__.Notify(557887)
			}
			__antithesis_instrumentation__.Notify(557865)
			if r, ok := fkActionMap[con.FK.OnDelete]; ok {
				__antithesis_instrumentation__.Notify(557888)
				confdeltype = r
			} else {
				__antithesis_instrumentation__.Notify(557889)
			}
			__antithesis_instrumentation__.Notify(557866)
			if r, ok := fkMatchMap[con.FK.Match]; ok {
				__antithesis_instrumentation__.Notify(557890)
				confmatchtype = r
			} else {
				__antithesis_instrumentation__.Notify(557891)
			}
			__antithesis_instrumentation__.Notify(557867)
			if conkey, err = colIDArrayToDatum(con.FK.OriginColumnIDs); err != nil {
				__antithesis_instrumentation__.Notify(557892)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557893)
			}
			__antithesis_instrumentation__.Notify(557868)
			if confkey, err = colIDArrayToDatum(con.FK.ReferencedColumnIDs); err != nil {
				__antithesis_instrumentation__.Notify(557894)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557895)
			}
			__antithesis_instrumentation__.Notify(557869)
			var buf bytes.Buffer
			if err := showForeignKeyConstraint(
				&buf, db.GetName(),
				table, con.FK,
				tableLookup,
				p.extendedEvalCtx.SessionData().SearchPath,
			); err != nil {
				__antithesis_instrumentation__.Notify(557896)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557897)
			}
			__antithesis_instrumentation__.Notify(557870)
			condef = tree.NewDString(buf.String())

		case descpb.ConstraintTypeUnique:
			__antithesis_instrumentation__.Notify(557871)
			contype = conTypeUnique
			f := tree.NewFmtCtx(tree.FmtSimple)
			if con.Index != nil {
				__antithesis_instrumentation__.Notify(557898)
				oid = h.UniqueConstraintOid(db.GetID(), scName, table.GetID(), con.Index.ID)
				conindid = h.IndexOid(table.GetID(), con.Index.ID)
				var err error
				if conkey, err = colIDArrayToDatum(con.Index.KeyColumnIDs); err != nil {
					__antithesis_instrumentation__.Notify(557901)
					return err
				} else {
					__antithesis_instrumentation__.Notify(557902)
				}
				__antithesis_instrumentation__.Notify(557899)
				f.WriteString("UNIQUE (")
				if err := catformat.FormatIndexElements(
					ctx, table, con.Index, f, p.SemaCtx(), p.SessionData(),
				); err != nil {
					__antithesis_instrumentation__.Notify(557903)
					return err
				} else {
					__antithesis_instrumentation__.Notify(557904)
				}
				__antithesis_instrumentation__.Notify(557900)
				f.WriteByte(')')
				if con.Index.IsPartial() {
					__antithesis_instrumentation__.Notify(557905)
					pred, err := schemaexpr.FormatExprForDisplay(ctx, table, con.Index.Predicate, p.SemaCtx(), p.SessionData(), tree.FmtPGCatalog)
					if err != nil {
						__antithesis_instrumentation__.Notify(557907)
						return err
					} else {
						__antithesis_instrumentation__.Notify(557908)
					}
					__antithesis_instrumentation__.Notify(557906)
					f.WriteString(fmt.Sprintf(" WHERE (%s)", pred))
				} else {
					__antithesis_instrumentation__.Notify(557909)
				}
			} else {
				__antithesis_instrumentation__.Notify(557910)
				if con.UniqueWithoutIndexConstraint != nil {
					__antithesis_instrumentation__.Notify(557911)
					oid = h.UniqueWithoutIndexConstraintOid(
						db.GetID(), scName, table.GetID(), con.UniqueWithoutIndexConstraint,
					)
					f.WriteString("UNIQUE WITHOUT INDEX (")
					colNames, err := table.NamesForColumnIDs(con.UniqueWithoutIndexConstraint.ColumnIDs)
					if err != nil {
						__antithesis_instrumentation__.Notify(557914)
						return err
					} else {
						__antithesis_instrumentation__.Notify(557915)
					}
					__antithesis_instrumentation__.Notify(557912)
					f.WriteString(strings.Join(colNames, ", "))
					f.WriteByte(')')
					if con.UniqueWithoutIndexConstraint.Validity != descpb.ConstraintValidity_Validated {
						__antithesis_instrumentation__.Notify(557916)
						f.WriteString(" NOT VALID")
					} else {
						__antithesis_instrumentation__.Notify(557917)
					}
					__antithesis_instrumentation__.Notify(557913)
					if con.UniqueWithoutIndexConstraint.Predicate != "" {
						__antithesis_instrumentation__.Notify(557918)
						pred, err := schemaexpr.FormatExprForDisplay(ctx, table, con.UniqueWithoutIndexConstraint.Predicate, p.SemaCtx(), p.SessionData(), tree.FmtPGCatalog)
						if err != nil {
							__antithesis_instrumentation__.Notify(557920)
							return err
						} else {
							__antithesis_instrumentation__.Notify(557921)
						}
						__antithesis_instrumentation__.Notify(557919)
						f.WriteString(fmt.Sprintf(" WHERE (%s)", pred))
					} else {
						__antithesis_instrumentation__.Notify(557922)
					}
				} else {
					__antithesis_instrumentation__.Notify(557923)
					return errors.AssertionFailedf(
						"Index or UniqueWithoutIndexConstraint must be non-nil for a unique constraint",
					)
				}
			}
			__antithesis_instrumentation__.Notify(557872)
			condef = tree.NewDString(f.CloseAndGetString())

		case descpb.ConstraintTypeCheck:
			__antithesis_instrumentation__.Notify(557873)
			oid = h.CheckConstraintOid(db.GetID(), scName, table.GetID(), con.CheckConstraint)
			contype = conTypeCheck
			if conkey, err = colIDArrayToDatum(con.CheckConstraint.ColumnIDs); err != nil {
				__antithesis_instrumentation__.Notify(557924)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557925)
			}
			__antithesis_instrumentation__.Notify(557874)
			displayExpr, err := schemaexpr.FormatExprForDisplay(ctx, table, con.Details, &p.semaCtx, p.SessionData(), tree.FmtPGCatalog)
			if err != nil {
				__antithesis_instrumentation__.Notify(557926)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557927)
			}
			__antithesis_instrumentation__.Notify(557875)
			consrc = tree.NewDString(fmt.Sprintf("(%s)", displayExpr))
			conbin = consrc
			validity := ""
			if con.CheckConstraint.Validity != descpb.ConstraintValidity_Validated {
				__antithesis_instrumentation__.Notify(557928)
				validity = " NOT VALID"
			} else {
				__antithesis_instrumentation__.Notify(557929)
			}
			__antithesis_instrumentation__.Notify(557876)
			condef = tree.NewDString(fmt.Sprintf("CHECK ((%s))%s", displayExpr, validity))
		default:
			__antithesis_instrumentation__.Notify(557877)
		}
		__antithesis_instrumentation__.Notify(557859)

		if err := addRow(
			oid,
			dNameOrNull(conName),
			namespaceOid,
			contype,
			tree.DBoolFalse,
			tree.DBoolFalse,
			tree.MakeDBool(tree.DBool(!con.Unvalidated)),
			tblOid,
			oidZero,
			conindid,
			confrelid,
			confupdtype,
			confdeltype,
			confmatchtype,
			tree.DBoolTrue,
			zeroVal,
			tree.DBoolTrue,
			conkey,
			confkey,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
			conbin,
			consrc,
			condef,

			tree.DNull,
		); err != nil {
			__antithesis_instrumentation__.Notify(557930)
			return err
		} else {
			__antithesis_instrumentation__.Notify(557931)
		}
	}
	__antithesis_instrumentation__.Notify(557855)
	return nil
}

type oneAtATimeSchemaResolver struct {
	ctx context.Context
	p   *planner
}

func (r oneAtATimeSchemaResolver) getDatabaseByID(
	id descpb.ID,
) (catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(557932)
	_, desc, err := r.p.Descriptors().GetImmutableDatabaseByID(
		r.ctx, r.p.txn, id, tree.DatabaseLookupFlags{Required: true},
	)
	return desc, err
}

func (r oneAtATimeSchemaResolver) getTableByID(id descpb.ID) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(557933)
	table, err := r.p.LookupTableByID(r.ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(557935)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(557936)
	}
	__antithesis_instrumentation__.Notify(557934)
	return table, nil
}

func (r oneAtATimeSchemaResolver) getSchemaByID(id descpb.ID) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(557937)
	return r.p.Descriptors().Direct().MustGetSchemaDescByID(r.ctx, r.p.txn, id)
}

func makeAllRelationsVirtualTableWithDescriptorIDIndex(
	comment string,
	schemaDef string,
	virtualOpts virtualOpts,
	includesIndexEntries bool,
	populateFromTable func(ctx context.Context, p *planner, h oidHasher, db catalog.DatabaseDescriptor,
		scName string, table catalog.TableDescriptor, lookup simpleSchemaResolver,
		addRow func(...tree.Datum) error,
	) error,
) virtualSchemaTable {
	__antithesis_instrumentation__.Notify(557938)
	populateAll := func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557940)
		h := makeOidHasher()
		return forEachTableDescWithTableLookup(ctx, p, dbContext, virtualOpts,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor, lookup tableLookupFn) error {
				__antithesis_instrumentation__.Notify(557941)
				return populateFromTable(ctx, p, h, db, scName, table, lookup, addRow)
			})
	}
	__antithesis_instrumentation__.Notify(557939)
	return virtualSchemaTable{
		comment: comment,
		schema:  schemaDef,
		indexes: []virtualIndex{
			{
				partial: includesIndexEntries,
				populate: func(ctx context.Context, constraint tree.Datum, p *planner, db catalog.DatabaseDescriptor,
					addRow func(...tree.Datum) error) (bool, error) {
					__antithesis_instrumentation__.Notify(557942)
					var id descpb.ID
					d := tree.UnwrapDatum(p.EvalContext(), constraint)
					if d == tree.DNull {
						__antithesis_instrumentation__.Notify(557950)
						return false, nil
					} else {
						__antithesis_instrumentation__.Notify(557951)
					}
					__antithesis_instrumentation__.Notify(557943)
					switch t := d.(type) {
					case *tree.DOid:
						__antithesis_instrumentation__.Notify(557952)
						id = descpb.ID(t.DInt)
					case *tree.DInt:
						__antithesis_instrumentation__.Notify(557953)
						id = descpb.ID(*t)
					default:
						__antithesis_instrumentation__.Notify(557954)
						return false, errors.AssertionFailedf("unexpected type %T for table id column in virtual table %s",
							d, schemaDef)
					}
					__antithesis_instrumentation__.Notify(557944)
					table, err := p.LookupTableByID(ctx, id)
					if err != nil {
						__antithesis_instrumentation__.Notify(557955)
						if sqlerrors.IsUndefinedRelationError(err) {
							__antithesis_instrumentation__.Notify(557957)

							return false, nil
						} else {
							__antithesis_instrumentation__.Notify(557958)
						}
						__antithesis_instrumentation__.Notify(557956)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(557959)
					}
					__antithesis_instrumentation__.Notify(557945)

					canSeeDescriptor, err := userCanSeeDescriptor(ctx, p, table, db, true)
					if err != nil {
						__antithesis_instrumentation__.Notify(557960)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(557961)
					}
					__antithesis_instrumentation__.Notify(557946)
					if (!table.IsVirtualTable() && func() bool {
						__antithesis_instrumentation__.Notify(557962)
						return table.GetParentID() != db.GetID() == true
					}() == true) || func() bool {
						__antithesis_instrumentation__.Notify(557963)
						return table.Dropped() == true
					}() == true || func() bool {
						__antithesis_instrumentation__.Notify(557964)
						return !canSeeDescriptor == true
					}() == true {
						__antithesis_instrumentation__.Notify(557965)
						return false, nil
					} else {
						__antithesis_instrumentation__.Notify(557966)
					}
					__antithesis_instrumentation__.Notify(557947)
					h := makeOidHasher()
					scResolver := oneAtATimeSchemaResolver{p: p, ctx: ctx}
					sc, err := p.Descriptors().GetImmutableSchemaByID(
						ctx, p.txn, table.GetParentSchemaID(), tree.SchemaLookupFlags{})
					if err != nil {
						__antithesis_instrumentation__.Notify(557967)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(557968)
					}
					__antithesis_instrumentation__.Notify(557948)
					if err := populateFromTable(ctx, p, h, db, sc.GetName(), table, scResolver,
						addRow); err != nil {
						__antithesis_instrumentation__.Notify(557969)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(557970)
					}
					__antithesis_instrumentation__.Notify(557949)
					return true, nil
				},
			},
		},
		populate: populateAll,
	}
}

var pgCatalogConstraintTable = makeAllRelationsVirtualTableWithDescriptorIDIndex(
	`table constraints (incomplete - see also information_schema.table_constraints)
https://www.postgresql.org/docs/9.5/catalog-pg-constraint.html`,
	vtable.PGCatalogConstraint,
	hideVirtual,
	false,
	populateTableConstraints)

func colIDArrayToDatum(arr []descpb.ColumnID) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(557971)
	if len(arr) == 0 {
		__antithesis_instrumentation__.Notify(557974)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(557975)
	}
	__antithesis_instrumentation__.Notify(557972)
	d := tree.NewDArray(types.Int2)
	for _, val := range arr {
		__antithesis_instrumentation__.Notify(557976)
		if err := d.Append(tree.NewDInt(tree.DInt(val))); err != nil {
			__antithesis_instrumentation__.Notify(557977)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(557978)
		}
	}
	__antithesis_instrumentation__.Notify(557973)
	return d, nil
}

func colIDArrayToVector(arr []descpb.ColumnID) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(557979)
	dArr, err := colIDArrayToDatum(arr)
	if err != nil {
		__antithesis_instrumentation__.Notify(557982)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(557983)
	}
	__antithesis_instrumentation__.Notify(557980)
	if dArr == tree.DNull {
		__antithesis_instrumentation__.Notify(557984)
		return dArr, nil
	} else {
		__antithesis_instrumentation__.Notify(557985)
	}
	__antithesis_instrumentation__.Notify(557981)
	return tree.NewDIntVectorFromDArray(tree.MustBeDArray(dArr)), nil
}

var pgCatalogConversionTable = virtualSchemaTable{
	comment: `encoding conversions (empty - unimplemented)
https://www.postgresql.org/docs/9.6/catalog-pg-conversion.html`,
	schema: vtable.PGCatalogConversion,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557986)
		return nil
	},
	unimplemented: true,
}

var pgCatalogDatabaseTable = virtualSchemaTable{
	comment: `available databases (incomplete)
https://www.postgresql.org/docs/9.5/catalog-pg-database.html`,
	schema: vtable.PGCatalogDatabase,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557987)
		return forEachDatabaseDesc(ctx, p, nil, false,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(557988)
				return addRow(
					dbOid(db.GetID()),
					tree.NewDName(db.GetName()),
					getOwnerOID(db),

					builtins.DatEncodingUTFId,
					builtins.DatEncodingEnUTF8,
					builtins.DatEncodingEnUTF8,
					tree.DBoolFalse,
					tree.DBoolTrue,
					negOneVal,
					oidZero,
					tree.DNull,
					tree.DNull,
					oidZero,
					tree.DNull,
				)
			})
	},
}

var pgCatalogDefaultACLTable = virtualSchemaTable{
	comment: `default ACLs; these are the privileges that will be assigned to newly created objects
https://www.postgresql.org/docs/13/catalog-pg-default-acl.html`,
	schema: vtable.PGCatalogDefaultACL,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(557989)
		h := makeOidHasher()
		f := func(defaultPrivilegesForRole catpb.DefaultPrivilegesForRole) error {
			__antithesis_instrumentation__.Notify(557991)
			objectTypes := tree.GetAlterDefaultPrivilegesTargetObjects()
			for _, objectType := range objectTypes {
				__antithesis_instrumentation__.Notify(557993)
				privs, ok := defaultPrivilegesForRole.DefaultPrivilegesPerObject[objectType]
				if !ok || func() bool {
					__antithesis_instrumentation__.Notify(557999)
					return len(privs.Users) == 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(558000)

					if objectType == tree.Types {
						__antithesis_instrumentation__.Notify(558001)

						if (!defaultPrivilegesForRole.IsExplicitRole() || func() bool {
							__antithesis_instrumentation__.Notify(558002)
							return catprivilege.GetRoleHasAllPrivilegesOnTargetObject(&defaultPrivilegesForRole, tree.Types) == true
						}() == true) && func() bool {
							__antithesis_instrumentation__.Notify(558003)
							return catprivilege.GetPublicHasUsageOnTypes(&defaultPrivilegesForRole) == true
						}() == true {
							__antithesis_instrumentation__.Notify(558004)
							continue
						} else {
							__antithesis_instrumentation__.Notify(558005)
						}
					} else {
						__antithesis_instrumentation__.Notify(558006)
						if !defaultPrivilegesForRole.IsExplicitRole() || func() bool {
							__antithesis_instrumentation__.Notify(558007)
							return catprivilege.GetRoleHasAllPrivilegesOnTargetObject(&defaultPrivilegesForRole, objectType) == true
						}() == true {
							__antithesis_instrumentation__.Notify(558008)
							continue
						} else {
							__antithesis_instrumentation__.Notify(558009)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(558010)
				}
				__antithesis_instrumentation__.Notify(557994)

				var c string
				switch objectType {
				case tree.Tables:
					__antithesis_instrumentation__.Notify(558011)
					c = "r"
				case tree.Sequences:
					__antithesis_instrumentation__.Notify(558012)
					c = "S"
				case tree.Types:
					__antithesis_instrumentation__.Notify(558013)
					c = "T"
				case tree.Schemas:
					__antithesis_instrumentation__.Notify(558014)
					c = "n"
				default:
					__antithesis_instrumentation__.Notify(558015)
				}
				__antithesis_instrumentation__.Notify(557995)
				privilegeObjectType := targetObjectToPrivilegeObject[objectType]
				arr := tree.NewDArray(types.String)
				for _, userPrivs := range privs.Users {
					__antithesis_instrumentation__.Notify(558016)
					var user string
					if userPrivs.UserProto.Decode().IsPublicRole() {
						__antithesis_instrumentation__.Notify(558018)

						user = ""
					} else {
						__antithesis_instrumentation__.Notify(558019)
						user = userPrivs.UserProto.Decode().Normalized()
					}
					__antithesis_instrumentation__.Notify(558017)

					privileges := privilege.ListFromBitField(
						userPrivs.Privileges, privilegeObjectType,
					)
					grantOptions := privilege.ListFromBitField(
						userPrivs.WithGrantOption, privilegeObjectType,
					)
					defaclItem := createDefACLItem(user, privileges, grantOptions, privilegeObjectType)
					if err := arr.Append(
						tree.NewDString(defaclItem)); err != nil {
						__antithesis_instrumentation__.Notify(558020)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558021)
					}
				}
				__antithesis_instrumentation__.Notify(557996)

				if defaultPrivilegesForRole.IsExplicitRole() {
					__antithesis_instrumentation__.Notify(558022)
					if objectType == tree.Types {
						__antithesis_instrumentation__.Notify(558023)
						if !catprivilege.GetRoleHasAllPrivilegesOnTargetObject(&defaultPrivilegesForRole, tree.Types) && func() bool {
							__antithesis_instrumentation__.Notify(558025)
							return catprivilege.GetPublicHasUsageOnTypes(&defaultPrivilegesForRole) == true
						}() == true {
							__antithesis_instrumentation__.Notify(558026)
							defaclItem := createDefACLItem(
								"", privilege.List{privilege.USAGE}, privilege.List{}, privilegeObjectType,
							)
							if err := arr.Append(tree.NewDString(defaclItem)); err != nil {
								__antithesis_instrumentation__.Notify(558027)
								return err
							} else {
								__antithesis_instrumentation__.Notify(558028)
							}
						} else {
							__antithesis_instrumentation__.Notify(558029)
						}
						__antithesis_instrumentation__.Notify(558024)
						if !catprivilege.GetPublicHasUsageOnTypes(&defaultPrivilegesForRole) && func() bool {
							__antithesis_instrumentation__.Notify(558030)
							return defaultPrivilegesForRole.GetExplicitRole().RoleHasAllPrivilegesOnTypes == true
						}() == true {
							__antithesis_instrumentation__.Notify(558031)
							defaclItem := createDefACLItem(
								defaultPrivilegesForRole.GetExplicitRole().UserProto.Decode().Normalized(),
								privilege.List{privilege.ALL}, privilege.List{}, privilegeObjectType,
							)
							if err := arr.Append(tree.NewDString(defaclItem)); err != nil {
								__antithesis_instrumentation__.Notify(558032)
								return err
							} else {
								__antithesis_instrumentation__.Notify(558033)
							}
						} else {
							__antithesis_instrumentation__.Notify(558034)
						}
					} else {
						__antithesis_instrumentation__.Notify(558035)
					}
				} else {
					__antithesis_instrumentation__.Notify(558036)
				}
				__antithesis_instrumentation__.Notify(557997)

				schemaName := ""

				normalizedName := ""
				roleOid := oidZero
				if defaultPrivilegesForRole.IsExplicitRole() {
					__antithesis_instrumentation__.Notify(558037)
					roleOid = h.UserOid(defaultPrivilegesForRole.GetExplicitRole().UserProto.Decode())
					normalizedName = defaultPrivilegesForRole.GetExplicitRole().UserProto.Decode().Normalized()
				} else {
					__antithesis_instrumentation__.Notify(558038)
				}
				__antithesis_instrumentation__.Notify(557998)
				rowOid := h.DBSchemaRoleOid(
					dbContext.GetID(),
					schemaName,
					normalizedName,
				)
				if err := addRow(
					rowOid,
					roleOid,
					oidZero,
					tree.NewDString(c),
					arr,
				); err != nil {
					__antithesis_instrumentation__.Notify(558039)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558040)
				}
			}
			__antithesis_instrumentation__.Notify(557992)
			return nil
		}
		__antithesis_instrumentation__.Notify(557990)
		return dbContext.GetDefaultPrivilegeDescriptor().ForEachDefaultPrivilegeForRole(f)
	},
}

func createDefACLItem(
	user string,
	privileges privilege.List,
	grantOptions privilege.List,
	privilegeObjectType privilege.ObjectType,
) string {
	__antithesis_instrumentation__.Notify(558041)
	return fmt.Sprintf(`%s=%s/%s`,
		user,
		privileges.ListToACL(
			grantOptions,
			privilegeObjectType,
		),

		"",
	)
}

var (
	depTypeNormal        = tree.NewDString("n")
	depTypeAuto          = tree.NewDString("a")
	depTypeInternal      = tree.NewDString("i")
	depTypeExtension     = tree.NewDString("e")
	depTypeAutoExtension = tree.NewDString("x")
	depTypePin           = tree.NewDString("p")

	_ = depTypeAuto
	_ = depTypeInternal
	_ = depTypeExtension
	_ = depTypeAutoExtension
	_ = depTypePin

	pgAuthIDTableName      = tree.MakeTableNameWithSchema("", tree.Name(pgCatalogName), tree.Name("pg_authid"))
	pgConstraintsTableName = tree.MakeTableNameWithSchema("", tree.Name(pgCatalogName), tree.Name("pg_constraint"))
	pgClassTableName       = tree.MakeTableNameWithSchema("", tree.Name(pgCatalogName), tree.Name("pg_class"))
	pgDatabaseTableName    = tree.MakeTableNameWithSchema("", tree.Name(pgCatalogName), tree.Name("pg_database"))
	pgRewriteTableName     = tree.MakeTableNameWithSchema("", tree.Name(pgCatalogName), tree.Name("pg_rewrite"))
)

var pgCatalogDependTable = virtualSchemaTable{
	comment: `dependency relationships (incomplete)
https://www.postgresql.org/docs/9.5/catalog-pg-depend.html`,
	schema: vtable.PGCatalogDepend,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558042)
		vt := p.getVirtualTabler()
		pgConstraintsDesc, err := vt.getVirtualTableDesc(&pgConstraintsTableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(558046)
			return errors.New("could not find pg_catalog.pg_constraint")
		} else {
			__antithesis_instrumentation__.Notify(558047)
		}
		__antithesis_instrumentation__.Notify(558043)
		pgClassDesc, err := vt.getVirtualTableDesc(&pgClassTableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(558048)
			return errors.New("could not find pg_catalog.pg_class")
		} else {
			__antithesis_instrumentation__.Notify(558049)
		}
		__antithesis_instrumentation__.Notify(558044)
		pgRewriteDesc, err := vt.getVirtualTableDesc(&pgRewriteTableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(558050)
			return errors.New("could not find pg_catalog.pg_rewrite")
		} else {
			__antithesis_instrumentation__.Notify(558051)
		}
		__antithesis_instrumentation__.Notify(558045)
		h := makeOidHasher()
		return forEachTableDescWithTableLookup(ctx, p, dbContext, hideVirtual, func(
			db catalog.DatabaseDescriptor,
			scName string,
			table catalog.TableDescriptor,
			tableLookup tableLookupFn,
		) error {
			__antithesis_instrumentation__.Notify(558052)
			pgConstraintTableOid := tableOid(pgConstraintsDesc.GetID())
			pgClassTableOid := tableOid(pgClassDesc.GetID())
			pgRewriteTableOid := tableOid(pgRewriteDesc.GetID())
			if table.IsSequence() && func() bool {
				__antithesis_instrumentation__.Notify(558058)
				return !table.GetSequenceOpts().SequenceOwner.Equal(descpb.TableDescriptor_SequenceOpts_SequenceOwner{}) == true
			}() == true {
				__antithesis_instrumentation__.Notify(558059)
				refObjID := tableOid(table.GetSequenceOpts().SequenceOwner.OwnerTableID)
				refObjSubID := tree.NewDInt(tree.DInt(table.GetSequenceOpts().SequenceOwner.OwnerColumnID))
				objID := tableOid(table.GetID())
				return addRow(
					pgConstraintTableOid,
					objID,
					zeroVal,
					pgClassTableOid,
					refObjID,
					refObjSubID,
					depTypeAuto,
				)
			} else {
				__antithesis_instrumentation__.Notify(558060)
			}
			__antithesis_instrumentation__.Notify(558053)

			reportViewDependency := func(dep *descpb.TableDescriptor_Reference) error {
				__antithesis_instrumentation__.Notify(558061)
				refObjOid := tableOid(table.GetID())
				objID := h.rewriteOid(table.GetID(), dep.ID)
				for _, colID := range dep.ColumnIDs {
					__antithesis_instrumentation__.Notify(558063)
					if err := addRow(
						pgRewriteTableOid,
						objID,
						zeroVal,
						pgClassTableOid,
						refObjOid,
						tree.NewDInt(tree.DInt(colID)),
						depTypeNormal,
					); err != nil {
						__antithesis_instrumentation__.Notify(558064)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558065)
					}
				}
				__antithesis_instrumentation__.Notify(558062)

				return nil
			}
			__antithesis_instrumentation__.Notify(558054)

			if table.IsTable() || func() bool {
				__antithesis_instrumentation__.Notify(558066)
				return table.IsView() == true
			}() == true {
				__antithesis_instrumentation__.Notify(558067)
				if err := table.ForeachDependedOnBy(reportViewDependency); err != nil {
					__antithesis_instrumentation__.Notify(558068)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558069)
				}
			} else {
				__antithesis_instrumentation__.Notify(558070)
			}
			__antithesis_instrumentation__.Notify(558055)

			conInfo, err := table.GetConstraintInfoWithLookup(tableLookup.getTableByID)
			if err != nil {
				__antithesis_instrumentation__.Notify(558071)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558072)
			}
			__antithesis_instrumentation__.Notify(558056)
			for _, con := range conInfo {
				__antithesis_instrumentation__.Notify(558073)
				if con.Kind != descpb.ConstraintTypeFK {
					__antithesis_instrumentation__.Notify(558077)
					continue
				} else {
					__antithesis_instrumentation__.Notify(558078)
				}
				__antithesis_instrumentation__.Notify(558074)

				referencedTable, err := tableLookup.getTableByID(con.FK.ReferencedTableID)
				if err != nil {
					__antithesis_instrumentation__.Notify(558079)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558080)
				}
				__antithesis_instrumentation__.Notify(558075)
				refObjID := oidZero
				if refConstraint, err := tabledesc.FindFKReferencedUniqueConstraint(
					referencedTable, con.FK.ReferencedColumnIDs,
				); err != nil {
					__antithesis_instrumentation__.Notify(558081)

					log.Warningf(ctx, "broken fk reference: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(558082)
					if idx, ok := refConstraint.(*descpb.IndexDescriptor); ok {
						__antithesis_instrumentation__.Notify(558083)
						refObjID = h.IndexOid(con.ReferencedTable.ID, idx.ID)
					} else {
						__antithesis_instrumentation__.Notify(558084)
					}
				}
				__antithesis_instrumentation__.Notify(558076)
				constraintOid := h.ForeignKeyConstraintOid(db.GetID(), scName, table.GetID(), con.FK)

				if err := addRow(
					pgConstraintTableOid,
					constraintOid,
					zeroVal,
					pgClassTableOid,
					refObjID,
					zeroVal,
					depTypeNormal,
				); err != nil {
					__antithesis_instrumentation__.Notify(558085)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558086)
				}
			}
			__antithesis_instrumentation__.Notify(558057)
			return nil
		})
	},
}

func getComments(ctx context.Context, p *planner) ([]tree.Datums, error) {
	__antithesis_instrumentation__.Notify(558087)
	return p.extendedEvalCtx.ExecCfg.InternalExecutor.QueryBuffered(
		ctx,
		"select-comments",
		p.EvalContext().Txn,
		`SELECT COALESCE(pc.object_id, sc.object_id) AS object_id,
              COALESCE(pc.sub_id, sc.sub_id) AS sub_id,
              COALESCE(pc.comment, sc.comment) AS comment,
              COALESCE(pc.type, sc.type) AS type
         FROM (SELECT * FROM system.comments) AS sc
    FULL JOIN (SELECT * FROM crdb_internal.predefined_comments) AS pc
           ON (pc.object_id = sc.object_id AND pc.sub_id = sc.sub_id AND pc.type = sc.type)`)
}

var pgCatalogDescriptionTable = virtualSchemaTable{
	comment: `object comments
https://www.postgresql.org/docs/9.5/catalog-pg-description.html`,
	schema: vtable.PGCatalogDescription,
	populate: func(
		ctx context.Context,
		p *planner,
		dbContext catalog.DatabaseDescriptor,
		addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558088)

		comments, err := getComments(ctx, p)
		if err != nil {
			__antithesis_instrumentation__.Notify(558091)
			return err
		} else {
			__antithesis_instrumentation__.Notify(558092)
		}
		__antithesis_instrumentation__.Notify(558089)
		for _, comment := range comments {
			__antithesis_instrumentation__.Notify(558093)
			objID := comment[0]
			objSubID := comment[1]
			description := comment[2]
			commentType := keys.CommentType(tree.MustBeDInt(comment[3]))

			classOid := oidZero

			switch commentType {
			case keys.DatabaseCommentType:
				__antithesis_instrumentation__.Notify(558095)

				continue
			case keys.SchemaCommentType:
				__antithesis_instrumentation__.Notify(558096)
				objID = tree.NewDOid(tree.MustBeDInt(objID))
				classOid = tree.NewDOid(catconstants.PgCatalogNamespaceTableID)
			case keys.ColumnCommentType, keys.TableCommentType:
				__antithesis_instrumentation__.Notify(558097)
				objID = tree.NewDOid(tree.MustBeDInt(objID))
				classOid = tree.NewDOid(catconstants.PgCatalogClassTableID)
			case keys.ConstraintCommentType:
				__antithesis_instrumentation__.Notify(558098)
				tableDesc, err := p.Descriptors().GetImmutableTableByID(
					ctx,
					p.txn,
					descpb.ID(tree.MustBeDInt(objID)),
					tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveRequireTableDesc))
				if err != nil {
					__antithesis_instrumentation__.Notify(558105)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558106)
				}
				__antithesis_instrumentation__.Notify(558099)
				schema, err := p.Descriptors().GetImmutableSchemaByID(
					ctx,
					p.txn,
					tableDesc.GetParentSchemaID(),
					tree.CommonLookupFlags{
						Required: true,
					})
				if err != nil {
					__antithesis_instrumentation__.Notify(558107)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558108)
				}
				__antithesis_instrumentation__.Notify(558100)
				constraints, err := tableDesc.GetConstraintInfo()
				if err != nil {
					__antithesis_instrumentation__.Notify(558109)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558110)
				}
				__antithesis_instrumentation__.Notify(558101)
				var constraint descpb.ConstraintDetail
				for _, constraintToCheck := range constraints {
					__antithesis_instrumentation__.Notify(558111)
					if constraintToCheck.ConstraintID == descpb.ConstraintID(tree.MustBeDInt(objSubID)) {
						__antithesis_instrumentation__.Notify(558112)
						constraint = constraintToCheck
						break
					} else {
						__antithesis_instrumentation__.Notify(558113)
					}
				}
				__antithesis_instrumentation__.Notify(558102)
				objID = getOIDFromConstraint(constraint, dbContext.GetID(), schema.GetName(), tableDesc)
				objSubID = tree.DZero
				classOid = tree.NewDOid(catconstants.PgCatalogConstraintTableID)
			case keys.IndexCommentType:
				__antithesis_instrumentation__.Notify(558103)
				objID = makeOidHasher().IndexOid(
					descpb.ID(tree.MustBeDInt(objID)),
					descpb.IndexID(tree.MustBeDInt(objSubID)))
				objSubID = tree.DZero
				classOid = tree.NewDOid(catconstants.PgCatalogClassTableID)
			default:
				__antithesis_instrumentation__.Notify(558104)
			}
			__antithesis_instrumentation__.Notify(558094)
			if err := addRow(
				objID,
				classOid,
				objSubID,
				description); err != nil {
				__antithesis_instrumentation__.Notify(558114)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558115)
			}
		}
		__antithesis_instrumentation__.Notify(558090)
		return nil
	},
}

func getOIDFromConstraint(
	constraint descpb.ConstraintDetail,
	dbID descpb.ID,
	schemaName string,
	tableDesc catalog.TableDescriptor,
) *tree.DOid {
	__antithesis_instrumentation__.Notify(558116)
	hasher := makeOidHasher()
	tableID := tableDesc.GetID()
	var oid *tree.DOid
	if constraint.CheckConstraint != nil {
		__antithesis_instrumentation__.Notify(558118)
		oid = hasher.CheckConstraintOid(
			dbID,
			schemaName,
			tableID,
			constraint.CheckConstraint)
	} else {
		__antithesis_instrumentation__.Notify(558119)
		if constraint.FK != nil {
			__antithesis_instrumentation__.Notify(558120)
			oid = hasher.ForeignKeyConstraintOid(
				dbID,
				schemaName,
				tableID,
				constraint.FK,
			)
		} else {
			__antithesis_instrumentation__.Notify(558121)
			if constraint.UniqueWithoutIndexConstraint != nil {
				__antithesis_instrumentation__.Notify(558122)
				oid = hasher.UniqueWithoutIndexConstraintOid(
					dbID,
					schemaName,
					tableID,
					constraint.UniqueWithoutIndexConstraint,
				)
			} else {
				__antithesis_instrumentation__.Notify(558123)
				if constraint.Index != nil {
					__antithesis_instrumentation__.Notify(558124)
					if constraint.Index.ID == tableDesc.GetPrimaryIndexID() {
						__antithesis_instrumentation__.Notify(558125)
						oid = hasher.PrimaryKeyConstraintOid(
							dbID,
							schemaName,
							tableID,
							constraint.Index,
						)
					} else {
						__antithesis_instrumentation__.Notify(558126)
						oid = hasher.UniqueConstraintOid(
							dbID,
							schemaName,
							tableID,
							constraint.Index.ID,
						)
					}
				} else {
					__antithesis_instrumentation__.Notify(558127)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(558117)
	return oid
}

var pgCatalogSharedDescriptionTable = virtualSchemaTable{
	comment: `shared object comments
https://www.postgresql.org/docs/9.5/catalog-pg-shdescription.html`,
	schema: vtable.PGCatalogSharedDescription,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558128)

		comments, err := getComments(ctx, p)
		if err != nil {
			__antithesis_instrumentation__.Notify(558131)
			return err
		} else {
			__antithesis_instrumentation__.Notify(558132)
		}
		__antithesis_instrumentation__.Notify(558129)
		for _, comment := range comments {
			__antithesis_instrumentation__.Notify(558133)
			commentType := keys.CommentType(tree.MustBeDInt(comment[3]))
			if commentType != keys.DatabaseCommentType {
				__antithesis_instrumentation__.Notify(558135)

				continue
			} else {
				__antithesis_instrumentation__.Notify(558136)
			}
			__antithesis_instrumentation__.Notify(558134)
			classOid := tree.NewDOid(catconstants.PgCatalogDatabaseTableID)
			objID := descpb.ID(tree.MustBeDInt(comment[0]))
			if err := addRow(
				tableOid(objID),
				classOid,
				comment[2]); err != nil {
				__antithesis_instrumentation__.Notify(558137)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558138)
			}
		}
		__antithesis_instrumentation__.Notify(558130)
		return nil
	},
}

var pgCatalogEnumTable = virtualSchemaTable{
	comment: `enum types and labels (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-enum.html`,
	schema: vtable.PGCatalogEnum,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558139)
		h := makeOidHasher()

		return forEachTypeDesc(ctx, p, dbContext, func(_ catalog.DatabaseDescriptor, _ string, typDesc catalog.TypeDescriptor) error {
			__antithesis_instrumentation__.Notify(558140)
			switch typDesc.GetKind() {
			case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
				__antithesis_instrumentation__.Notify(558143)

			default:
				__antithesis_instrumentation__.Notify(558144)
				return nil
			}
			__antithesis_instrumentation__.Notify(558141)

			typOID := tree.NewDOid(tree.DInt(typedesc.TypeIDToOID(typDesc.GetID())))
			for i := 0; i < typDesc.NumEnumMembers(); i++ {
				__antithesis_instrumentation__.Notify(558145)
				if err := addRow(
					h.EnumEntryOid(typOID, typDesc.GetMemberPhysicalRepresentation(i)),
					typOID,
					tree.NewDFloat(tree.DFloat(float64(i))),
					tree.NewDString(typDesc.GetMemberLogicalRepresentation(i)),
				); err != nil {
					__antithesis_instrumentation__.Notify(558146)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558147)
				}
			}
			__antithesis_instrumentation__.Notify(558142)
			return nil
		})
	},
}

var pgCatalogEventTriggerTable = virtualSchemaTable{
	comment: `event triggers (empty - feature does not exist)
https://www.postgresql.org/docs/9.6/catalog-pg-event-trigger.html`,
	schema: vtable.PGCatalogEventTrigger,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558148)

		return nil
	},
	unimplemented: true,
}

var pgCatalogExtensionTable = virtualSchemaTable{
	comment: `installed extensions (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-extension.html`,
	schema: vtable.PGCatalogExtension,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558149)

		return nil
	},
	unimplemented: true,
}

var pgCatalogForeignDataWrapperTable = virtualSchemaTable{
	comment: `foreign data wrappers (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-foreign-data-wrapper.html`,
	schema: vtable.PGCatalogForeignDataWrapper,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558150)

		return nil
	},
	unimplemented: true,
}

var pgCatalogForeignServerTable = virtualSchemaTable{
	comment: `foreign servers (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-foreign-server.html`,
	schema: vtable.PGCatalogForeignServer,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558151)

		return nil
	},
	unimplemented: true,
}

var pgCatalogForeignTableTable = virtualSchemaTable{
	comment: `foreign tables (empty  - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-foreign-table.html`,
	schema: vtable.PGCatalogForeignTable,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558152)

		return nil
	},
	unimplemented: true,
}

func makeZeroedOidVector(size int) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(558153)
	oidArray := tree.NewDArray(types.Oid)
	for i := 0; i < size; i++ {
		__antithesis_instrumentation__.Notify(558155)
		if err := oidArray.Append(oidZero); err != nil {
			__antithesis_instrumentation__.Notify(558156)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(558157)
		}
	}
	__antithesis_instrumentation__.Notify(558154)
	return tree.NewDOidVectorFromDArray(oidArray), nil
}

var pgCatalogIndexTable = virtualSchemaTable{
	comment: `indexes (incomplete)
https://www.postgresql.org/docs/9.5/catalog-pg-index.html`,
	schema: vtable.PGCatalogIndex,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558158)
		h := makeOidHasher()
		return forEachTableDesc(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(558159)
				tableOid := tableOid(table.GetID())
				return catalog.ForEachIndex(table, catalog.IndexOpts{}, func(index catalog.Index) error {
					__antithesis_instrumentation__.Notify(558160)
					isMutation, isWriteOnly :=
						table.GetIndexMutationCapabilities(index.GetID())
					isReady := isMutation && func() bool {
						__antithesis_instrumentation__.Notify(558167)
						return isWriteOnly == true
					}() == true

					collationOids := tree.NewDArray(types.Oid)
					indoption := tree.NewDArray(types.Int)

					colIDs := make([]descpb.ColumnID, 0, index.NumKeyColumns())
					exprs := make([]string, 0, index.NumKeyColumns())
					for i := index.IndexDesc().ExplicitColumnStartIdx(); i < index.NumKeyColumns(); i++ {
						__antithesis_instrumentation__.Notify(558168)
						columnID := index.GetKeyColumnID(i)
						col, err := table.FindColumnWithID(columnID)
						if err != nil {
							__antithesis_instrumentation__.Notify(558173)
							return err
						} else {
							__antithesis_instrumentation__.Notify(558174)
						}
						__antithesis_instrumentation__.Notify(558169)

						if col.IsExpressionIndexColumn() {
							__antithesis_instrumentation__.Notify(558175)
							colIDs = append(colIDs, 0)
							formattedExpr, err := schemaexpr.FormatExprForDisplay(
								ctx, table, col.GetComputeExpr(), p.SemaCtx(), p.SessionData(), tree.FmtPGCatalog,
							)
							if err != nil {
								__antithesis_instrumentation__.Notify(558177)
								return err
							} else {
								__antithesis_instrumentation__.Notify(558178)
							}
							__antithesis_instrumentation__.Notify(558176)
							exprs = append(exprs, fmt.Sprintf("(%s)", formattedExpr))
						} else {
							__antithesis_instrumentation__.Notify(558179)
							colIDs = append(colIDs, columnID)
						}
						__antithesis_instrumentation__.Notify(558170)
						if err := collationOids.Append(typColl(col.GetType(), h)); err != nil {
							__antithesis_instrumentation__.Notify(558180)
							return err
						} else {
							__antithesis_instrumentation__.Notify(558181)
						}
						__antithesis_instrumentation__.Notify(558171)

						var thisIndOption tree.DInt
						if index.GetKeyColumnDirection(i) == descpb.IndexDescriptor_ASC {
							__antithesis_instrumentation__.Notify(558182)
							thisIndOption = indoptionNullsFirst
						} else {
							__antithesis_instrumentation__.Notify(558183)
							thisIndOption = indoptionDesc
						}
						__antithesis_instrumentation__.Notify(558172)
						if err := indoption.Append(tree.NewDInt(thisIndOption)); err != nil {
							__antithesis_instrumentation__.Notify(558184)
							return err
						} else {
							__antithesis_instrumentation__.Notify(558185)
						}
					}
					__antithesis_instrumentation__.Notify(558161)

					indnkeyatts := len(colIDs)
					for i := 0; i < index.NumSecondaryStoredColumns(); i++ {
						__antithesis_instrumentation__.Notify(558186)
						colIDs = append(colIDs, index.GetStoredColumnID(i))
					}
					__antithesis_instrumentation__.Notify(558162)

					indnatts := len(colIDs)
					indkey, err := colIDArrayToVector(colIDs)
					if err != nil {
						__antithesis_instrumentation__.Notify(558187)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558188)
					}
					__antithesis_instrumentation__.Notify(558163)
					collationOidVector := tree.NewDOidVectorFromDArray(collationOids)
					indoptionIntVector := tree.NewDIntVectorFromDArray(indoption)

					indclass, err := makeZeroedOidVector(indnkeyatts)
					if err != nil {
						__antithesis_instrumentation__.Notify(558189)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558190)
					}
					__antithesis_instrumentation__.Notify(558164)
					indpred := tree.DNull
					if index.IsPartial() {
						__antithesis_instrumentation__.Notify(558191)
						formattedPred, err := schemaexpr.FormatExprForDisplay(
							ctx, table, index.GetPredicate(), p.SemaCtx(), p.SessionData(), tree.FmtPGCatalog,
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(558193)
							return err
						} else {
							__antithesis_instrumentation__.Notify(558194)
						}
						__antithesis_instrumentation__.Notify(558192)
						indpred = tree.NewDString(formattedPred)
					} else {
						__antithesis_instrumentation__.Notify(558195)
					}
					__antithesis_instrumentation__.Notify(558165)
					indexprs := tree.DNull
					if len(exprs) > 0 {
						__antithesis_instrumentation__.Notify(558196)
						indexprs = tree.NewDString(strings.Join(exprs, " "))
					} else {
						__antithesis_instrumentation__.Notify(558197)
					}
					__antithesis_instrumentation__.Notify(558166)
					return addRow(
						h.IndexOid(table.GetID(), index.GetID()),
						tableOid,
						tree.NewDInt(tree.DInt(indnatts)),
						tree.MakeDBool(tree.DBool(index.IsUnique())),
						tree.MakeDBool(tree.DBool(index.Primary())),
						tree.DBoolFalse,
						tree.MakeDBool(tree.DBool(index.IsUnique())),
						tree.DBoolFalse,
						tree.MakeDBool(tree.DBool(!isMutation)),
						tree.DBoolFalse,
						tree.MakeDBool(tree.DBool(isReady)),
						tree.DBoolTrue,
						tree.DBoolFalse,
						indkey,
						collationOidVector,
						indclass,
						indoptionIntVector,
						indexprs,
						indpred,
						tree.NewDInt(tree.DInt(indnkeyatts)),
					)
				})
			})
	},
}

var pgCatalogIndexesTable = virtualSchemaTable{
	comment: `index creation statements
https://www.postgresql.org/docs/9.5/view-pg-indexes.html`,
	schema: vtable.PGCatalogIndexes,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558198)
		h := makeOidHasher()
		return forEachTableDescWithTableLookup(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor, _ tableLookupFn) error {
				__antithesis_instrumentation__.Notify(558199)
				scNameName := tree.NewDName(scName)
				tblName := tree.NewDName(table.GetName())
				return catalog.ForEachIndex(table, catalog.IndexOpts{}, func(index catalog.Index) error {
					__antithesis_instrumentation__.Notify(558200)
					def, err := indexDefFromDescriptor(ctx, p, db, scName, table, index)
					if err != nil {
						__antithesis_instrumentation__.Notify(558202)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558203)
					}
					__antithesis_instrumentation__.Notify(558201)
					return addRow(
						h.IndexOid(table.GetID(), index.GetID()),
						scNameName,
						tblName,
						tree.NewDName(index.GetName()),
						tree.DNull,
						tree.NewDString(def),
					)
				})
			})
	},
}

func indexDefFromDescriptor(
	ctx context.Context,
	p *planner,
	db catalog.DatabaseDescriptor,
	schemaName string,
	table catalog.TableDescriptor,
	index catalog.Index,
) (string, error) {
	__antithesis_instrumentation__.Notify(558204)
	tableName := tree.MakeTableNameWithSchema(tree.Name(db.GetName()), tree.Name(schemaName), tree.Name(table.GetName()))
	partitionStr := ""
	fmtStr, err := catformat.IndexForDisplay(
		ctx,
		table,
		&tableName,
		index,
		partitionStr,
		tree.FmtPGCatalog,
		p.SemaCtx(),
		p.SessionData(),
		catformat.IndexDisplayShowCreate,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(558206)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(558207)
	}
	__antithesis_instrumentation__.Notify(558205)
	return fmtStr, nil
}

var pgCatalogInheritsTable = virtualSchemaTable{
	comment: `table inheritance hierarchy (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-inherits.html`,
	schema: vtable.PGCatalogInherits,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558208)

		return nil
	},
	unimplemented: true,
}

var pgCatalogLanguageTable = virtualSchemaTable{
	comment: `available languages (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-language.html`,
	schema: vtable.PGCatalogLanguage,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558209)

		return nil
	},
	unimplemented: true,
}

var pgCatalogLocksTable = virtualSchemaTable{
	comment: `locks held by active processes (empty - feature does not exist)
https://www.postgresql.org/docs/9.6/view-pg-locks.html`,
	schema: vtable.PGCatalogLocks,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558210)
		return nil
	},
	unimplemented: true,
}

var pgCatalogMatViewsTable = virtualSchemaTable{
	comment: `available materialized views (empty - feature does not exist)
https://www.postgresql.org/docs/9.6/view-pg-matviews.html`,
	schema: vtable.PGCatalogMatViews,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558211)
		return forEachTableDesc(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, desc catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(558212)
				if !desc.MaterializedView() {
					__antithesis_instrumentation__.Notify(558214)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(558215)
				}
				__antithesis_instrumentation__.Notify(558213)

				return addRow(
					tree.NewDName(scName),
					tree.NewDName(desc.GetName()),
					getOwnerName(desc),
					tree.DNull,
					tree.MakeDBool(len(desc.PublicNonPrimaryIndexes()) > 0),
					tree.DBoolTrue,
					tree.NewDString(desc.GetViewQuery()),
				)
			})
	},
}

var pgCatalogNamespaceTable = virtualSchemaTable{
	comment: `available namespaces (incomplete; namespaces and databases are congruent in CockroachDB)
https://www.postgresql.org/docs/9.5/catalog-pg-namespace.html`,
	schema: vtable.PGCatalogNamespace,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558216)
		h := makeOidHasher()
		return forEachDatabaseDesc(ctx, p, dbContext, true,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(558217)
				return forEachSchema(ctx, p, db, func(sc catalog.SchemaDescriptor) error {
					__antithesis_instrumentation__.Notify(558218)
					ownerOID := tree.DNull
					if sc.SchemaKind() == catalog.SchemaUserDefined {
						__antithesis_instrumentation__.Notify(558220)
						ownerOID = getOwnerOID(sc)
					} else {
						__antithesis_instrumentation__.Notify(558221)
						if sc.SchemaKind() == catalog.SchemaPublic {
							__antithesis_instrumentation__.Notify(558222)

							ownerOID = h.UserOid(security.MakeSQLUsernameFromPreNormalizedString("admin"))
						} else {
							__antithesis_instrumentation__.Notify(558223)
						}
					}
					__antithesis_instrumentation__.Notify(558219)
					return addRow(
						h.NamespaceOid(db.GetID(), sc.GetName()),
						tree.NewDString(sc.GetName()),
						ownerOID,
						tree.DNull,
					)
				})
			})
	},
}

var (
	infixKind   = tree.NewDString("b")
	prefixKind  = tree.NewDString("l")
	postfixKind = tree.NewDString("r")

	_ = postfixKind
)

var pgCatalogOpclassTable = virtualSchemaTable{
	comment: `opclass (empty - Operator classes not supported yet)
https://www.postgresql.org/docs/12/catalog-pg-opclass.html`,
	schema: vtable.PGCatalogOpclass,
	populate: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558224)
		return nil
	},
	unimplemented: true,
}

var pgCatalogOperatorTable = virtualSchemaTable{
	comment: `operators (incomplete)
https://www.postgresql.org/docs/9.5/catalog-pg-operator.html`,
	schema: vtable.PGCatalogOperator,
	populate: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558225)
		h := makeOidHasher()
		nspOid := h.NamespaceOid(db.GetID(), pgCatalogName)
		addOp := func(opName string, kind tree.Datum, params tree.TypeList, returnTyper tree.ReturnTyper) error {
			__antithesis_instrumentation__.Notify(558230)
			var leftType, rightType *tree.DOid
			switch params.Length() {
			case 1:
				__antithesis_instrumentation__.Notify(558232)
				leftType = oidZero
				rightType = tree.NewDOid(tree.DInt(params.Types()[0].Oid()))
			case 2:
				__antithesis_instrumentation__.Notify(558233)
				leftType = tree.NewDOid(tree.DInt(params.Types()[0].Oid()))
				rightType = tree.NewDOid(tree.DInt(params.Types()[1].Oid()))
			default:
				__antithesis_instrumentation__.Notify(558234)
				panic(errors.AssertionFailedf("unexpected operator %s with %d params",
					opName, params.Length()))
			}
			__antithesis_instrumentation__.Notify(558231)
			returnType := tree.NewDOid(tree.DInt(returnTyper(nil).Oid()))
			err := addRow(
				h.OperatorOid(opName, leftType, rightType, returnType),

				tree.NewDString(opName),
				nspOid,
				tree.DNull,
				kind,
				tree.DBoolFalse,
				tree.DBoolFalse,
				leftType,
				rightType,
				returnType,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
			)
			return err
		}
		__antithesis_instrumentation__.Notify(558226)
		for cmpOp, overloads := range tree.CmpOps {
			__antithesis_instrumentation__.Notify(558235)

			if cmpOp == treecmp.In {
				__antithesis_instrumentation__.Notify(558237)
				continue
			} else {
				__antithesis_instrumentation__.Notify(558238)
			}
			__antithesis_instrumentation__.Notify(558236)
			for _, overload := range overloads {
				__antithesis_instrumentation__.Notify(558239)
				params, returnType := tree.GetParamsAndReturnType(overload)
				if err := addOp(cmpOp.String(), infixKind, params, returnType); err != nil {
					__antithesis_instrumentation__.Notify(558241)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558242)
				}
				__antithesis_instrumentation__.Notify(558240)
				if inverse, ok := tree.CmpOpInverse(cmpOp); ok {
					__antithesis_instrumentation__.Notify(558243)
					if err := addOp(inverse.String(), infixKind, params, returnType); err != nil {
						__antithesis_instrumentation__.Notify(558244)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558245)
					}
				} else {
					__antithesis_instrumentation__.Notify(558246)
				}
			}
		}
		__antithesis_instrumentation__.Notify(558227)
		for binOp, overloads := range tree.BinOps {
			__antithesis_instrumentation__.Notify(558247)
			for _, overload := range overloads {
				__antithesis_instrumentation__.Notify(558248)
				params, returnType := tree.GetParamsAndReturnType(overload)
				if err := addOp(binOp.String(), infixKind, params, returnType); err != nil {
					__antithesis_instrumentation__.Notify(558249)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558250)
				}
			}
		}
		__antithesis_instrumentation__.Notify(558228)
		for unaryOp, overloads := range tree.UnaryOps {
			__antithesis_instrumentation__.Notify(558251)
			for _, overload := range overloads {
				__antithesis_instrumentation__.Notify(558252)
				params, returnType := tree.GetParamsAndReturnType(overload)
				if err := addOp(unaryOp.String(), prefixKind, params, returnType); err != nil {
					__antithesis_instrumentation__.Notify(558253)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558254)
				}
			}
		}
		__antithesis_instrumentation__.Notify(558229)
		return nil
	},
}

func newSingletonStringArray(s string) tree.Datum {
	__antithesis_instrumentation__.Notify(558255)
	return &tree.DArray{ParamTyp: types.String, Array: tree.Datums{tree.NewDString(s)}}
}

var (
	proArgModeInOut    = newSingletonStringArray("b")
	proArgModeIn       = newSingletonStringArray("i")
	proArgModeOut      = newSingletonStringArray("o")
	proArgModeTable    = newSingletonStringArray("t")
	proArgModeVariadic = newSingletonStringArray("v")

	_ = proArgModeInOut
	_ = proArgModeIn
	_ = proArgModeOut
	_ = proArgModeTable
)

var pgCatalogPreparedXactsTable = virtualSchemaTable{
	comment: `prepared transactions (empty - feature does not exist)
https://www.postgresql.org/docs/9.6/view-pg-prepared-xacts.html`,
	schema: vtable.PGCatalogPreparedXacts,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558256)
		return nil
	},
	unimplemented: true,
}

var pgCatalogPreparedStatementsTable = virtualSchemaTable{
	comment: `prepared statements
https://www.postgresql.org/docs/9.6/view-pg-prepared-statements.html`,
	schema: vtable.PGCatalogPreparedStatements,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558257)
		for name, stmt := range p.preparedStatements.List() {
			__antithesis_instrumentation__.Notify(558259)
			placeholderTypes := stmt.PrepareMetadata.PlaceholderTypesInfo.Types
			paramTypes := tree.NewDArray(types.RegType)
			paramTypes.Array = make(tree.Datums, len(placeholderTypes))
			paramNames := make([]string, len(placeholderTypes))

			for i, placeholderType := range placeholderTypes {
				__antithesis_instrumentation__.Notify(558264)
				paramTypes.Array[i] = tree.NewDOidWithName(
					tree.DInt(placeholderType.Oid()),
					placeholderType,
					placeholderType.SQLStandardName(),
				)
				paramNames[i] = placeholderType.Name()
			}
			__antithesis_instrumentation__.Notify(558260)

			argumentsStr := ""
			if len(paramNames) > 0 {
				__antithesis_instrumentation__.Notify(558265)
				argumentsStr = fmt.Sprintf(" (%s)", strings.Join(paramNames, ", "))
			} else {
				__antithesis_instrumentation__.Notify(558266)
			}
			__antithesis_instrumentation__.Notify(558261)

			fromSQL := tree.DBoolFalse
			if stmt.origin == PreparedStatementOriginSQL {
				__antithesis_instrumentation__.Notify(558267)
				fromSQL = tree.DBoolTrue
			} else {
				__antithesis_instrumentation__.Notify(558268)
			}
			__antithesis_instrumentation__.Notify(558262)

			ts, err := tree.MakeDTimestampTZ(stmt.createdAt, time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(558269)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558270)
			}
			__antithesis_instrumentation__.Notify(558263)
			if err := addRow(
				tree.NewDString(name),
				tree.NewDString(fmt.Sprintf("PREPARE %s%s AS %s", name, argumentsStr, stmt.SQL)),
				ts,
				paramTypes,
				fromSQL,
			); err != nil {
				__antithesis_instrumentation__.Notify(558271)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558272)
			}
		}
		__antithesis_instrumentation__.Notify(558258)
		return nil
	},
}

var pgCatalogProcTable = virtualSchemaTable{
	comment: `built-in functions (incomplete)
https://www.postgresql.org/docs/9.5/catalog-pg-proc.html`,
	schema: vtable.PGCatalogProc,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558273)
		h := makeOidHasher()
		return forEachDatabaseDesc(ctx, p, dbContext, false,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(558274)
				nspOid := h.NamespaceOid(db.GetID(), pgCatalogName)
				for _, name := range builtins.AllBuiltinNames {
					__antithesis_instrumentation__.Notify(558276)

					var first rune
					for _, c := range name {
						__antithesis_instrumentation__.Notify(558279)
						first = c
						break
					}
					__antithesis_instrumentation__.Notify(558277)
					if unicode.IsUpper(first) {
						__antithesis_instrumentation__.Notify(558280)
						continue
					} else {
						__antithesis_instrumentation__.Notify(558281)
					}
					__antithesis_instrumentation__.Notify(558278)
					props, overloads := builtins.GetBuiltinProperties(name)
					isAggregate := props.Class == tree.AggregateClass
					isWindow := props.Class == tree.WindowClass
					for _, builtin := range overloads {
						__antithesis_instrumentation__.Notify(558282)
						dName := tree.NewDName(name)
						dSrc := tree.NewDString(name)

						var retType tree.Datum
						isRetSet := false
						if fixedRetType := builtin.FixedReturnType(); fixedRetType != nil {
							__antithesis_instrumentation__.Notify(558286)
							var retOid oid.Oid
							if fixedRetType.Family() == types.TupleFamily && func() bool {
								__antithesis_instrumentation__.Notify(558288)
								return builtin.IsGenerator() == true
							}() == true {
								__antithesis_instrumentation__.Notify(558289)
								isRetSet = true

								retOid = oid.T_anyelement
								if len(fixedRetType.TupleContents()) == 1 {
									__antithesis_instrumentation__.Notify(558290)

									retOid = fixedRetType.TupleContents()[0].Oid()
								} else {
									__antithesis_instrumentation__.Notify(558291)
								}
							} else {
								__antithesis_instrumentation__.Notify(558292)
								retOid = fixedRetType.Oid()
							}
							__antithesis_instrumentation__.Notify(558287)
							retType = tree.NewDOid(tree.DInt(retOid))
						} else {
							__antithesis_instrumentation__.Notify(558293)
						}
						__antithesis_instrumentation__.Notify(558283)

						argTypes := builtin.Types
						dArgTypes := tree.NewDArray(types.Oid)
						for _, argType := range argTypes.Types() {
							__antithesis_instrumentation__.Notify(558294)
							if err := dArgTypes.Append(tree.NewDOid(tree.DInt(argType.Oid()))); err != nil {
								__antithesis_instrumentation__.Notify(558295)
								return err
							} else {
								__antithesis_instrumentation__.Notify(558296)
							}
						}
						__antithesis_instrumentation__.Notify(558284)

						var argmodes tree.Datum
						var variadicType tree.Datum
						switch v := argTypes.(type) {
						case tree.VariadicType:
							__antithesis_instrumentation__.Notify(558297)
							if len(v.FixedTypes) == 0 {
								__antithesis_instrumentation__.Notify(558301)
								argmodes = proArgModeVariadic
							} else {
								__antithesis_instrumentation__.Notify(558302)
								ary := tree.NewDArray(types.String)
								for range v.FixedTypes {
									__antithesis_instrumentation__.Notify(558305)
									if err := ary.Append(tree.NewDString("i")); err != nil {
										__antithesis_instrumentation__.Notify(558306)
										return err
									} else {
										__antithesis_instrumentation__.Notify(558307)
									}
								}
								__antithesis_instrumentation__.Notify(558303)
								if err := ary.Append(tree.NewDString("v")); err != nil {
									__antithesis_instrumentation__.Notify(558308)
									return err
								} else {
									__antithesis_instrumentation__.Notify(558309)
								}
								__antithesis_instrumentation__.Notify(558304)
								argmodes = ary
							}
							__antithesis_instrumentation__.Notify(558298)
							variadicType = tree.NewDOid(tree.DInt(v.VarType.Oid()))
						case tree.HomogeneousType:
							__antithesis_instrumentation__.Notify(558299)
							argmodes = proArgModeVariadic
							argType := types.Any
							oid := argType.Oid()
							variadicType = tree.NewDOid(tree.DInt(oid))
						default:
							__antithesis_instrumentation__.Notify(558300)
							argmodes = tree.DNull
							variadicType = oidZero
						}
						__antithesis_instrumentation__.Notify(558285)
						provolatile, proleakproof := builtin.Volatility.ToPostgres()

						err := addRow(
							h.BuiltinOid(name, &builtin),
							dName,
							nspOid,
							tree.DNull,
							oidZero,
							tree.DNull,
							tree.DNull,
							variadicType,
							tree.DNull,
							tree.MakeDBool(tree.DBool(isAggregate)),
							tree.MakeDBool(tree.DBool(isWindow)),
							tree.DBoolFalse,
							tree.MakeDBool(tree.DBool(proleakproof)),
							tree.DBoolFalse,
							tree.MakeDBool(tree.DBool(isRetSet)),
							tree.NewDString(provolatile),
							tree.DNull,
							tree.NewDInt(tree.DInt(builtin.Types.Length())),
							tree.NewDInt(tree.DInt(0)),
							retType,
							tree.NewDOidVectorFromDArray(dArgTypes),
							tree.DNull,
							argmodes,
							tree.DNull,
							tree.DNull,
							tree.DNull,
							dSrc,
							tree.DNull,
							tree.DNull,
							tree.DNull,

							tree.DNull,
							tree.DNull,
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(558310)
							return err
						} else {
							__antithesis_instrumentation__.Notify(558311)
						}
					}
				}
				__antithesis_instrumentation__.Notify(558275)
				return nil
			})
	},
}

var pgCatalogRangeTable = virtualSchemaTable{
	comment: `range types (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-range.html`,
	schema: vtable.PGCatalogRange,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558312)

		return nil
	},
	unimplemented: true,
}

var pgCatalogRewriteTable = virtualSchemaTable{
	comment: `rewrite rules (only for referencing on pg_depend for table-view dependencies)
https://www.postgresql.org/docs/9.5/catalog-pg-rewrite.html`,
	schema: vtable.PGCatalogRewrite,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558313)
		h := makeOidHasher()
		ruleName := tree.NewDString("_RETURN")
		evType := tree.NewDString(string(evTypeSelect))
		return forEachTableDescWithTableLookup(ctx, p, dbContext, hideVirtual, func(
			db catalog.DatabaseDescriptor,
			scName string,
			table catalog.TableDescriptor,
			tableLookup tableLookupFn,
		) error {
			__antithesis_instrumentation__.Notify(558314)
			if !table.IsTable() && func() bool {
				__antithesis_instrumentation__.Notify(558316)
				return !table.IsView() == true
			}() == true {
				__antithesis_instrumentation__.Notify(558317)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(558318)
			}
			__antithesis_instrumentation__.Notify(558315)

			return table.ForeachDependedOnBy(func(dep *descpb.TableDescriptor_Reference) error {
				__antithesis_instrumentation__.Notify(558319)
				rewriteOid := h.rewriteOid(table.GetID(), dep.ID)
				evClass := tableOid(dep.ID)
				return addRow(
					rewriteOid,
					ruleName,
					evClass,
					evType,
					tree.DNull,
					tree.DBoolTrue,
					tree.DNull,
					tree.DNull,
				)
			})
		})
	},
}

var pgCatalogRolesTable = virtualSchemaTable{
	comment: `database roles
https://www.postgresql.org/docs/9.5/view-pg-roles.html`,
	schema: vtable.PGCatalogRoles,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558320)

		h := makeOidHasher()
		return forEachRole(ctx, p,
			func(username security.SQLUsername, isRole bool, options roleOptions, settings tree.Datum) error {
				__antithesis_instrumentation__.Notify(558321)
				isRoot := tree.DBool(username.IsRootUser() || func() bool {
					__antithesis_instrumentation__.Notify(558327)
					return username.IsAdminRole() == true
				}() == true)

				roleInherits := tree.DBool(true)
				noLogin, err := options.noLogin()
				if err != nil {
					__antithesis_instrumentation__.Notify(558328)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558329)
				}
				__antithesis_instrumentation__.Notify(558322)
				roleCanLogin := isRoot || func() bool {
					__antithesis_instrumentation__.Notify(558330)
					return !noLogin == true
				}() == true
				createDB, err := options.createDB()
				if err != nil {
					__antithesis_instrumentation__.Notify(558331)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558332)
				}
				__antithesis_instrumentation__.Notify(558323)
				rolValidUntil, err := options.validUntil(p)
				if err != nil {
					__antithesis_instrumentation__.Notify(558333)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558334)
				}
				__antithesis_instrumentation__.Notify(558324)
				createRole, err := options.createRole()
				if err != nil {
					__antithesis_instrumentation__.Notify(558335)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558336)
				}
				__antithesis_instrumentation__.Notify(558325)
				isSuper, err := userIsSuper(ctx, p, username)
				if err != nil {
					__antithesis_instrumentation__.Notify(558337)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558338)
				}
				__antithesis_instrumentation__.Notify(558326)

				return addRow(
					h.UserOid(username),
					tree.NewDName(username.Normalized()),
					tree.MakeDBool(isRoot || func() bool {
						__antithesis_instrumentation__.Notify(558339)
						return isSuper == true
					}() == true),
					tree.MakeDBool(roleInherits),
					tree.MakeDBool(isRoot || func() bool {
						__antithesis_instrumentation__.Notify(558340)
						return createRole == true
					}() == true),
					tree.MakeDBool(isRoot || func() bool {
						__antithesis_instrumentation__.Notify(558341)
						return createDB == true
					}() == true),
					tree.DBoolFalse,
					tree.MakeDBool(roleCanLogin),
					tree.DBoolFalse,
					negOneVal,
					passwdStarString,
					rolValidUntil,
					tree.DBoolFalse,
					settings,
				)
			})
	},
}

var pgCatalogSecLabelsTable = virtualSchemaTable{
	comment: `security labels (empty)
https://www.postgresql.org/docs/9.6/view-pg-seclabels.html`,
	schema: vtable.PGCatalogSecLabels,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558342)
		return nil
	},
	unimplemented: true,
}

var pgCatalogSequenceTable = virtualSchemaTable{
	comment: `sequences (see also information_schema.sequences)
https://www.postgresql.org/docs/9.5/catalog-pg-sequence.html`,
	schema: vtable.PGCatalogSequence,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558343)
		return forEachTableDesc(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(558344)
				if !table.IsSequence() {
					__antithesis_instrumentation__.Notify(558346)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(558347)
				}
				__antithesis_instrumentation__.Notify(558345)
				opts := table.GetSequenceOpts()
				return addRow(
					tableOid(table.GetID()),
					tree.NewDOid(tree.DInt(oid.T_int8)),
					tree.NewDInt(tree.DInt(opts.Start)),
					tree.NewDInt(tree.DInt(opts.Increment)),
					tree.NewDInt(tree.DInt(opts.MaxValue)),
					tree.NewDInt(tree.DInt(opts.MinValue)),
					tree.NewDInt(1),
					tree.DBoolFalse,
				)
			})
	},
}

var (
	varTypeString   = tree.NewDString("string")
	settingsCtxUser = tree.NewDString("user")
)

var pgCatalogSettingsTable = virtualSchemaTable{
	comment: `session variables (incomplete)
https://www.postgresql.org/docs/9.5/catalog-pg-settings.html`,
	schema: vtable.PGCatalogSettings,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558348)
		for _, vName := range varNames {
			__antithesis_instrumentation__.Notify(558350)
			gen := varGen[vName]
			if gen.Hidden {
				__antithesis_instrumentation__.Notify(558354)
				continue
			} else {
				__antithesis_instrumentation__.Notify(558355)
			}
			__antithesis_instrumentation__.Notify(558351)
			value, err := gen.Get(&p.extendedEvalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(558356)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558357)
			}
			__antithesis_instrumentation__.Notify(558352)
			valueDatum := tree.NewDString(value)
			var bootDatum tree.Datum = tree.DNull
			var resetDatum tree.Datum = tree.DNull
			if gen.Set == nil && func() bool {
				__antithesis_instrumentation__.Notify(558358)
				return gen.RuntimeSet == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(558359)

				bootDatum = valueDatum
				resetDatum = bootDatum
			} else {
				__antithesis_instrumentation__.Notify(558360)
				if gen.GlobalDefault != nil {
					__antithesis_instrumentation__.Notify(558362)
					globalDefVal := gen.GlobalDefault(&p.EvalContext().Settings.SV)
					bootDatum = tree.NewDString(globalDefVal)
				} else {
					__antithesis_instrumentation__.Notify(558363)
				}
				__antithesis_instrumentation__.Notify(558361)
				if hasDefault, defVal := getSessionVarDefaultString(
					vName,
					gen,
					p.sessionDataMutatorIterator.sessionDataMutatorBase,
				); hasDefault {
					__antithesis_instrumentation__.Notify(558364)
					resetDatum = tree.NewDString(defVal)
				} else {
					__antithesis_instrumentation__.Notify(558365)
				}
			}
			__antithesis_instrumentation__.Notify(558353)
			if err := addRow(
				tree.NewDString(strings.ToLower(vName)),
				valueDatum,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				settingsCtxUser,
				varTypeString,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				bootDatum,
				resetDatum,
				tree.DNull,
				tree.DNull,
				tree.DBoolFalse,
			); err != nil {
				__antithesis_instrumentation__.Notify(558366)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558367)
			}
		}
		__antithesis_instrumentation__.Notify(558349)
		return nil
	},
}

var pgCatalogShdependTable = virtualSchemaTable{
	comment: `Shared Dependencies (Roles depending on objects). 
https://www.postgresql.org/docs/9.6/catalog-pg-shdepend.html`,
	schema: vtable.PGCatalogShdepend,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558368)
		vt := p.getVirtualTabler()
		h := makeOidHasher()

		pgClassDesc, err := vt.getVirtualTableDesc(&pgClassTableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(558376)
			return errors.New("could not find pg_catalog.pg_class")
		} else {
			__antithesis_instrumentation__.Notify(558377)
		}
		__antithesis_instrumentation__.Notify(558369)
		pgClassOid := tableOid(pgClassDesc.GetID())

		pgAuthIDDesc, err := vt.getVirtualTableDesc(&pgAuthIDTableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(558378)
			return errors.New("could not find pg_catalog.pg_authid")
		} else {
			__antithesis_instrumentation__.Notify(558379)
		}
		__antithesis_instrumentation__.Notify(558370)
		pgAuthIDOid := tableOid(pgAuthIDDesc.GetID())

		pgDatabaseDesc, err := vt.getVirtualTableDesc(&pgDatabaseTableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(558380)
			return errors.New("could not find pg_catalog.pg_database")
		} else {
			__antithesis_instrumentation__.Notify(558381)
		}
		__antithesis_instrumentation__.Notify(558371)
		pgDatabaseOid := tableOid(pgDatabaseDesc.GetID())

		pinnedRoles := map[string]security.SQLUsername{
			security.RootUser:  security.RootUserName(),
			security.AdminRole: security.AdminRoleName(),
		}

		addSharedDependency := func(
			dbID *tree.DOid, classID *tree.DOid, objID *tree.DOid, refClassID *tree.DOid, user, owner security.SQLUsername,
		) error {
			__antithesis_instrumentation__.Notify(558382)

			if _, ok := pinnedRoles[user.Normalized()]; ok {
				__antithesis_instrumentation__.Notify(558385)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(558386)
			}
			__antithesis_instrumentation__.Notify(558383)

			depType := sharedDependencyACL
			if owner.Normalized() == user.Normalized() {
				__antithesis_instrumentation__.Notify(558387)
				depType = sharedDependencyOwner
			} else {
				__antithesis_instrumentation__.Notify(558388)
			}
			__antithesis_instrumentation__.Notify(558384)

			return addRow(
				dbID,
				classID,
				objID,
				zeroVal,
				refClassID,
				h.UserOid(user),
				tree.NewDString(string(depType)),
			)
		}
		__antithesis_instrumentation__.Notify(558372)

		if err = forEachTableDesc(ctx, p, dbContext, virtualMany,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(558389)
				owner := table.GetPrivileges().Owner()
				for _, u := range table.GetPrivileges().Show(privilege.Table) {
					__antithesis_instrumentation__.Notify(558391)
					if err := addSharedDependency(
						dbOid(db.GetID()),
						pgClassOid,
						tableOid(table.GetID()),
						pgAuthIDOid,
						u.User,
						owner,
					); err != nil {
						__antithesis_instrumentation__.Notify(558392)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558393)
					}
				}
				__antithesis_instrumentation__.Notify(558390)
				return nil
			},
		); err != nil {
			__antithesis_instrumentation__.Notify(558394)
			return err
		} else {
			__antithesis_instrumentation__.Notify(558395)
		}
		__antithesis_instrumentation__.Notify(558373)

		if err = forEachDatabaseDesc(ctx, p, nil, false,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(558396)
				owner := db.GetPrivileges().Owner()
				for _, u := range db.GetPrivileges().Show(privilege.Database) {
					__antithesis_instrumentation__.Notify(558398)
					if err := addSharedDependency(
						tree.NewDOid(0),
						pgDatabaseOid,
						dbOid(db.GetID()),
						pgAuthIDOid,
						u.User,
						owner,
					); err != nil {
						__antithesis_instrumentation__.Notify(558399)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558400)
					}
				}
				__antithesis_instrumentation__.Notify(558397)
				return nil
			},
		); err != nil {
			__antithesis_instrumentation__.Notify(558401)
			return err
		} else {
			__antithesis_instrumentation__.Notify(558402)
		}
		__antithesis_instrumentation__.Notify(558374)

		for _, role := range pinnedRoles {
			__antithesis_instrumentation__.Notify(558403)
			if err := addRow(
				tree.NewDOid(0),
				tree.NewDOid(0),
				tree.NewDOid(0),
				zeroVal,
				pgAuthIDOid,
				h.UserOid(role),
				tree.NewDString(string(sharedDependencyPin)),
			); err != nil {
				__antithesis_instrumentation__.Notify(558404)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558405)
			}
		}
		__antithesis_instrumentation__.Notify(558375)

		return nil
	},
}

var pgCatalogTablesTable = virtualSchemaTable{
	comment: `tables summary (see also information_schema.tables, pg_catalog.pg_class)
https://www.postgresql.org/docs/9.5/view-pg-tables.html`,
	schema: vtable.PGCatalogTables,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558406)

		return forEachTableDesc(ctx, p, dbContext, virtualMany,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(558407)
				if !table.IsTable() {
					__antithesis_instrumentation__.Notify(558409)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(558410)
				}
				__antithesis_instrumentation__.Notify(558408)
				return addRow(
					tree.NewDName(scName),
					tree.NewDName(table.GetName()),
					getOwnerName(table),
					tree.DNull,
					tree.MakeDBool(tree.DBool(table.IsPhysicalTable())),
					tree.DBoolFalse,
					tree.DBoolFalse,
					tree.DBoolFalse,
				)
			})
	},
}

var pgCatalogTablespaceTable = virtualSchemaTable{
	comment: `available tablespaces (incomplete; concept inapplicable to CockroachDB)
https://www.postgresql.org/docs/9.5/catalog-pg-tablespace.html`,
	schema: vtable.PGCatalogTablespace,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558411)
		return addRow(
			oidZero,
			tree.NewDString("pg_default"),
			tree.DNull,
			tree.DNull,
			tree.DNull,
			tree.DNull,
		)
	},
}

var pgCatalogTriggerTable = virtualSchemaTable{
	comment: `triggers (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-trigger.html`,
	schema: vtable.PGCatalogTrigger,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558412)

		return nil
	},
	unimplemented: true,
}

var (
	typTypeBase      = tree.NewDString("b")
	typTypeComposite = tree.NewDString("c")
	typTypeDomain    = tree.NewDString("d")
	typTypeEnum      = tree.NewDString("e")
	typTypePseudo    = tree.NewDString("p")
	typTypeRange     = tree.NewDString("r")

	_ = typTypeDomain
	_ = typTypePseudo
	_ = typTypeRange

	typCategoryArray       = tree.NewDString("A")
	typCategoryBoolean     = tree.NewDString("B")
	typCategoryComposite   = tree.NewDString("C")
	typCategoryDateTime    = tree.NewDString("D")
	typCategoryEnum        = tree.NewDString("E")
	typCategoryGeometric   = tree.NewDString("G")
	typCategoryNetworkAddr = tree.NewDString("I")
	typCategoryNumeric     = tree.NewDString("N")
	typCategoryPseudo      = tree.NewDString("P")
	typCategoryRange       = tree.NewDString("R")
	typCategoryString      = tree.NewDString("S")
	typCategoryTimespan    = tree.NewDString("T")
	typCategoryUserDefined = tree.NewDString("U")
	typCategoryBitString   = tree.NewDString("V")
	typCategoryUnknown     = tree.NewDString("X")

	_ = typCategoryEnum
	_ = typCategoryGeometric
	_ = typCategoryRange
	_ = typCategoryBitString

	typDelim = tree.NewDString(",")
)

func tableIDToTypeOID(table catalog.TableDescriptor) tree.Datum {
	__antithesis_instrumentation__.Notify(558413)

	if !table.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(558415)
		return tree.NewDOid(tree.DInt(typedesc.TypeIDToOID(table.GetID())))
	} else {
		__antithesis_instrumentation__.Notify(558416)
	}
	__antithesis_instrumentation__.Notify(558414)

	return tree.NewDOid(tree.DInt(table.GetID()))
}

func addPGTypeRowForTable(
	h oidHasher,
	db catalog.DatabaseDescriptor,
	scName string,
	table catalog.TableDescriptor,
	addRow func(...tree.Datum) error,
) error {
	__antithesis_instrumentation__.Notify(558417)
	nspOid := h.NamespaceOid(db.GetID(), scName)
	return addRow(
		tableIDToTypeOID(table),
		tree.NewDName(table.GetName()),
		nspOid,
		getOwnerOID(table),
		negOneVal,
		tree.DBoolFalse,
		typTypeComposite,
		typCategoryComposite,
		tree.DBoolFalse,
		tree.DBoolTrue,
		typDelim,
		tableOid(table.GetID()),
		oidZero,

		oidZero,

		h.RegProc("record_in"),
		h.RegProc("record_out"),
		h.RegProc("record_recv"),
		h.RegProc("record_send"),
		oidZero,
		oidZero,
		oidZero,

		tree.DNull,
		tree.DNull,
		tree.DBoolFalse,
		oidZero,
		negOneVal,
		zeroVal,
		oidZero,
		tree.DNull,
		tree.DNull,
		tree.DNull,
	)
}

func addPGTypeRow(
	h oidHasher, nspOid tree.Datum, owner tree.Datum, typ *types.T, addRow func(...tree.Datum) error,
) error {
	__antithesis_instrumentation__.Notify(558418)
	cat := typCategory(typ)
	typType := typTypeBase
	typElem := oidZero
	typArray := oidZero
	builtinPrefix := builtins.PGIOBuiltinPrefix(typ)
	switch typ.Family() {
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(558422)
		switch typ.Oid() {
		case oid.T_int2vector:
			__antithesis_instrumentation__.Notify(558425)

			typElem = tree.NewDOid(tree.DInt(oid.T_int2))
			typArray = tree.NewDOid(tree.DInt(oid.T__int2vector))
		case oid.T_oidvector:
			__antithesis_instrumentation__.Notify(558426)

			typElem = tree.NewDOid(tree.DInt(oid.T_oid))
			typArray = tree.NewDOid(tree.DInt(oid.T__oidvector))
		case oid.T_anyarray:
			__antithesis_instrumentation__.Notify(558427)

		default:
			__antithesis_instrumentation__.Notify(558428)
			builtinPrefix = "array_"
			typElem = tree.NewDOid(tree.DInt(typ.ArrayContents().Oid()))
		}
	case types.VoidFamily:
		__antithesis_instrumentation__.Notify(558423)

	default:
		__antithesis_instrumentation__.Notify(558424)
		typArray = tree.NewDOid(tree.DInt(types.CalcArrayOid(typ)))
	}
	__antithesis_instrumentation__.Notify(558419)
	if typ.Family() == types.EnumFamily {
		__antithesis_instrumentation__.Notify(558429)
		builtinPrefix = "enum_"
		typType = typTypeEnum
	} else {
		__antithesis_instrumentation__.Notify(558430)
	}
	__antithesis_instrumentation__.Notify(558420)
	if cat == typCategoryPseudo {
		__antithesis_instrumentation__.Notify(558431)
		typType = typTypePseudo
	} else {
		__antithesis_instrumentation__.Notify(558432)
	}
	__antithesis_instrumentation__.Notify(558421)
	typname := typ.PGName()

	return addRow(
		tree.NewDOid(tree.DInt(typ.Oid())),
		tree.NewDName(typname),
		nspOid,
		owner,
		typLen(typ),
		typByVal(typ),
		typType,
		cat,
		tree.DBoolFalse,
		tree.DBoolTrue,
		typDelim,
		oidZero,
		typElem,
		typArray,

		h.RegProc(builtinPrefix+"in"),
		h.RegProc(builtinPrefix+"out"),
		h.RegProc(builtinPrefix+"recv"),
		h.RegProc(builtinPrefix+"send"),
		oidZero,
		oidZero,
		oidZero,

		tree.DNull,
		tree.DNull,
		tree.DBoolFalse,
		oidZero,
		negOneVal,
		zeroVal,
		typColl(typ, h),
		tree.DNull,
		tree.DNull,
		tree.DNull,
	)
}

func getSchemaAndTypeByTypeID(
	ctx context.Context, p *planner, id descpb.ID,
) (string, catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(558433)
	typDesc, err := p.Descriptors().GetImmutableTypeByID(ctx, p.txn, id, tree.ObjectLookupFlags{})
	if err != nil {
		__antithesis_instrumentation__.Notify(558436)

		if !(errors.Is(err, catalog.ErrDescriptorNotFound) || func() bool {
			__antithesis_instrumentation__.Notify(558438)
			return pgerror.GetPGCode(err) == pgcode.UndefinedObject == true
		}() == true) {
			__antithesis_instrumentation__.Notify(558439)
			return "", nil, err
		} else {
			__antithesis_instrumentation__.Notify(558440)
		}
		__antithesis_instrumentation__.Notify(558437)
		return "", nil, nil
	} else {
		__antithesis_instrumentation__.Notify(558441)
	}
	__antithesis_instrumentation__.Notify(558434)

	sc, err := p.Descriptors().GetImmutableSchemaByID(
		ctx,
		p.txn,
		typDesc.GetParentSchemaID(),
		tree.SchemaLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(558442)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(558443)
	}
	__antithesis_instrumentation__.Notify(558435)
	return sc.GetName(), typDesc, nil
}

var pgCatalogTypeTable = virtualSchemaTable{
	comment: `scalar types (incomplete)
https://www.postgresql.org/docs/9.5/catalog-pg-type.html`,
	schema: vtable.PGCatalogType,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558444)
		h := makeOidHasher()
		return forEachDatabaseDesc(ctx, p, dbContext, false,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(558445)
				nspOid := h.NamespaceOid(db.GetID(), pgCatalogName)

				for _, typ := range types.OidToType {
					__antithesis_instrumentation__.Notify(558448)
					if err := addPGTypeRow(h, nspOid, tree.DNull, typ, addRow); err != nil {
						__antithesis_instrumentation__.Notify(558449)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558450)
					}
				}
				__antithesis_instrumentation__.Notify(558446)

				if err := forEachTableDescWithTableLookup(
					ctx,
					p,
					dbContext,
					virtualCurrentDB,
					func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor, lookup tableLookupFn) error {
						__antithesis_instrumentation__.Notify(558451)
						return addPGTypeRowForTable(h, db, scName, table, addRow)
					},
				); err != nil {
					__antithesis_instrumentation__.Notify(558452)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558453)
				}
				__antithesis_instrumentation__.Notify(558447)

				return forEachTypeDesc(
					ctx,
					p,
					db,
					func(_ catalog.DatabaseDescriptor, scName string, typDesc catalog.TypeDescriptor) error {
						__antithesis_instrumentation__.Notify(558454)
						nspOid := h.NamespaceOid(db.GetID(), scName)
						typ, err := typDesc.MakeTypesT(ctx, tree.NewQualifiedTypeName(db.GetName(), scName, typDesc.GetName()), p)
						if err != nil {
							__antithesis_instrumentation__.Notify(558456)
							return err
						} else {
							__antithesis_instrumentation__.Notify(558457)
						}
						__antithesis_instrumentation__.Notify(558455)
						return addPGTypeRow(h, nspOid, getOwnerOID(typDesc), typ, addRow)
					},
				)
			},
		)
	},
	indexes: []virtualIndex{
		{
			partial: false,
			populate: func(ctx context.Context, constraint tree.Datum, p *planner, db catalog.DatabaseDescriptor,
				addRow func(...tree.Datum) error) (bool, error) {
				__antithesis_instrumentation__.Notify(558458)

				h := makeOidHasher()
				nspOid := h.NamespaceOid(db.GetID(), pgCatalogName)
				coid := tree.MustBeDOid(constraint)
				ooid := oid.Oid(int(coid.DInt))

				typ, ok := types.OidToType[ooid]
				if ok {
					__antithesis_instrumentation__.Notify(558466)
					if err := addPGTypeRow(h, nspOid, tree.DNull, typ, addRow); err != nil {
						__antithesis_instrumentation__.Notify(558468)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(558469)
					}
					__antithesis_instrumentation__.Notify(558467)
					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(558470)
				}
				__antithesis_instrumentation__.Notify(558459)

				if !types.IsOIDUserDefinedType(ooid) {
					__antithesis_instrumentation__.Notify(558471)

					return false, nil
				} else {
					__antithesis_instrumentation__.Notify(558472)
				}
				__antithesis_instrumentation__.Notify(558460)

				id, err := typedesc.UserDefinedTypeOIDToID(ooid)
				if err != nil {
					__antithesis_instrumentation__.Notify(558473)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(558474)
				}
				__antithesis_instrumentation__.Notify(558461)

				scName, typDesc, err := getSchemaAndTypeByTypeID(ctx, p, id)
				if err != nil || func() bool {
					__antithesis_instrumentation__.Notify(558475)
					return typDesc == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(558476)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(558477)
				}
				__antithesis_instrumentation__.Notify(558462)

				if typDesc.GetKind() == descpb.TypeDescriptor_TABLE_IMPLICIT_RECORD_TYPE {
					__antithesis_instrumentation__.Notify(558478)
					table, err := p.Descriptors().GetImmutableTableByID(ctx, p.txn, id, tree.ObjectLookupFlags{})
					if err != nil {
						__antithesis_instrumentation__.Notify(558482)
						if errors.Is(err, catalog.ErrDescriptorNotFound) || func() bool {
							__antithesis_instrumentation__.Notify(558484)
							return pgerror.GetPGCode(err) == pgcode.UndefinedObject == true
						}() == true || func() bool {
							__antithesis_instrumentation__.Notify(558485)
							return pgerror.GetPGCode(err) == pgcode.UndefinedTable == true
						}() == true {
							__antithesis_instrumentation__.Notify(558486)
							return false, nil
						} else {
							__antithesis_instrumentation__.Notify(558487)
						}
						__antithesis_instrumentation__.Notify(558483)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(558488)
					}
					__antithesis_instrumentation__.Notify(558479)
					sc, err := p.Descriptors().GetImmutableSchemaByID(
						ctx,
						p.txn,
						table.GetParentSchemaID(),
						tree.SchemaLookupFlags{Required: true},
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(558489)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(558490)
					}
					__antithesis_instrumentation__.Notify(558480)
					if err := addPGTypeRowForTable(
						h,
						db,
						sc.GetName(),
						table,
						addRow,
					); err != nil {
						__antithesis_instrumentation__.Notify(558491)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(558492)
					}
					__antithesis_instrumentation__.Notify(558481)
					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(558493)
				}
				__antithesis_instrumentation__.Notify(558463)

				nspOid = h.NamespaceOid(db.GetID(), scName)
				typ, err = typDesc.MakeTypesT(ctx, tree.NewUnqualifiedTypeName(typDesc.GetName()), p)
				if err != nil {
					__antithesis_instrumentation__.Notify(558494)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(558495)
				}
				__antithesis_instrumentation__.Notify(558464)
				if err := addPGTypeRow(h, nspOid, getOwnerOID(typDesc), typ, addRow); err != nil {
					__antithesis_instrumentation__.Notify(558496)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(558497)
				}
				__antithesis_instrumentation__.Notify(558465)

				return true, nil
			},
		},
	},
}

var pgCatalogUserTable = virtualSchemaTable{
	comment: `database users
https://www.postgresql.org/docs/9.5/view-pg-user.html`,
	schema: vtable.PGCatalogUser,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558498)
		h := makeOidHasher()
		return forEachRole(ctx, p,
			func(username security.SQLUsername, isRole bool, options roleOptions, settings tree.Datum) error {
				__antithesis_instrumentation__.Notify(558499)
				if isRole {
					__antithesis_instrumentation__.Notify(558504)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(558505)
				}
				__antithesis_instrumentation__.Notify(558500)
				isRoot := tree.DBool(username.IsRootUser())
				createDB, err := options.createDB()
				if err != nil {
					__antithesis_instrumentation__.Notify(558506)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558507)
				}
				__antithesis_instrumentation__.Notify(558501)
				validUntil, err := options.validUntil(p)
				if err != nil {
					__antithesis_instrumentation__.Notify(558508)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558509)
				}
				__antithesis_instrumentation__.Notify(558502)
				isSuper, err := userIsSuper(ctx, p, username)
				if err != nil {
					__antithesis_instrumentation__.Notify(558510)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558511)
				}
				__antithesis_instrumentation__.Notify(558503)

				return addRow(
					tree.NewDName(username.Normalized()),
					h.UserOid(username),
					tree.MakeDBool(isRoot || func() bool {
						__antithesis_instrumentation__.Notify(558512)
						return createDB == true
					}() == true),
					tree.MakeDBool(isRoot || func() bool {
						__antithesis_instrumentation__.Notify(558513)
						return isSuper == true
					}() == true),
					tree.DBoolFalse,
					tree.DBoolFalse,
					passwdStarString,
					validUntil,
					settings,
				)
			})
	},
}

var pgCatalogUserMappingTable = virtualSchemaTable{
	comment: `local to remote user mapping (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-user-mapping.html`,
	schema: vtable.PGCatalogUserMapping,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558514)

		return nil
	},
	unimplemented: true,
}

var pgCatalogStatActivityTable = virtualSchemaTable{
	comment: `backend access statistics (empty - monitoring works differently in CockroachDB)
https://www.postgresql.org/docs/9.6/monitoring-stats.html#PG-STAT-ACTIVITY-VIEW`,
	schema: vtable.PGCatalogStatActivity,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558515)
		return nil
	},
	unimplemented: true,
}

var pgCatalogSecurityLabelTable = virtualSchemaTable{
	comment: `security labels (empty - feature does not exist)
https://www.postgresql.org/docs/9.5/catalog-pg-seclabel.html`,
	schema: vtable.PGCatalogSecurityLabel,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558516)
		return nil
	},
	unimplemented: true,
}

var pgCatalogSharedSecurityLabelTable = virtualSchemaTable{
	comment: `shared security labels (empty - feature not supported)
https://www.postgresql.org/docs/9.5/catalog-pg-shseclabel.html`,
	schema: vtable.PGCatalogSharedSecurityLabel,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558517)
		return nil
	},
	unimplemented: true,
}

var pgCatalogDbRoleSettingTable = virtualSchemaTable{
	comment: `contains the default values that have been configured for session variables
https://www.postgresql.org/docs/13/catalog-pg-db-role-setting.html`,
	schema: vtable.PgCatalogDbRoleSetting,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558518)
		rows, err := p.extendedEvalCtx.ExecCfg.InternalExecutor.QueryBufferedEx(
			ctx,
			"select-db-role-settings",
			p.EvalContext().Txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			`SELECT database_id, role_name, settings FROM system.public.database_role_settings`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(558521)
			return err
		} else {
			__antithesis_instrumentation__.Notify(558522)
		}
		__antithesis_instrumentation__.Notify(558519)
		h := makeOidHasher()
		for _, row := range rows {
			__antithesis_instrumentation__.Notify(558523)
			databaseID := tree.MustBeDOid(row[0])
			roleName := tree.MustBeDString(row[1])
			roleID := oidZero
			if roleName != "" {
				__antithesis_instrumentation__.Notify(558525)
				roleID = h.UserOid(security.MakeSQLUsernameFromPreNormalizedString(string(roleName)))
			} else {
				__antithesis_instrumentation__.Notify(558526)
			}
			__antithesis_instrumentation__.Notify(558524)
			settings := tree.MustBeDArray(row[2])
			if err := addRow(
				settings,
				databaseID,
				roleID,
			); err != nil {
				__antithesis_instrumentation__.Notify(558527)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558528)
			}
		}
		__antithesis_instrumentation__.Notify(558520)
		return nil
	},
}

var pgCatalogShadowTable = virtualSchemaTable{
	comment: `pg_shadow lists properties for roles that are marked as rolcanlogin in pg_authid
https://www.postgresql.org/docs/13/view-pg-shadow.html`,
	schema: vtable.PgCatalogShadow,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558529)
		h := makeOidHasher()
		return forEachRole(ctx, p, func(username security.SQLUsername, isRole bool, options roleOptions, settings tree.Datum) error {
			__antithesis_instrumentation__.Notify(558530)
			noLogin, err := options.noLogin()
			if err != nil {
				__antithesis_instrumentation__.Notify(558536)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558537)
			}
			__antithesis_instrumentation__.Notify(558531)
			if noLogin {
				__antithesis_instrumentation__.Notify(558538)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(558539)
			}
			__antithesis_instrumentation__.Notify(558532)

			isRoot := tree.DBool(username.IsRootUser() || func() bool {
				__antithesis_instrumentation__.Notify(558540)
				return username.IsAdminRole() == true
			}() == true)
			createDB, err := options.createDB()
			if err != nil {
				__antithesis_instrumentation__.Notify(558541)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558542)
			}
			__antithesis_instrumentation__.Notify(558533)
			rolValidUntil, err := options.validUntil(p)
			if err != nil {
				__antithesis_instrumentation__.Notify(558543)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558544)
			}
			__antithesis_instrumentation__.Notify(558534)
			isSuper, err := userIsSuper(ctx, p, username)
			if err != nil {
				__antithesis_instrumentation__.Notify(558545)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558546)
			}
			__antithesis_instrumentation__.Notify(558535)

			return addRow(
				tree.NewDName(username.Normalized()),
				h.UserOid(username),
				tree.MakeDBool(isRoot || func() bool {
					__antithesis_instrumentation__.Notify(558547)
					return createDB == true
				}() == true),
				tree.MakeDBool(isRoot || func() bool {
					__antithesis_instrumentation__.Notify(558548)
					return isSuper == true
				}() == true),
				tree.DBoolFalse,
				tree.DBoolFalse,
				passwdStarString,
				rolValidUntil,
				settings,
			)
		})
	},
}

var pgCatalogStatisticExtTable = virtualSchemaTable{
	comment: `pg_statistic_ext has the statistics objects created with CREATE STATISTICS
https://www.postgresql.org/docs/13/catalog-pg-statistic-ext.html`,
	schema: vtable.PgCatalogStatisticExt,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558549)
		query := `SELECT "statisticID", name, "tableID", "columnIDs" FROM system.table_statistics;`
		rows, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryBuffered(
			ctx, "read-statistics-objects", p.txn, query,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(558552)
			return err
		} else {
			__antithesis_instrumentation__.Notify(558553)
		}
		__antithesis_instrumentation__.Notify(558550)

		for _, row := range rows {
			__antithesis_instrumentation__.Notify(558554)
			statisticsID := tree.MustBeDInt(row[0])
			name := tree.MustBeDString(row[1])
			tableID := tree.MustBeDInt(row[2])
			columnIDs := tree.MustBeDArray(row[3])

			if err := addRow(
				tree.NewDOid(statisticsID),
				tree.NewDOid(tableID),
				&name,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				columnIDs,
				tree.DNull,
			); err != nil {
				__antithesis_instrumentation__.Notify(558555)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558556)
			}
		}
		__antithesis_instrumentation__.Notify(558551)
		return nil
	},
}

var pgCatalogSequencesTable = virtualSchemaTable{
	comment: `pg_sequences is very similar as pg_sequence.
https://www.postgresql.org/docs/13/view-pg-sequences.html
`,
	schema: vtable.PgCatalogSequences,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558557)
		return forEachTableDesc(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(558558)
				if !table.IsSequence() {
					__antithesis_instrumentation__.Notify(558562)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(558563)
				}
				__antithesis_instrumentation__.Notify(558559)
				opts := table.GetSequenceOpts()
				lastValue := tree.DNull
				sequenceValue, err := p.GetSequenceValue(ctx, p.execCfg.Codec, table)
				if err != nil {
					__antithesis_instrumentation__.Notify(558564)
					return err
				} else {
					__antithesis_instrumentation__.Notify(558565)
				}
				__antithesis_instrumentation__.Notify(558560)

				if sequenceValue != opts.Start-opts.Increment {
					__antithesis_instrumentation__.Notify(558566)
					lastValue = tree.NewDInt(tree.DInt(sequenceValue))
				} else {
					__antithesis_instrumentation__.Notify(558567)
				}
				__antithesis_instrumentation__.Notify(558561)

				return addRow(
					tree.NewDString(scName),
					tree.NewDString(table.GetName()),
					getOwnerName(table),
					tree.NewDOid(tree.DInt(oid.T_int8)),
					tree.NewDInt(tree.DInt(opts.Start)),
					tree.NewDInt(tree.DInt(opts.MinValue)),
					tree.NewDInt(tree.DInt(opts.MaxValue)),
					tree.NewDInt(tree.DInt(opts.Increment)),
					tree.DBoolFalse,
					tree.NewDInt(tree.DInt(opts.CacheSize)),
					lastValue,
				)
			},
		)
	},
}

var pgCatalogInitPrivsTable = virtualSchemaTable{
	comment: "pg_init_privs was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogInitPrivs,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558568)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatProgressCreateIndexTable = virtualSchemaTable{
	comment: "pg_stat_progress_create_index was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatProgressCreateIndex,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558569)
		return nil
	},
	unimplemented: true,
}

var pgCatalogOpfamilyTable = virtualSchemaTable{
	comment: "pg_opfamily was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogOpfamily,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558570)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatioAllSequencesTable = virtualSchemaTable{
	comment: "pg_statio_all_sequences was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatioAllSequences,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558571)
		return nil
	},
	unimplemented: true,
}

var pgCatalogPoliciesTable = virtualSchemaTable{
	comment: "pg_policies was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogPolicies,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558572)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatsExtTable = virtualSchemaTable{
	comment: "pg_stats_ext was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatsExt,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558573)
		return nil
	},
	unimplemented: true,
}

var pgCatalogUserMappingsTable = virtualSchemaTable{
	comment: "pg_user_mappings was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogUserMappings,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558574)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatGssapiTable = virtualSchemaTable{
	comment: "pg_stat_gssapi was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatGssapi,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558575)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatDatabaseTable = virtualSchemaTable{
	comment: "pg_stat_database was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatDatabase,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558576)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatioUserIndexesTable = virtualSchemaTable{
	comment: "pg_statio_user_indexes was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatioUserIndexes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558577)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatSslTable = virtualSchemaTable{
	comment: "pg_stat_ssl was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatSsl,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558578)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatioAllIndexesTable = virtualSchemaTable{
	comment: "pg_statio_all_indexes was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatioAllIndexes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558579)
		return nil
	},
	unimplemented: true,
}

var pgCatalogTimezoneAbbrevsTable = virtualSchemaTable{
	comment: "pg_timezone_abbrevs was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogTimezoneAbbrevs,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558580)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatSysTablesTable = virtualSchemaTable{
	comment: "pg_stat_sys_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatSysTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558581)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatioSysSequencesTable = virtualSchemaTable{
	comment: "pg_statio_sys_sequences was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatioSysSequences,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558582)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatAllTablesTable = virtualSchemaTable{
	comment: "pg_stat_all_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatAllTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558583)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatioSysTablesTable = virtualSchemaTable{
	comment: "pg_statio_sys_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatioSysTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558584)
		return nil
	},
	unimplemented: true,
}

var pgCatalogTsConfigTable = virtualSchemaTable{
	comment: "pg_ts_config was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogTsConfig,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558585)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatsTable = virtualSchemaTable{
	comment: "pg_stats was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStats,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558586)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatProgressBasebackupTable = virtualSchemaTable{
	comment: "pg_stat_progress_basebackup was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatProgressBasebackup,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558587)
		return nil
	},
	unimplemented: true,
}

var pgCatalogPolicyTable = virtualSchemaTable{
	comment: "pg_policy was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogPolicy,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558588)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatArchiverTable = virtualSchemaTable{
	comment: "pg_stat_archiver was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatArchiver,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558589)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatXactUserFunctionsTable = virtualSchemaTable{
	comment: "pg_stat_xact_user_functions was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatXactUserFunctions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558590)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatUserFunctionsTable = virtualSchemaTable{
	comment: "pg_stat_user_functions was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatUserFunctions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558591)
		return nil
	},
	unimplemented: true,
}

var pgCatalogPublicationTable = virtualSchemaTable{
	comment: "pg_publication was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogPublication,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558592)
		return nil
	},
	unimplemented: true,
}

var pgCatalogAmprocTable = virtualSchemaTable{
	comment: "pg_amproc was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogAmproc,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558593)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatProgressAnalyzeTable = virtualSchemaTable{
	comment: "pg_stat_progress_analyze was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatProgressAnalyze,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558594)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatXactAllTablesTable = virtualSchemaTable{
	comment: "pg_stat_xact_all_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatXactAllTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558595)
		return nil
	},
	unimplemented: true,
}

var pgCatalogHbaFileRulesTable = virtualSchemaTable{
	comment: "pg_hba_file_rules was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogHbaFileRules,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558596)
		return nil
	},
	unimplemented: true,
}

var pgCatalogCursorsTable = virtualSchemaTable{
	comment: `contains currently active SQL cursors created with DECLARE
https://www.postgresql.org/docs/14/view-pg-cursors.html`,
	schema: vtable.PgCatalogCursors,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558597)
		for name, c := range p.sqlCursors.list() {
			__antithesis_instrumentation__.Notify(558599)
			tz, err := tree.MakeDTimestampTZ(c.created, time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(558601)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558602)
			}
			__antithesis_instrumentation__.Notify(558600)
			if err := addRow(
				tree.NewDString(name),
				tree.NewDString(c.statement),
				tree.DBoolFalse,
				tree.DBoolFalse,
				tree.DBoolFalse,
				tz,
			); err != nil {
				__antithesis_instrumentation__.Notify(558603)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558604)
			}
		}
		__antithesis_instrumentation__.Notify(558598)
		return nil
	},
}

var pgCatalogStatSlruTable = virtualSchemaTable{
	comment: "pg_stat_slru was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatSlru,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558605)
		return nil
	},
	unimplemented: true,
}

var pgCatalogFileSettingsTable = virtualSchemaTable{
	comment: "pg_file_settings was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogFileSettings,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558606)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatUserIndexesTable = virtualSchemaTable{
	comment: "pg_stat_user_indexes was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatUserIndexes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558607)
		return nil
	},
	unimplemented: true,
}

var pgCatalogRulesTable = virtualSchemaTable{
	comment: "pg_rules was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogRules,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558608)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatioUserSequencesTable = virtualSchemaTable{
	comment: "pg_statio_user_sequences was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatioUserSequences,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558609)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatAllIndexesTable = virtualSchemaTable{
	comment: "pg_stat_all_indexes was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatAllIndexes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558610)
		return nil
	},
	unimplemented: true,
}

var pgCatalogTsConfigMapTable = virtualSchemaTable{
	comment: "pg_ts_config_map was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogTsConfigMap,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558611)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatBgwriterTable = virtualSchemaTable{
	comment: "pg_stat_bgwriter was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatBgwriter,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558612)
		return nil
	},
	unimplemented: true,
}

var pgCatalogTransformTable = virtualSchemaTable{
	comment: "pg_transform was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogTransform,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558613)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatXactUserTablesTable = virtualSchemaTable{
	comment: "pg_stat_xact_user_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatXactUserTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558614)
		return nil
	},
	unimplemented: true,
}

var pgCatalogPublicationTablesTable = virtualSchemaTable{
	comment: "pg_publication_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogPublicationTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558615)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatProgressClusterTable = virtualSchemaTable{
	comment: "pg_stat_progress_cluster was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatProgressCluster,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558616)
		return nil
	},
	unimplemented: true,
}

var pgCatalogGroupTable = virtualSchemaTable{
	comment: "pg_group was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogGroup,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558617)
		return nil
	},
	unimplemented: true,
}

var pgCatalogLargeobjectMetadataTable = virtualSchemaTable{
	comment: "pg_largeobject_metadata was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogLargeobjectMetadata,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558618)
		return nil
	},
	unimplemented: true,
}

var pgCatalogReplicationSlotsTable = virtualSchemaTable{
	comment: "pg_replication_slots was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogReplicationSlots,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558619)
		return nil
	},
	unimplemented: true,
}

var pgCatalogSubscriptionRelTable = virtualSchemaTable{
	comment: "pg_subscription_rel was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogSubscriptionRel,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558620)
		return nil
	},
	unimplemented: true,
}

var pgCatalogTsParserTable = virtualSchemaTable{
	comment: "pg_ts_parser was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogTsParser,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558621)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatisticExtDataTable = virtualSchemaTable{
	comment: "pg_statistic_ext_data was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatisticExtData,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558622)
		return nil
	},
	unimplemented: true,
}

var pgCatalogPartitionedTableTable = virtualSchemaTable{
	comment: "pg_partitioned_table was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogPartitionedTable,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558623)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatioSysIndexesTable = virtualSchemaTable{
	comment: "pg_statio_sys_indexes was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatioSysIndexes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558624)
		return nil
	},
	unimplemented: true,
}

var pgCatalogConfigTable = virtualSchemaTable{
	comment: "pg_config was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogConfig,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558625)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatioUserTablesTable = virtualSchemaTable{
	comment: "pg_statio_user_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatioUserTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558626)
		return nil
	},
	unimplemented: true,
}

var pgCatalogTimezoneNamesTable = virtualSchemaTable{
	comment: "pg_timezone_names was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogTimezoneNames,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558627)
		return nil
	},
	unimplemented: true,
}

var pgCatalogTsDictTable = virtualSchemaTable{
	comment: "pg_ts_dict was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogTsDict,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558628)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatUserTablesTable = virtualSchemaTable{
	comment: "pg_stat_user_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatUserTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558629)
		return nil
	},
	unimplemented: true,
}

var pgCatalogSubscriptionTable = virtualSchemaTable{
	comment: "pg_subscription was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogSubscription,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558630)
		return nil
	},
	unimplemented: true,
}

var pgCatalogShmemAllocationsTable = virtualSchemaTable{
	comment: "pg_shmem_allocations was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogShmemAllocations,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558631)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatWalReceiverTable = virtualSchemaTable{
	comment: "pg_stat_wal_receiver was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatWalReceiver,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558632)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatSubscriptionTable = virtualSchemaTable{
	comment: "pg_stat_subscription was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatSubscription,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558633)
		return nil
	},
	unimplemented: true,
}

var pgCatalogLargeobjectTable = virtualSchemaTable{
	comment: "pg_largeobject was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogLargeobject,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558634)
		return nil
	},
	unimplemented: true,
}

var pgCatalogReplicationOriginStatusTable = virtualSchemaTable{
	comment: "pg_replication_origin_status was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogReplicationOriginStatus,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558635)
		return nil
	},
	unimplemented: true,
}

var pgCatalogAmopTable = virtualSchemaTable{
	comment: "pg_amop was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogAmop,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558636)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatProgressVacuumTable = virtualSchemaTable{
	comment: "pg_stat_progress_vacuum was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatProgressVacuum,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558637)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatSysIndexesTable = virtualSchemaTable{
	comment: "pg_stat_sys_indexes was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatSysIndexes,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558638)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatioAllTablesTable = virtualSchemaTable{
	comment: "pg_statio_all_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatioAllTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558639)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatDatabaseConflictsTable = virtualSchemaTable{
	comment: "pg_stat_database_conflicts was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatDatabaseConflicts,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558640)
		return nil
	},
	unimplemented: true,
}

var pgCatalogReplicationOriginTable = virtualSchemaTable{
	comment: "pg_replication_origin was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogReplicationOrigin,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558641)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatisticTable = virtualSchemaTable{
	comment: "pg_statistic was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatistic,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558642)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatXactSysTablesTable = virtualSchemaTable{
	comment: "pg_stat_xact_sys_tables was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatXactSysTables,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558643)
		return nil
	},
	unimplemented: true,
}

var pgCatalogTsTemplateTable = virtualSchemaTable{
	comment: "pg_ts_template was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogTsTemplate,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558644)
		return nil
	},
	unimplemented: true,
}

var pgCatalogStatReplicationTable = virtualSchemaTable{
	comment: "pg_stat_replication was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogStatReplication,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558645)
		return nil
	},
	unimplemented: true,
}

var pgCatalogPublicationRelTable = virtualSchemaTable{
	comment: "pg_publication_rel was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogPublicationRel,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558646)
		return nil
	},
	unimplemented: true,
}

var pgCatalogAvailableExtensionVersionsTable = virtualSchemaTable{
	comment: "pg_available_extension_versions was created for compatibility and is currently unimplemented",
	schema:  vtable.PgCatalogAvailableExtensionVersions,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558647)
		return nil
	},
	unimplemented: true,
}

func typOid(typ *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(558648)
	return tree.NewDOid(tree.DInt(typ.Oid()))
}

func typLen(typ *types.T) *tree.DInt {
	__antithesis_instrumentation__.Notify(558649)
	if sz, variable := tree.DatumTypeSize(typ); !variable {
		__antithesis_instrumentation__.Notify(558651)
		return tree.NewDInt(tree.DInt(sz))
	} else {
		__antithesis_instrumentation__.Notify(558652)
	}
	__antithesis_instrumentation__.Notify(558650)
	return negOneVal
}

func typByVal(typ *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(558653)
	_, variable := tree.DatumTypeSize(typ)
	return tree.MakeDBool(tree.DBool(!variable))
}

func typColl(typ *types.T, h oidHasher) tree.Datum {
	__antithesis_instrumentation__.Notify(558654)
	switch typ.Family() {
	case types.AnyFamily:
		__antithesis_instrumentation__.Notify(558657)
		return oidZero
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(558658)
		return h.CollationOid(tree.DefaultCollationTag)
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(558659)
		return h.CollationOid(typ.Locale())
	default:
		__antithesis_instrumentation__.Notify(558660)
	}
	__antithesis_instrumentation__.Notify(558655)

	if typ.Equivalent(types.StringArray) {
		__antithesis_instrumentation__.Notify(558661)
		return h.CollationOid(tree.DefaultCollationTag)
	} else {
		__antithesis_instrumentation__.Notify(558662)
	}
	__antithesis_instrumentation__.Notify(558656)
	return oidZero
}

var datumToTypeCategory = map[types.Family]*tree.DString{
	types.AnyFamily:         typCategoryPseudo,
	types.BitFamily:         typCategoryBitString,
	types.BoolFamily:        typCategoryBoolean,
	types.BytesFamily:       typCategoryUserDefined,
	types.DateFamily:        typCategoryDateTime,
	types.EnumFamily:        typCategoryEnum,
	types.TimeFamily:        typCategoryDateTime,
	types.TimeTZFamily:      typCategoryDateTime,
	types.FloatFamily:       typCategoryNumeric,
	types.IntFamily:         typCategoryNumeric,
	types.IntervalFamily:    typCategoryTimespan,
	types.Box2DFamily:       typCategoryUserDefined,
	types.GeographyFamily:   typCategoryUserDefined,
	types.GeometryFamily:    typCategoryUserDefined,
	types.JsonFamily:        typCategoryUserDefined,
	types.DecimalFamily:     typCategoryNumeric,
	types.StringFamily:      typCategoryString,
	types.TimestampFamily:   typCategoryDateTime,
	types.TimestampTZFamily: typCategoryDateTime,
	types.ArrayFamily:       typCategoryArray,
	types.TupleFamily:       typCategoryPseudo,
	types.OidFamily:         typCategoryNumeric,
	types.UuidFamily:        typCategoryUserDefined,
	types.INetFamily:        typCategoryNetworkAddr,
	types.UnknownFamily:     typCategoryUnknown,
	types.VoidFamily:        typCategoryPseudo,
}

func typCategory(typ *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(558663)

	if typ.Family() == types.ArrayFamily && func() bool {
		__antithesis_instrumentation__.Notify(558665)
		return typ.ArrayContents().Family() == types.AnyFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(558666)
		return typCategoryPseudo
	} else {
		__antithesis_instrumentation__.Notify(558667)
	}
	__antithesis_instrumentation__.Notify(558664)
	return datumToTypeCategory[typ.Family()]
}

var pgCatalogViewsTable = virtualSchemaTable{
	comment: `view definitions (incomplete - see also information_schema.views)
https://www.postgresql.org/docs/9.5/view-pg-views.html`,
	schema: vtable.PGCatalogViews,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558668)

		return forEachTableDesc(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, desc catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(558669)
				if !desc.IsView() || func() bool {
					__antithesis_instrumentation__.Notify(558671)
					return desc.MaterializedView() == true
				}() == true {
					__antithesis_instrumentation__.Notify(558672)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(558673)
				}
				__antithesis_instrumentation__.Notify(558670)

				return addRow(
					tree.NewDName(scName),
					tree.NewDName(desc.GetName()),
					getOwnerName(desc),
					tree.NewDString(desc.GetViewQuery()),
				)
			})
	},
}

var pgCatalogAggregateTable = virtualSchemaTable{
	comment: `aggregated built-in functions (incomplete)
https://www.postgresql.org/docs/9.6/catalog-pg-aggregate.html`,
	schema: vtable.PGCatalogAggregate,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558674)
		h := makeOidHasher()
		return forEachDatabaseDesc(ctx, p, dbContext, false,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(558675)
				for _, name := range builtins.AllAggregateBuiltinNames {
					__antithesis_instrumentation__.Notify(558677)
					if name == builtins.AnyNotNull {
						__antithesis_instrumentation__.Notify(558679)

						continue
					} else {
						__antithesis_instrumentation__.Notify(558680)
					}
					__antithesis_instrumentation__.Notify(558678)
					_, overloads := builtins.GetBuiltinProperties(name)
					for _, overload := range overloads {
						__antithesis_instrumentation__.Notify(558681)
						params, _ := tree.GetParamsAndReturnType(overload)
						sortOperatorOid := oidZero
						aggregateKind := tree.NewDString("n")
						aggNumDirectArgs := zeroVal
						if params.Length() != 0 {
							__antithesis_instrumentation__.Notify(558683)
							argType := tree.NewDOid(tree.DInt(params.Types()[0].Oid()))
							returnType := tree.NewDOid(tree.DInt(oid.T_bool))
							switch name {

							case "max", "bool_or":
								__antithesis_instrumentation__.Notify(558684)
								sortOperatorOid = h.OperatorOid(">", argType, argType, returnType)
							case "min", "bool_and", "every":
								__antithesis_instrumentation__.Notify(558685)
								sortOperatorOid = h.OperatorOid("<", argType, argType, returnType)

							case "rank", "percent_rank", "cume_dist", "dense_rank":
								__antithesis_instrumentation__.Notify(558686)
								aggregateKind = tree.NewDString("h")
								aggNumDirectArgs = tree.NewDInt(1)
							case "mode":
								__antithesis_instrumentation__.Notify(558687)
								aggregateKind = tree.NewDString("o")
							default:
								__antithesis_instrumentation__.Notify(558688)
								if strings.HasPrefix(name, "percentile_") {
									__antithesis_instrumentation__.Notify(558689)
									aggregateKind = tree.NewDString("o")
									aggNumDirectArgs = tree.NewDInt(1)
								} else {
									__antithesis_instrumentation__.Notify(558690)
								}
							}
						} else {
							__antithesis_instrumentation__.Notify(558691)
						}
						__antithesis_instrumentation__.Notify(558682)
						regprocForZeroOid := tree.NewDOidWithName(tree.DInt(0), types.RegProc, "-")
						err := addRow(
							h.BuiltinOid(name, &overload).AsRegProc(name),
							aggregateKind,
							aggNumDirectArgs,
							regprocForZeroOid,
							regprocForZeroOid,
							regprocForZeroOid,
							regprocForZeroOid,
							regprocForZeroOid,
							regprocForZeroOid,
							regprocForZeroOid,
							regprocForZeroOid,
							tree.DBoolFalse,
							tree.DBoolFalse,
							sortOperatorOid,
							tree.DNull,
							tree.DNull,
							tree.DNull,
							tree.DNull,
							tree.DNull,
							tree.DNull,

							tree.DNull,
							tree.DNull,
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(558692)
							return err
						} else {
							__antithesis_instrumentation__.Notify(558693)
						}
					}
				}
				__antithesis_instrumentation__.Notify(558676)
				return nil
			})
	},
}

func init() {
	h := makeOidHasher()
	tree.OidToBuiltinName = make(map[oid.Oid]string, len(tree.FunDefs))
	for name, def := range tree.FunDefs {
		for _, o := range def.Definition {
			if overload, ok := o.(*tree.Overload); ok {
				builtinOid := h.BuiltinOid(name, overload)
				id := oid.Oid(builtinOid.DInt)
				tree.OidToBuiltinName[id] = name
				overload.Oid = id
			}
		}
	}
}

type oidHasher struct {
	h hash.Hash32
}

func makeOidHasher() oidHasher {
	__antithesis_instrumentation__.Notify(558694)
	return oidHasher{h: fnv.New32()}
}

func (h oidHasher) writeStr(s string) {
	__antithesis_instrumentation__.Notify(558695)
	if _, err := h.h.Write([]byte(s)); err != nil {
		__antithesis_instrumentation__.Notify(558696)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(558697)
	}
}

func (h oidHasher) writeBytes(b []byte) {
	__antithesis_instrumentation__.Notify(558698)
	if _, err := h.h.Write(b); err != nil {
		__antithesis_instrumentation__.Notify(558699)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(558700)
	}
}

func (h oidHasher) writeUInt8(i uint8) {
	__antithesis_instrumentation__.Notify(558701)
	if err := binary.Write(h.h, binary.BigEndian, i); err != nil {
		__antithesis_instrumentation__.Notify(558702)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(558703)
	}
}

func (h oidHasher) writeUInt32(i uint32) {
	__antithesis_instrumentation__.Notify(558704)
	if err := binary.Write(h.h, binary.BigEndian, i); err != nil {
		__antithesis_instrumentation__.Notify(558705)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(558706)
	}
}

func (h oidHasher) writeUInt64(i uint64) {
	__antithesis_instrumentation__.Notify(558707)
	if err := binary.Write(h.h, binary.BigEndian, i); err != nil {
		__antithesis_instrumentation__.Notify(558708)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(558709)
	}
}

func (h oidHasher) writeOID(oid *tree.DOid) {
	__antithesis_instrumentation__.Notify(558710)
	h.writeUInt64(uint64(oid.DInt))
}

type oidTypeTag uint8

const (
	_ oidTypeTag = iota
	namespaceTypeTag
	indexTypeTag
	columnTypeTag
	checkConstraintTypeTag
	fkConstraintTypeTag
	pKeyConstraintTypeTag
	uniqueConstraintTypeTag
	functionTypeTag
	userTypeTag
	collationTypeTag
	operatorTypeTag
	enumEntryTypeTag
	rewriteTypeTag
	dbSchemaRoleTypeTag
)

func (h oidHasher) writeTypeTag(tag oidTypeTag) {
	__antithesis_instrumentation__.Notify(558711)
	h.writeUInt8(uint8(tag))
}

func (h oidHasher) getOid() *tree.DOid {
	__antithesis_instrumentation__.Notify(558712)
	i := h.h.Sum32()
	h.h.Reset()
	return tree.NewDOid(tree.DInt(i))
}

func (h oidHasher) writeDB(dbID descpb.ID) {
	__antithesis_instrumentation__.Notify(558713)
	h.writeUInt32(uint32(dbID))
}

func (h oidHasher) writeSchema(scName string) {
	__antithesis_instrumentation__.Notify(558714)
	h.writeStr(scName)
}

func (h oidHasher) writeTable(tableID descpb.ID) {
	__antithesis_instrumentation__.Notify(558715)
	h.writeUInt32(uint32(tableID))
}

func (h oidHasher) writeIndex(indexID descpb.IndexID) {
	__antithesis_instrumentation__.Notify(558716)
	h.writeUInt32(uint32(indexID))
}

func (h oidHasher) writeUniqueConstraint(uc *descpb.UniqueWithoutIndexConstraint) {
	__antithesis_instrumentation__.Notify(558717)
	h.writeUInt32(uint32(uc.TableID))
	h.writeStr(uc.Name)
}

func (h oidHasher) writeCheckConstraint(check *descpb.TableDescriptor_CheckConstraint) {
	__antithesis_instrumentation__.Notify(558718)
	h.writeStr(check.Name)
	h.writeStr(check.Expr)
}

func (h oidHasher) writeForeignKeyConstraint(fk *descpb.ForeignKeyConstraint) {
	__antithesis_instrumentation__.Notify(558719)
	h.writeUInt32(uint32(fk.ReferencedTableID))
	h.writeStr(fk.Name)
}

func (h oidHasher) NamespaceOid(dbID descpb.ID, scName string) *tree.DOid {
	__antithesis_instrumentation__.Notify(558720)
	h.writeTypeTag(namespaceTypeTag)
	h.writeDB(dbID)
	h.writeSchema(scName)
	return h.getOid()
}

func (h oidHasher) IndexOid(tableID descpb.ID, indexID descpb.IndexID) *tree.DOid {
	__antithesis_instrumentation__.Notify(558721)
	h.writeTypeTag(indexTypeTag)
	h.writeTable(tableID)
	h.writeIndex(indexID)
	return h.getOid()
}

func (h oidHasher) ColumnOid(tableID descpb.ID, columnID descpb.ColumnID) *tree.DOid {
	__antithesis_instrumentation__.Notify(558722)
	h.writeTypeTag(columnTypeTag)
	h.writeUInt32(uint32(tableID))
	h.writeUInt32(uint32(columnID))
	return h.getOid()
}

func (h oidHasher) CheckConstraintOid(
	dbID descpb.ID, scName string, tableID descpb.ID, check *descpb.TableDescriptor_CheckConstraint,
) *tree.DOid {
	__antithesis_instrumentation__.Notify(558723)
	h.writeTypeTag(checkConstraintTypeTag)
	h.writeDB(dbID)
	h.writeSchema(scName)
	h.writeTable(tableID)
	h.writeCheckConstraint(check)
	return h.getOid()
}

func (h oidHasher) PrimaryKeyConstraintOid(
	dbID descpb.ID, scName string, tableID descpb.ID, pkey *descpb.IndexDescriptor,
) *tree.DOid {
	__antithesis_instrumentation__.Notify(558724)
	h.writeTypeTag(pKeyConstraintTypeTag)
	h.writeDB(dbID)
	h.writeSchema(scName)
	h.writeTable(tableID)
	h.writeIndex(pkey.ID)
	return h.getOid()
}

func (h oidHasher) ForeignKeyConstraintOid(
	dbID descpb.ID, scName string, tableID descpb.ID, fk *descpb.ForeignKeyConstraint,
) *tree.DOid {
	__antithesis_instrumentation__.Notify(558725)
	h.writeTypeTag(fkConstraintTypeTag)
	h.writeDB(dbID)
	h.writeSchema(scName)
	h.writeTable(tableID)
	h.writeForeignKeyConstraint(fk)
	return h.getOid()
}

func (h oidHasher) UniqueWithoutIndexConstraintOid(
	dbID descpb.ID, scName string, tableID descpb.ID, uc *descpb.UniqueWithoutIndexConstraint,
) *tree.DOid {
	__antithesis_instrumentation__.Notify(558726)
	h.writeTypeTag(uniqueConstraintTypeTag)
	h.writeDB(dbID)
	h.writeSchema(scName)
	h.writeTable(tableID)
	h.writeUniqueConstraint(uc)
	return h.getOid()
}

func (h oidHasher) UniqueConstraintOid(
	dbID descpb.ID, scName string, tableID descpb.ID, indexID descpb.IndexID,
) *tree.DOid {
	__antithesis_instrumentation__.Notify(558727)
	h.writeTypeTag(uniqueConstraintTypeTag)
	h.writeDB(dbID)
	h.writeSchema(scName)
	h.writeTable(tableID)
	h.writeIndex(indexID)
	return h.getOid()
}

func (h oidHasher) BuiltinOid(name string, builtin *tree.Overload) *tree.DOid {
	__antithesis_instrumentation__.Notify(558728)
	h.writeTypeTag(functionTypeTag)
	h.writeStr(name)
	h.writeStr(builtin.Types.String())
	h.writeStr(builtin.FixedReturnType().String())
	return h.getOid()
}

func (h oidHasher) RegProc(name string) tree.Datum {
	__antithesis_instrumentation__.Notify(558729)
	_, overloads := builtins.GetBuiltinProperties(name)
	if len(overloads) == 0 {
		__antithesis_instrumentation__.Notify(558731)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(558732)
	}
	__antithesis_instrumentation__.Notify(558730)
	return h.BuiltinOid(name, &overloads[0]).AsRegProc(name)
}

func (h oidHasher) UserOid(username security.SQLUsername) *tree.DOid {
	__antithesis_instrumentation__.Notify(558733)
	h.writeTypeTag(userTypeTag)
	h.writeStr(username.Normalized())
	return h.getOid()
}

func (h oidHasher) CollationOid(collation string) *tree.DOid {
	__antithesis_instrumentation__.Notify(558734)
	h.writeTypeTag(collationTypeTag)
	h.writeStr(collation)
	return h.getOid()
}

func (h oidHasher) OperatorOid(name string, leftType, rightType, returnType *tree.DOid) *tree.DOid {
	__antithesis_instrumentation__.Notify(558735)
	h.writeTypeTag(operatorTypeTag)
	h.writeStr(name)
	h.writeOID(leftType)
	h.writeOID(rightType)
	h.writeOID(returnType)
	return h.getOid()
}

func (h oidHasher) EnumEntryOid(typOID *tree.DOid, physicalRep []byte) *tree.DOid {
	__antithesis_instrumentation__.Notify(558736)
	h.writeTypeTag(enumEntryTypeTag)
	h.writeOID(typOID)
	h.writeBytes(physicalRep)
	return h.getOid()
}

func (h oidHasher) rewriteOid(source descpb.ID, depended descpb.ID) *tree.DOid {
	__antithesis_instrumentation__.Notify(558737)
	h.writeTypeTag(rewriteTypeTag)
	h.writeUInt32(uint32(source))
	h.writeUInt32(uint32(depended))
	return h.getOid()
}

func (h oidHasher) DBSchemaRoleOid(
	dbID descpb.ID, scName string, normalizedRole string,
) *tree.DOid {
	__antithesis_instrumentation__.Notify(558738)
	h.writeTypeTag(dbSchemaRoleTypeTag)
	h.writeDB(dbID)
	h.writeSchema(scName)
	h.writeStr(normalizedRole)
	return h.getOid()
}

func tableOid(id descpb.ID) *tree.DOid {
	__antithesis_instrumentation__.Notify(558739)
	return tree.NewDOid(tree.DInt(id))
}

func dbOid(id descpb.ID) *tree.DOid {
	__antithesis_instrumentation__.Notify(558740)
	return tree.NewDOid(tree.DInt(id))
}

func stringOid(s string) *tree.DOid {
	__antithesis_instrumentation__.Notify(558741)
	h := makeOidHasher()
	h.writeStr(s)
	return h.getOid()
}
