package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type alterTableSetLocalityNode struct {
	n         tree.AlterTableLocality
	tableDesc *tabledesc.Mutable
	dbDesc    catalog.DatabaseDescriptor
}

func (p *planner) AlterTableLocality(
	ctx context.Context, n *tree.AlterTableLocality,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(244738)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER TABLE",
	); err != nil {
		__antithesis_instrumentation__.Notify(244745)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244746)
	}
	__antithesis_instrumentation__.Notify(244739)

	_, tableDesc, err := p.ResolveMutableTableDescriptorEx(
		ctx, n.Name, !n.IfExists, tree.ResolveRequireTableDesc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(244747)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244748)
	}
	__antithesis_instrumentation__.Notify(244740)
	if tableDesc == nil {
		__antithesis_instrumentation__.Notify(244749)
		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(244750)
	}
	__antithesis_instrumentation__.Notify(244741)

	if err := p.checkPrivilegesForMultiRegionOp(ctx, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(244751)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244752)
	}
	__antithesis_instrumentation__.Notify(244742)

	_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(
		ctx,
		p.txn,
		tableDesc.GetParentID(),
		tree.DatabaseLookupFlags{
			AvoidLeased: true,
			Required:    true,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(244753)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244754)
	}
	__antithesis_instrumentation__.Notify(244743)

	if !dbDesc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(244755)
		return nil, errors.WithHint(pgerror.Newf(
			pgcode.InvalidTableDefinition,
			"cannot alter a table's LOCALITY if its database is not multi-region enabled",
		),
			"database must first be multi-region enabled using ALTER DATABASE ... SET PRIMARY REGION <region>",
		)
	} else {
		__antithesis_instrumentation__.Notify(244756)
	}
	__antithesis_instrumentation__.Notify(244744)

	return &alterTableSetLocalityNode{
		n:         *n,
		tableDesc: tableDesc,
		dbDesc:    dbDesc,
	}, nil
}

func (n *alterTableSetLocalityNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(244757)
	return false, nil
}
func (n *alterTableSetLocalityNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(244758)
	return tree.Datums{}
}
func (n *alterTableSetLocalityNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(244759)
}

func (n *alterTableSetLocalityNode) alterTableLocalityGlobalToRegionalByTable(
	params runParams,
) error {
	__antithesis_instrumentation__.Notify(244760)
	if !n.tableDesc.IsLocalityGlobal() {
		__antithesis_instrumentation__.Notify(244766)
		f := params.p.EvalContext().FmtCtx(tree.FmtSimple)
		if err := multiregion.FormatTableLocalityConfig(n.tableDesc.LocalityConfig, f); err != nil {
			__antithesis_instrumentation__.Notify(244768)

			return err
		} else {
			__antithesis_instrumentation__.Notify(244769)
		}
		__antithesis_instrumentation__.Notify(244767)
		return errors.AssertionFailedf(
			"invalid call %q on incorrect table locality %s",
			"alter table locality GLOBAL to REGIONAL BY TABLE",
			f.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(244770)
	}
	__antithesis_instrumentation__.Notify(244761)

	_, dbDesc, err := params.p.Descriptors().GetImmutableDatabaseByID(
		params.ctx, params.p.txn, n.tableDesc.ParentID,
		tree.DatabaseLookupFlags{
			Required:    true,
			AvoidLeased: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(244771)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244772)
	}
	__antithesis_instrumentation__.Notify(244762)

	regionEnumID, err := dbDesc.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(244773)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244774)
	}
	__antithesis_instrumentation__.Notify(244763)

	if err := params.p.alterTableDescLocalityToRegionalByTable(
		params.ctx, n.n.Locality.TableRegion, n.tableDesc, regionEnumID,
	); err != nil {
		__antithesis_instrumentation__.Notify(244775)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244776)
	}
	__antithesis_instrumentation__.Notify(244764)

	if err := n.writeNewTableLocalityAndZoneConfig(
		params,
		n.dbDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(244777)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244778)
	}
	__antithesis_instrumentation__.Notify(244765)

	return nil
}

