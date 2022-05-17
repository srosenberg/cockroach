package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type renameTableNode struct {
	n            *tree.RenameTable
	oldTn, newTn *tree.TableName
	tableDesc    *tabledesc.Mutable
}

func (p *planner) RenameTable(ctx context.Context, n *tree.RenameTable) (planNode, error) {
	__antithesis_instrumentation__.Notify(565903)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"RENAME TABLE/VIEW/SEQUENCE",
	); err != nil {
		__antithesis_instrumentation__.Notify(565912)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565913)
	}
	__antithesis_instrumentation__.Notify(565904)

	oldTn := n.Name.ToTableName()
	newTn := n.NewName.ToTableName()
	toRequire := tree.ResolveRequireTableOrViewDesc
	if n.IsView {
		__antithesis_instrumentation__.Notify(565914)
		toRequire = tree.ResolveRequireViewDesc
	} else {
		__antithesis_instrumentation__.Notify(565915)
		if n.IsSequence {
			__antithesis_instrumentation__.Notify(565916)
			toRequire = tree.ResolveRequireSequenceDesc
		} else {
			__antithesis_instrumentation__.Notify(565917)
		}
	}
	__antithesis_instrumentation__.Notify(565905)

	_, tableDesc, err := p.ResolveMutableTableDescriptor(ctx, &oldTn, !n.IfExists, toRequire)
	if err != nil {
		__antithesis_instrumentation__.Notify(565918)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565919)
	}
	__antithesis_instrumentation__.Notify(565906)
	if tableDesc == nil {
		__antithesis_instrumentation__.Notify(565920)

		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(565921)
	}
	__antithesis_instrumentation__.Notify(565907)

	if err := checkViewMatchesMaterialized(tableDesc, n.IsView, n.IsMaterialized); err != nil {
		__antithesis_instrumentation__.Notify(565922)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565923)
	}
	__antithesis_instrumentation__.Notify(565908)

	if tableDesc.State != descpb.DescriptorState_PUBLIC {
		__antithesis_instrumentation__.Notify(565924)
		return nil, sqlerrors.NewUndefinedRelationError(&oldTn)
	} else {
		__antithesis_instrumentation__.Notify(565925)
	}
	__antithesis_instrumentation__.Notify(565909)

	if err := p.CheckPrivilege(ctx, tableDesc, privilege.DROP); err != nil {
		__antithesis_instrumentation__.Notify(565926)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565927)
	}
	__antithesis_instrumentation__.Notify(565910)

	for _, dependent := range tableDesc.DependedOnBy {
		__antithesis_instrumentation__.Notify(565928)
		if !dependent.ByID {
			__antithesis_instrumentation__.Notify(565929)
			return nil, p.dependentViewError(
				ctx, string(tableDesc.DescriptorType()), oldTn.String(),
				tableDesc.ParentID, dependent.ID, "rename",
			)
		} else {
			__antithesis_instrumentation__.Notify(565930)
		}
	}
	__antithesis_instrumentation__.Notify(565911)

	return &renameTableNode{n: n, oldTn: &oldTn, newTn: &newTn, tableDesc: tableDesc}, nil
}

func (n *renameTableNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(565931) }

