package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"go/constant"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/lib/pq/oid"
)

type createTableNode struct {
	n          *tree.CreateTable
	dbDesc     catalog.DatabaseDescriptor
	sourcePlan planNode
}

func (n *createTableNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(463768) }

func (p *planner) getNonTemporarySchemaForCreate(
	ctx context.Context, db catalog.DatabaseDescriptor, scName string,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(463769)
	res, err := p.Descriptors().GetMutableSchemaByName(
		ctx, p.txn, db, scName, tree.SchemaLookupFlags{
			Required:       true,
			RequireMutable: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(463771)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463772)
	}
	__antithesis_instrumentation__.Notify(463770)
	switch res.SchemaKind() {
	case catalog.SchemaPublic, catalog.SchemaUserDefined:
		__antithesis_instrumentation__.Notify(463773)
		return res, nil
	case catalog.SchemaVirtual:
		__antithesis_instrumentation__.Notify(463774)
		return nil, pgerror.Newf(pgcode.InsufficientPrivilege, "schema cannot be modified: %q", scName)
	default:
		__antithesis_instrumentation__.Notify(463775)
		return nil, errors.AssertionFailedf(
			"invalid schema kind for getNonTemporarySchemaForCreate: %d", res.SchemaKind())
	}
}

func getSchemaForCreateTable(
	params runParams,
	db catalog.DatabaseDescriptor,
	persistence tree.Persistence,
	tableName *tree.TableName,
	kind tree.RequiredTableKind,
	ifNotExists bool,
) (schema catalog.SchemaDescriptor, err error) {
	__antithesis_instrumentation__.Notify(463776)

	if tableName.Schema() == tree.PublicSchema {
		__antithesis_instrumentation__.Notify(463783)
		if _, ok := types.PublicSchemaAliases[tableName.Object()]; ok {
			__antithesis_instrumentation__.Notify(463784)
			return nil, sqlerrors.NewTypeAlreadyExistsError(tableName.String())
		} else {
			__antithesis_instrumentation__.Notify(463785)
		}
	} else {
		__antithesis_instrumentation__.Notify(463786)
	}
	__antithesis_instrumentation__.Notify(463777)

	if persistence.IsTemporary() {
		__antithesis_instrumentation__.Notify(463787)
		if !params.SessionData().TempTablesEnabled {
			__antithesis_instrumentation__.Notify(463789)
			return nil, errors.WithTelemetry(
				pgerror.WithCandidateCode(
					errors.WithHint(
						errors.WithIssueLink(
							errors.Newf("temporary tables are only supported experimentally"),
							errors.IssueLink{IssueURL: build.MakeIssueURL(46260)},
						),
						"You can enable temporary tables by running `SET experimental_enable_temp_tables = 'on'`.",
					),
					pgcode.FeatureNotSupported,
				),
				"sql.schema.temp_tables_disabled",
			)
		} else {
			__antithesis_instrumentation__.Notify(463790)
		}
		__antithesis_instrumentation__.Notify(463788)

		var err error
		schema, err = params.p.getOrCreateTemporarySchema(params.ctx, db)
		if err != nil {
			__antithesis_instrumentation__.Notify(463791)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463792)
		}
	} else {
		__antithesis_instrumentation__.Notify(463793)

		var err error
		schema, err = params.p.getNonTemporarySchemaForCreate(params.ctx, db, tableName.Schema())
		if err != nil {
			__antithesis_instrumentation__.Notify(463795)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463796)
		}
		__antithesis_instrumentation__.Notify(463794)
		if schema.SchemaKind() == catalog.SchemaUserDefined {
			__antithesis_instrumentation__.Notify(463797)
			sqltelemetry.IncrementUserDefinedSchemaCounter(sqltelemetry.UserDefinedSchemaUsedByObject)
		} else {
			__antithesis_instrumentation__.Notify(463798)
		}
	}
	__antithesis_instrumentation__.Notify(463778)

	if persistence.IsUnlogged() {
		__antithesis_instrumentation__.Notify(463799)
		telemetry.Inc(sqltelemetry.CreateUnloggedTableCounter)
		params.p.BufferClientNotice(
			params.ctx,
			pgnotice.Newf("UNLOGGED TABLE will behave as a regular table in CockroachDB"),
		)
	} else {
		__antithesis_instrumentation__.Notify(463800)
	}
	__antithesis_instrumentation__.Notify(463779)

	if err := params.p.canCreateOnSchema(
		params.ctx, schema.GetID(), db.GetID(), params.p.User(), skipCheckPublicSchema); err != nil {
		__antithesis_instrumentation__.Notify(463801)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463802)
	}
	__antithesis_instrumentation__.Notify(463780)

	desc, err := params.p.Descriptors().Direct().GetDescriptorCollidingWithObject(
		params.ctx,
		params.p.txn,
		db.GetID(),
		schema.GetID(),
		tableName.Table(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(463803)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463804)
	}
	__antithesis_instrumentation__.Notify(463781)
	if desc != nil {
		__antithesis_instrumentation__.Notify(463805)

		{
			__antithesis_instrumentation__.Notify(463808)
			mismatchedType := true
			if tableDescriptor, ok := desc.(catalog.TableDescriptor); ok {
				__antithesis_instrumentation__.Notify(463810)
				mismatchedType = false
				switch kind {
				case tree.ResolveRequireTableDesc:
					__antithesis_instrumentation__.Notify(463811)
					mismatchedType = !tableDescriptor.IsTable()
				case tree.ResolveRequireViewDesc:
					__antithesis_instrumentation__.Notify(463812)
					mismatchedType = !tableDescriptor.IsView()
				case tree.ResolveRequireSequenceDesc:
					__antithesis_instrumentation__.Notify(463813)
					mismatchedType = !tableDescriptor.IsSequence()
				default:
					__antithesis_instrumentation__.Notify(463814)
				}

			} else {
				__antithesis_instrumentation__.Notify(463815)
			}
			__antithesis_instrumentation__.Notify(463809)

			if mismatchedType && func() bool {
				__antithesis_instrumentation__.Notify(463816)
				return ifNotExists == true
			}() == true {
				__antithesis_instrumentation__.Notify(463817)
				return nil, pgerror.Newf(pgcode.WrongObjectType,
					"%q is not a %s",
					tableName.Table(),
					kind)
			} else {
				__antithesis_instrumentation__.Notify(463818)
			}
		}
		__antithesis_instrumentation__.Notify(463806)

		if desc.Dropped() {
			__antithesis_instrumentation__.Notify(463819)
			return nil, pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"%s %q is being dropped, try again later",
				kind,
				tableName.Table())
		} else {
			__antithesis_instrumentation__.Notify(463820)
		}
		__antithesis_instrumentation__.Notify(463807)

		return schema, sqlerrors.MakeObjectAlreadyExistsError(desc.DescriptorProto(), tableName.FQString())
	} else {
		__antithesis_instrumentation__.Notify(463821)
	}
	__antithesis_instrumentation__.Notify(463782)

	return schema, nil
}
func hasPrimaryKeySerialType(params runParams, colDef *tree.ColumnTableDef) (bool, error) {
	__antithesis_instrumentation__.Notify(463822)
	if colDef.IsSerial || func() bool {
		__antithesis_instrumentation__.Notify(463825)
		return colDef.GeneratedIdentity.IsGeneratedAsIdentity == true
	}() == true {
		__antithesis_instrumentation__.Notify(463826)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(463827)
	}
	__antithesis_instrumentation__.Notify(463823)

	if funcExpr, ok := colDef.DefaultExpr.Expr.(*tree.FuncExpr); ok {
		__antithesis_instrumentation__.Notify(463828)
		var name string

		switch t := funcExpr.Func.FunctionReference.(type) {
		case *tree.FunctionDefinition:
			__antithesis_instrumentation__.Notify(463830)
			name = t.Name
		case *tree.UnresolvedName:
			__antithesis_instrumentation__.Notify(463831)
			fn, err := t.ResolveFunction(params.SessionData().SearchPath)
			if err != nil {
				__antithesis_instrumentation__.Notify(463833)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(463834)
			}
			__antithesis_instrumentation__.Notify(463832)
			name = fn.Name
		}
		__antithesis_instrumentation__.Notify(463829)

		if name == "nextval" {
			__antithesis_instrumentation__.Notify(463835)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(463836)
		}
	} else {
		__antithesis_instrumentation__.Notify(463837)
	}
	__antithesis_instrumentation__.Notify(463824)

	return false, nil
}

func (n *createTableNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(463838)
	telemetry.Inc(sqltelemetry.SchemaChangeCreateCounter("table"))

	colsWithPrimaryKeyConstraint := make(map[tree.Name]bool)

	for _, def := range n.n.Defs {
		__antithesis_instrumentation__.Notify(463853)
		switch v := def.(type) {
		case *tree.UniqueConstraintTableDef:
			__antithesis_instrumentation__.Notify(463854)
			if v.PrimaryKey {
				__antithesis_instrumentation__.Notify(463856)
				for _, indexEle := range v.IndexTableDef.Columns {
					__antithesis_instrumentation__.Notify(463857)
					colsWithPrimaryKeyConstraint[indexEle.Column] = true
				}
			} else {
				__antithesis_instrumentation__.Notify(463858)
			}

		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(463855)
			if v.PrimaryKey.IsPrimaryKey {
				__antithesis_instrumentation__.Notify(463859)
				colsWithPrimaryKeyConstraint[v.Name] = true
			} else {
				__antithesis_instrumentation__.Notify(463860)
			}
		}
	}
	__antithesis_instrumentation__.Notify(463839)

	for _, def := range n.n.Defs {
		__antithesis_instrumentation__.Notify(463861)
		switch v := def.(type) {
		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(463862)
			if _, ok := colsWithPrimaryKeyConstraint[v.Name]; ok {
				__antithesis_instrumentation__.Notify(463863)
				primaryKeySerial, err := hasPrimaryKeySerialType(params, v)
				if err != nil {
					__antithesis_instrumentation__.Notify(463865)
					return err
				} else {
					__antithesis_instrumentation__.Notify(463866)
				}
				__antithesis_instrumentation__.Notify(463864)

				if primaryKeySerial {
					__antithesis_instrumentation__.Notify(463867)
					params.p.BufferClientNotice(
						params.ctx,
						pgnotice.Newf("using sequential values in a primary key does not perform as well as using random UUIDs. See %s", docs.URL("serial.html")),
					)
					break
				} else {
					__antithesis_instrumentation__.Notify(463868)
				}
			} else {
				__antithesis_instrumentation__.Notify(463869)
			}
		}
	}
	__antithesis_instrumentation__.Notify(463840)

	schema, err := getSchemaForCreateTable(params, n.dbDesc, n.n.Persistence, &n.n.Table,
		tree.ResolveRequireTableDesc, n.n.IfNotExists)
	if err != nil {
		__antithesis_instrumentation__.Notify(463870)
		if sqlerrors.IsRelationAlreadyExistsError(err) && func() bool {
			__antithesis_instrumentation__.Notify(463872)
			return n.n.IfNotExists == true
		}() == true {
			__antithesis_instrumentation__.Notify(463873)
			params.p.BufferClientNotice(
				params.ctx,
				pgnotice.Newf("relation %q already exists, skipping", n.n.Table.Table()),
			)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(463874)
		}
		__antithesis_instrumentation__.Notify(463871)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463875)
	}
	__antithesis_instrumentation__.Notify(463841)
	if n.n.Persistence.IsTemporary() {
		__antithesis_instrumentation__.Notify(463876)
		telemetry.Inc(sqltelemetry.CreateTempTableCounter)

		switch n.n.OnCommit {
		case tree.CreateTableOnCommitUnset, tree.CreateTableOnCommitPreserveRows:
			__antithesis_instrumentation__.Notify(463877)
		default:
			__antithesis_instrumentation__.Notify(463878)
			return errors.AssertionFailedf("ON COMMIT value %d is unrecognized", n.n.OnCommit)
		}
	} else {
		__antithesis_instrumentation__.Notify(463879)
		if n.n.OnCommit != tree.CreateTableOnCommitUnset {
			__antithesis_instrumentation__.Notify(463880)
			return pgerror.Newf(
				pgcode.InvalidTableDefinition,
				"ON COMMIT can only be used on temporary tables",
			)
		} else {
			__antithesis_instrumentation__.Notify(463881)
		}
	}
	__antithesis_instrumentation__.Notify(463842)

	if n.n.PartitionByTable.ContainsPartitions() {
		__antithesis_instrumentation__.Notify(463882)
		for _, def := range n.n.Defs {
			__antithesis_instrumentation__.Notify(463883)
			if d, ok := def.(*tree.IndexTableDef); ok {
				__antithesis_instrumentation__.Notify(463884)
				if d.PartitionByIndex == nil && func() bool {
					__antithesis_instrumentation__.Notify(463885)
					return !n.n.PartitionByTable.All == true
				}() == true {
					__antithesis_instrumentation__.Notify(463886)
					params.p.BufferClientNotice(
						params.ctx,
						errors.WithHint(
							pgnotice.Newf("creating non-partitioned index on partitioned table may not be performant"),
							"Consider modifying the index such that it is also partitioned.",
						),
					)
				} else {
					__antithesis_instrumentation__.Notify(463887)
				}
			} else {
				__antithesis_instrumentation__.Notify(463888)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(463889)
	}
	__antithesis_instrumentation__.Notify(463843)

	id, err := descidgen.GenerateUniqueDescID(params.ctx, params.p.ExecCfg().DB, params.p.ExecCfg().Codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(463890)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463891)
	}
	__antithesis_instrumentation__.Notify(463844)

	var desc *tabledesc.Mutable
	var affected map[descpb.ID]*tabledesc.Mutable

	var creationTime hlc.Timestamp
	privs := catprivilege.CreatePrivilegesFromDefaultPrivileges(
		n.dbDesc.GetDefaultPrivilegeDescriptor(),
		schema.GetDefaultPrivilegeDescriptor(),
		n.dbDesc.GetID(),
		params.SessionData().User(),
		tree.Tables,
		n.dbDesc.GetPrivileges(),
	)
	if n.n.As() {
		__antithesis_instrumentation__.Notify(463892)
		asCols := planColumns(n.sourcePlan)
		if !n.n.AsHasUserSpecifiedPrimaryKey() {
			__antithesis_instrumentation__.Notify(463895)

			asCols = asCols[:len(asCols)-1]
		} else {
			__antithesis_instrumentation__.Notify(463896)
		}
		__antithesis_instrumentation__.Notify(463893)
		desc, err = newTableDescIfAs(
			params, n.n, n.dbDesc, schema, id, creationTime, asCols, privs, params.p.EvalContext(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(463897)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463898)
		}
		__antithesis_instrumentation__.Notify(463894)

		if params.extendedEvalCtx.TxnIsSingleStmt {
			__antithesis_instrumentation__.Notify(463899)
			desc.State = descpb.DescriptorState_ADD
		} else {
			__antithesis_instrumentation__.Notify(463900)
		}
	} else {
		__antithesis_instrumentation__.Notify(463901)
		affected = make(map[descpb.ID]*tabledesc.Mutable)
		desc, err = newTableDesc(params, n.n, n.dbDesc, schema, id, creationTime, privs, affected)
		if err != nil {
			__antithesis_instrumentation__.Notify(463903)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463904)
		}
		__antithesis_instrumentation__.Notify(463902)

		if desc.Adding() {
			__antithesis_instrumentation__.Notify(463905)

			refs, err := desc.FindAllReferences()
			if err != nil {
				__antithesis_instrumentation__.Notify(463908)
				return err
			} else {
				__antithesis_instrumentation__.Notify(463909)
			}
			__antithesis_instrumentation__.Notify(463906)
			var foundExternalReference bool
			for id := range refs {
				__antithesis_instrumentation__.Notify(463910)
				if t, err := params.p.Descriptors().GetUncommittedMutableTableByID(id); err != nil {
					__antithesis_instrumentation__.Notify(463911)
					return err
				} else {
					__antithesis_instrumentation__.Notify(463912)
					if t == nil || func() bool {
						__antithesis_instrumentation__.Notify(463913)
						return !t.IsNew() == true
					}() == true {
						__antithesis_instrumentation__.Notify(463914)
						foundExternalReference = true
						break
					} else {
						__antithesis_instrumentation__.Notify(463915)
					}
				}
			}
			__antithesis_instrumentation__.Notify(463907)
			if !foundExternalReference {
				__antithesis_instrumentation__.Notify(463916)
				desc.State = descpb.DescriptorState_PUBLIC
			} else {
				__antithesis_instrumentation__.Notify(463917)
			}
		} else {
			__antithesis_instrumentation__.Notify(463918)
		}
	}
	__antithesis_instrumentation__.Notify(463845)

	if err := params.p.createDescriptorWithID(
		params.ctx,
		catalogkeys.MakeObjectNameKey(params.ExecCfg().Codec, n.dbDesc.GetID(), schema.GetID(), n.n.Table.Table()),
		id,
		desc,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(463919)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463920)
	}
	__antithesis_instrumentation__.Notify(463846)

	for _, updated := range affected {
		__antithesis_instrumentation__.Notify(463921)
		if err := params.p.writeSchemaChange(
			params.ctx, updated, descpb.InvalidMutationID,
			fmt.Sprintf("updating referenced FK table %s(%d) for table %s(%d)",
				updated.Name, updated.ID, desc.Name, desc.ID,
			),
		); err != nil {
			__antithesis_instrumentation__.Notify(463922)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463923)
		}
	}
	__antithesis_instrumentation__.Notify(463847)

	if err := params.p.addBackRefsFromAllTypesInTable(params.ctx, desc); err != nil {
		__antithesis_instrumentation__.Notify(463924)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463925)
	}
	__antithesis_instrumentation__.Notify(463848)

	if err := validateDescriptor(params.ctx, params.p, desc); err != nil {
		__antithesis_instrumentation__.Notify(463926)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463927)
	}
	__antithesis_instrumentation__.Notify(463849)

	if desc.LocalityConfig != nil {
		__antithesis_instrumentation__.Notify(463928)
		_, dbDesc, err := params.p.Descriptors().GetImmutableDatabaseByID(
			params.ctx,
			params.p.txn,
			desc.ParentID,
			tree.DatabaseLookupFlags{
				Required:    true,
				AvoidLeased: true,
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(463932)
			return errors.Wrap(err, "error resolving database for multi-region")
		} else {
			__antithesis_instrumentation__.Notify(463933)
		}
		__antithesis_instrumentation__.Notify(463929)

		regionConfig, err := SynthesizeRegionConfig(params.ctx, params.p.txn, dbDesc.GetID(), params.p.Descriptors())
		if err != nil {
			__antithesis_instrumentation__.Notify(463934)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463935)
		}
		__antithesis_instrumentation__.Notify(463930)

		if err := ApplyZoneConfigForMultiRegionTable(
			params.ctx,
			params.p.txn,
			params.p.ExecCfg(),
			regionConfig,
			desc,
			ApplyZoneConfigForMultiRegionTableOptionTableAndIndexes,
		); err != nil {
			__antithesis_instrumentation__.Notify(463936)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463937)
		}
		__antithesis_instrumentation__.Notify(463931)

		if desc.GetMultiRegionEnumDependencyIfExists() {
			__antithesis_instrumentation__.Notify(463938)
			regionEnumID, err := dbDesc.MultiRegionEnumID()
			if err != nil {
				__antithesis_instrumentation__.Notify(463941)
				return err
			} else {
				__antithesis_instrumentation__.Notify(463942)
			}
			__antithesis_instrumentation__.Notify(463939)
			typeDesc, err := params.p.Descriptors().GetMutableTypeVersionByID(
				params.ctx,
				params.p.txn,
				regionEnumID,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(463943)
				return errors.Wrap(err, "error resolving multi-region enum")
			} else {
				__antithesis_instrumentation__.Notify(463944)
			}
			__antithesis_instrumentation__.Notify(463940)
			typeDesc.AddReferencingDescriptorID(desc.ID)
			err = params.p.writeTypeSchemaChange(
				params.ctx, typeDesc, "add REGIONAL BY TABLE back reference")
			if err != nil {
				__antithesis_instrumentation__.Notify(463945)
				return errors.Wrap(err, "error adding backreference to multi-region enum")
			} else {
				__antithesis_instrumentation__.Notify(463946)
			}
		} else {
			__antithesis_instrumentation__.Notify(463947)
		}
	} else {
		__antithesis_instrumentation__.Notify(463948)
	}
	__antithesis_instrumentation__.Notify(463850)

	if err := params.p.logEvent(params.ctx,
		desc.ID,
		&eventpb.CreateTable{
			TableName: n.n.Table.FQString(),
		}); err != nil {
		__antithesis_instrumentation__.Notify(463949)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463950)
	}
	__antithesis_instrumentation__.Notify(463851)

	if n.n.As() && func() bool {
		__antithesis_instrumentation__.Notify(463951)
		return !params.extendedEvalCtx.TxnIsSingleStmt == true
	}() == true {
		__antithesis_instrumentation__.Notify(463952)
		err = func() error {
			__antithesis_instrumentation__.Notify(463954)

			prevMode := params.p.Txn().ConfigureStepping(params.ctx, kv.SteppingEnabled)
			defer func() {
				__antithesis_instrumentation__.Notify(463960)
				_ = params.p.Txn().ConfigureStepping(params.ctx, prevMode)
			}()
			__antithesis_instrumentation__.Notify(463955)

			internal := params.p.SessionData().Internal
			ri, err := row.MakeInserter(
				params.ctx,
				params.p.txn,
				params.ExecCfg().Codec,
				desc.ImmutableCopy().(catalog.TableDescriptor),
				desc.PublicColumns(),
				params.p.alloc,
				&params.ExecCfg().Settings.SV,
				internal,
				params.ExecCfg().GetRowMetrics(internal),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(463961)
				return err
			} else {
				__antithesis_instrumentation__.Notify(463962)
			}
			__antithesis_instrumentation__.Notify(463956)
			ti := tableInserterPool.Get().(*tableInserter)
			*ti = tableInserter{ri: ri}
			tw := tableWriter(ti)
			defer func() {
				__antithesis_instrumentation__.Notify(463963)
				tw.close(params.ctx)
				*ti = tableInserter{}
				tableInserterPool.Put(ti)
			}()
			__antithesis_instrumentation__.Notify(463957)
			if err := tw.init(params.ctx, params.p.txn, params.p.EvalContext(), &params.p.EvalContext().Settings.SV); err != nil {
				__antithesis_instrumentation__.Notify(463964)
				return err
			} else {
				__antithesis_instrumentation__.Notify(463965)
			}
			__antithesis_instrumentation__.Notify(463958)

			rowBuffer := make(tree.Datums, len(desc.Columns))

			for {
				__antithesis_instrumentation__.Notify(463966)
				if err := params.p.cancelChecker.Check(); err != nil {
					__antithesis_instrumentation__.Notify(463969)
					return err
				} else {
					__antithesis_instrumentation__.Notify(463970)
				}
				__antithesis_instrumentation__.Notify(463967)
				if next, err := n.sourcePlan.Next(params); !next {
					__antithesis_instrumentation__.Notify(463971)
					if err != nil {
						__antithesis_instrumentation__.Notify(463974)
						return err
					} else {
						__antithesis_instrumentation__.Notify(463975)
					}
					__antithesis_instrumentation__.Notify(463972)
					if err := tw.finalize(params.ctx); err != nil {
						__antithesis_instrumentation__.Notify(463976)
						return err
					} else {
						__antithesis_instrumentation__.Notify(463977)
					}
					__antithesis_instrumentation__.Notify(463973)
					break
				} else {
					__antithesis_instrumentation__.Notify(463978)
				}
				__antithesis_instrumentation__.Notify(463968)

				copy(rowBuffer, n.sourcePlan.Values())

				var pm row.PartialIndexUpdateHelper
				if err := tw.row(params.ctx, rowBuffer, pm, params.extendedEvalCtx.Tracing.KVTracingEnabled()); err != nil {
					__antithesis_instrumentation__.Notify(463979)
					return err
				} else {
					__antithesis_instrumentation__.Notify(463980)
				}
			}
			__antithesis_instrumentation__.Notify(463959)
			return nil
		}()
		__antithesis_instrumentation__.Notify(463953)
		if err != nil {
			__antithesis_instrumentation__.Notify(463981)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463982)
		}
	} else {
		__antithesis_instrumentation__.Notify(463983)
	}
	__antithesis_instrumentation__.Notify(463852)

	return nil
}