func (n *alterTableSetLocalityNode) alterTableLocalityToGlobal(params runParams) error {
	__antithesis_instrumentation__.Notify(244779)
	regionEnumID, err := n.dbDesc.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(244783)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244784)
	}
	__antithesis_instrumentation__.Notify(244780)
	err = params.p.alterTableDescLocalityToGlobal(params.ctx, n.tableDesc, regionEnumID)
	if err != nil {
		__antithesis_instrumentation__.Notify(244785)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244786)
	}
	__antithesis_instrumentation__.Notify(244781)

	if err := n.writeNewTableLocalityAndZoneConfig(
		params,
		n.dbDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(244787)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244788)
	}
	__antithesis_instrumentation__.Notify(244782)

	return nil
}

func (n *alterTableSetLocalityNode) alterTableLocalityRegionalByTableToRegionalByTable(
	params runParams,
) error {
	__antithesis_instrumentation__.Notify(244789)
	const operation string = "alter table locality REGIONAL BY TABLE to REGIONAL BY TABLE"
	if !n.tableDesc.IsLocalityRegionalByTable() {
		__antithesis_instrumentation__.Notify(244795)
		return errors.AssertionFailedf(
			"invalid call %q on incorrect table locality. %v",
			operation,
			n.tableDesc.LocalityConfig,
		)
	} else {
		__antithesis_instrumentation__.Notify(244796)
	}
	__antithesis_instrumentation__.Notify(244790)

	_, dbDesc, err := params.p.Descriptors().GetImmutableDatabaseByID(
		params.ctx, params.p.txn, n.tableDesc.ParentID,
		tree.DatabaseLookupFlags{
			Required:    true,
			AvoidLeased: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(244797)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244798)
	}
	__antithesis_instrumentation__.Notify(244791)

	regionEnumID, err := dbDesc.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(244799)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244800)
	}
	__antithesis_instrumentation__.Notify(244792)

	if err := params.p.alterTableDescLocalityToRegionalByTable(
		params.ctx, n.n.Locality.TableRegion, n.tableDesc, regionEnumID,
	); err != nil {
		__antithesis_instrumentation__.Notify(244801)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244802)
	}
	__antithesis_instrumentation__.Notify(244793)

	if err := n.writeNewTableLocalityAndZoneConfig(
		params,
		n.dbDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(244803)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244804)
	}
	__antithesis_instrumentation__.Notify(244794)

	return nil
}