func (n *renameTableNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(565932)
	p := params.p
	ctx := params.ctx
	tableDesc := n.tableDesc
	oldTn := n.oldTn
	oldNameKey := descpb.NameInfo{
		ParentID:       tableDesc.GetParentID(),
		ParentSchemaID: tableDesc.GetParentSchemaID(),
		Name:           tableDesc.GetName(),
	}

	var targetDbDesc catalog.DatabaseDescriptor
	var targetSchemaDesc catalog.SchemaDescriptor

	newTn := n.newTn
	if !newTn.ExplicitSchema && func() bool {
		__antithesis_instrumentation__.Notify(565943)
		return !newTn.ExplicitCatalog == true
	}() == true {
		__antithesis_instrumentation__.Notify(565944)
		newTn.ObjectNamePrefix = oldTn.ObjectNamePrefix
		var err error
		targetDbDesc, err = p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn,
			string(oldTn.CatalogName), tree.DatabaseLookupFlags{Required: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(565946)
			return err
		} else {
			__antithesis_instrumentation__.Notify(565947)
		}
		__antithesis_instrumentation__.Notify(565945)

		targetSchemaDesc, err = p.Descriptors().GetMutableSchemaByName(
			ctx, p.txn, targetDbDesc, oldTn.Schema(), tree.SchemaLookupFlags{
				Required:       true,
				RequireMutable: true,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(565948)
			return err
		} else {
			__antithesis_instrumentation__.Notify(565949)
		}
	} else {
		__antithesis_instrumentation__.Notify(565950)

		params.p.BufferClientNotice(
			ctx,
			errors.WithHintf(
				pgnotice.Newf("renaming tables with a qualification is deprecated"),
				"use ALTER TABLE %s RENAME TO %s instead",
				n.n.Name.String(),
				newTn.Table(),
			),
		)

		newUn := newTn.ToUnresolvedObjectName()
		var prefix tree.ObjectNamePrefix
		var err error
		targetDbDesc, targetSchemaDesc, prefix, err = p.ResolveTargetObject(ctx, newUn)
		if err != nil {
			__antithesis_instrumentation__.Notify(565952)
			return err
		} else {
			__antithesis_instrumentation__.Notify(565953)
		}
		__antithesis_instrumentation__.Notify(565951)
		newTn.ObjectNamePrefix = prefix
	}
	__antithesis_instrumentation__.Notify(565933)

	if err := p.CheckPrivilege(ctx, targetDbDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(565954)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565955)
	}
	__antithesis_instrumentation__.Notify(565934)

	if newTn.Catalog() == oldTn.Catalog() && func() bool {
		__antithesis_instrumentation__.Notify(565956)
		return newTn.Schema() != oldTn.Schema() == true
	}() == true {
		__antithesis_instrumentation__.Notify(565957)
		return errors.WithHint(
			pgerror.Newf(pgcode.InvalidName, "cannot change schema of table with RENAME"),
			"use ALTER TABLE ... SET SCHEMA instead",
		)
	} else {
		__antithesis_instrumentation__.Notify(565958)
	}
	__antithesis_instrumentation__.Notify(565935)

	if oldTn.Catalog() != newTn.Catalog() {
		__antithesis_instrumentation__.Notify(565959)

		if p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) {
			__antithesis_instrumentation__.Notify(565963)
			return pgerror.Newf(pgcode.FeatureNotSupported,
				"cannot change database of table using alter table rename to")
		} else {
			__antithesis_instrumentation__.Notify(565964)
		}
		__antithesis_instrumentation__.Notify(565960)

		if oldTn.Schema() != string(tree.PublicSchemaName) || func() bool {
			__antithesis_instrumentation__.Notify(565965)
			return newTn.Schema() != string(tree.PublicSchemaName) == true
		}() == true {
			__antithesis_instrumentation__.Notify(565966)
			return pgerror.Newf(pgcode.InvalidName,
				"cannot change database of table unless both the old and new schemas are the public schema in each database")
		} else {
			__antithesis_instrumentation__.Notify(565967)
		}
		__antithesis_instrumentation__.Notify(565961)

		columns := make([]descpb.ColumnDescriptor, 0, len(tableDesc.Columns)+len(tableDesc.Mutations))
		columns = append(columns, tableDesc.Columns...)
		for _, m := range tableDesc.Mutations {
			__antithesis_instrumentation__.Notify(565968)
			if col := m.GetColumn(); col != nil {
				__antithesis_instrumentation__.Notify(565969)
				columns = append(columns, *col)
			} else {
				__antithesis_instrumentation__.Notify(565970)
			}
		}
		__antithesis_instrumentation__.Notify(565962)
		for _, c := range columns {
			__antithesis_instrumentation__.Notify(565971)
			if c.Type.UserDefined() {
				__antithesis_instrumentation__.Notify(565972)
				return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
					"cannot change database of table if any of its column types are user-defined")
			} else {
				__antithesis_instrumentation__.Notify(565973)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(565974)
	}
	__antithesis_instrumentation__.Notify(565936)

	if oldTn.Catalog() != newTn.Catalog() {
		__antithesis_instrumentation__.Notify(565975)
		err := n.checkForCrossDbReferences(ctx, p, targetDbDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(565976)
			return err
		} else {
			__antithesis_instrumentation__.Notify(565977)
		}
	} else {
		__antithesis_instrumentation__.Notify(565978)
	}
	__antithesis_instrumentation__.Notify(565937)

	if oldTn.Catalog() == newTn.Catalog() && func() bool {
		__antithesis_instrumentation__.Notify(565979)
		return oldTn.Schema() == newTn.Schema() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(565980)
		return oldTn.Table() == newTn.Table() == true
	}() == true {
		__antithesis_instrumentation__.Notify(565981)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(565982)
	}
	__antithesis_instrumentation__.Notify(565938)

	err := p.Descriptors().Direct().CheckObjectCollision(
		params.ctx,
		params.p.txn,
		targetDbDesc.GetID(),
		targetSchemaDesc.GetID(),
		newTn,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(565983)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565984)
	}
	__antithesis_instrumentation__.Notify(565939)

	tableDesc.SetName(newTn.Table())
	tableDesc.ParentID = targetDbDesc.GetID()

	if err := validateDescriptor(ctx, p, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(565985)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565986)
	}
	__antithesis_instrumentation__.Notify(565940)

	b := p.txn.NewBatch()
	p.renameNamespaceEntry(ctx, b, oldNameKey, tableDesc)

	if err := p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(565987)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565988)
	}
	__antithesis_instrumentation__.Notify(565941)

	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(565989)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565990)
	}
	__antithesis_instrumentation__.Notify(565942)

	return p.logEvent(ctx,
		tableDesc.ID,
		&eventpb.RenameTable{
			TableName:    oldTn.FQString(),
			NewTableName: newTn.FQString(),
		})
}