func (*createTableNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(463984)
	return false, nil
}
func (*createTableNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(463985)
	return tree.Datums{}
}

func (n *createTableNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(463986)
	if n.sourcePlan != nil {
		__antithesis_instrumentation__.Notify(463987)
		n.sourcePlan.Close(ctx)
		n.sourcePlan = nil
	} else {
		__antithesis_instrumentation__.Notify(463988)
	}
}

func qualifyFKColErrorWithDB(
	ctx context.Context,
	db catalog.DatabaseDescriptor,
	sc catalog.SchemaDescriptor,
	tbl catalog.TableDescriptor,
	col string,
) string {
	__antithesis_instrumentation__.Notify(463989)
	return tree.ErrString(tree.NewUnresolvedName(
		db.GetName(),
		sc.GetName(),
		tbl.GetName(),
		col,
	))
}

type TableState int

const (
	NewTable TableState = iota

	EmptyTable

	NonEmptyTable
)

func addUniqueWithoutIndexColumnTableDef(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	sessionData *sessiondata.SessionData,
	d *tree.ColumnTableDef,
	desc *tabledesc.Mutable,
	ts TableState,
	validationBehavior tree.ValidationBehavior,
) error {
	__antithesis_instrumentation__.Notify(463990)
	if !sessionData.EnableUniqueWithoutIndexConstraints {
		__antithesis_instrumentation__.Notify(463993)
		return pgerror.New(pgcode.FeatureNotSupported,
			"unique constraints without an index are not yet supported",
		)
	} else {
		__antithesis_instrumentation__.Notify(463994)
	}
	__antithesis_instrumentation__.Notify(463991)

	if err := ResolveUniqueWithoutIndexConstraint(
		ctx,
		desc,
		string(d.Unique.ConstraintName),
		[]string{string(d.Name)},
		"",
		ts,
		validationBehavior,
	); err != nil {
		__antithesis_instrumentation__.Notify(463995)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463996)
	}
	__antithesis_instrumentation__.Notify(463992)
	return nil
}

func addUniqueWithoutIndexTableDef(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	sessionData *sessiondata.SessionData,
	d *tree.UniqueConstraintTableDef,
	desc *tabledesc.Mutable,
	tn tree.TableName,
	ts TableState,
	validationBehavior tree.ValidationBehavior,
	semaCtx *tree.SemaContext,
) error {
	__antithesis_instrumentation__.Notify(463997)
	if !sessionData.EnableUniqueWithoutIndexConstraints {
		__antithesis_instrumentation__.Notify(464004)
		return pgerror.New(pgcode.FeatureNotSupported,
			"unique constraints without an index are not yet supported",
		)
	} else {
		__antithesis_instrumentation__.Notify(464005)
	}
	__antithesis_instrumentation__.Notify(463998)
	if len(d.Storing) > 0 {
		__antithesis_instrumentation__.Notify(464006)
		return pgerror.New(pgcode.FeatureNotSupported,
			"unique constraints without an index cannot store columns",
		)
	} else {
		__antithesis_instrumentation__.Notify(464007)
	}
	__antithesis_instrumentation__.Notify(463999)
	if d.PartitionByIndex.ContainsPartitions() {
		__antithesis_instrumentation__.Notify(464008)
		return pgerror.New(pgcode.FeatureNotSupported,
			"partitioned unique constraints without an index are not supported",
		)
	} else {
		__antithesis_instrumentation__.Notify(464009)
	}
	__antithesis_instrumentation__.Notify(464000)

	var predicate string
	if d.Predicate != nil {
		__antithesis_instrumentation__.Notify(464010)
		var err error
		predicate, err = schemaexpr.ValidateUniqueWithoutIndexPredicate(
			ctx, tn, desc, d.Predicate, semaCtx,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(464011)
			return err
		} else {
			__antithesis_instrumentation__.Notify(464012)
		}
	} else {
		__antithesis_instrumentation__.Notify(464013)
	}
	__antithesis_instrumentation__.Notify(464001)

	colNames := make([]string, len(d.Columns))
	for i := range colNames {
		__antithesis_instrumentation__.Notify(464014)
		colNames[i] = string(d.Columns[i].Column)
	}
	__antithesis_instrumentation__.Notify(464002)
	if err := ResolveUniqueWithoutIndexConstraint(
		ctx, desc, string(d.Name), colNames, predicate, ts, validationBehavior,
	); err != nil {
		__antithesis_instrumentation__.Notify(464015)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464016)
	}
	__antithesis_instrumentation__.Notify(464003)
	return nil
}

func ResolveUniqueWithoutIndexConstraint(
	ctx context.Context,
	tbl *tabledesc.Mutable,
	constraintName string,
	colNames []string,
	predicate string,
	ts TableState,
	validationBehavior tree.ValidationBehavior,
) error {
	__antithesis_instrumentation__.Notify(464017)
	var colSet catalog.TableColSet
	cols := make([]catalog.Column, len(colNames))
	for i, name := range colNames {
		__antithesis_instrumentation__.Notify(464024)
		col, err := tbl.FindActiveOrNewColumnByName(tree.Name(name))
		if err != nil {
			__antithesis_instrumentation__.Notify(464027)
			return err
		} else {
			__antithesis_instrumentation__.Notify(464028)
		}
		__antithesis_instrumentation__.Notify(464025)

		if colSet.Contains(col.GetID()) {
			__antithesis_instrumentation__.Notify(464029)
			return pgerror.Newf(pgcode.DuplicateColumn,
				"column %q appears twice in unique constraint", col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(464030)
		}
		__antithesis_instrumentation__.Notify(464026)
		colSet.Add(col.GetID())
		cols[i] = col
	}
	__antithesis_instrumentation__.Notify(464018)

	constraintInfo, err := tbl.GetConstraintInfo()
	if err != nil {
		__antithesis_instrumentation__.Notify(464031)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464032)
	}
	__antithesis_instrumentation__.Notify(464019)
	if constraintName == "" {
		__antithesis_instrumentation__.Notify(464033)
		constraintName = tabledesc.GenerateUniqueName(
			fmt.Sprintf("unique_%s", strings.Join(colNames, "_")),
			func(p string) bool {
				__antithesis_instrumentation__.Notify(464034)
				_, ok := constraintInfo[p]
				return ok
			},
		)
	} else {
		__antithesis_instrumentation__.Notify(464035)
		if _, ok := constraintInfo[constraintName]; ok {
			__antithesis_instrumentation__.Notify(464036)
			return pgerror.Newf(pgcode.DuplicateObject, "duplicate constraint name: %q", constraintName)
		} else {
			__antithesis_instrumentation__.Notify(464037)
		}
	}
	__antithesis_instrumentation__.Notify(464020)

	columnIDs := make(descpb.ColumnIDs, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(464038)
		columnIDs[i] = col.GetID()
	}
	__antithesis_instrumentation__.Notify(464021)

	validity := descpb.ConstraintValidity_Validated
	if ts != NewTable {
		__antithesis_instrumentation__.Notify(464039)
		if validationBehavior == tree.ValidationSkip {
			__antithesis_instrumentation__.Notify(464040)
			validity = descpb.ConstraintValidity_Unvalidated
		} else {
			__antithesis_instrumentation__.Notify(464041)
			validity = descpb.ConstraintValidity_Validating
		}
	} else {
		__antithesis_instrumentation__.Notify(464042)
	}
	__antithesis_instrumentation__.Notify(464022)

	uc := descpb.UniqueWithoutIndexConstraint{
		Name:         constraintName,
		TableID:      tbl.ID,
		ColumnIDs:    columnIDs,
		Predicate:    predicate,
		Validity:     validity,
		ConstraintID: tbl.NextConstraintID,
	}
	tbl.NextConstraintID++
	if ts == NewTable {
		__antithesis_instrumentation__.Notify(464043)
		tbl.UniqueWithoutIndexConstraints = append(tbl.UniqueWithoutIndexConstraints, uc)
	} else {
		__antithesis_instrumentation__.Notify(464044)
		tbl.AddUniqueWithoutIndexMutation(&uc, descpb.DescriptorMutation_ADD)
	}
	__antithesis_instrumentation__.Notify(464023)

	return nil
}