func (n *alterTableSetLocalityNode) alterTableLocalityToRegionalByRow(
	params runParams, newLocality *tree.Locality,
) error {
	__antithesis_instrumentation__.Notify(244805)
	mayNeedImplicitCRDBRegionCol := false
	var mutationIdxAllowedInSameTxn *int
	var newColumnName *tree.Name

	partColName := newLocality.RegionalByRowColumn

	primaryIndexColIdxStart := 0
	if n.tableDesc.IsLocalityRegionalByRow() {
		__antithesis_instrumentation__.Notify(244812)
		as := n.tableDesc.LocalityConfig.GetRegionalByRow().As

		defaultColumnSpecified := as == nil && func() bool {
			__antithesis_instrumentation__.Notify(244814)
			return partColName == tree.RegionalByRowRegionNotSpecifiedName == true
		}() == true
		sameAsColumnSpecified := as != nil && func() bool {
			__antithesis_instrumentation__.Notify(244815)
			return *as == string(partColName) == true
		}() == true
		if defaultColumnSpecified || func() bool {
			__antithesis_instrumentation__.Notify(244816)
			return sameAsColumnSpecified == true
		}() == true {
			__antithesis_instrumentation__.Notify(244817)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(244818)
		}
		__antithesis_instrumentation__.Notify(244813)

		primaryIndexColIdxStart = int(n.tableDesc.PrimaryIndex.Partitioning.NumImplicitColumns)
	} else {
		__antithesis_instrumentation__.Notify(244819)
	}
	__antithesis_instrumentation__.Notify(244806)

	for _, idx := range n.tableDesc.AllIndexes() {
		__antithesis_instrumentation__.Notify(244820)
		if idx.IsSharded() {
			__antithesis_instrumentation__.Notify(244821)
			return pgerror.Newf(
				pgcode.FeatureNotSupported,
				"cannot convert %s to REGIONAL BY ROW as the table contains hash sharded indexes",
				tree.Name(n.tableDesc.GetName()),
			)
		} else {
			__antithesis_instrumentation__.Notify(244822)
		}
	}
	__antithesis_instrumentation__.Notify(244807)

	if newLocality.RegionalByRowColumn == tree.RegionalByRowRegionNotSpecifiedName {
		__antithesis_instrumentation__.Notify(244823)
		partColName = tree.RegionalByRowRegionDefaultColName
		mayNeedImplicitCRDBRegionCol = true
	} else {
		__antithesis_instrumentation__.Notify(244824)
	}
	__antithesis_instrumentation__.Notify(244808)

	partCol, err := n.tableDesc.FindColumnWithName(partColName)
	createDefaultRegionCol := mayNeedImplicitCRDBRegionCol && func() bool {
		__antithesis_instrumentation__.Notify(244825)
		return sqlerrors.IsUndefinedColumnError(err) == true
	}() == true
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(244826)
		return !createDefaultRegionCol == true
	}() == true {
		__antithesis_instrumentation__.Notify(244827)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244828)
	}
	__antithesis_instrumentation__.Notify(244809)

	enumTypeID, err := n.dbDesc.MultiRegionEnumID()
	if err != nil {
		__antithesis_instrumentation__.Notify(244829)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244830)
	}
	__antithesis_instrumentation__.Notify(244810)
	enumOID := typedesc.TypeIDToOID(enumTypeID)

	var newColumnID *descpb.ColumnID
	var newColumnDefaultExpr *string

	if !createDefaultRegionCol {
		__antithesis_instrumentation__.Notify(244831)

		if !partCol.Public() {
			__antithesis_instrumentation__.Notify(244834)
			return colinfo.NewUndefinedColumnError(string(partColName))
		} else {
			__antithesis_instrumentation__.Notify(244835)
		}
		__antithesis_instrumentation__.Notify(244832)

		if partCol.GetType().Oid() != typedesc.TypeIDToOID(enumTypeID) {
			__antithesis_instrumentation__.Notify(244836)
			return pgerror.Newf(
				pgcode.InvalidTableDefinition,
				"cannot use column %s for REGIONAL BY ROW table as it does not have the %s type",
				partColName,
				tree.RegionEnum,
			)
		} else {
			__antithesis_instrumentation__.Notify(244837)
		}
		__antithesis_instrumentation__.Notify(244833)

		if partCol.IsNullable() {
			__antithesis_instrumentation__.Notify(244838)
			return errors.WithHintf(
				pgerror.Newf(
					pgcode.InvalidTableDefinition,
					"cannot use column %s for REGIONAL BY ROW table as it may contain NULL values",
					partColName,
				),
				"Add the NOT NULL constraint first using ALTER TABLE %s ALTER COLUMN %s SET NOT NULL.",
				tree.Name(n.tableDesc.Name),
				partColName,
			)
		} else {
			__antithesis_instrumentation__.Notify(244839)
		}
	} else {
		__antithesis_instrumentation__.Notify(244840)

		primaryRegion, err := n.dbDesc.PrimaryRegionName()
		if err != nil {
			__antithesis_instrumentation__.Notify(244846)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244847)
		}
		__antithesis_instrumentation__.Notify(244841)

		defaultColDef := &tree.AlterTableAddColumn{
			ColumnDef: regionalByRowDefaultColDef(
				enumOID,
				regionalByRowRegionDefaultExpr(enumOID, tree.Name(primaryRegion)),
				maybeRegionalByRowOnUpdateExpr(params.EvalContext(), enumOID),
			),
		}
		tn, err := params.p.getQualifiedTableName(params.ctx, n.tableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(244848)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244849)
		}
		__antithesis_instrumentation__.Notify(244842)
		if err := params.p.addColumnImpl(
			params,
			&alterTableNode{
				tableDesc: n.tableDesc,
				n: &tree.AlterTable{
					Cmds: []tree.AlterTableCmd{defaultColDef},
				},
			},
			tn,
			n.tableDesc,
			defaultColDef,
		); err != nil {
			__antithesis_instrumentation__.Notify(244850)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244851)
		}
		__antithesis_instrumentation__.Notify(244843)

		mutationIdx := len(n.tableDesc.Mutations) - 1
		mutationIdxAllowedInSameTxn = &mutationIdx
		newColumnName = &partColName

		version := params.ExecCfg().Settings.Version.ActiveVersion(params.ctx)
		if err := n.tableDesc.AllocateIDs(params.ctx, version); err != nil {
			__antithesis_instrumentation__.Notify(244852)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244853)
		}
		__antithesis_instrumentation__.Notify(244844)

		col := n.tableDesc.Mutations[mutationIdx].GetColumn()
		finalDefaultExpr, err := schemaexpr.SanitizeVarFreeExpr(
			params.ctx,
			regionalByRowGatewayRegionDefaultExpr(enumOID),
			col.Type,
			"REGIONAL BY ROW DEFAULT",
			params.p.SemaCtx(),
			tree.VolatilityVolatile,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(244854)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244855)
		}
		__antithesis_instrumentation__.Notify(244845)
		s := tree.Serialize(finalDefaultExpr)
		newColumnDefaultExpr = &s
		newColumnID = &col.ID
	}
	__antithesis_instrumentation__.Notify(244811)
	return n.alterTableLocalityFromOrToRegionalByRow(
		params,
		tabledesc.LocalityConfigRegionalByRow(newLocality.RegionalByRowColumn),
		mutationIdxAllowedInSameTxn,
		newColumnName,
		newColumnID,
		newColumnDefaultExpr,
		n.tableDesc.PrimaryIndex.KeyColumnNames[primaryIndexColIdxStart:],
		n.tableDesc.PrimaryIndex.KeyColumnDirections[primaryIndexColIdxStart:],
	)
}

