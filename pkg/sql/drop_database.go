package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type dropDatabaseNode struct {
	n      *tree.DropDatabase
	dbDesc *dbdesc.Mutable
	d      *dropCascadeState
}

func (p *planner) DropDatabase(ctx context.Context, n *tree.DropDatabase) (planNode, error) {
	__antithesis_instrumentation__.Notify(468779)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"DROP DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(468790)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468791)
	}
	__antithesis_instrumentation__.Notify(468780)

	if n.Name == "" {
		__antithesis_instrumentation__.Notify(468792)
		return nil, errEmptyDatabaseName
	} else {
		__antithesis_instrumentation__.Notify(468793)
	}
	__antithesis_instrumentation__.Notify(468781)

	if string(n.Name) == p.SessionData().Database && func() bool {
		__antithesis_instrumentation__.Notify(468794)
		return p.SessionData().SafeUpdates == true
	}() == true {
		__antithesis_instrumentation__.Notify(468795)
		return nil, pgerror.DangerousStatementf("DROP DATABASE on current database")
	} else {
		__antithesis_instrumentation__.Notify(468796)
	}
	__antithesis_instrumentation__.Notify(468782)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Name),
		tree.DatabaseLookupFlags{Required: !n.IfExists})
	if err != nil {
		__antithesis_instrumentation__.Notify(468797)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468798)
	}
	__antithesis_instrumentation__.Notify(468783)
	if dbDesc == nil {
		__antithesis_instrumentation__.Notify(468799)

		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(468800)
	}
	__antithesis_instrumentation__.Notify(468784)

	if err := p.CheckPrivilege(ctx, dbDesc, privilege.DROP); err != nil {
		__antithesis_instrumentation__.Notify(468801)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468802)
	}
	__antithesis_instrumentation__.Notify(468785)

	schemas, err := p.Descriptors().GetSchemasForDatabase(ctx, p.txn, dbDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(468803)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468804)
	}
	__antithesis_instrumentation__.Notify(468786)

	d := newDropCascadeState()

	for _, schema := range schemas {
		__antithesis_instrumentation__.Notify(468805)
		res, err := p.Descriptors().GetSchemaByName(
			ctx, p.txn, dbDesc, schema, tree.SchemaLookupFlags{
				Required:       true,
				RequireMutable: true,
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(468807)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(468808)
		}
		__antithesis_instrumentation__.Notify(468806)
		if err := d.collectObjectsInSchema(ctx, p, dbDesc, res); err != nil {
			__antithesis_instrumentation__.Notify(468809)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(468810)
		}
	}
	__antithesis_instrumentation__.Notify(468787)

	if len(d.objectNamesToDelete) > 0 {
		__antithesis_instrumentation__.Notify(468811)
		switch n.DropBehavior {
		case tree.DropRestrict:
			__antithesis_instrumentation__.Notify(468812)
			return nil, pgerror.Newf(pgcode.DependentObjectsStillExist,
				"database %q is not empty and RESTRICT was specified",
				tree.ErrNameString(dbDesc.GetName()))
		case tree.DropDefault:
			__antithesis_instrumentation__.Notify(468813)

			if p.SessionData().SafeUpdates {
				__antithesis_instrumentation__.Notify(468815)
				return nil, pgerror.DangerousStatementf(
					"DROP DATABASE on non-empty database without explicit CASCADE")
			} else {
				__antithesis_instrumentation__.Notify(468816)
			}
		default:
			__antithesis_instrumentation__.Notify(468814)
		}
	} else {
		__antithesis_instrumentation__.Notify(468817)
	}
	__antithesis_instrumentation__.Notify(468788)

	if err := d.resolveCollectedObjects(ctx, p, dbDesc); err != nil {
		__antithesis_instrumentation__.Notify(468818)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468819)
	}
	__antithesis_instrumentation__.Notify(468789)

	return &dropDatabaseNode{
		n:      n,
		dbDesc: dbDesc,
		d:      d,
	}, nil
}