func (n *renameTableNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(565991)
	return false, nil
}
func (n *renameTableNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(565992)
	return tree.Datums{}
}
func (n *renameTableNode) Close(context.Context) { __antithesis_instrumentation__.Notify(565993) }

func (p *planner) dependentViewError(
	ctx context.Context, typeName, objName string, parentID, viewID descpb.ID, op string,
) error {
	__antithesis_instrumentation__.Notify(565994)
	viewDesc, err := p.Descriptors().Direct().MustGetTableDescByID(ctx, p.txn, viewID)
	if err != nil {
		__antithesis_instrumentation__.Notify(565997)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565998)
	}
	__antithesis_instrumentation__.Notify(565995)
	viewName := viewDesc.GetName()
	if viewDesc.GetParentID() != parentID {
		__antithesis_instrumentation__.Notify(565999)
		viewFQName, err := p.getQualifiedTableName(ctx, viewDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(566001)
			log.Warningf(ctx, "unable to retrieve name of view %d: %v", viewID, err)
			return sqlerrors.NewDependentObjectErrorf(
				"cannot %s %s %q because a view depends on it",
				op, typeName, objName)
		} else {
			__antithesis_instrumentation__.Notify(566002)
		}
		__antithesis_instrumentation__.Notify(566000)
		viewName = viewFQName.FQString()
	} else {
		__antithesis_instrumentation__.Notify(566003)
	}
	__antithesis_instrumentation__.Notify(565996)
	return errors.WithHintf(
		sqlerrors.NewDependentObjectErrorf("cannot %s %s %q because view %q depends on it",
			op, typeName, objName, viewName),
		"you can drop %s instead.", viewName)
}