func (n *alterTableSetLocalityNode) alterTableLocalityFromOrToRegionalByRow(
	params runParams,
	newLocalityConfig catpb.LocalityConfig,
	mutationIdxAllowedInSameTxn *int,
	newColumnName *tree.Name,
	newColumnID *descpb.ColumnID,
	newColumnDefaultExpr *string,
	pkColumnNames []string,
	pkColumnDirections []descpb.IndexDescriptor_Direction,
) error {
	__antithesis_instrumentation__.Notify(244856)

	cols := make([]tree.IndexElem, len(pkColumnNames))
	for i, col := range pkColumnNames {
		__antithesis_instrumentation__.Notify(244859)
		cols[i] = tree.IndexElem{
			Column: tree.Name(col),
		}
		switch dir := pkColumnDirections[i]; dir {
		case descpb.IndexDescriptor_ASC:
			__antithesis_instrumentation__.Notify(244860)
			cols[i].Direction = tree.Ascending
		case descpb.IndexDescriptor_DESC:
			__antithesis_instrumentation__.Notify(244861)
			cols[i].Direction = tree.Descending
		default:
			__antithesis_instrumentation__.Notify(244862)
			return errors.AssertionFailedf("unknown direction: %v", dir)
		}
	}
	__antithesis_instrumentation__.Notify(244857)

	if err := params.p.AlterPrimaryKey(
		params.ctx,
		n.tableDesc,
		tree.AlterTableAlterPrimaryKey{
			Name:    tree.Name(n.tableDesc.PrimaryIndex.Name),
			Columns: cols,
		},
		&alterPrimaryKeyLocalitySwap{
			localityConfigSwap: descpb.PrimaryKeySwap_LocalityConfigSwap{
				OldLocalityConfig:                 *n.tableDesc.LocalityConfig,
				NewLocalityConfig:                 newLocalityConfig,
				NewRegionalByRowColumnID:          newColumnID,
				NewRegionalByRowColumnDefaultExpr: newColumnDefaultExpr,
			},
			mutationIdxAllowedInSameTxn: mutationIdxAllowedInSameTxn,
			newColumnName:               newColumnName,
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(244863)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244864)
	}
	__antithesis_instrumentation__.Notify(244858)

	return params.p.writeSchemaChange(
		params.ctx,
		n.tableDesc,
		n.tableDesc.ClusterVersion().NextMutationID,
		tree.AsStringWithFQNames(&n.n, params.Ann()),
	)
}