func ResolveFK(
	ctx context.Context,
	txn *kv.Txn,
	sc resolver.SchemaResolver,
	parentDB catalog.DatabaseDescriptor,
	parentSchema catalog.SchemaDescriptor,
	tbl *tabledesc.Mutable,
	d *tree.ForeignKeyConstraintTableDef,
	backrefs map[descpb.ID]*tabledesc.Mutable,
	ts TableState,
	validationBehavior tree.ValidationBehavior,
	evalCtx *tree.EvalContext,
) error {
	__antithesis_instrumentation__.Notify(464045)
	var originColSet catalog.TableColSet
	originCols := make([]catalog.Column, len(d.FromCols))
	for i, fromCol := range d.FromCols {
		__antithesis_instrumentation__.Notify(464065)
		col, err := tbl.FindActiveOrNewColumnByName(fromCol)
		if err != nil {
			__antithesis_instrumentation__.Notify(464069)
			return err
		} else {
			__antithesis_instrumentation__.Notify(464070)
		}
		__antithesis_instrumentation__.Notify(464066)
		if err := col.CheckCanBeOutboundFKRef(); err != nil {
			__antithesis_instrumentation__.Notify(464071)
			return err
		} else {
			__antithesis_instrumentation__.Notify(464072)
		}
		__antithesis_instrumentation__.Notify(464067)

		if originColSet.Contains(col.GetID()) {
			__antithesis_instrumentation__.Notify(464073)
			return pgerror.Newf(pgcode.InvalidForeignKey,
				"foreign key contains duplicate column %q", col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(464074)
		}
		__antithesis_instrumentation__.Notify(464068)
		originColSet.Add(col.GetID())
		originCols[i] = col
	}
	__antithesis_instrumentation__.Notify(464046)

	_, target, err := resolver.ResolveMutableExistingTableObject(ctx, sc, &d.Table, true, tree.ResolveRequireTableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(464075)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464076)
	}
	__antithesis_instrumentation__.Notify(464047)
	if target.ParentID != tbl.ParentID {
		__antithesis_instrumentation__.Notify(464077)
		if !allowCrossDatabaseFKs.Get(&evalCtx.Settings.SV) {
			__antithesis_instrumentation__.Notify(464078)
			return errors.WithHintf(
				pgerror.Newf(pgcode.InvalidForeignKey,
					"foreign references between databases are not allowed (see the '%s' cluster setting)",
					allowCrossDatabaseFKsSetting),
				crossDBReferenceDeprecationHint(),
			)
		} else {
			__antithesis_instrumentation__.Notify(464079)
		}
	} else {
		__antithesis_instrumentation__.Notify(464080)
	}
	__antithesis_instrumentation__.Notify(464048)
	if tbl.Temporary != target.Temporary {
		__antithesis_instrumentation__.Notify(464081)
		persistenceType := "permanent"
		if tbl.Temporary {
			__antithesis_instrumentation__.Notify(464083)
			persistenceType = "temporary"
		} else {
			__antithesis_instrumentation__.Notify(464084)
		}
		__antithesis_instrumentation__.Notify(464082)
		return pgerror.Newf(
			pgcode.InvalidTableDefinition,
			"constraints on %s tables may reference only %s tables",
			persistenceType,
			persistenceType,
		)
	} else {
		__antithesis_instrumentation__.Notify(464085)
	}
	__antithesis_instrumentation__.Notify(464049)
	if target.ID == tbl.ID {
		__antithesis_instrumentation__.Notify(464086)

		target = tbl
	} else {
		__antithesis_instrumentation__.Notify(464087)

		if ts == NewTable {
			__antithesis_instrumentation__.Notify(464089)
			tbl.State = descpb.DescriptorState_ADD
		} else {
			__antithesis_instrumentation__.Notify(464090)
		}
		__antithesis_instrumentation__.Notify(464088)

		if prev, ok := backrefs[target.ID]; ok {
			__antithesis_instrumentation__.Notify(464091)
			target = prev
		} else {
			__antithesis_instrumentation__.Notify(464092)
			backrefs[target.ID] = target
		}
	}
	__antithesis_instrumentation__.Notify(464050)

	referencedColNames := d.ToCols

	if len(referencedColNames) == 0 {
		__antithesis_instrumentation__.Notify(464093)
		numImplicitCols := target.GetPrimaryIndex().GetPartitioning().NumImplicitColumns()
		referencedColNames = make(
			tree.NameList,
			0,
			target.GetPrimaryIndex().NumKeyColumns()-numImplicitCols,
		)
		for i := numImplicitCols; i < target.GetPrimaryIndex().NumKeyColumns(); i++ {
			__antithesis_instrumentation__.Notify(464094)
			referencedColNames = append(
				referencedColNames,
				tree.Name(target.GetPrimaryIndex().GetKeyColumnName(i)),
			)
		}
	} else {
		__antithesis_instrumentation__.Notify(464095)
	}
	__antithesis_instrumentation__.Notify(464051)

	referencedCols, err := tabledesc.FindPublicColumnsWithNames(target, referencedColNames)
	if err != nil {
		__antithesis_instrumentation__.Notify(464096)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464097)
	}
	__antithesis_instrumentation__.Notify(464052)

	for i := range referencedCols {
		__antithesis_instrumentation__.Notify(464098)
		if err := referencedCols[i].CheckCanBeInboundFKRef(); err != nil {
			__antithesis_instrumentation__.Notify(464099)
			return err
		} else {
			__antithesis_instrumentation__.Notify(464100)
		}
	}
	__antithesis_instrumentation__.Notify(464053)

	if len(referencedCols) != len(originCols) {
		__antithesis_instrumentation__.Notify(464101)
		return pgerror.Newf(pgcode.Syntax,
			"%d columns must reference exactly %d columns in referenced table (found %d)",
			len(originCols), len(originCols), len(referencedCols))
	} else {
		__antithesis_instrumentation__.Notify(464102)
	}
	__antithesis_instrumentation__.Notify(464054)

	for i := range originCols {
		__antithesis_instrumentation__.Notify(464103)
		if s, t := originCols[i], referencedCols[i]; !s.GetType().Equivalent(t.GetType()) {
			__antithesis_instrumentation__.Notify(464105)
			return pgerror.Newf(pgcode.DatatypeMismatch,
				"type of %q (%s) does not match foreign key %q.%q (%s)",
				s.GetName(), s.GetType().String(), target.Name, t.GetName(), t.GetType().String())
		} else {
			__antithesis_instrumentation__.Notify(464106)
		}
		__antithesis_instrumentation__.Notify(464104)

		if s, t := originCols[i], referencedCols[i]; !s.GetType().Identical(t.GetType()) {
			__antithesis_instrumentation__.Notify(464107)
			notice := pgnotice.Newf(
				"type of foreign key column %q (%s) is not identical to referenced column %q.%q (%s)",
				s.ColName(), s.GetType().String(), target.Name, t.GetName(), t.GetType().String())
			evalCtx.ClientNoticeSender.BufferClientNotice(ctx, notice)
		} else {
			__antithesis_instrumentation__.Notify(464108)
		}
	}
	__antithesis_instrumentation__.Notify(464055)

	constraintInfo, err := tbl.GetConstraintInfo()
	if err != nil {
		__antithesis_instrumentation__.Notify(464109)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464110)
	}
	__antithesis_instrumentation__.Notify(464056)
	constraintName := string(d.Name)
	if constraintName == "" {
		__antithesis_instrumentation__.Notify(464111)
		constraintName = tabledesc.GenerateUniqueName(
			tabledesc.ForeignKeyConstraintName(tbl.GetName(), d.FromCols.ToStrings()),
			func(p string) bool {
				__antithesis_instrumentation__.Notify(464112)
				_, ok := constraintInfo[p]
				return ok
			},
		)
	} else {
		__antithesis_instrumentation__.Notify(464113)
		if _, ok := constraintInfo[constraintName]; ok {
			__antithesis_instrumentation__.Notify(464114)
			return pgerror.Newf(pgcode.DuplicateObject, "duplicate constraint name: %q", constraintName)
		} else {
			__antithesis_instrumentation__.Notify(464115)
		}
	}
	__antithesis_instrumentation__.Notify(464057)

	originColumnIDs := make(descpb.ColumnIDs, len(originCols))
	for i, col := range originCols {
		__antithesis_instrumentation__.Notify(464116)
		originColumnIDs[i] = col.GetID()
	}
	__antithesis_instrumentation__.Notify(464058)

	targetColIDs := make(descpb.ColumnIDs, len(referencedCols))
	for i := range referencedCols {
		__antithesis_instrumentation__.Notify(464117)
		targetColIDs[i] = referencedCols[i].GetID()
	}
	__antithesis_instrumentation__.Notify(464059)

	if d.Actions.Delete == tree.SetNull || func() bool {
		__antithesis_instrumentation__.Notify(464118)
		return d.Actions.Update == tree.SetNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(464119)
		for _, originColumn := range originCols {
			__antithesis_instrumentation__.Notify(464120)
			if !originColumn.IsNullable() {
				__antithesis_instrumentation__.Notify(464121)
				col := qualifyFKColErrorWithDB(ctx, parentDB, parentSchema, tbl, originColumn.GetName())
				return pgerror.Newf(pgcode.InvalidForeignKey,
					"cannot add a SET NULL cascading action on column %q which has a NOT NULL constraint", col,
				)
			} else {
				__antithesis_instrumentation__.Notify(464122)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(464123)
	}
	__antithesis_instrumentation__.Notify(464060)

	if d.Actions.Delete == tree.SetDefault || func() bool {
		__antithesis_instrumentation__.Notify(464124)
		return d.Actions.Update == tree.SetDefault == true
	}() == true {
		__antithesis_instrumentation__.Notify(464125)
		for _, originColumn := range originCols {
			__antithesis_instrumentation__.Notify(464126)

			if !originColumn.HasDefault() && func() bool {
				__antithesis_instrumentation__.Notify(464127)
				return !originColumn.IsNullable() == true
			}() == true {
				__antithesis_instrumentation__.Notify(464128)
				col := qualifyFKColErrorWithDB(ctx, parentDB, parentSchema, tbl, originColumn.GetName())
				return pgerror.Newf(pgcode.InvalidForeignKey,
					"cannot add a SET DEFAULT cascading action on column %q which has a "+
						"NOT NULL constraint and a NULL default expression", col,
				)
			} else {
				__antithesis_instrumentation__.Notify(464129)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(464130)
	}
	__antithesis_instrumentation__.Notify(464061)

	_, err = tabledesc.FindFKReferencedUniqueConstraint(target, targetColIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(464131)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464132)
	}
	__antithesis_instrumentation__.Notify(464062)

	var validity descpb.ConstraintValidity
	if ts != NewTable {
		__antithesis_instrumentation__.Notify(464133)
		if validationBehavior == tree.ValidationSkip {
			__antithesis_instrumentation__.Notify(464134)
			validity = descpb.ConstraintValidity_Unvalidated
		} else {
			__antithesis_instrumentation__.Notify(464135)
			validity = descpb.ConstraintValidity_Validating
		}
	} else {
		__antithesis_instrumentation__.Notify(464136)
	}
	__antithesis_instrumentation__.Notify(464063)

	ref := descpb.ForeignKeyConstraint{
		OriginTableID:       tbl.ID,
		OriginColumnIDs:     originColumnIDs,
		ReferencedColumnIDs: targetColIDs,
		ReferencedTableID:   target.ID,
		Name:                constraintName,
		Validity:            validity,
		OnDelete:            descpb.ForeignKeyReferenceActionValue[d.Actions.Delete],
		OnUpdate:            descpb.ForeignKeyReferenceActionValue[d.Actions.Update],
		Match:               descpb.CompositeKeyMatchMethodValue[d.Match],
		ConstraintID:        tbl.NextConstraintID,
	}
	tbl.NextConstraintID++
	if ts == NewTable {
		__antithesis_instrumentation__.Notify(464137)
		tbl.OutboundFKs = append(tbl.OutboundFKs, ref)
		target.InboundFKs = append(target.InboundFKs, ref)
	} else {
		__antithesis_instrumentation__.Notify(464138)
		tbl.AddForeignKeyMutation(&ref, descpb.DescriptorMutation_ADD)
	}
	__antithesis_instrumentation__.Notify(464064)

	return nil
}

func CreatePartitioning(
	ctx context.Context,
	st *cluster.Settings,
	evalCtx *tree.EvalContext,
	tableDesc catalog.TableDescriptor,
	indexDesc descpb.IndexDescriptor,
	partBy *tree.PartitionBy,
	allowedNewColumnNames []tree.Name,
	allowImplicitPartitioning bool,
) (newImplicitCols []catalog.Column, newPartitioning catpb.PartitioningDescriptor, err error) {
	__antithesis_instrumentation__.Notify(464139)
	if partBy == nil {
		__antithesis_instrumentation__.Notify(464141)
		if indexDesc.Partitioning.NumImplicitColumns > 0 {
			__antithesis_instrumentation__.Notify(464143)
			return nil, newPartitioning, unimplemented.Newf(
				"ALTER ... PARTITION BY NOTHING",
				"cannot alter to PARTITION BY NOTHING if the object has implicit column partitioning",
			)
		} else {
			__antithesis_instrumentation__.Notify(464144)
		}
		__antithesis_instrumentation__.Notify(464142)

		return nil, newPartitioning, nil
	} else {
		__antithesis_instrumentation__.Notify(464145)
	}
	__antithesis_instrumentation__.Notify(464140)
	return CreatePartitioningCCL(
		ctx,
		st,
		evalCtx,
		tableDesc.FindColumnWithName,
		int(indexDesc.Partitioning.NumImplicitColumns),
		indexDesc.KeyColumnNames,
		partBy,
		allowedNewColumnNames,
		allowImplicitPartitioning,
	)
}

var CreatePartitioningCCL = func(
	ctx context.Context,
	st *cluster.Settings,
	evalCtx *tree.EvalContext,
	columnLookupFn func(tree.Name) (catalog.Column, error),
	oldNumImplicitColumns int,
	oldKeyColumnNames []string,
	partBy *tree.PartitionBy,
	allowedNewColumnNames []tree.Name,
	allowImplicitPartitioning bool,
) (newImplicitCols []catalog.Column, newPartitioning catpb.PartitioningDescriptor, err error) {
	__antithesis_instrumentation__.Notify(464146)
	return nil, catpb.PartitioningDescriptor{}, sqlerrors.NewCCLRequiredError(errors.New(
		"creating or manipulating partitions requires a CCL binary"))
}

func getFinalSourceQuery(
	params runParams, source *tree.Select, evalCtx *tree.EvalContext,
) (string, error) {
	__antithesis_instrumentation__.Notify(464147)

	f := evalCtx.FmtCtx(
		tree.FmtSerializable,
		tree.FmtReformatTableNames(
			func(_ *tree.FmtCtx, tn *tree.TableName) {
				__antithesis_instrumentation__.Notify(464151)

				if tn.SchemaName != "" {
					__antithesis_instrumentation__.Notify(464152)

					tn.ExplicitSchema = true
					tn.ExplicitCatalog = true
				} else {
					__antithesis_instrumentation__.Notify(464153)
				}
			}),
	)
	__antithesis_instrumentation__.Notify(464148)
	f.FormatNode(source)
	f.Close()

	ctx := evalCtx.FmtCtx(
		tree.FmtSerializable,
		tree.FmtPlaceholderFormat(func(ctx *tree.FmtCtx, placeholder *tree.Placeholder) {
			__antithesis_instrumentation__.Notify(464154)
			d, err := placeholder.Eval(evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(464156)
				panic(errors.NewAssertionErrorWithWrappedErrf(err, "failed to serialize placeholder"))
			} else {
				__antithesis_instrumentation__.Notify(464157)
			}
			__antithesis_instrumentation__.Notify(464155)
			d.Format(ctx)
		}),
	)
	__antithesis_instrumentation__.Notify(464149)
	ctx.FormatNode(source)

	sequenceReplacedQuery, err := replaceSeqNamesWithIDs(params.ctx, params.p, ctx.CloseAndGetString())
	if err != nil {
		__antithesis_instrumentation__.Notify(464158)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(464159)
	}
	__antithesis_instrumentation__.Notify(464150)
	return sequenceReplacedQuery, nil
}

func newTableDescIfAs(
	params runParams,
	p *tree.CreateTable,
	db catalog.DatabaseDescriptor,
	sc catalog.SchemaDescriptor,
	id descpb.ID,
	creationTime hlc.Timestamp,
	resultColumns []colinfo.ResultColumn,
	privileges *catpb.PrivilegeDescriptor,
	evalContext *tree.EvalContext,
) (desc *tabledesc.Mutable, err error) {
	__antithesis_instrumentation__.Notify(464160)
	if err := validateUniqueConstraintParamsForCreateTableAs(p); err != nil {
		__antithesis_instrumentation__.Notify(464166)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464167)
	}
	__antithesis_instrumentation__.Notify(464161)

	colResIndex := 0

	for _, defs := range p.Defs {
		__antithesis_instrumentation__.Notify(464168)
		var d *tree.ColumnTableDef
		var ok bool
		if d, ok = defs.(*tree.ColumnTableDef); ok {
			__antithesis_instrumentation__.Notify(464169)
			d.Type = resultColumns[colResIndex].Typ
			colResIndex++
		} else {
			__antithesis_instrumentation__.Notify(464170)
		}
	}
	__antithesis_instrumentation__.Notify(464162)

	if len(p.Defs) == 0 {
		__antithesis_instrumentation__.Notify(464171)
		for _, colRes := range resultColumns {
			__antithesis_instrumentation__.Notify(464172)
			var d *tree.ColumnTableDef
			var ok bool
			var tableDef tree.TableDef = &tree.ColumnTableDef{
				Name:   tree.Name(colRes.Name),
				Type:   colRes.Typ,
				Hidden: colRes.Hidden,
			}
			if d, ok = tableDef.(*tree.ColumnTableDef); !ok {
				__antithesis_instrumentation__.Notify(464174)
				return nil, errors.Errorf("failed to cast type to ColumnTableDef\n")
			} else {
				__antithesis_instrumentation__.Notify(464175)
			}
			__antithesis_instrumentation__.Notify(464173)
			d.Nullable.Nullability = tree.SilentNull
			p.Defs = append(p.Defs, tableDef)
		}
	} else {
		__antithesis_instrumentation__.Notify(464176)
	}
	__antithesis_instrumentation__.Notify(464163)

	desc, err = newTableDesc(
		params,
		p,
		db, sc, id,
		creationTime,
		privileges,
		nil,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(464177)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464178)
	}
	__antithesis_instrumentation__.Notify(464164)
	createQuery, err := getFinalSourceQuery(params, p.AsSource, evalContext)
	if err != nil {
		__antithesis_instrumentation__.Notify(464179)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464180)
	}
	__antithesis_instrumentation__.Notify(464165)
	desc.CreateQuery = createQuery
	return desc, nil
}