func (n *dropDatabaseNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(468820)
	telemetry.Inc(sqltelemetry.SchemaChangeDropCounter("database"))

	ctx := params.ctx
	p := params.p

	if err := n.d.dropAllCollectedObjects(ctx, p); err != nil {
		__antithesis_instrumentation__.Notify(468828)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468829)
	}
	__antithesis_instrumentation__.Notify(468821)

	var schemasIDsToDelete []descpb.ID
	for _, schemaWithDbDesc := range n.d.schemasToDelete {
		__antithesis_instrumentation__.Notify(468830)
		schemaToDelete := schemaWithDbDesc.schema
		switch schemaToDelete.SchemaKind() {
		case catalog.SchemaTemporary, catalog.SchemaPublic:
			__antithesis_instrumentation__.Notify(468831)

			key := catalogkeys.MakeSchemaNameKey(p.ExecCfg().Codec, n.dbDesc.GetID(), schemaToDelete.GetName())
			if err := p.txn.Del(ctx, key); err != nil {
				__antithesis_instrumentation__.Notify(468836)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468837)
			}
		case catalog.SchemaUserDefined:
			__antithesis_instrumentation__.Notify(468832)

			mutDesc, ok := schemaToDelete.(*schemadesc.Mutable)
			if !ok {
				__antithesis_instrumentation__.Notify(468838)
				return errors.AssertionFailedf("expected Mutable, found %T", schemaToDelete)
			} else {
				__antithesis_instrumentation__.Notify(468839)
			}
			__antithesis_instrumentation__.Notify(468833)
			if err := params.p.dropSchemaImpl(ctx, n.dbDesc, mutDesc); err != nil {
				__antithesis_instrumentation__.Notify(468840)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468841)
			}
			__antithesis_instrumentation__.Notify(468834)
			schemasIDsToDelete = append(schemasIDsToDelete, schemaToDelete.GetID())
		default:
			__antithesis_instrumentation__.Notify(468835)
		}
	}
	__antithesis_instrumentation__.Notify(468822)

	if err := p.createDropDatabaseJob(
		ctx,
		n.dbDesc.GetID(),
		schemasIDsToDelete,
		n.d.getDroppedTableDetails(),
		n.d.typesToDelete,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(468842)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468843)
	}
	__antithesis_instrumentation__.Notify(468823)

	n.dbDesc.SetDropped()
	b := p.txn.NewBatch()
	p.dropNamespaceEntry(ctx, b, n.dbDesc)

	if err := p.writeDatabaseChangeToBatch(ctx, n.dbDesc, b); err != nil {
		__antithesis_instrumentation__.Notify(468844)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468845)
	}
	__antithesis_instrumentation__.Notify(468824)
	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(468846)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468847)
	}
	__antithesis_instrumentation__.Notify(468825)

	metadataUpdater := p.ExecCfg().DescMetadaUpdaterFactory.NewMetadataUpdater(
		ctx,
		p.txn,
		p.SessionData())
	err := metadataUpdater.DeleteDescriptorComment(
		int64(n.dbDesc.GetID()),
		0,
		keys.DatabaseCommentType)
	if err != nil {
		__antithesis_instrumentation__.Notify(468848)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468849)
	}
	__antithesis_instrumentation__.Notify(468826)

	err = metadataUpdater.DeleteDatabaseRoleSettings(ctx, n.dbDesc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(468850)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468851)
	}
	__antithesis_instrumentation__.Notify(468827)

	return p.logEvent(ctx,
		n.dbDesc.GetID(),
		&eventpb.DropDatabase{
			DatabaseName:         n.n.Name.String(),
			DroppedSchemaObjects: n.d.droppedNames,
		})
}

func (*dropDatabaseNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(468852)
	return false, nil
}
func (*dropDatabaseNode) Close(context.Context) { __antithesis_instrumentation__.Notify(468853) }
func (*dropDatabaseNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(468854)
	return tree.Datums{}
}

func filterImplicitlyDeletedObjects(
	tables []toDelete, implicitDeleteObjects map[descpb.ID]*tabledesc.Mutable,
) []toDelete {
	__antithesis_instrumentation__.Notify(468855)
	filteredDeleteList := make([]toDelete, 0, len(tables))
	for _, toDel := range tables {
		__antithesis_instrumentation__.Notify(468857)
		if _, found := implicitDeleteObjects[toDel.desc.ID]; !found {
			__antithesis_instrumentation__.Notify(468858)
			filteredDeleteList = append(filteredDeleteList, toDel)
		} else {
			__antithesis_instrumentation__.Notify(468859)
		}
	}
	__antithesis_instrumentation__.Notify(468856)
	return filteredDeleteList
}