func (n *alterTableSetLocalityNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(244865)
	newLocality := n.n.Locality
	existingLocality := n.tableDesc.LocalityConfig

	existingLocalityTelemetryName, err := existingLocality.TelemetryName()
	if err != nil {
		__antithesis_instrumentation__.Notify(244869)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244870)
	}
	__antithesis_instrumentation__.Notify(244866)
	telemetry.Inc(
		sqltelemetry.AlterTableLocalityCounter(
			existingLocalityTelemetryName,
			newLocality.TelemetryName(),
		),
	)

	if err := params.p.validateZoneConfigForMultiRegionTableWasNotModifiedByUser(
		params.ctx,
		n.dbDesc,
		n.tableDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(244871)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244872)
	}
	__antithesis_instrumentation__.Notify(244867)

	switch existingLocality.Locality.(type) {
	case *catpb.LocalityConfig_Global_:
		__antithesis_instrumentation__.Notify(244873)
		switch newLocality.LocalityLevel {
		case tree.LocalityLevelGlobal:
			__antithesis_instrumentation__.Notify(244877)
			if err := n.alterTableLocalityToGlobal(params); err != nil {
				__antithesis_instrumentation__.Notify(244881)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244882)
			}
		case tree.LocalityLevelRow:
			__antithesis_instrumentation__.Notify(244878)
			if err := n.alterTableLocalityToRegionalByRow(
				params,
				newLocality,
			); err != nil {
				__antithesis_instrumentation__.Notify(244883)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244884)
			}
		case tree.LocalityLevelTable:
			__antithesis_instrumentation__.Notify(244879)
			if err := n.alterTableLocalityGlobalToRegionalByTable(params); err != nil {
				__antithesis_instrumentation__.Notify(244885)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244886)
			}
		default:
			__antithesis_instrumentation__.Notify(244880)
			return errors.AssertionFailedf("unknown table locality: %v", newLocality)
		}
	case *catpb.LocalityConfig_RegionalByTable_:
		__antithesis_instrumentation__.Notify(244874)
		switch newLocality.LocalityLevel {
		case tree.LocalityLevelGlobal:
			__antithesis_instrumentation__.Notify(244887)
			if err := n.alterTableLocalityToGlobal(params); err != nil {
				__antithesis_instrumentation__.Notify(244891)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244892)
			}
		case tree.LocalityLevelRow:
			__antithesis_instrumentation__.Notify(244888)
			if err := n.alterTableLocalityToRegionalByRow(
				params,
				newLocality,
			); err != nil {
				__antithesis_instrumentation__.Notify(244893)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244894)
			}
		case tree.LocalityLevelTable:
			__antithesis_instrumentation__.Notify(244889)
			if err := n.alterTableLocalityRegionalByTableToRegionalByTable(params); err != nil {
				__antithesis_instrumentation__.Notify(244895)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244896)
			}
		default:
			__antithesis_instrumentation__.Notify(244890)
			return errors.AssertionFailedf("unknown table locality: %v", newLocality)
		}
	case *catpb.LocalityConfig_RegionalByRow_:
		__antithesis_instrumentation__.Notify(244875)
		explicitColStart := n.tableDesc.PrimaryIndex.Partitioning.NumImplicitColumns
		switch newLocality.LocalityLevel {
		case tree.LocalityLevelGlobal:
			__antithesis_instrumentation__.Notify(244897)
			return n.alterTableLocalityFromOrToRegionalByRow(
				params,
				tabledesc.LocalityConfigGlobal(),
				nil,
				nil,
				nil,
				nil,
				n.tableDesc.PrimaryIndex.KeyColumnNames[explicitColStart:],
				n.tableDesc.PrimaryIndex.KeyColumnDirections[explicitColStart:],
			)
		case tree.LocalityLevelRow:
			__antithesis_instrumentation__.Notify(244898)
			if err := n.alterTableLocalityToRegionalByRow(
				params,
				newLocality,
			); err != nil {
				__antithesis_instrumentation__.Notify(244901)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244902)
			}
		case tree.LocalityLevelTable:
			__antithesis_instrumentation__.Notify(244899)
			return n.alterTableLocalityFromOrToRegionalByRow(
				params,
				tabledesc.LocalityConfigRegionalByTable(n.n.Locality.TableRegion),
				nil,
				nil,
				nil,
				nil,
				n.tableDesc.PrimaryIndex.KeyColumnNames[explicitColStart:],
				n.tableDesc.PrimaryIndex.KeyColumnDirections[explicitColStart:],
			)
		default:
			__antithesis_instrumentation__.Notify(244900)
			return errors.AssertionFailedf("unknown table locality: %v", newLocality)
		}
	default:
		__antithesis_instrumentation__.Notify(244876)
		return errors.AssertionFailedf("unknown table locality: %v", existingLocality)
	}
	__antithesis_instrumentation__.Notify(244868)

	return params.p.logEvent(params.ctx,
		n.tableDesc.ID,
		&eventpb.AlterTable{
			TableName: n.n.Name.String(),
		})
}