type newTableDescOptions struct {
	bypassLocalityOnNonMultiRegionDatabaseCheck bool
}

type NewTableDescOption func(o *newTableDescOptions)

func NewTableDescOptionBypassLocalityOnNonMultiRegionDatabaseCheck() NewTableDescOption {
	__antithesis_instrumentation__.Notify(464181)
	return func(o *newTableDescOptions) {
		__antithesis_instrumentation__.Notify(464182)
		o.bypassLocalityOnNonMultiRegionDatabaseCheck = true
	}
}

func NewTableDesc(
	ctx context.Context,
	txn *kv.Txn,
	vt resolver.SchemaResolver,
	st *cluster.Settings,
	n *tree.CreateTable,
	db catalog.DatabaseDescriptor,
	sc catalog.SchemaDescriptor,
	id descpb.ID,
	regionConfig *multiregion.RegionConfig,
	creationTime hlc.Timestamp,
	privileges *catpb.PrivilegeDescriptor,
	affected map[descpb.ID]*tabledesc.Mutable,
	semaCtx *tree.SemaContext,
	evalCtx *tree.EvalContext,
	sessionData *sessiondata.SessionData,
	persistence tree.Persistence,
	inOpts ...NewTableDescOption,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(464183)

	cdd := make([]*tabledesc.ColumnDefDescs, len(n.Defs))

	var opts newTableDescOptions
	for _, o := range inOpts {
		__antithesis_instrumentation__.Notify(464212)
		o(&opts)
	}
	__antithesis_instrumentation__.Notify(464184)

	var dbID descpb.ID
	if db != nil {
		__antithesis_instrumentation__.Notify(464213)
		dbID = db.GetID()
	} else {
		__antithesis_instrumentation__.Notify(464214)
	}
	__antithesis_instrumentation__.Notify(464185)
	desc := tabledesc.InitTableDescriptor(
		id, dbID, sc.GetID(), n.Table.Table(), creationTime, privileges, persistence,
	)

	if err := paramparse.SetStorageParameters(
		ctx,
		semaCtx,
		evalCtx,
		n.StorageParams,
		paramparse.NewTableStorageParamObserver(&desc),
	); err != nil {
		__antithesis_instrumentation__.Notify(464215)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464216)
	}
	__antithesis_instrumentation__.Notify(464186)

	indexEncodingVersion := descpb.StrictIndexColumnIDGuaranteesVersion
	isRegionalByRow := n.Locality != nil && func() bool {
		__antithesis_instrumentation__.Notify(464217)
		return n.Locality.LocalityLevel == tree.LocalityLevelRow == true
	}() == true

	var partitionAllBy *tree.PartitionBy
	primaryIndexColumnSet := make(map[string]struct{})

	if n.Locality != nil && func() bool {
		__antithesis_instrumentation__.Notify(464218)
		return regionConfig == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(464219)
		return !opts.bypassLocalityOnNonMultiRegionDatabaseCheck == true
	}() == true {
		__antithesis_instrumentation__.Notify(464220)
		return nil, errors.WithHint(pgerror.Newf(
			pgcode.InvalidTableDefinition,
			"cannot set LOCALITY on a table in a database that is not multi-region enabled",
		),
			"database must first be multi-region enabled using ALTER DATABASE ... SET PRIMARY REGION <region>",
		)
	} else {
		__antithesis_instrumentation__.Notify(464221)
	}
	__antithesis_instrumentation__.Notify(464187)

	if n.Locality != nil || func() bool {
		__antithesis_instrumentation__.Notify(464222)
		return regionConfig != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(464223)

		if n.PartitionByTable.ContainsPartitioningClause() {
			__antithesis_instrumentation__.Notify(464225)
			return nil, pgerror.New(
				pgcode.FeatureNotSupported,
				"multi-region tables containing PARTITION BY are not supported",
			)
		} else {
			__antithesis_instrumentation__.Notify(464226)
		}
		__antithesis_instrumentation__.Notify(464224)
		for _, def := range n.Defs {
			__antithesis_instrumentation__.Notify(464227)
			switch d := def.(type) {
			case *tree.IndexTableDef:
				__antithesis_instrumentation__.Notify(464228)
				if d.PartitionByIndex.ContainsPartitioningClause() {
					__antithesis_instrumentation__.Notify(464230)
					return nil, pgerror.New(
						pgcode.FeatureNotSupported,
						"multi-region tables with an INDEX containing PARTITION BY are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(464231)
				}
			case *tree.UniqueConstraintTableDef:
				__antithesis_instrumentation__.Notify(464229)
				if d.PartitionByIndex.ContainsPartitioningClause() {
					__antithesis_instrumentation__.Notify(464232)
					return nil, pgerror.New(
						pgcode.FeatureNotSupported,
						"multi-region tables with an UNIQUE constraint containing PARTITION BY are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(464233)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(464234)
	}
	__antithesis_instrumentation__.Notify(464188)

	if isRegionalByRow {
		__antithesis_instrumentation__.Notify(464235)
		regionalByRowCol := tree.RegionalByRowRegionDefaultColName
		if n.Locality.RegionalByRowColumn != "" {
			__antithesis_instrumentation__.Notify(464239)
			regionalByRowCol = n.Locality.RegionalByRowColumn
		} else {
			__antithesis_instrumentation__.Notify(464240)
		}
		__antithesis_instrumentation__.Notify(464236)

		regionalByRowColExists := false
		for _, def := range n.Defs {
			__antithesis_instrumentation__.Notify(464241)
			switch d := def.(type) {
			case *tree.ColumnTableDef:
				__antithesis_instrumentation__.Notify(464242)
				if d.Name == regionalByRowCol {
					__antithesis_instrumentation__.Notify(464243)
					regionalByRowColExists = true
					t, err := tree.ResolveType(ctx, d.Type, vt)
					if err != nil {
						__antithesis_instrumentation__.Notify(464246)
						return nil, errors.Wrap(err, "error resolving REGIONAL BY ROW column type")
					} else {
						__antithesis_instrumentation__.Notify(464247)
					}
					__antithesis_instrumentation__.Notify(464244)
					if t.Oid() != typedesc.TypeIDToOID(regionConfig.RegionEnumID()) {
						__antithesis_instrumentation__.Notify(464248)
						err = pgerror.Newf(
							pgcode.InvalidTableDefinition,
							"cannot use column %s which has type %s in REGIONAL BY ROW",
							d.Name,
							t.SQLString(),
						)
						if t, terr := vt.ResolveTypeByOID(
							ctx,
							typedesc.TypeIDToOID(regionConfig.RegionEnumID()),
						); terr == nil {
							__antithesis_instrumentation__.Notify(464250)
							if n.Locality.RegionalByRowColumn != tree.RegionalByRowRegionNotSpecifiedName {
								__antithesis_instrumentation__.Notify(464251)

								err = errors.WithDetailf(
									err,
									"REGIONAL BY ROW AS must reference a column of type %s",
									t.Name(),
								)
							} else {
								__antithesis_instrumentation__.Notify(464252)

								err = errors.WithDetailf(
									err,
									"Column %s must be of type %s",
									t.Name(),
									tree.RegionEnum,
								)
							}
						} else {
							__antithesis_instrumentation__.Notify(464253)
						}
						__antithesis_instrumentation__.Notify(464249)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(464254)
					}
					__antithesis_instrumentation__.Notify(464245)
					break
				} else {
					__antithesis_instrumentation__.Notify(464255)
				}
			}
		}
		__antithesis_instrumentation__.Notify(464237)

		if !regionalByRowColExists {
			__antithesis_instrumentation__.Notify(464256)
			if n.Locality.RegionalByRowColumn != tree.RegionalByRowRegionNotSpecifiedName {
				__antithesis_instrumentation__.Notify(464258)
				return nil, pgerror.Newf(
					pgcode.UndefinedColumn,
					"column %s in REGIONAL BY ROW AS does not exist",
					regionalByRowCol.String(),
				)
			} else {
				__antithesis_instrumentation__.Notify(464259)
			}
			__antithesis_instrumentation__.Notify(464257)
			oid := typedesc.TypeIDToOID(regionConfig.RegionEnumID())
			n.Defs = append(
				n.Defs,
				regionalByRowDefaultColDef(
					oid,
					regionalByRowGatewayRegionDefaultExpr(oid),
					maybeRegionalByRowOnUpdateExpr(evalCtx, oid),
				),
			)
			cdd = append(cdd, nil)
		} else {
			__antithesis_instrumentation__.Notify(464260)
		}
		__antithesis_instrumentation__.Notify(464238)

		desc.PartitionAllBy = true
		partitionAllBy = partitionByForRegionalByRow(
			*regionConfig,
			regionalByRowCol,
		)

		primaryIndexColumnSet[string(regionalByRowCol)] = struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(464261)
	}
	__antithesis_instrumentation__.Notify(464189)

	if ttl := desc.GetRowLevelTTL(); ttl != nil {
		__antithesis_instrumentation__.Notify(464262)
		if err := checkTTLEnabledForCluster(ctx, st); err != nil {
			__antithesis_instrumentation__.Notify(464265)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464266)
		}
		__antithesis_instrumentation__.Notify(464263)
		hasRowLevelTTLColumn := false
		for _, def := range n.Defs {
			__antithesis_instrumentation__.Notify(464267)
			switch def := def.(type) {
			case *tree.ColumnTableDef:
				__antithesis_instrumentation__.Notify(464268)
				if def.Name == colinfo.TTLDefaultExpirationColumnName {
					__antithesis_instrumentation__.Notify(464269)

					if def.Type.SQLString() != types.TimestampTZ.SQLString() {
						__antithesis_instrumentation__.Notify(464271)
						return nil, pgerror.Newf(
							pgcode.InvalidTableDefinition,
							`table %s has TTL defined, but column %s is not a %s`,
							def.Name,
							colinfo.TTLDefaultExpirationColumnName,
							types.TimestampTZ.SQLString(),
						)
					} else {
						__antithesis_instrumentation__.Notify(464272)
					}
					__antithesis_instrumentation__.Notify(464270)
					hasRowLevelTTLColumn = true
					break
				} else {
					__antithesis_instrumentation__.Notify(464273)
				}
			}
		}
		__antithesis_instrumentation__.Notify(464264)
		if !hasRowLevelTTLColumn {
			__antithesis_instrumentation__.Notify(464274)
			col, err := rowLevelTTLAutomaticColumnDef(ttl)
			if err != nil {
				__antithesis_instrumentation__.Notify(464276)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464277)
			}
			__antithesis_instrumentation__.Notify(464275)
			n.Defs = append(n.Defs, col)
			cdd = append(cdd, nil)
		} else {
			__antithesis_instrumentation__.Notify(464278)
		}
	} else {
		__antithesis_instrumentation__.Notify(464279)
	}
	__antithesis_instrumentation__.Notify(464190)

	if n.PartitionByTable.ContainsPartitioningClause() {
		__antithesis_instrumentation__.Notify(464280)

		if n.PartitionByTable.PartitionBy != nil {
			__antithesis_instrumentation__.Notify(464282)
			for _, field := range n.PartitionByTable.PartitionBy.Fields {
				__antithesis_instrumentation__.Notify(464283)
				primaryIndexColumnSet[string(field)] = struct{}{}
			}
		} else {
			__antithesis_instrumentation__.Notify(464284)
		}
		__antithesis_instrumentation__.Notify(464281)
		if n.PartitionByTable.All {
			__antithesis_instrumentation__.Notify(464285)
			if !evalCtx.SessionData().ImplicitColumnPartitioningEnabled {
				__antithesis_instrumentation__.Notify(464287)
				return nil, errors.WithHint(
					pgerror.New(
						pgcode.FeatureNotSupported,
						"PARTITION ALL BY LIST/RANGE is currently experimental",
					),
					"to enable, use SET experimental_enable_implicit_column_partitioning = true",
				)
			} else {
				__antithesis_instrumentation__.Notify(464288)
			}
			__antithesis_instrumentation__.Notify(464286)
			desc.PartitionAllBy = true
			partitionAllBy = n.PartitionByTable.PartitionBy
		} else {
			__antithesis_instrumentation__.Notify(464289)
		}
	} else {
		__antithesis_instrumentation__.Notify(464290)
	}
	__antithesis_instrumentation__.Notify(464191)

	allowImplicitPartitioning := (sessionData != nil && func() bool {
		__antithesis_instrumentation__.Notify(464291)
		return sessionData.ImplicitColumnPartitioningEnabled == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(464292)
		return (n.Locality != nil && func() bool {
			__antithesis_instrumentation__.Notify(464293)
			return n.Locality.LocalityLevel == tree.LocalityLevelRow == true
		}() == true) == true
	}() == true

	type implicitColumnDefIdx struct {
		idx *descpb.IndexDescriptor
		def *tree.ColumnTableDef
	}
	var implicitColumnDefIdxs []implicitColumnDefIdx

	for i, def := range n.Defs {
		__antithesis_instrumentation__.Notify(464294)
		if d, ok := def.(*tree.ColumnTableDef); ok {
			__antithesis_instrumentation__.Notify(464295)
			if d.IsComputed() {
				__antithesis_instrumentation__.Notify(464304)
				d.Computed.Expr = schemaexpr.MaybeRewriteComputedColumn(d.Computed.Expr, evalCtx.SessionData())
			} else {
				__antithesis_instrumentation__.Notify(464305)
			}
			__antithesis_instrumentation__.Notify(464296)

			defType, err := tree.ResolveType(ctx, d.Type, semaCtx.GetTypeResolver())
			if err != nil {
				__antithesis_instrumentation__.Notify(464306)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464307)
			}
			__antithesis_instrumentation__.Notify(464297)
			if !desc.IsVirtualTable() {
				__antithesis_instrumentation__.Notify(464308)
				switch defType.Oid() {
				case oid.T_int2vector, oid.T_oidvector:
					__antithesis_instrumentation__.Notify(464309)
					return nil, pgerror.Newf(
						pgcode.FeatureNotSupported,
						"VECTOR column types are unsupported",
					)
				default:
					__antithesis_instrumentation__.Notify(464310)
				}
			} else {
				__antithesis_instrumentation__.Notify(464311)
			}
			__antithesis_instrumentation__.Notify(464298)
			if d.PrimaryKey.Sharded {
				__antithesis_instrumentation__.Notify(464312)
				if n.PartitionByTable.ContainsPartitions() && func() bool {
					__antithesis_instrumentation__.Notify(464317)
					return !n.PartitionByTable.All == true
				}() == true {
					__antithesis_instrumentation__.Notify(464318)
					return nil, pgerror.New(pgcode.FeatureNotSupported, "hash sharded indexes cannot be explicitly partitioned")
				} else {
					__antithesis_instrumentation__.Notify(464319)
				}
				__antithesis_instrumentation__.Notify(464313)
				buckets, err := tabledesc.EvalShardBucketCount(ctx, semaCtx, evalCtx, d.PrimaryKey.ShardBuckets, d.PrimaryKey.StorageParams)
				if err != nil {
					__antithesis_instrumentation__.Notify(464320)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464321)
				}
				__antithesis_instrumentation__.Notify(464314)
				shardCol, err := maybeCreateAndAddShardCol(int(buckets), &desc,
					[]string{string(d.Name)}, true,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(464322)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464323)
				}
				__antithesis_instrumentation__.Notify(464315)
				primaryIndexColumnSet[shardCol.GetName()] = struct{}{}
				checkConstraint, err := makeShardCheckConstraintDef(int(buckets), shardCol)
				if err != nil {
					__antithesis_instrumentation__.Notify(464324)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464325)
				}
				__antithesis_instrumentation__.Notify(464316)

				n.Defs = append(n.Defs, checkConstraint)
				cdd = append(cdd, nil)
			} else {
				__antithesis_instrumentation__.Notify(464326)
			}
			__antithesis_instrumentation__.Notify(464299)
			if d.IsVirtual() && func() bool {
				__antithesis_instrumentation__.Notify(464327)
				return d.HasColumnFamily() == true
			}() == true {
				__antithesis_instrumentation__.Notify(464328)
				return nil, pgerror.Newf(pgcode.Syntax, "virtual columns cannot have family specifications")
			} else {
				__antithesis_instrumentation__.Notify(464329)
			}
			__antithesis_instrumentation__.Notify(464300)

			cdd[i], err = tabledesc.MakeColumnDefDescs(ctx, d, semaCtx, evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(464330)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464331)
			}
			__antithesis_instrumentation__.Notify(464301)
			col := cdd[i].ColumnDescriptor
			idx := cdd[i].PrimaryKeyOrUniqueIndexDescriptor

			if !descpb.IsVirtualTable(id) {
				__antithesis_instrumentation__.Notify(464332)
				incTelemetryForNewColumn(d, col)
			} else {
				__antithesis_instrumentation__.Notify(464333)
			}
			__antithesis_instrumentation__.Notify(464302)

			desc.AddColumn(col)

			if idx != nil {
				__antithesis_instrumentation__.Notify(464334)
				idx.Version = indexEncodingVersion
				implicitColumnDefIdxs = append(implicitColumnDefIdxs, implicitColumnDefIdx{idx: idx, def: d})
			} else {
				__antithesis_instrumentation__.Notify(464335)
			}
			__antithesis_instrumentation__.Notify(464303)

			if d.HasColumnFamily() {
				__antithesis_instrumentation__.Notify(464336)

				err := desc.AddColumnToFamilyMaybeCreate(col.Name, string(d.Family.Name), true, true)
				if err != nil {
					__antithesis_instrumentation__.Notify(464337)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464338)
				}
			} else {
				__antithesis_instrumentation__.Notify(464339)
			}
		} else {
			__antithesis_instrumentation__.Notify(464340)
		}
	}
	__antithesis_instrumentation__.Notify(464192)

	for _, implicitColumnDefIdx := range implicitColumnDefIdxs {
		__antithesis_instrumentation__.Notify(464341)
		if implicitColumnDefIdx.def.PrimaryKey.IsPrimaryKey {
			__antithesis_instrumentation__.Notify(464342)
			if err := desc.AddPrimaryIndex(*implicitColumnDefIdx.idx); err != nil {
				__antithesis_instrumentation__.Notify(464344)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464345)
			}
			__antithesis_instrumentation__.Notify(464343)
			primaryIndexColumnSet[string(implicitColumnDefIdx.def.Name)] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(464346)

			if desc.PartitionAllBy {
				__antithesis_instrumentation__.Notify(464348)
				var err error
				newImplicitCols, newPartitioning, err := CreatePartitioning(
					ctx,
					st,
					evalCtx,
					&desc,
					*implicitColumnDefIdx.idx,
					partitionAllBy,
					nil,
					allowImplicitPartitioning,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(464350)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464351)
				}
				__antithesis_instrumentation__.Notify(464349)
				tabledesc.UpdateIndexPartitioning(implicitColumnDefIdx.idx, false, newImplicitCols, newPartitioning)
			} else {
				__antithesis_instrumentation__.Notify(464352)
			}
			__antithesis_instrumentation__.Notify(464347)

			if err := desc.AddSecondaryIndex(*implicitColumnDefIdx.idx); err != nil {
				__antithesis_instrumentation__.Notify(464353)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464354)
			}
		}
	}
	__antithesis_instrumentation__.Notify(464193)

	sourceInfo := colinfo.NewSourceInfoForSingleTable(
		n.Table, colinfo.ResultColumnsFromColumns(desc.GetID(), desc.PublicColumns()),
	)

	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(464355)
		col := &desc.Columns[i]
		if col.IsComputed() {
			__antithesis_instrumentation__.Notify(464356)
			expr, err := parser.ParseExpr(*col.ComputeExpr)
			if err != nil {
				__antithesis_instrumentation__.Notify(464359)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464360)
			}
			__antithesis_instrumentation__.Notify(464357)

			deqExpr, err := schemaexpr.DequalifyColumnRefs(ctx, sourceInfo, expr)
			if err != nil {
				__antithesis_instrumentation__.Notify(464361)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464362)
			}
			__antithesis_instrumentation__.Notify(464358)
			col.ComputeExpr = &deqExpr
		} else {
			__antithesis_instrumentation__.Notify(464363)
		}
	}
	__antithesis_instrumentation__.Notify(464194)

	setupShardedIndexForNewTable := func(
		d tree.IndexTableDef, idx *descpb.IndexDescriptor,
	) (columns tree.IndexElemList, _ error) {
		__antithesis_instrumentation__.Notify(464364)
		if d.PartitionByIndex.ContainsPartitions() {
			__antithesis_instrumentation__.Notify(464372)
			return nil, pgerror.New(pgcode.FeatureNotSupported, "hash sharded indexes cannot be explicitly partitioned")
		} else {
			__antithesis_instrumentation__.Notify(464373)
		}
		__antithesis_instrumentation__.Notify(464365)
		if desc.IsPartitionAllBy() && func() bool {
			__antithesis_instrumentation__.Notify(464374)
			return anyColumnIsPartitioningField(d.Columns, partitionAllBy) == true
		}() == true {
			__antithesis_instrumentation__.Notify(464375)
			return nil, pgerror.New(
				pgcode.FeatureNotSupported,
				`hash sharded indexes cannot include implicit partitioning columns from "PARTITION ALL BY" or "LOCALITY REGIONAL BY ROW"`,
			)
		} else {
			__antithesis_instrumentation__.Notify(464376)
		}
		__antithesis_instrumentation__.Notify(464366)
		shardCol, newColumns, err := setupShardedIndex(
			ctx,
			evalCtx,
			semaCtx,
			d.Columns,
			d.Sharded.ShardBuckets,
			&desc,
			idx,
			d.StorageParams,
			true)
		if err != nil {
			__antithesis_instrumentation__.Notify(464377)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464378)
		}
		__antithesis_instrumentation__.Notify(464367)

		buckets, err := tabledesc.EvalShardBucketCount(ctx, semaCtx, evalCtx, d.Sharded.ShardBuckets, d.StorageParams)
		if err != nil {
			__antithesis_instrumentation__.Notify(464379)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464380)
		}
		__antithesis_instrumentation__.Notify(464368)
		checkConstraint, err := makeShardCheckConstraintDef(int(buckets), shardCol)
		if err != nil {
			__antithesis_instrumentation__.Notify(464381)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464382)
		}
		__antithesis_instrumentation__.Notify(464369)

		ckBuilder := schemaexpr.MakeCheckConstraintBuilder(ctx, n.Table, &desc, semaCtx)
		checkConstraintDesc, err := ckBuilder.Build(checkConstraint)
		if err != nil {
			__antithesis_instrumentation__.Notify(464383)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464384)
		}
		__antithesis_instrumentation__.Notify(464370)
		for _, def := range n.Defs {
			__antithesis_instrumentation__.Notify(464385)
			if inputCheckConstraint, ok := def.(*tree.CheckConstraintTableDef); ok {
				__antithesis_instrumentation__.Notify(464386)
				inputCheckConstraintDesc, err := ckBuilder.Build(inputCheckConstraint)
				if err != nil {
					__antithesis_instrumentation__.Notify(464388)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464389)
				}
				__antithesis_instrumentation__.Notify(464387)
				if checkConstraintDesc.Expr == inputCheckConstraintDesc.Expr {
					__antithesis_instrumentation__.Notify(464390)
					return newColumns, nil
				} else {
					__antithesis_instrumentation__.Notify(464391)
				}
			} else {
				__antithesis_instrumentation__.Notify(464392)
			}
		}
		__antithesis_instrumentation__.Notify(464371)

		n.Defs = append(n.Defs, checkConstraint)
		cdd = append(cdd, nil)

		return newColumns, nil
	}
	__antithesis_instrumentation__.Notify(464195)

	for _, def := range n.Defs {
		__antithesis_instrumentation__.Notify(464393)
		switch d := def.(type) {
		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(464394)
			if d.IsComputed() {
				__antithesis_instrumentation__.Notify(464395)
				serializedExpr, _, err := schemaexpr.ValidateComputedColumnExpression(
					ctx, &desc, d, &n.Table, "computed column", semaCtx,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(464398)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464399)
				}
				__antithesis_instrumentation__.Notify(464396)
				col, err := desc.FindColumnWithName(d.Name)
				if err != nil {
					__antithesis_instrumentation__.Notify(464400)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464401)
				}
				__antithesis_instrumentation__.Notify(464397)
				col.ColumnDesc().ComputeExpr = &serializedExpr
			} else {
				__antithesis_instrumentation__.Notify(464402)
			}
		}
	}
	__antithesis_instrumentation__.Notify(464196)

	for _, def := range n.Defs {
		__antithesis_instrumentation__.Notify(464403)
		switch d := def.(type) {
		case *tree.ColumnTableDef, *tree.LikeTableDef:
			__antithesis_instrumentation__.Notify(464404)

		case *tree.IndexTableDef:
			__antithesis_instrumentation__.Notify(464405)

			if d.Name != "" {
				__antithesis_instrumentation__.Notify(464430)
				if idx, _ := desc.FindIndexWithName(d.Name.String()); idx != nil {
					__antithesis_instrumentation__.Notify(464431)
					return nil, pgerror.Newf(pgcode.DuplicateRelation, "duplicate index name: %q", d.Name)
				} else {
					__antithesis_instrumentation__.Notify(464432)
				}
			} else {
				__antithesis_instrumentation__.Notify(464433)
			}
			__antithesis_instrumentation__.Notify(464406)
			if err := validateColumnsAreAccessible(&desc, d.Columns); err != nil {
				__antithesis_instrumentation__.Notify(464434)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464435)
			}
			__antithesis_instrumentation__.Notify(464407)
			if err := replaceExpressionElemsWithVirtualCols(
				ctx,
				&desc,
				&n.Table,
				d.Columns,
				d.Inverted,
				true,
				semaCtx,
			); err != nil {
				__antithesis_instrumentation__.Notify(464436)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464437)
			}
			__antithesis_instrumentation__.Notify(464408)
			idx := descpb.IndexDescriptor{
				Name:             string(d.Name),
				StoreColumnNames: d.Storing.ToStrings(),
				Version:          indexEncodingVersion,
			}
			if d.Inverted {
				__antithesis_instrumentation__.Notify(464438)
				idx.Type = descpb.IndexDescriptor_INVERTED
			} else {
				__antithesis_instrumentation__.Notify(464439)
			}
			__antithesis_instrumentation__.Notify(464409)
			columns := d.Columns
			if d.Sharded != nil {
				__antithesis_instrumentation__.Notify(464440)
				var err error
				columns, err = setupShardedIndexForNewTable(*d, &idx)
				if err != nil {
					__antithesis_instrumentation__.Notify(464441)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464442)
				}
			} else {
				__antithesis_instrumentation__.Notify(464443)
			}
			__antithesis_instrumentation__.Notify(464410)
			if err := idx.FillColumns(columns); err != nil {
				__antithesis_instrumentation__.Notify(464444)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464445)
			}
			__antithesis_instrumentation__.Notify(464411)
			if d.Inverted {
				__antithesis_instrumentation__.Notify(464446)
				column, err := desc.FindColumnWithName(tree.Name(idx.InvertedColumnName()))
				if err != nil {
					__antithesis_instrumentation__.Notify(464448)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464449)
				}
				__antithesis_instrumentation__.Notify(464447)
				switch column.GetType().Family() {
				case types.GeometryFamily:
					__antithesis_instrumentation__.Notify(464450)
					config, err := geoindex.GeometryIndexConfigForSRID(column.GetType().GeoSRIDOrZero())
					if err != nil {
						__antithesis_instrumentation__.Notify(464454)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(464455)
					}
					__antithesis_instrumentation__.Notify(464451)
					idx.GeoConfig = *config
				case types.GeographyFamily:
					__antithesis_instrumentation__.Notify(464452)
					idx.GeoConfig = *geoindex.DefaultGeographyIndexConfig()
				default:
					__antithesis_instrumentation__.Notify(464453)
				}
			} else {
				__antithesis_instrumentation__.Notify(464456)
			}
			__antithesis_instrumentation__.Notify(464412)

			var idxPartitionBy *tree.PartitionBy
			if desc.PartitionAllBy && func() bool {
				__antithesis_instrumentation__.Notify(464457)
				return d.PartitionByIndex.ContainsPartitions() == true
			}() == true {
				__antithesis_instrumentation__.Notify(464458)
				return nil, pgerror.New(
					pgcode.FeatureNotSupported,
					"cannot define PARTITION BY on an index if the table is implicitly partitioned with PARTITION ALL BY or LOCALITY REGIONAL BY ROW definition",
				)
			} else {
				__antithesis_instrumentation__.Notify(464459)
			}
			__antithesis_instrumentation__.Notify(464413)
			if desc.PartitionAllBy {
				__antithesis_instrumentation__.Notify(464460)
				idxPartitionBy = partitionAllBy
			} else {
				__antithesis_instrumentation__.Notify(464461)
				if d.PartitionByIndex.ContainsPartitions() {
					__antithesis_instrumentation__.Notify(464462)
					idxPartitionBy = d.PartitionByIndex.PartitionBy
				} else {
					__antithesis_instrumentation__.Notify(464463)
				}
			}
			__antithesis_instrumentation__.Notify(464414)
			if idxPartitionBy != nil {
				__antithesis_instrumentation__.Notify(464464)
				var err error
				newImplicitCols, newPartitioning, err := CreatePartitioning(
					ctx,
					st,
					evalCtx,
					&desc,
					idx,
					idxPartitionBy,
					nil,
					allowImplicitPartitioning,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(464466)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464467)
				}
				__antithesis_instrumentation__.Notify(464465)
				tabledesc.UpdateIndexPartitioning(&idx, false, newImplicitCols, newPartitioning)
			} else {
				__antithesis_instrumentation__.Notify(464468)
			}
			__antithesis_instrumentation__.Notify(464415)

			if d.Predicate != nil {
				__antithesis_instrumentation__.Notify(464469)
				expr, err := schemaexpr.ValidatePartialIndexPredicate(
					ctx, &desc, d.Predicate, &n.Table, semaCtx,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(464471)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464472)
				}
				__antithesis_instrumentation__.Notify(464470)
				idx.Predicate = expr
			} else {
				__antithesis_instrumentation__.Notify(464473)
			}
			__antithesis_instrumentation__.Notify(464416)
			if err := paramparse.SetStorageParameters(
				ctx,
				semaCtx,
				evalCtx,
				d.StorageParams,
				&paramparse.IndexStorageParamObserver{IndexDesc: &idx},
			); err != nil {
				__antithesis_instrumentation__.Notify(464474)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464475)
			}
			__antithesis_instrumentation__.Notify(464417)

			if err := desc.AddSecondaryIndex(idx); err != nil {
				__antithesis_instrumentation__.Notify(464476)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464477)
			}
		case *tree.UniqueConstraintTableDef:
			__antithesis_instrumentation__.Notify(464418)
			if d.WithoutIndex {
				__antithesis_instrumentation__.Notify(464478)

				break
			} else {
				__antithesis_instrumentation__.Notify(464479)
			}
			__antithesis_instrumentation__.Notify(464419)

			if d.Name != "" {
				__antithesis_instrumentation__.Notify(464480)
				if idx, _ := desc.FindIndexWithName(d.Name.String()); idx != nil {
					__antithesis_instrumentation__.Notify(464481)
					return nil, pgerror.Newf(pgcode.DuplicateRelation, "duplicate index name: %q", d.Name)
				} else {
					__antithesis_instrumentation__.Notify(464482)
				}
			} else {
				__antithesis_instrumentation__.Notify(464483)
			}
			__antithesis_instrumentation__.Notify(464420)
			if err := validateColumnsAreAccessible(&desc, d.Columns); err != nil {
				__antithesis_instrumentation__.Notify(464484)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464485)
			}
			__antithesis_instrumentation__.Notify(464421)
			if err := replaceExpressionElemsWithVirtualCols(
				ctx,
				&desc,
				&n.Table,
				d.Columns,
				false,
				true,
				semaCtx,
			); err != nil {
				__antithesis_instrumentation__.Notify(464486)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464487)
			}
			__antithesis_instrumentation__.Notify(464422)
			idx := descpb.IndexDescriptor{
				Name:             string(d.Name),
				Unique:           true,
				StoreColumnNames: d.Storing.ToStrings(),
				Version:          indexEncodingVersion,
			}
			columns := d.Columns
			if d.Sharded != nil {
				__antithesis_instrumentation__.Notify(464488)
				if d.PrimaryKey && func() bool {
					__antithesis_instrumentation__.Notify(464490)
					return n.PartitionByTable.ContainsPartitions() == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(464491)
					return !n.PartitionByTable.All == true
				}() == true {
					__antithesis_instrumentation__.Notify(464492)
					return nil, pgerror.New(
						pgcode.FeatureNotSupported,
						"hash sharded indexes cannot be explicitly partitioned",
					)
				} else {
					__antithesis_instrumentation__.Notify(464493)
				}
				__antithesis_instrumentation__.Notify(464489)
				var err error
				columns, err = setupShardedIndexForNewTable(d.IndexTableDef, &idx)
				if err != nil {
					__antithesis_instrumentation__.Notify(464494)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464495)
				}
			} else {
				__antithesis_instrumentation__.Notify(464496)
			}
			__antithesis_instrumentation__.Notify(464423)
			if err := idx.FillColumns(columns); err != nil {
				__antithesis_instrumentation__.Notify(464497)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464498)
			}
			__antithesis_instrumentation__.Notify(464424)

			if d.PrimaryKey && func() bool {
				__antithesis_instrumentation__.Notify(464499)
				return d.PartitionByIndex.ContainsPartitioningClause() == true
			}() == true {
				__antithesis_instrumentation__.Notify(464500)
				return nil, errors.AssertionFailedf(
					"PRIMARY KEY partitioning should be defined at table level",
				)
			} else {
				__antithesis_instrumentation__.Notify(464501)
			}
			__antithesis_instrumentation__.Notify(464425)

			if !d.PrimaryKey {
				__antithesis_instrumentation__.Notify(464502)
				if desc.PartitionAllBy && func() bool {
					__antithesis_instrumentation__.Notify(464505)
					return d.PartitionByIndex.ContainsPartitions() == true
				}() == true {
					__antithesis_instrumentation__.Notify(464506)
					return nil, pgerror.New(
						pgcode.FeatureNotSupported,
						"cannot define PARTITION BY on an unique constraint if the table is implicitly partitioned with PARTITION ALL BY or LOCALITY REGIONAL BY ROW definition",
					)
				} else {
					__antithesis_instrumentation__.Notify(464507)
				}
				__antithesis_instrumentation__.Notify(464503)
				var idxPartitionBy *tree.PartitionBy
				if desc.PartitionAllBy {
					__antithesis_instrumentation__.Notify(464508)
					idxPartitionBy = partitionAllBy
				} else {
					__antithesis_instrumentation__.Notify(464509)
					if d.PartitionByIndex.ContainsPartitions() {
						__antithesis_instrumentation__.Notify(464510)
						idxPartitionBy = d.PartitionByIndex.PartitionBy
					} else {
						__antithesis_instrumentation__.Notify(464511)
					}
				}
				__antithesis_instrumentation__.Notify(464504)

				if idxPartitionBy != nil {
					__antithesis_instrumentation__.Notify(464512)
					var err error
					newImplicitCols, newPartitioning, err := CreatePartitioning(
						ctx,
						st,
						evalCtx,
						&desc,
						idx,
						idxPartitionBy,
						nil,
						allowImplicitPartitioning,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(464514)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(464515)
					}
					__antithesis_instrumentation__.Notify(464513)
					tabledesc.UpdateIndexPartitioning(&idx, false, newImplicitCols, newPartitioning)
				} else {
					__antithesis_instrumentation__.Notify(464516)
				}
			} else {
				__antithesis_instrumentation__.Notify(464517)
			}
			__antithesis_instrumentation__.Notify(464426)
			if d.Predicate != nil {
				__antithesis_instrumentation__.Notify(464518)
				expr, err := schemaexpr.ValidatePartialIndexPredicate(
					ctx, &desc, d.Predicate, &n.Table, semaCtx,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(464520)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464521)
				}
				__antithesis_instrumentation__.Notify(464519)
				idx.Predicate = expr
			} else {
				__antithesis_instrumentation__.Notify(464522)
			}
			__antithesis_instrumentation__.Notify(464427)
			if d.PrimaryKey {
				__antithesis_instrumentation__.Notify(464523)
				if err := desc.AddPrimaryIndex(idx); err != nil {
					__antithesis_instrumentation__.Notify(464525)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464526)
				}
				__antithesis_instrumentation__.Notify(464524)
				for _, c := range columns {
					__antithesis_instrumentation__.Notify(464527)
					primaryIndexColumnSet[string(c.Column)] = struct{}{}
				}
			} else {
				__antithesis_instrumentation__.Notify(464528)
				if err := desc.AddSecondaryIndex(idx); err != nil {
					__antithesis_instrumentation__.Notify(464529)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464530)
				}
			}
		case *tree.CheckConstraintTableDef, *tree.ForeignKeyConstraintTableDef, *tree.FamilyTableDef:
			__antithesis_instrumentation__.Notify(464428)

		default:
			__antithesis_instrumentation__.Notify(464429)
			return nil, errors.Errorf("unsupported table def: %T", def)
		}
	}
	__antithesis_instrumentation__.Notify(464197)

	if desc.GetPrimaryIndex().NumKeyColumns() == 0 && func() bool {
		__antithesis_instrumentation__.Notify(464531)
		return desc.IsPhysicalTable() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(464532)
		return evalCtx != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(464533)
		return evalCtx.SessionData() != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(464534)
		return evalCtx.SessionData().RequireExplicitPrimaryKeys == true
	}() == true {
		__antithesis_instrumentation__.Notify(464535)
		return nil, errors.Errorf(
			"no primary key specified for table %s (require_explicit_primary_keys = true)", desc.Name)
	} else {
		__antithesis_instrumentation__.Notify(464536)
	}
	__antithesis_instrumentation__.Notify(464198)

	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(464537)
		if _, ok := primaryIndexColumnSet[desc.Columns[i].Name]; ok {
			__antithesis_instrumentation__.Notify(464538)
			if !st.Version.IsActive(ctx, clusterversion.Start22_1) {
				__antithesis_instrumentation__.Notify(464540)
				if desc.Columns[i].Virtual {
					__antithesis_instrumentation__.Notify(464541)
					return nil, pgerror.Newf(
						pgcode.FeatureNotSupported,
						"cannot use virtual column %q in primary key", desc.Columns[i].Name,
					)
				} else {
					__antithesis_instrumentation__.Notify(464542)
				}
			} else {
				__antithesis_instrumentation__.Notify(464543)
			}
			__antithesis_instrumentation__.Notify(464539)
			desc.Columns[i].Nullable = false
		} else {
			__antithesis_instrumentation__.Notify(464544)
		}
	}
	__antithesis_instrumentation__.Notify(464199)

	columnsInExplicitFamilies := map[string]bool{}
	for _, def := range n.Defs {
		__antithesis_instrumentation__.Notify(464545)
		if d, ok := def.(*tree.FamilyTableDef); ok {
			__antithesis_instrumentation__.Notify(464546)
			fam := descpb.ColumnFamilyDescriptor{
				Name:        string(d.Name),
				ColumnNames: d.Columns.ToStrings(),
			}
			for _, c := range fam.ColumnNames {
				__antithesis_instrumentation__.Notify(464548)
				columnsInExplicitFamilies[c] = true
			}
			__antithesis_instrumentation__.Notify(464547)
			desc.AddFamily(fam)
		} else {
			__antithesis_instrumentation__.Notify(464549)
		}
	}
	__antithesis_instrumentation__.Notify(464200)
	version := st.Version.ActiveVersionOrEmpty(ctx)
	if err := desc.AllocateIDs(ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(464550)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464551)
	}
	__antithesis_instrumentation__.Notify(464201)

	for _, idx := range desc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(464552)

		if idx.NumSecondaryStoredColumns() > 1 && func() bool {
			__antithesis_instrumentation__.Notify(464553)
			return len(desc.Families) > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(464554)
			telemetry.Inc(sqltelemetry.SecondaryIndexColumnFamiliesCounter)
		} else {
			__antithesis_instrumentation__.Notify(464555)
		}
	}
	__antithesis_instrumentation__.Notify(464202)

	if n.PartitionByTable.ContainsPartitions() || func() bool {
		__antithesis_instrumentation__.Notify(464556)
		return desc.PartitionAllBy == true
	}() == true {
		__antithesis_instrumentation__.Notify(464557)
		partitionBy := partitionAllBy
		if partitionBy == nil {
			__antithesis_instrumentation__.Notify(464559)
			partitionBy = n.PartitionByTable.PartitionBy
		} else {
			__antithesis_instrumentation__.Notify(464560)
		}
		__antithesis_instrumentation__.Notify(464558)

		if partitionBy != nil {
			__antithesis_instrumentation__.Notify(464561)
			newPrimaryIndex := desc.GetPrimaryIndex().IndexDescDeepCopy()
			newImplicitCols, newPartitioning, err := CreatePartitioning(
				ctx,
				st,
				evalCtx,
				&desc,
				newPrimaryIndex,
				partitionBy,
				nil,
				allowImplicitPartitioning,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(464563)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464564)
			}
			__antithesis_instrumentation__.Notify(464562)
			isIndexAltered := tabledesc.UpdateIndexPartitioning(&newPrimaryIndex, true, newImplicitCols, newPartitioning)
			if isIndexAltered {
				__antithesis_instrumentation__.Notify(464565)

				if numImplicitCols := newPrimaryIndex.Partitioning.NumImplicitColumns; numImplicitCols > 0 {
					__antithesis_instrumentation__.Notify(464567)
					for _, idx := range desc.PublicNonPrimaryIndexes() {
						__antithesis_instrumentation__.Notify(464568)
						if idx.GetEncodingType() != descpb.SecondaryIndexEncoding {
							__antithesis_instrumentation__.Notify(464572)
							continue
						} else {
							__antithesis_instrumentation__.Notify(464573)
						}
						__antithesis_instrumentation__.Notify(464569)
						colIDs := idx.CollectKeyColumnIDs()
						colIDs.UnionWith(idx.CollectSecondaryStoredColumnIDs())
						colIDs.UnionWith(idx.CollectKeySuffixColumnIDs())
						missingExtraColumnIDs := make([]descpb.ColumnID, 0, numImplicitCols)
						for _, implicitPrimaryColID := range newPrimaryIndex.KeyColumnIDs[:numImplicitCols] {
							__antithesis_instrumentation__.Notify(464574)
							if !colIDs.Contains(implicitPrimaryColID) {
								__antithesis_instrumentation__.Notify(464575)
								missingExtraColumnIDs = append(missingExtraColumnIDs, implicitPrimaryColID)
							} else {
								__antithesis_instrumentation__.Notify(464576)
							}
						}
						__antithesis_instrumentation__.Notify(464570)
						if len(missingExtraColumnIDs) == 0 {
							__antithesis_instrumentation__.Notify(464577)
							continue
						} else {
							__antithesis_instrumentation__.Notify(464578)
						}
						__antithesis_instrumentation__.Notify(464571)
						newIdxDesc := idx.IndexDescDeepCopy()
						newIdxDesc.KeySuffixColumnIDs = append(newIdxDesc.KeySuffixColumnIDs, missingExtraColumnIDs...)
						desc.SetPublicNonPrimaryIndex(idx.Ordinal(), newIdxDesc)
					}
				} else {
					__antithesis_instrumentation__.Notify(464579)
				}
				__antithesis_instrumentation__.Notify(464566)
				desc.SetPrimaryIndex(newPrimaryIndex)
			} else {
				__antithesis_instrumentation__.Notify(464580)
			}
		} else {
			__antithesis_instrumentation__.Notify(464581)
		}
	} else {
		__antithesis_instrumentation__.Notify(464582)
	}
	__antithesis_instrumentation__.Notify(464203)

	colIdx := 0
	for i := range n.Defs {
		__antithesis_instrumentation__.Notify(464583)
		if _, ok := n.Defs[i].(*tree.ColumnTableDef); ok {
			__antithesis_instrumentation__.Notify(464584)
			if cdd[i] != nil {
				__antithesis_instrumentation__.Notify(464586)
				if err := cdd[i].ForEachTypedExpr(func(expr tree.TypedExpr) error {
					__antithesis_instrumentation__.Notify(464587)
					changedSeqDescs, err := maybeAddSequenceDependencies(
						ctx, st, vt, &desc, &desc.Columns[colIdx], expr, affected)
					if err != nil {
						__antithesis_instrumentation__.Notify(464590)
						return err
					} else {
						__antithesis_instrumentation__.Notify(464591)
					}
					__antithesis_instrumentation__.Notify(464588)
					for _, changedSeqDesc := range changedSeqDescs {
						__antithesis_instrumentation__.Notify(464592)
						affected[changedSeqDesc.ID] = changedSeqDesc
					}
					__antithesis_instrumentation__.Notify(464589)
					return nil
				}); err != nil {
					__antithesis_instrumentation__.Notify(464593)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464594)
				}
			} else {
				__antithesis_instrumentation__.Notify(464595)
			}
			__antithesis_instrumentation__.Notify(464585)
			colIdx++
		} else {
			__antithesis_instrumentation__.Notify(464596)
		}
	}
	__antithesis_instrumentation__.Notify(464204)

	fkResolver := &fkSelfResolver{
		SchemaResolver: vt,
		prefix: catalog.ResolvedObjectPrefix{
			Database: db,
			Schema:   sc,
		},
		newTableDesc: &desc,
		newTableName: &n.Table,
	}

	ckBuilder := schemaexpr.MakeCheckConstraintBuilder(ctx, n.Table, &desc, semaCtx)
	for _, def := range n.Defs {
		__antithesis_instrumentation__.Notify(464597)
		switch d := def.(type) {
		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(464598)
			if d.Unique.WithoutIndex {
				__antithesis_instrumentation__.Notify(464605)
				if err := addUniqueWithoutIndexColumnTableDef(
					ctx, evalCtx, sessionData, d, &desc, NewTable, tree.ValidationDefault,
				); err != nil {
					__antithesis_instrumentation__.Notify(464606)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464607)
				}
			} else {
				__antithesis_instrumentation__.Notify(464608)
			}

		case *tree.UniqueConstraintTableDef:
			__antithesis_instrumentation__.Notify(464599)
			if d.WithoutIndex {
				__antithesis_instrumentation__.Notify(464609)
				if err := addUniqueWithoutIndexTableDef(
					ctx, evalCtx, sessionData, d, &desc, n.Table, NewTable, tree.ValidationDefault, semaCtx,
				); err != nil {
					__antithesis_instrumentation__.Notify(464610)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464611)
				}
			} else {
				__antithesis_instrumentation__.Notify(464612)
			}

		case *tree.IndexTableDef, *tree.FamilyTableDef, *tree.LikeTableDef:
			__antithesis_instrumentation__.Notify(464600)

		case *tree.CheckConstraintTableDef:
			__antithesis_instrumentation__.Notify(464601)
			ck, err := ckBuilder.Build(d)
			if err != nil {
				__antithesis_instrumentation__.Notify(464613)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464614)
			}
			__antithesis_instrumentation__.Notify(464602)
			desc.Checks = append(desc.Checks, ck)

		case *tree.ForeignKeyConstraintTableDef:
			__antithesis_instrumentation__.Notify(464603)
			if err := ResolveFK(
				ctx, txn, fkResolver, db, sc, &desc, d, affected, NewTable,
				tree.ValidationDefault, evalCtx,
			); err != nil {
				__antithesis_instrumentation__.Notify(464615)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(464616)
			}

		default:
			__antithesis_instrumentation__.Notify(464604)
			return nil, errors.Errorf("unsupported table def: %T", def)
		}
	}
	__antithesis_instrumentation__.Notify(464205)

	var onUpdateErr error
	tabledesc.ValidateOnUpdate(&desc, func(err error) {
		__antithesis_instrumentation__.Notify(464617)
		onUpdateErr = err
	})
	__antithesis_instrumentation__.Notify(464206)
	if onUpdateErr != nil {
		__antithesis_instrumentation__.Notify(464618)
		return nil, onUpdateErr
	} else {
		__antithesis_instrumentation__.Notify(464619)
	}
	__antithesis_instrumentation__.Notify(464207)

	if err := desc.AllocateIDs(ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(464620)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464621)
	}
	__antithesis_instrumentation__.Notify(464208)

	if desc.IsPhysicalTable() && func() bool {
		__antithesis_instrumentation__.Notify(464622)
		return !catalog.IsSystemDescriptor(&desc) == true
	}() == true {
		__antithesis_instrumentation__.Notify(464623)
		ts := evalCtx.GetTxnTimestamp(time.Microsecond).UnixNano()
		_ = catalog.ForEachNonDropIndex(&desc, func(idx catalog.Index) error {
			__antithesis_instrumentation__.Notify(464624)
			idx.IndexDesc().CreatedAtNanos = ts
			return nil
		})
	} else {
		__antithesis_instrumentation__.Notify(464625)
	}
	__antithesis_instrumentation__.Notify(464209)

	if err := catalog.ForEachNonDropIndex(&desc, func(idx catalog.Index) error {
		__antithesis_instrumentation__.Notify(464626)
		if idx.IsSharded() {
			__antithesis_instrumentation__.Notify(464630)
			telemetry.Inc(sqltelemetry.HashShardedIndexCounter)
		} else {
			__antithesis_instrumentation__.Notify(464631)
		}
		__antithesis_instrumentation__.Notify(464627)
		if idx.GetType() == descpb.IndexDescriptor_INVERTED {
			__antithesis_instrumentation__.Notify(464632)
			telemetry.Inc(sqltelemetry.InvertedIndexCounter)
			geoConfig := idx.GetGeoConfig()
			if !geoindex.IsEmptyConfig(&geoConfig) {
				__antithesis_instrumentation__.Notify(464636)
				if geoindex.IsGeographyConfig(&geoConfig) {
					__antithesis_instrumentation__.Notify(464637)
					telemetry.Inc(sqltelemetry.GeographyInvertedIndexCounter)
				} else {
					__antithesis_instrumentation__.Notify(464638)
					if geoindex.IsGeometryConfig(&geoConfig) {
						__antithesis_instrumentation__.Notify(464639)
						telemetry.Inc(sqltelemetry.GeometryInvertedIndexCounter)
					} else {
						__antithesis_instrumentation__.Notify(464640)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(464641)
			}
			__antithesis_instrumentation__.Notify(464633)
			if idx.IsPartial() {
				__antithesis_instrumentation__.Notify(464642)
				telemetry.Inc(sqltelemetry.PartialInvertedIndexCounter)
			} else {
				__antithesis_instrumentation__.Notify(464643)
			}
			__antithesis_instrumentation__.Notify(464634)
			if idx.NumKeyColumns() > 1 {
				__antithesis_instrumentation__.Notify(464644)
				telemetry.Inc(sqltelemetry.MultiColumnInvertedIndexCounter)
			} else {
				__antithesis_instrumentation__.Notify(464645)
			}
			__antithesis_instrumentation__.Notify(464635)
			if idx.GetPartitioning().NumColumns() != 0 {
				__antithesis_instrumentation__.Notify(464646)
				telemetry.Inc(sqltelemetry.PartitionedInvertedIndexCounter)
			} else {
				__antithesis_instrumentation__.Notify(464647)
			}
		} else {
			__antithesis_instrumentation__.Notify(464648)
		}
		__antithesis_instrumentation__.Notify(464628)
		if idx.IsPartial() {
			__antithesis_instrumentation__.Notify(464649)
			telemetry.Inc(sqltelemetry.PartialIndexCounter)
		} else {
			__antithesis_instrumentation__.Notify(464650)
		}
		__antithesis_instrumentation__.Notify(464629)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(464651)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464652)
	}
	__antithesis_instrumentation__.Notify(464210)

	if regionConfig != nil || func() bool {
		__antithesis_instrumentation__.Notify(464653)
		return n.Locality != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(464654)
		localityTelemetryName := "unspecified"
		if n.Locality != nil {
			__antithesis_instrumentation__.Notify(464656)
			localityTelemetryName = n.Locality.TelemetryName()
		} else {
			__antithesis_instrumentation__.Notify(464657)
		}
		__antithesis_instrumentation__.Notify(464655)
		telemetry.Inc(sqltelemetry.CreateTableLocalityCounter(localityTelemetryName))
		if n.Locality == nil {
			__antithesis_instrumentation__.Notify(464658)

			desc.SetTableLocalityRegionalByTable(tree.PrimaryRegionNotSpecifiedName)
		} else {
			__antithesis_instrumentation__.Notify(464659)
			if n.Locality.LocalityLevel == tree.LocalityLevelTable {
				__antithesis_instrumentation__.Notify(464660)
				desc.SetTableLocalityRegionalByTable(n.Locality.TableRegion)
			} else {
				__antithesis_instrumentation__.Notify(464661)
				if n.Locality.LocalityLevel == tree.LocalityLevelGlobal {
					__antithesis_instrumentation__.Notify(464662)
					desc.SetTableLocalityGlobal()
				} else {
					__antithesis_instrumentation__.Notify(464663)
					if n.Locality.LocalityLevel == tree.LocalityLevelRow {
						__antithesis_instrumentation__.Notify(464664)
						desc.SetTableLocalityRegionalByRow(n.Locality.RegionalByRowColumn)
					} else {
						__antithesis_instrumentation__.Notify(464665)
						return nil, errors.Newf("unknown locality level: %v", n.Locality.LocalityLevel)
					}
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(464666)
	}
	__antithesis_instrumentation__.Notify(464211)

	return &desc, nil
}

func newTableDesc(
	params runParams,
	n *tree.CreateTable,
	db catalog.DatabaseDescriptor,
	sc catalog.SchemaDescriptor,
	id descpb.ID,
	creationTime hlc.Timestamp,
	privileges *catpb.PrivilegeDescriptor,
	affected map[descpb.ID]*tabledesc.Mutable,
) (ret *tabledesc.Mutable, err error) {
	__antithesis_instrumentation__.Notify(464667)
	if err := validateUniqueConstraintParamsForCreateTable(n); err != nil {
		__antithesis_instrumentation__.Notify(464677)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464678)
	}
	__antithesis_instrumentation__.Notify(464668)

	newDefs, err := replaceLikeTableOpts(n, params)
	if err != nil {
		__antithesis_instrumentation__.Notify(464679)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464680)
	}
	__antithesis_instrumentation__.Notify(464669)

	if newDefs != nil {
		__antithesis_instrumentation__.Notify(464681)

		n.Defs = newDefs
	} else {
		__antithesis_instrumentation__.Notify(464682)
	}
	__antithesis_instrumentation__.Notify(464670)

	colNameToOwnedSeq, err := createSequencesForSerialColumns(
		params.ctx, params.p, params.SessionData(), db, sc, n,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(464683)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464684)
	}
	__antithesis_instrumentation__.Notify(464671)

	var regionConfig *multiregion.RegionConfig
	if db.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(464685)
		conf, err := SynthesizeRegionConfig(params.ctx, params.p.txn, db.GetID(), params.p.Descriptors())
		if err != nil {
			__antithesis_instrumentation__.Notify(464687)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464688)
		}
		__antithesis_instrumentation__.Notify(464686)
		regionConfig = &conf
	} else {
		__antithesis_instrumentation__.Notify(464689)
	}
	__antithesis_instrumentation__.Notify(464672)

	params.p.runWithOptions(resolveFlags{skipCache: true, contextDatabaseID: db.GetID()}, func() {
		__antithesis_instrumentation__.Notify(464690)
		ret, err = NewTableDesc(
			params.ctx,
			params.p.txn,
			params.p,
			params.p.ExecCfg().Settings,
			n,
			db,
			sc,
			id,
			regionConfig,
			creationTime,
			privileges,
			affected,
			&params.p.semaCtx,
			params.EvalContext(),
			params.SessionData(),
			n.Persistence,
		)
	})
	__antithesis_instrumentation__.Notify(464673)
	if err != nil {
		__antithesis_instrumentation__.Notify(464691)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464692)
	}
	__antithesis_instrumentation__.Notify(464674)

	for colName, seqDesc := range colNameToOwnedSeq {
		__antithesis_instrumentation__.Notify(464693)

		affectedSeqDesc := affected[seqDesc.ID]
		if err := setSequenceOwner(affectedSeqDesc, colName, ret); err != nil {
			__antithesis_instrumentation__.Notify(464694)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464695)
		}
	}
	__antithesis_instrumentation__.Notify(464675)

	if ttl := ret.RowLevelTTL; ttl != nil {
		__antithesis_instrumentation__.Notify(464696)
		j, err := CreateRowLevelTTLScheduledJob(
			params.ctx,
			params.ExecCfg(),
			params.p.txn,
			params.p.User(),
			ret.GetID(),
			ttl,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(464698)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464699)
		}
		__antithesis_instrumentation__.Notify(464697)
		ttl.ScheduleID = j.ScheduleID()
	} else {
		__antithesis_instrumentation__.Notify(464700)
	}
	__antithesis_instrumentation__.Notify(464676)
	return ret, nil
}