func (p *planner) accumulateAllObjectsToDelete(
	ctx context.Context, objects []toDelete,
) ([]*tabledesc.Mutable, map[descpb.ID]*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(468860)
	implicitDeleteObjects := make(map[descpb.ID]*tabledesc.Mutable)
	for _, toDel := range objects {
		__antithesis_instrumentation__.Notify(468864)
		err := p.accumulateCascadingViews(ctx, implicitDeleteObjects, toDel.desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(468866)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(468867)
		}
		__antithesis_instrumentation__.Notify(468865)

		if toDel.desc.IsTable() {
			__antithesis_instrumentation__.Notify(468868)
			err := p.accumulateOwnedSequences(ctx, implicitDeleteObjects, toDel.desc)
			if err != nil {
				__antithesis_instrumentation__.Notify(468869)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(468870)
			}
		} else {
			__antithesis_instrumentation__.Notify(468871)
		}
	}
	__antithesis_instrumentation__.Notify(468861)
	allObjectsToDelete := make([]*tabledesc.Mutable, 0,
		len(objects)+len(implicitDeleteObjects))
	for _, desc := range implicitDeleteObjects {
		__antithesis_instrumentation__.Notify(468872)
		allObjectsToDelete = append(allObjectsToDelete, desc)
	}
	__antithesis_instrumentation__.Notify(468862)
	for _, toDel := range objects {
		__antithesis_instrumentation__.Notify(468873)
		if _, found := implicitDeleteObjects[toDel.desc.ID]; !found {
			__antithesis_instrumentation__.Notify(468874)
			allObjectsToDelete = append(allObjectsToDelete, toDel.desc)
		} else {
			__antithesis_instrumentation__.Notify(468875)
		}
	}
	__antithesis_instrumentation__.Notify(468863)
	return allObjectsToDelete, implicitDeleteObjects, nil
}

func (p *planner) accumulateOwnedSequences(
	ctx context.Context, dependentObjects map[descpb.ID]*tabledesc.Mutable, desc *tabledesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(468876)
	for colID := range desc.GetColumns() {
		__antithesis_instrumentation__.Notify(468878)
		for _, seqID := range desc.GetColumns()[colID].OwnsSequenceIds {
			__antithesis_instrumentation__.Notify(468879)
			ownedSeqDesc, err := p.Descriptors().GetMutableTableVersionByID(ctx, seqID, p.txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(468881)

				if errors.Is(err, catalog.ErrDescriptorDropped) || func() bool {
					__antithesis_instrumentation__.Notify(468883)
					return pgerror.GetPGCode(err) == pgcode.UndefinedTable == true
				}() == true {
					__antithesis_instrumentation__.Notify(468884)
					log.Infof(ctx,
						"swallowing error for owned sequence that was not found %s", err.Error())
					continue
				} else {
					__antithesis_instrumentation__.Notify(468885)
				}
				__antithesis_instrumentation__.Notify(468882)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468886)
			}
			__antithesis_instrumentation__.Notify(468880)
			dependentObjects[seqID] = ownedSeqDesc
		}
	}
	__antithesis_instrumentation__.Notify(468877)
	return nil
}

func (p *planner) accumulateCascadingViews(
	ctx context.Context, dependentObjects map[descpb.ID]*tabledesc.Mutable, desc *tabledesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(468887)
	for _, ref := range desc.DependedOnBy {
		__antithesis_instrumentation__.Notify(468889)
		dependentDesc, err := p.Descriptors().GetMutableTableVersionByID(ctx, ref.ID, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(468892)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468893)
		}
		__antithesis_instrumentation__.Notify(468890)
		if !dependentDesc.IsView() {
			__antithesis_instrumentation__.Notify(468894)
			continue
		} else {
			__antithesis_instrumentation__.Notify(468895)
		}
		__antithesis_instrumentation__.Notify(468891)
		dependentObjects[ref.ID] = dependentDesc
		if err := p.accumulateCascadingViews(ctx, dependentObjects, dependentDesc); err != nil {
			__antithesis_instrumentation__.Notify(468896)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468897)
		}
	}
	__antithesis_instrumentation__.Notify(468888)
	return nil
}