func (n *renameTableNode) checkForCrossDbReferences(
	ctx context.Context, p *planner, targetDbDesc catalog.DatabaseDescriptor,
) error {
	__antithesis_instrumentation__.Notify(566004)
	tableDesc := n.tableDesc

	checkFkForCrossDbDep := func(fk *descpb.ForeignKeyConstraint, refTableID bool) error {
		__antithesis_instrumentation__.Notify(566009)
		if allowCrossDatabaseFKs.Get(&p.execCfg.Settings.SV) {
			__antithesis_instrumentation__.Notify(566014)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(566015)
		}
		__antithesis_instrumentation__.Notify(566010)
		tableID := fk.ReferencedTableID
		if !refTableID {
			__antithesis_instrumentation__.Notify(566016)
			tableID = fk.OriginTableID
		} else {
			__antithesis_instrumentation__.Notify(566017)
		}
		__antithesis_instrumentation__.Notify(566011)

		referencedTable, err := p.Descriptors().GetImmutableTableByID(ctx, p.txn, tableID,
			tree.ObjectLookupFlags{
				CommonLookupFlags: tree.CommonLookupFlags{
					Required:    true,
					AvoidLeased: true,
				},
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(566018)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566019)
		}
		__antithesis_instrumentation__.Notify(566012)

		if referencedTable.GetParentID() == targetDbDesc.GetID() {
			__antithesis_instrumentation__.Notify(566020)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(566021)
		}
		__antithesis_instrumentation__.Notify(566013)

		return errors.WithHintf(
			pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"a foreign key constraint %q will exist between databases after rename "+
					"(see the '%s' cluster setting)",
				fk.Name,
				allowCrossDatabaseFKsSetting),
			crossDBReferenceDeprecationHint(),
		)
	}
	__antithesis_instrumentation__.Notify(566005)

	type crossDBDepType int
	const owner, reference crossDBDepType = 0, 1
	checkDepForCrossDbRef := func(depID descpb.ID, depType crossDBDepType) error {
		__antithesis_instrumentation__.Notify(566022)
		dependentObject, err := p.Descriptors().GetImmutableTableByID(ctx, p.txn, depID,
			tree.ObjectLookupFlags{
				CommonLookupFlags: tree.CommonLookupFlags{
					Required:    true,
					AvoidLeased: true,
				}})
		if err != nil {
			__antithesis_instrumentation__.Notify(566026)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566027)
		}
		__antithesis_instrumentation__.Notify(566023)

		if dependentObject.GetParentID() == targetDbDesc.GetID() {
			__antithesis_instrumentation__.Notify(566028)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(566029)
		}
		__antithesis_instrumentation__.Notify(566024)

		switch {
		case tableDesc.IsTable():
			__antithesis_instrumentation__.Notify(566030)

			switch {
			case dependentObject.IsView():
				__antithesis_instrumentation__.Notify(566035)
				if !allowCrossDatabaseViews.Get(&p.execCfg.Settings.SV) {
					__antithesis_instrumentation__.Notify(566039)
					return errors.WithHintf(
						pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
							"a view %q reference to this table will refer to another databases after rename "+
								"(see the '%s' cluster setting)",
							dependentObject.GetName(),
							allowCrossDatabaseViewsSetting),
						crossDBReferenceDeprecationHint(),
					)
				} else {
					__antithesis_instrumentation__.Notify(566040)
				}
			case dependentObject.IsSequence() && func() bool {
				__antithesis_instrumentation__.Notify(566041)
				return depType == owner == true
			}() == true:
				__antithesis_instrumentation__.Notify(566036)
				if !allowCrossDatabaseSeqOwner.Get(&p.execCfg.Settings.SV) {
					__antithesis_instrumentation__.Notify(566042)
					return errors.WithHintf(
						pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
							"a sequence %q will be OWNED BY a table in a different database after rename "+
								"(see the '%s' cluster setting)",
							dependentObject.GetName(),
							allowCrossDatabaseSeqOwnerSetting),
						crossDBReferenceDeprecationHint(),
					)
				} else {
					__antithesis_instrumentation__.Notify(566043)
				}
			case dependentObject.IsSequence() && func() bool {
				__antithesis_instrumentation__.Notify(566044)
				return depType == reference == true
			}() == true:
				__antithesis_instrumentation__.Notify(566037)
				if !allowCrossDatabaseSeqReferences.Get(&p.execCfg.Settings.SV) {
					__antithesis_instrumentation__.Notify(566045)
					return errors.WithHintf(
						pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
							"a sequence %q will be referenced by a table in a different database after rename "+
								"(see the '%s' cluster setting)",
							dependentObject.GetName(),
							allowCrossDatabaseSeqOwnerSetting),
						crossDBReferenceDeprecationHint(),
					)
				} else {
					__antithesis_instrumentation__.Notify(566046)
				}
			default:
				__antithesis_instrumentation__.Notify(566038)
			}
		case tableDesc.IsView():
			__antithesis_instrumentation__.Notify(566031)
			if !allowCrossDatabaseViews.Get(&p.execCfg.Settings.SV) {
				__antithesis_instrumentation__.Notify(566047)

				return errors.WithHintf(
					pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"this view will reference a table %q in another databases after rename "+
							"(see the '%s' cluster setting)",
						dependentObject.GetName(),
						allowCrossDatabaseViewsSetting),
					crossDBReferenceDeprecationHint(),
				)
			} else {
				__antithesis_instrumentation__.Notify(566048)
			}
		case tableDesc.IsSequence() && func() bool {
			__antithesis_instrumentation__.Notify(566049)
			return depType == reference == true
		}() == true:
			__antithesis_instrumentation__.Notify(566032)
			if !allowCrossDatabaseSeqReferences.Get(&p.execCfg.Settings.SV) {
				__antithesis_instrumentation__.Notify(566050)

				return errors.WithHintf(
					pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"this sequence will be referenced by a table %q in a different database after rename "+
							"(see the '%s' cluster setting)",
						dependentObject.GetName(),
						allowCrossDatabaseSeqReferencesSetting),
					crossDBReferenceDeprecationHint(),
				)
			} else {
				__antithesis_instrumentation__.Notify(566051)
			}
		case tableDesc.IsSequence() && func() bool {
			__antithesis_instrumentation__.Notify(566052)
			return depType == owner == true
		}() == true:
			__antithesis_instrumentation__.Notify(566033)
			if !allowCrossDatabaseSeqOwner.Get(&p.execCfg.Settings.SV) {
				__antithesis_instrumentation__.Notify(566053)

				return errors.WithHintf(
					pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"this sequence will be OWNED BY a table %q in a different database after rename "+
							"(see the '%s' cluster setting)",
						dependentObject.GetName(),
						allowCrossDatabaseSeqReferencesSetting),
					crossDBReferenceDeprecationHint(),
				)
			} else {
				__antithesis_instrumentation__.Notify(566054)
			}
		default:
			__antithesis_instrumentation__.Notify(566034)
		}
		__antithesis_instrumentation__.Notify(566025)
		return nil
	}
	__antithesis_instrumentation__.Notify(566006)

	checkTypeDepForCrossDbRef := func(depID descpb.ID) error {
		__antithesis_instrumentation__.Notify(566055)
		if allowCrossDatabaseViews.Get(&p.execCfg.Settings.SV) {
			__antithesis_instrumentation__.Notify(566059)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(566060)
		}
		__antithesis_instrumentation__.Notify(566056)
		dependentObject, err := p.Descriptors().GetImmutableTypeByID(ctx, p.txn, depID,
			tree.ObjectLookupFlags{
				CommonLookupFlags: tree.CommonLookupFlags{
					Required:    true,
					AvoidLeased: true,
				}})
		if err != nil {
			__antithesis_instrumentation__.Notify(566061)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566062)
		}
		__antithesis_instrumentation__.Notify(566057)

		if dependentObject.GetParentID() == targetDbDesc.GetID() {
			__antithesis_instrumentation__.Notify(566063)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(566064)
		}
		__antithesis_instrumentation__.Notify(566058)
		return errors.WithHintf(
			pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"this view will reference a type %q in another databases after rename "+
					"(see the '%s' cluster setting)",
				dependentObject.GetName(),
				allowCrossDatabaseViewsSetting),
			crossDBReferenceDeprecationHint(),
		)
	}
	__antithesis_instrumentation__.Notify(566007)

	if tableDesc.IsTable() {
		__antithesis_instrumentation__.Notify(566065)
		err := tableDesc.ForeachOutboundFK(func(fk *descpb.ForeignKeyConstraint) error {
			__antithesis_instrumentation__.Notify(566071)
			return checkFkForCrossDbDep(fk, true)
		})
		__antithesis_instrumentation__.Notify(566066)
		if err != nil {
			__antithesis_instrumentation__.Notify(566072)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566073)
		}
		__antithesis_instrumentation__.Notify(566067)

		err = tableDesc.ForeachInboundFK(func(fk *descpb.ForeignKeyConstraint) error {
			__antithesis_instrumentation__.Notify(566074)
			return checkFkForCrossDbDep(fk, false)
		})
		__antithesis_instrumentation__.Notify(566068)
		if err != nil {
			__antithesis_instrumentation__.Notify(566075)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566076)
		}
		__antithesis_instrumentation__.Notify(566069)

		for _, columnDesc := range tableDesc.Columns {
			__antithesis_instrumentation__.Notify(566077)
			for _, ownsSequenceID := range columnDesc.OwnsSequenceIds {
				__antithesis_instrumentation__.Notify(566079)
				if err := checkDepForCrossDbRef(ownsSequenceID, owner); err != nil {
					__antithesis_instrumentation__.Notify(566080)
					return err
				} else {
					__antithesis_instrumentation__.Notify(566081)
				}
			}
			__antithesis_instrumentation__.Notify(566078)
			for _, seqID := range columnDesc.UsesSequenceIds {
				__antithesis_instrumentation__.Notify(566082)
				if err := checkDepForCrossDbRef(seqID, reference); err != nil {
					__antithesis_instrumentation__.Notify(566083)
					return err
				} else {
					__antithesis_instrumentation__.Notify(566084)
				}
			}
		}
		__antithesis_instrumentation__.Notify(566070)

		if !allowCrossDatabaseViews.Get(&p.execCfg.Settings.SV) {
			__antithesis_instrumentation__.Notify(566085)
			err := tableDesc.ForeachDependedOnBy(func(dep *descpb.TableDescriptor_Reference) error {
				__antithesis_instrumentation__.Notify(566087)
				return checkDepForCrossDbRef(dep.ID, reference)
			})
			__antithesis_instrumentation__.Notify(566086)
			if err != nil {
				__antithesis_instrumentation__.Notify(566088)
				return err
			} else {
				__antithesis_instrumentation__.Notify(566089)
			}
		} else {
			__antithesis_instrumentation__.Notify(566090)
		}
	} else {
		__antithesis_instrumentation__.Notify(566091)
		if tableDesc.IsView() {
			__antithesis_instrumentation__.Notify(566092)

			dependsOn := tableDesc.GetDependsOn()
			for _, dependency := range dependsOn {
				__antithesis_instrumentation__.Notify(566094)
				if err := checkDepForCrossDbRef(dependency, reference); err != nil {
					__antithesis_instrumentation__.Notify(566095)
					return err
				} else {
					__antithesis_instrumentation__.Notify(566096)
				}
			}
			__antithesis_instrumentation__.Notify(566093)

			dependsOnTypes := tableDesc.GetDependsOnTypes()
			for _, dependency := range dependsOnTypes {
				__antithesis_instrumentation__.Notify(566097)
				if err := checkTypeDepForCrossDbRef(dependency); err != nil {
					__antithesis_instrumentation__.Notify(566098)
					return err
				} else {
					__antithesis_instrumentation__.Notify(566099)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(566100)
			if tableDesc.IsSequence() {
				__antithesis_instrumentation__.Notify(566101)

				sequenceOpts := tableDesc.GetSequenceOpts()
				if sequenceOpts.SequenceOwner.OwnerTableID != descpb.InvalidID {
					__antithesis_instrumentation__.Notify(566103)
					err := checkDepForCrossDbRef(sequenceOpts.SequenceOwner.OwnerTableID, owner)
					if err != nil {
						__antithesis_instrumentation__.Notify(566104)
						return err
					} else {
						__antithesis_instrumentation__.Notify(566105)
					}
				} else {
					__antithesis_instrumentation__.Notify(566106)
				}
				__antithesis_instrumentation__.Notify(566102)

				for _, sequenceReferences := range tableDesc.GetDependedOnBy() {
					__antithesis_instrumentation__.Notify(566107)
					err := checkDepForCrossDbRef(sequenceReferences.ID, reference)
					if err != nil {
						__antithesis_instrumentation__.Notify(566108)
						return err
					} else {
						__antithesis_instrumentation__.Notify(566109)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(566110)
			}
		}
	}
	__antithesis_instrumentation__.Notify(566008)
	return nil
}