func newRowLevelTTLScheduledJob(
	env scheduledjobs.JobSchedulerEnv,
	owner security.SQLUsername,
	tblID descpb.ID,
	ttl *catpb.RowLevelTTL,
) (*jobs.ScheduledJob, error) {
	__antithesis_instrumentation__.Notify(464701)
	sj := jobs.NewScheduledJob(env)
	sj.SetScheduleLabel(fmt.Sprintf("row-level-ttl-%d", tblID))
	sj.SetOwner(owner)
	sj.SetScheduleDetails(jobspb.ScheduleDetails{
		Wait: jobspb.ScheduleDetails_WAIT,

		OnError: jobspb.ScheduleDetails_RETRY_SCHED,
	})

	if err := sj.SetSchedule(ttl.DeletionCronOrDefault()); err != nil {
		__antithesis_instrumentation__.Notify(464704)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464705)
	}
	__antithesis_instrumentation__.Notify(464702)
	args := &catpb.ScheduledRowLevelTTLArgs{
		TableID: tblID,
	}
	any, err := pbtypes.MarshalAny(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(464706)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464707)
	}
	__antithesis_instrumentation__.Notify(464703)
	sj.SetExecutionDetails(
		tree.ScheduledRowLevelTTLExecutor.InternalName(),
		jobspb.ExecutionArguments{Args: any},
	)
	return sj, nil
}