func (n *alterTableSetLocalityNode) writeNewTableLocalityAndZoneConfig(
	params runParams, dbDesc catalog.DatabaseDescriptor,
) error {
	__antithesis_instrumentation__.Notify(244903)

	if err := params.p.writeSchemaChange(
		params.ctx,
		n.tableDesc,
		descpb.InvalidMutationID,
		tree.AsStringWithFQNames(&n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(244907)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244908)
	}
	__antithesis_instrumentation__.Notify(244904)

	regionConfig, err := SynthesizeRegionConfig(params.ctx, params.p.txn, dbDesc.GetID(), params.p.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(244909)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244910)
	}
	__antithesis_instrumentation__.Notify(244905)

	if err := ApplyZoneConfigForMultiRegionTable(
		params.ctx,
		params.p.txn,
		params.p.ExecCfg(),
		regionConfig,
		n.tableDesc,
		ApplyZoneConfigForMultiRegionTableOptionTableAndIndexes,
	); err != nil {
		__antithesis_instrumentation__.Notify(244911)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244912)
	}
	__antithesis_instrumentation__.Notify(244906)

	return nil
}

func (p *planner) alterTableDescLocalityToRegionalByTable(
	ctx context.Context, region tree.Name, tableDesc *tabledesc.Mutable, regionEnumID descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(244913)
	if tableDesc.GetMultiRegionEnumDependencyIfExists() {
		__antithesis_instrumentation__.Notify(244916)
		typesDependedOn := []descpb.ID{regionEnumID}
		if err := p.removeTypeBackReferences(ctx, typesDependedOn, tableDesc.GetID(),
			fmt.Sprintf("remove back ref on mr-enum %d for table %d", regionEnumID, tableDesc.GetID()),
		); err != nil {
			__antithesis_instrumentation__.Notify(244917)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244918)
		}
	} else {
		__antithesis_instrumentation__.Notify(244919)
	}
	__antithesis_instrumentation__.Notify(244914)
	tableDesc.SetTableLocalityRegionalByTable(region)
	if tableDesc.GetMultiRegionEnumDependencyIfExists() {
		__antithesis_instrumentation__.Notify(244920)
		return p.addTypeBackReference(
			ctx, regionEnumID, tableDesc.ID,
			fmt.Sprintf("add back ref on mr-enum %d for table %d", regionEnumID, tableDesc.GetID()),
		)
	} else {
		__antithesis_instrumentation__.Notify(244921)
	}
	__antithesis_instrumentation__.Notify(244915)
	return nil
}