func checkTTLEnabledForCluster(ctx context.Context, st *cluster.Settings) error {
	__antithesis_instrumentation__.Notify(464708)
	if !st.Version.IsActive(ctx, clusterversion.RowLevelTTL) {
		__antithesis_instrumentation__.Notify(464710)
		return pgerror.Newf(
			pgcode.FeatureNotSupported,
			"row level TTL is only available once the cluster is fully upgraded",
		)
	} else {
		__antithesis_instrumentation__.Notify(464711)
	}
	__antithesis_instrumentation__.Notify(464709)
	return nil
}

func CreateRowLevelTTLScheduledJob(
	ctx context.Context,
	execCfg *ExecutorConfig,
	txn *kv.Txn,
	owner security.SQLUsername,
	tblID descpb.ID,
	ttl *catpb.RowLevelTTL,
) (*jobs.ScheduledJob, error) {
	__antithesis_instrumentation__.Notify(464712)
	telemetry.Inc(sqltelemetry.RowLevelTTLCreated)
	env := JobSchedulerEnv(execCfg)
	j, err := newRowLevelTTLScheduledJob(env, owner, tblID, ttl)
	if err != nil {
		__antithesis_instrumentation__.Notify(464715)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464716)
	}
	__antithesis_instrumentation__.Notify(464713)
	if err := j.Create(ctx, execCfg.InternalExecutor, txn); err != nil {
		__antithesis_instrumentation__.Notify(464717)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464718)
	}
	__antithesis_instrumentation__.Notify(464714)
	return j, nil
}

func rowLevelTTLAutomaticColumnDef(ttl *catpb.RowLevelTTL) (*tree.ColumnTableDef, error) {
	__antithesis_instrumentation__.Notify(464719)
	def := &tree.ColumnTableDef{
		Name:   colinfo.TTLDefaultExpirationColumnName,
		Type:   types.TimestampTZ,
		Hidden: true,
	}
	intervalExpr, err := parser.ParseExpr(string(ttl.DurationExpr))
	if err != nil {
		__antithesis_instrumentation__.Notify(464721)
		return nil, errors.Wrapf(err, "unexpected expression for TTL duration")
	} else {
		__antithesis_instrumentation__.Notify(464722)
	}
	__antithesis_instrumentation__.Notify(464720)
	def.DefaultExpr.Expr = rowLevelTTLAutomaticColumnExpr(intervalExpr)
	def.OnUpdateExpr.Expr = rowLevelTTLAutomaticColumnExpr(intervalExpr)
	return def, nil
}

func rowLevelTTLAutomaticColumnExpr(intervalExpr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(464723)
	return &tree.BinaryExpr{
		Operator: treebin.MakeBinaryOperator(treebin.Plus),
		Left:     &tree.FuncExpr{Func: tree.WrapFunction("current_timestamp")},
		Right:    intervalExpr,
	}
}