func (p *planner) alterTableDescLocalityToGlobal(
	ctx context.Context, tableDesc *tabledesc.Mutable, regionEnumID descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(244922)
	if tableDesc.GetMultiRegionEnumDependencyIfExists() {
		__antithesis_instrumentation__.Notify(244924)
		typesDependedOn := []descpb.ID{regionEnumID}
		if err := p.removeTypeBackReferences(ctx, typesDependedOn, tableDesc.GetID(),
			fmt.Sprintf("remove back ref no mr-enum %d for table %d", regionEnumID, tableDesc.GetID()),
		); err != nil {
			__antithesis_instrumentation__.Notify(244925)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244926)
		}
	} else {
		__antithesis_instrumentation__.Notify(244927)
	}
	__antithesis_instrumentation__.Notify(244923)
	tableDesc.SetTableLocalityGlobal()
	return nil
}

func setNewLocalityConfig(
	ctx context.Context,
	desc *tabledesc.Mutable,
	txn *kv.Txn,
	b *kv.Batch,
	config catpb.LocalityConfig,
	kvTrace bool,
	descsCol *descs.Collection,
) error {
	__antithesis_instrumentation__.Notify(244928)
	getMultiRegionTypeDesc := func() (*typedesc.Mutable, error) {
		__antithesis_instrumentation__.Notify(244932)
		_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
			ctx, txn, desc.GetParentID(), tree.DatabaseLookupFlags{
				Required:    true,
				AvoidLeased: true,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(244935)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244936)
		}
		__antithesis_instrumentation__.Notify(244933)

		regionEnumID, err := dbDesc.MultiRegionEnumID()
		if err != nil {
			__antithesis_instrumentation__.Notify(244937)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244938)
		}
		__antithesis_instrumentation__.Notify(244934)
		return descsCol.GetMutableTypeVersionByID(ctx, txn, regionEnumID)
	}
	__antithesis_instrumentation__.Notify(244929)

	if desc.GetMultiRegionEnumDependencyIfExists() {
		__antithesis_instrumentation__.Notify(244939)
		typ, err := getMultiRegionTypeDesc()
		if err != nil {
			__antithesis_instrumentation__.Notify(244941)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244942)
		}
		__antithesis_instrumentation__.Notify(244940)
		typ.RemoveReferencingDescriptorID(desc.GetID())
		if err := descsCol.WriteDescToBatch(ctx, kvTrace, typ, b); err != nil {
			__antithesis_instrumentation__.Notify(244943)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244944)
		}
	} else {
		__antithesis_instrumentation__.Notify(244945)
	}
	__antithesis_instrumentation__.Notify(244930)
	desc.LocalityConfig = &config

	if desc.GetMultiRegionEnumDependencyIfExists() {
		__antithesis_instrumentation__.Notify(244946)
		typ, err := getMultiRegionTypeDesc()
		if err != nil {
			__antithesis_instrumentation__.Notify(244948)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244949)
		}
		__antithesis_instrumentation__.Notify(244947)
		typ.AddReferencingDescriptorID(desc.GetID())
		if err := descsCol.WriteDescToBatch(ctx, kvTrace, typ, b); err != nil {
			__antithesis_instrumentation__.Notify(244950)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244951)
		}
	} else {
		__antithesis_instrumentation__.Notify(244952)
	}
	__antithesis_instrumentation__.Notify(244931)
	return nil
}