func replaceLikeTableOpts(n *tree.CreateTable, params runParams) (tree.TableDefs, error) {
	__antithesis_instrumentation__.Notify(464724)
	var newDefs tree.TableDefs
	for i, def := range n.Defs {
		__antithesis_instrumentation__.Notify(464726)
		d, ok := def.(*tree.LikeTableDef)
		if !ok {
			__antithesis_instrumentation__.Notify(464735)
			if newDefs != nil {
				__antithesis_instrumentation__.Notify(464737)
				newDefs = append(newDefs, def)
			} else {
				__antithesis_instrumentation__.Notify(464738)
			}
			__antithesis_instrumentation__.Notify(464736)
			continue
		} else {
			__antithesis_instrumentation__.Notify(464739)
		}
		__antithesis_instrumentation__.Notify(464727)

		if newDefs == nil {
			__antithesis_instrumentation__.Notify(464740)
			newDefs = make(tree.TableDefs, 0, len(n.Defs))
			newDefs = append(newDefs, n.Defs[:i]...)
		} else {
			__antithesis_instrumentation__.Notify(464741)
		}
		__antithesis_instrumentation__.Notify(464728)
		_, td, err := params.p.ResolveMutableTableDescriptor(params.ctx, &d.Name, true, tree.ResolveRequireTableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(464742)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464743)
		}
		__antithesis_instrumentation__.Notify(464729)
		opts := tree.LikeTableOpt(0)

		for _, opt := range d.Options {
			__antithesis_instrumentation__.Notify(464744)
			if opt.Excluded {
				__antithesis_instrumentation__.Notify(464745)
				opts &^= opt.Opt
			} else {
				__antithesis_instrumentation__.Notify(464746)
				opts |= opt.Opt
			}
		}
		__antithesis_instrumentation__.Notify(464730)

		shouldCopyColumnDefaultSet := make(map[string]struct{})
		if opts.Has(tree.LikeTableOptIndexes) {
			__antithesis_instrumentation__.Notify(464747)
			for _, idx := range td.NonDropIndexes() {
				__antithesis_instrumentation__.Notify(464748)

				if idx.Primary() && func() bool {
					__antithesis_instrumentation__.Notify(464750)
					return td.IsPrimaryIndexDefaultRowID() == true
				}() == true {
					__antithesis_instrumentation__.Notify(464751)
					for i := 0; i < idx.NumKeyColumns(); i++ {
						__antithesis_instrumentation__.Notify(464752)
						shouldCopyColumnDefaultSet[idx.GetKeyColumnName(i)] = struct{}{}
					}
				} else {
					__antithesis_instrumentation__.Notify(464753)
				}
				__antithesis_instrumentation__.Notify(464749)

				for i := 0; i < idx.ExplicitColumnStartIdx(); i++ {
					__antithesis_instrumentation__.Notify(464754)
					for i := 0; i < idx.NumKeyColumns(); i++ {
						__antithesis_instrumentation__.Notify(464755)
						shouldCopyColumnDefaultSet[idx.GetKeyColumnName(i)] = struct{}{}
					}
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(464756)
		}
		__antithesis_instrumentation__.Notify(464731)

		defs := make(tree.TableDefs, 0)

		for i := range td.Columns {
			__antithesis_instrumentation__.Notify(464757)
			c := &td.Columns[i]
			if c.Inaccessible {
				__antithesis_instrumentation__.Notify(464763)

				continue
			} else {
				__antithesis_instrumentation__.Notify(464764)
			}
			__antithesis_instrumentation__.Notify(464758)
			def := tree.ColumnTableDef{
				Name:   tree.Name(c.Name),
				Type:   c.Type,
				Hidden: c.Hidden,
			}
			if c.Nullable {
				__antithesis_instrumentation__.Notify(464765)
				def.Nullable.Nullability = tree.Null
			} else {
				__antithesis_instrumentation__.Notify(464766)
				def.Nullable.Nullability = tree.NotNull
			}
			__antithesis_instrumentation__.Notify(464759)
			if c.DefaultExpr != nil {
				__antithesis_instrumentation__.Notify(464767)
				_, shouldCopyColumnDefault := shouldCopyColumnDefaultSet[c.Name]
				if opts.Has(tree.LikeTableOptDefaults) || func() bool {
					__antithesis_instrumentation__.Notify(464768)
					return shouldCopyColumnDefault == true
				}() == true {
					__antithesis_instrumentation__.Notify(464769)
					def.DefaultExpr.Expr, err = parser.ParseExpr(*c.DefaultExpr)
					if err != nil {
						__antithesis_instrumentation__.Notify(464770)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(464771)
					}
				} else {
					__antithesis_instrumentation__.Notify(464772)
				}
			} else {
				__antithesis_instrumentation__.Notify(464773)
			}
			__antithesis_instrumentation__.Notify(464760)
			if c.ComputeExpr != nil {
				__antithesis_instrumentation__.Notify(464774)
				if opts.Has(tree.LikeTableOptGenerated) {
					__antithesis_instrumentation__.Notify(464775)
					def.Computed.Computed = true
					def.Computed.Virtual = c.Virtual
					def.Computed.Expr, err = parser.ParseExpr(*c.ComputeExpr)
					if err != nil {
						__antithesis_instrumentation__.Notify(464776)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(464777)
					}
				} else {
					__antithesis_instrumentation__.Notify(464778)
				}
			} else {
				__antithesis_instrumentation__.Notify(464779)
			}
			__antithesis_instrumentation__.Notify(464761)
			if c.OnUpdateExpr != nil {
				__antithesis_instrumentation__.Notify(464780)
				if opts.Has(tree.LikeTableOptDefaults) {
					__antithesis_instrumentation__.Notify(464781)
					def.OnUpdateExpr.Expr, err = parser.ParseExpr(*c.OnUpdateExpr)
					if err != nil {
						__antithesis_instrumentation__.Notify(464782)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(464783)
					}
				} else {
					__antithesis_instrumentation__.Notify(464784)
				}
			} else {
				__antithesis_instrumentation__.Notify(464785)
			}
			__antithesis_instrumentation__.Notify(464762)
			defs = append(defs, &def)
		}
		__antithesis_instrumentation__.Notify(464732)
		if opts.Has(tree.LikeTableOptConstraints) {
			__antithesis_instrumentation__.Notify(464786)
			for _, c := range td.Checks {
				__antithesis_instrumentation__.Notify(464788)
				def := tree.CheckConstraintTableDef{
					Name:   tree.Name(c.Name),
					Hidden: c.Hidden,
				}
				def.Expr, err = parser.ParseExpr(c.Expr)
				if err != nil {
					__antithesis_instrumentation__.Notify(464790)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464791)
				}
				__antithesis_instrumentation__.Notify(464789)
				defs = append(defs, &def)
			}
			__antithesis_instrumentation__.Notify(464787)
			for _, c := range td.UniqueWithoutIndexConstraints {
				__antithesis_instrumentation__.Notify(464792)
				def := tree.UniqueConstraintTableDef{
					IndexTableDef: tree.IndexTableDef{
						Name:    tree.Name(c.Name),
						Columns: make(tree.IndexElemList, 0, len(c.ColumnIDs)),
					},
					WithoutIndex: true,
				}
				colNames, err := td.NamesForColumnIDs(c.ColumnIDs)
				if err != nil {
					__antithesis_instrumentation__.Notify(464795)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(464796)
				}
				__antithesis_instrumentation__.Notify(464793)
				for i := range colNames {
					__antithesis_instrumentation__.Notify(464797)
					def.Columns = append(def.Columns, tree.IndexElem{Column: tree.Name(colNames[i])})
				}
				__antithesis_instrumentation__.Notify(464794)
				defs = append(defs, &def)
				if c.IsPartial() {
					__antithesis_instrumentation__.Notify(464798)
					def.Predicate, err = parser.ParseExpr(c.Predicate)
					if err != nil {
						__antithesis_instrumentation__.Notify(464799)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(464800)
					}
				} else {
					__antithesis_instrumentation__.Notify(464801)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(464802)
		}
		__antithesis_instrumentation__.Notify(464733)
		if opts.Has(tree.LikeTableOptIndexes) {
			__antithesis_instrumentation__.Notify(464803)
			for _, idx := range td.NonDropIndexes() {
				__antithesis_instrumentation__.Notify(464804)
				indexDef := tree.IndexTableDef{
					Name:     tree.Name(idx.GetName()),
					Inverted: idx.GetType() == descpb.IndexDescriptor_INVERTED,
					Storing:  make(tree.NameList, 0, idx.NumSecondaryStoredColumns()),
					Columns:  make(tree.IndexElemList, 0, idx.NumKeyColumns()),
				}
				numColumns := idx.NumKeyColumns()
				if idx.IsSharded() {
					__antithesis_instrumentation__.Notify(464810)
					indexDef.Sharded = &tree.ShardedIndexDef{
						ShardBuckets: tree.NewDInt(tree.DInt(idx.GetSharded().ShardBuckets)),
					}
					numColumns = len(idx.GetSharded().ColumnNames)
				} else {
					__antithesis_instrumentation__.Notify(464811)
				}
				__antithesis_instrumentation__.Notify(464805)
				for j := 0; j < numColumns; j++ {
					__antithesis_instrumentation__.Notify(464812)
					name := idx.GetKeyColumnName(j)
					if idx.IsSharded() {
						__antithesis_instrumentation__.Notify(464817)
						name = idx.GetSharded().ColumnNames[j]
					} else {
						__antithesis_instrumentation__.Notify(464818)
					}
					__antithesis_instrumentation__.Notify(464813)
					elem := tree.IndexElem{
						Column:    tree.Name(name),
						Direction: tree.Ascending,
					}
					col, err := td.FindColumnWithID(idx.GetKeyColumnID(j))
					if err != nil {
						__antithesis_instrumentation__.Notify(464819)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(464820)
					}
					__antithesis_instrumentation__.Notify(464814)
					if col.IsExpressionIndexColumn() {
						__antithesis_instrumentation__.Notify(464821)
						elem.Column = ""
						elem.Expr, err = parser.ParseExpr(col.GetComputeExpr())
						if err != nil {
							__antithesis_instrumentation__.Notify(464822)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(464823)
						}
					} else {
						__antithesis_instrumentation__.Notify(464824)
					}
					__antithesis_instrumentation__.Notify(464815)
					if idx.GetKeyColumnDirection(j) == descpb.IndexDescriptor_DESC {
						__antithesis_instrumentation__.Notify(464825)
						elem.Direction = tree.Descending
					} else {
						__antithesis_instrumentation__.Notify(464826)
					}
					__antithesis_instrumentation__.Notify(464816)
					indexDef.Columns = append(indexDef.Columns, elem)
				}
				__antithesis_instrumentation__.Notify(464806)
				for j := 0; j < idx.NumSecondaryStoredColumns(); j++ {
					__antithesis_instrumentation__.Notify(464827)
					indexDef.Storing = append(indexDef.Storing, tree.Name(idx.GetStoredColumnName(j)))
				}
				__antithesis_instrumentation__.Notify(464807)
				var def tree.TableDef = &indexDef
				if idx.IsUnique() {
					__antithesis_instrumentation__.Notify(464828)
					def = &tree.UniqueConstraintTableDef{
						IndexTableDef: indexDef,
						PrimaryKey:    idx.Primary(),
					}
				} else {
					__antithesis_instrumentation__.Notify(464829)
				}
				__antithesis_instrumentation__.Notify(464808)
				if idx.IsPartial() {
					__antithesis_instrumentation__.Notify(464830)
					indexDef.Predicate, err = parser.ParseExpr(idx.GetPredicate())
					if err != nil {
						__antithesis_instrumentation__.Notify(464831)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(464832)
					}
				} else {
					__antithesis_instrumentation__.Notify(464833)
				}
				__antithesis_instrumentation__.Notify(464809)
				defs = append(defs, def)
			}
		} else {
			__antithesis_instrumentation__.Notify(464834)
		}
		__antithesis_instrumentation__.Notify(464734)
		newDefs = append(newDefs, defs...)
	}
	__antithesis_instrumentation__.Notify(464725)
	return newDefs, nil
}

func makeShardColumnDesc(colNames []string, buckets int) (*descpb.ColumnDescriptor, error) {
	__antithesis_instrumentation__.Notify(464835)
	col := &descpb.ColumnDescriptor{
		Hidden:   true,
		Nullable: false,
		Type:     types.Int,
		Virtual:  true,
	}
	col.Name = tabledesc.GetShardColumnName(colNames, int32(buckets))
	col.ComputeExpr = schemaexpr.MakeHashShardComputeExpr(colNames, buckets)
	return col, nil
}

func makeShardCheckConstraintDef(
	buckets int, shardCol catalog.Column,
) (*tree.CheckConstraintTableDef, error) {
	__antithesis_instrumentation__.Notify(464836)
	values := &tree.Tuple{}
	for i := 0; i < buckets; i++ {
		__antithesis_instrumentation__.Notify(464838)
		const negative = false
		values.Exprs = append(values.Exprs, tree.NewNumVal(
			constant.MakeInt64(int64(i)),
			strconv.Itoa(i),
			negative))
	}
	__antithesis_instrumentation__.Notify(464837)
	return &tree.CheckConstraintTableDef{
		Expr: &tree.ComparisonExpr{
			Operator: treecmp.MakeComparisonOperator(treecmp.In),
			Left: &tree.ColumnItem{
				ColumnName: tree.Name(shardCol.GetName()),
			},
			Right: values,
		},
		Hidden: true,
	}, nil
}

func incTelemetryForNewColumn(def *tree.ColumnTableDef, desc *descpb.ColumnDescriptor) {
	__antithesis_instrumentation__.Notify(464839)
	switch desc.Type.Family() {
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(464844)
		sqltelemetry.IncrementEnumCounter(sqltelemetry.EnumInTable)
	default:
		__antithesis_instrumentation__.Notify(464845)
		telemetry.Inc(sqltelemetry.SchemaNewTypeCounter(desc.Type.TelemetryName()))
	}
	__antithesis_instrumentation__.Notify(464840)
	if desc.IsComputed() {
		__antithesis_instrumentation__.Notify(464846)
		if desc.Virtual {
			__antithesis_instrumentation__.Notify(464847)
			telemetry.Inc(sqltelemetry.SchemaNewColumnTypeQualificationCounter("virtual"))
		} else {
			__antithesis_instrumentation__.Notify(464848)
			telemetry.Inc(sqltelemetry.SchemaNewColumnTypeQualificationCounter("computed"))
		}
	} else {
		__antithesis_instrumentation__.Notify(464849)
	}
	__antithesis_instrumentation__.Notify(464841)
	if desc.HasDefault() {
		__antithesis_instrumentation__.Notify(464850)
		telemetry.Inc(sqltelemetry.SchemaNewColumnTypeQualificationCounter("default_expr"))
	} else {
		__antithesis_instrumentation__.Notify(464851)
	}
	__antithesis_instrumentation__.Notify(464842)
	if def.Unique.IsUnique {
		__antithesis_instrumentation__.Notify(464852)
		if def.Unique.WithoutIndex {
			__antithesis_instrumentation__.Notify(464853)
			telemetry.Inc(sqltelemetry.SchemaNewColumnTypeQualificationCounter("unique_without_index"))
		} else {
			__antithesis_instrumentation__.Notify(464854)
			telemetry.Inc(sqltelemetry.SchemaNewColumnTypeQualificationCounter("unique"))
		}
	} else {
		__antithesis_instrumentation__.Notify(464855)
	}
	__antithesis_instrumentation__.Notify(464843)
	if desc.HasOnUpdate() {
		__antithesis_instrumentation__.Notify(464856)
		telemetry.Inc(sqltelemetry.SchemaNewColumnTypeQualificationCounter("on_update"))
	} else {
		__antithesis_instrumentation__.Notify(464857)
	}
}

func regionalByRowRegionDefaultExpr(oid oid.Oid, region tree.Name) tree.Expr {
	__antithesis_instrumentation__.Notify(464858)
	return &tree.CastExpr{
		Expr:       tree.NewDString(string(region)),
		Type:       &tree.OIDTypeReference{OID: oid},
		SyntaxMode: tree.CastShort,
	}
}

func regionalByRowGatewayRegionDefaultExpr(oid oid.Oid) tree.Expr {
	__antithesis_instrumentation__.Notify(464859)
	return &tree.CastExpr{
		Expr: &tree.FuncExpr{
			Func: tree.WrapFunction(builtins.DefaultToDatabasePrimaryRegionBuiltinName),
			Exprs: []tree.Expr{
				&tree.FuncExpr{
					Func: tree.WrapFunction(builtins.GatewayRegionBuiltinName),
				},
			},
		},
		Type:       &tree.OIDTypeReference{OID: oid},
		SyntaxMode: tree.CastShort,
	}
}

func maybeRegionalByRowOnUpdateExpr(evalCtx *tree.EvalContext, enumOid oid.Oid) tree.Expr {
	__antithesis_instrumentation__.Notify(464860)
	if evalCtx.SessionData().AutoRehomingEnabled {
		__antithesis_instrumentation__.Notify(464862)
		return &tree.CastExpr{
			Expr: &tree.FuncExpr{
				Func: tree.WrapFunction(builtins.RehomeRowBuiltinName),
			},
			Type:       &tree.OIDTypeReference{OID: enumOid},
			SyntaxMode: tree.CastShort,
		}
	} else {
		__antithesis_instrumentation__.Notify(464863)
	}
	__antithesis_instrumentation__.Notify(464861)
	return nil
}

func regionalByRowDefaultColDef(
	oid oid.Oid, defaultExpr tree.Expr, onUpdateExpr tree.Expr,
) *tree.ColumnTableDef {
	__antithesis_instrumentation__.Notify(464864)
	c := &tree.ColumnTableDef{
		Name:   tree.RegionalByRowRegionDefaultColName,
		Type:   &tree.OIDTypeReference{OID: oid},
		Hidden: true,
	}
	c.Nullable.Nullability = tree.NotNull
	c.DefaultExpr.Expr = defaultExpr
	c.OnUpdateExpr.Expr = onUpdateExpr

	return c
}

func setSequenceOwner(
	seqDesc *tabledesc.Mutable, colName tree.Name, table *tabledesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(464865)
	if !seqDesc.IsSequence() {
		__antithesis_instrumentation__.Notify(464870)
		return errors.Errorf("%s is not a sequence", seqDesc.Name)
	} else {
		__antithesis_instrumentation__.Notify(464871)
	}
	__antithesis_instrumentation__.Notify(464866)

	col, err := table.FindColumnWithName(colName)
	if err != nil {
		__antithesis_instrumentation__.Notify(464872)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464873)
	}
	__antithesis_instrumentation__.Notify(464867)
	found := false
	for _, seqID := range col.ColumnDesc().OwnsSequenceIds {
		__antithesis_instrumentation__.Notify(464874)
		if seqID == seqDesc.ID {
			__antithesis_instrumentation__.Notify(464875)
			found = true
			break
		} else {
			__antithesis_instrumentation__.Notify(464876)
		}
	}
	__antithesis_instrumentation__.Notify(464868)
	if !found {
		__antithesis_instrumentation__.Notify(464877)
		col.ColumnDesc().OwnsSequenceIds = append(col.ColumnDesc().OwnsSequenceIds, seqDesc.ID)
	} else {
		__antithesis_instrumentation__.Notify(464878)
	}
	__antithesis_instrumentation__.Notify(464869)
	seqDesc.SequenceOpts.SequenceOwner.OwnerTableID = table.ID
	seqDesc.SequenceOpts.SequenceOwner.OwnerColumnID = col.GetID()

	return nil
}

func validateUniqueConstraintParamsForCreateTable(n *tree.CreateTable) error {
	__antithesis_instrumentation__.Notify(464879)
	for _, def := range n.Defs {
		__antithesis_instrumentation__.Notify(464881)
		switch d := def.(type) {
		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(464882)
			if err := paramparse.ValidateUniqueConstraintParams(
				d.PrimaryKey.StorageParams,
				paramparse.UniqueConstraintParamContext{
					IsPrimaryKey: true,
					IsSharded:    d.PrimaryKey.Sharded,
				}); err != nil {
				__antithesis_instrumentation__.Notify(464885)
				return err
			} else {
				__antithesis_instrumentation__.Notify(464886)
			}
		case *tree.UniqueConstraintTableDef:
			__antithesis_instrumentation__.Notify(464883)
			if err := paramparse.ValidateUniqueConstraintParams(
				d.IndexTableDef.StorageParams,
				paramparse.UniqueConstraintParamContext{
					IsPrimaryKey: d.PrimaryKey,
					IsSharded:    d.Sharded != nil,
				},
			); err != nil {
				__antithesis_instrumentation__.Notify(464887)
				return err
			} else {
				__antithesis_instrumentation__.Notify(464888)
			}
		case *tree.IndexTableDef:
			__antithesis_instrumentation__.Notify(464884)
			if d.Sharded == nil && func() bool {
				__antithesis_instrumentation__.Notify(464889)
				return d.StorageParams.GetVal(`bucket_count`) != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(464890)
				return pgerror.New(
					pgcode.InvalidParameterValue,
					`"bucket_count" storage param should only be set with "USING HASH" for hash sharded index`,
				)
			} else {
				__antithesis_instrumentation__.Notify(464891)
			}
		}
	}
	__antithesis_instrumentation__.Notify(464880)
	return nil
}

func validateUniqueConstraintParamsForCreateTableAs(n *tree.CreateTable) error {
	__antithesis_instrumentation__.Notify(464892)

	const errMsg = `storage parameters are not supported on primary key for CREATE TABLE...AS... statement`
	for _, def := range n.Defs {
		__antithesis_instrumentation__.Notify(464894)
		switch d := def.(type) {
		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(464895)
			if len(d.PrimaryKey.StorageParams) > 0 {
				__antithesis_instrumentation__.Notify(464897)
				return pgerror.New(pgcode.FeatureNotSupported, errMsg)
			} else {
				__antithesis_instrumentation__.Notify(464898)
			}
		case *tree.UniqueConstraintTableDef:
			__antithesis_instrumentation__.Notify(464896)
			if d.PrimaryKey && func() bool {
				__antithesis_instrumentation__.Notify(464899)
				return len(d.StorageParams) > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(464900)
				return pgerror.New(pgcode.FeatureNotSupported, errMsg)
			} else {
				__antithesis_instrumentation__.Notify(464901)
			}
		}
	}
	__antithesis_instrumentation__.Notify(464893)
	return nil
}
