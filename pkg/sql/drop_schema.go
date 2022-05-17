package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type dropSchemaNode struct {
	n *tree.DropSchema
	d *dropCascadeState
}

var _ planNode = &dropSchemaNode{n: nil}

func (p *planner) DropSchema(ctx context.Context, n *tree.DropSchema) (planNode, error) {
	__antithesis_instrumentation__.Notify(469336)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"DROP SCHEMA",
	); err != nil {
		__antithesis_instrumentation__.Notify(469341)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469342)
	}
	__antithesis_instrumentation__.Notify(469337)

	isAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(469343)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469344)
	}
	__antithesis_instrumentation__.Notify(469338)

	d := newDropCascadeState()

	for _, schema := range n.Names {
		__antithesis_instrumentation__.Notify(469345)
		dbName := p.CurrentDatabase()
		if schema.ExplicitCatalog {
			__antithesis_instrumentation__.Notify(469351)
			dbName = schema.Catalog()
		} else {
			__antithesis_instrumentation__.Notify(469352)
		}
		__antithesis_instrumentation__.Notify(469346)
		scName := schema.Schema()

		db, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, dbName,
			tree.DatabaseLookupFlags{Required: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(469353)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469354)
		}
		__antithesis_instrumentation__.Notify(469347)

		sc, err := p.Descriptors().GetSchemaByName(
			ctx, p.txn, db, scName, tree.SchemaLookupFlags{
				Required:       false,
				RequireMutable: true,
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469355)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469356)
		}
		__antithesis_instrumentation__.Notify(469348)
		if sc == nil {
			__antithesis_instrumentation__.Notify(469357)
			if n.IfExists {
				__antithesis_instrumentation__.Notify(469359)
				continue
			} else {
				__antithesis_instrumentation__.Notify(469360)
			}
			__antithesis_instrumentation__.Notify(469358)
			return nil, pgerror.Newf(pgcode.InvalidSchemaName, "unknown schema %q", scName)
		} else {
			__antithesis_instrumentation__.Notify(469361)
		}
		__antithesis_instrumentation__.Notify(469349)

		if scName == tree.PublicSchema {
			__antithesis_instrumentation__.Notify(469362)
			return nil, pgerror.Newf(pgcode.InvalidSchemaName, "cannot drop schema %q", scName)
		} else {
			__antithesis_instrumentation__.Notify(469363)
		}
		__antithesis_instrumentation__.Notify(469350)

		switch sc.SchemaKind() {
		case catalog.SchemaPublic, catalog.SchemaVirtual, catalog.SchemaTemporary:
			__antithesis_instrumentation__.Notify(469364)
			return nil, pgerror.Newf(pgcode.InvalidSchemaName, "cannot drop schema %q", scName)
		case catalog.SchemaUserDefined:
			__antithesis_instrumentation__.Notify(469365)
			hasOwnership, err := p.HasOwnership(ctx, sc)
			if err != nil {
				__antithesis_instrumentation__.Notify(469371)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(469372)
			}
			__antithesis_instrumentation__.Notify(469366)
			if !(isAdmin || func() bool {
				__antithesis_instrumentation__.Notify(469373)
				return hasOwnership == true
			}() == true) {
				__antithesis_instrumentation__.Notify(469374)
				return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
					"must be owner of schema %s", tree.Name(sc.GetName()))
			} else {
				__antithesis_instrumentation__.Notify(469375)
			}
			__antithesis_instrumentation__.Notify(469367)
			namesBefore := len(d.objectNamesToDelete)
			if err := d.collectObjectsInSchema(ctx, p, db, sc); err != nil {
				__antithesis_instrumentation__.Notify(469376)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(469377)
			}
			__antithesis_instrumentation__.Notify(469368)

			if namesBefore != len(d.objectNamesToDelete) && func() bool {
				__antithesis_instrumentation__.Notify(469378)
				return n.DropBehavior != tree.DropCascade == true
			}() == true {
				__antithesis_instrumentation__.Notify(469379)
				return nil, pgerror.Newf(pgcode.DependentObjectsStillExist,
					"schema %q is not empty and CASCADE was not specified", scName)
			} else {
				__antithesis_instrumentation__.Notify(469380)
			}
			__antithesis_instrumentation__.Notify(469369)
			sqltelemetry.IncrementUserDefinedSchemaCounter(sqltelemetry.UserDefinedSchemaDrop)
		default:
			__antithesis_instrumentation__.Notify(469370)
			return nil, errors.AssertionFailedf("unknown schema kind %d", sc.SchemaKind())
		}

	}
	__antithesis_instrumentation__.Notify(469339)

	if err := d.resolveCollectedObjects(ctx, p, nil); err != nil {
		__antithesis_instrumentation__.Notify(469381)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469382)
	}
	__antithesis_instrumentation__.Notify(469340)

	return &dropSchemaNode{n: n, d: d}, nil
}

func (n *dropSchemaNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(469383)
	telemetry.Inc(sqltelemetry.SchemaChangeDropCounter("schema"))

	ctx := params.ctx
	p := params.p

	if err := n.d.dropAllCollectedObjects(ctx, p); err != nil {
		__antithesis_instrumentation__.Notify(469389)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469390)
	}
	__antithesis_instrumentation__.Notify(469384)

	schemaIDs := make([]descpb.ID, len(n.d.schemasToDelete))
	for i := range n.d.schemasToDelete {
		__antithesis_instrumentation__.Notify(469391)
		sc := n.d.schemasToDelete[i].schema
		schemaIDs[i] = sc.GetID()
		db := n.d.schemasToDelete[i].dbDesc

		mutDesc := sc.(*schemadesc.Mutable)
		if err := p.dropSchemaImpl(ctx, db, mutDesc); err != nil {
			__antithesis_instrumentation__.Notify(469392)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469393)
		}
	}
	__antithesis_instrumentation__.Notify(469385)

	for i := range n.d.schemasToDelete {
		__antithesis_instrumentation__.Notify(469394)
		sc := n.d.schemasToDelete[i].schema
		db := n.d.schemasToDelete[i].dbDesc
		if err := p.writeNonDropDatabaseChange(
			ctx, db,
			fmt.Sprintf("updating parent database %s for %s", db.GetName(), sc.GetName()),
		); err != nil {
			__antithesis_instrumentation__.Notify(469395)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469396)
		}
	}
	__antithesis_instrumentation__.Notify(469386)

	if err := p.createDropSchemaJob(
		schemaIDs,
		n.d.getDroppedTableDetails(),
		n.d.typesToDelete,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(469397)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469398)
	}
	__antithesis_instrumentation__.Notify(469387)

	for _, schemaToDelete := range n.d.schemasToDelete {
		__antithesis_instrumentation__.Notify(469399)
		sc := schemaToDelete.schema
		qualifiedSchemaName, err := p.getQualifiedSchemaName(params.ctx, sc)
		if err != nil {
			__antithesis_instrumentation__.Notify(469401)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469402)
		}
		__antithesis_instrumentation__.Notify(469400)

		if err := params.p.logEvent(params.ctx,
			sc.GetID(),
			&eventpb.DropSchema{
				SchemaName: qualifiedSchemaName.String(),
			}); err != nil {
			__antithesis_instrumentation__.Notify(469403)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469404)
		}
	}
	__antithesis_instrumentation__.Notify(469388)
	return nil
}

func (p *planner) dropSchemaImpl(
	ctx context.Context, parentDB *dbdesc.Mutable, sc *schemadesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(469405)

	if p.execCfg.Settings.Version.IsActive(ctx, clusterversion.AvoidDrainingNames) {
		__antithesis_instrumentation__.Notify(469409)
		delete(parentDB.Schemas, sc.GetName())
	} else {
		__antithesis_instrumentation__.Notify(469410)

		parentDB.AddSchemaToDatabase(sc.GetName(), descpb.DatabaseDescriptor_SchemaInfo{
			ID:      sc.GetID(),
			Dropped: true,
		})
	}
	__antithesis_instrumentation__.Notify(469406)

	sc.SetDropped()

	b := p.txn.NewBatch()
	p.dropNamespaceEntry(ctx, b, sc)

	if err := p.removeSchemaComment(ctx, sc.GetID()); err != nil {
		__antithesis_instrumentation__.Notify(469411)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469412)
	}
	__antithesis_instrumentation__.Notify(469407)

	if err := p.writeSchemaDesc(ctx, sc); err != nil {
		__antithesis_instrumentation__.Notify(469413)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469414)
	}
	__antithesis_instrumentation__.Notify(469408)

	return p.txn.Run(ctx, b)
}

func (p *planner) createDropSchemaJob(
	schemas []descpb.ID,
	tableDropDetails []jobspb.DroppedTableDetails,
	typesToDrop []*typedesc.Mutable,
	jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(469415)
	typeIDs := make([]descpb.ID, 0, len(typesToDrop))
	for _, t := range typesToDrop {
		__antithesis_instrumentation__.Notify(469417)
		typeIDs = append(typeIDs, t.ID)
	}
	__antithesis_instrumentation__.Notify(469416)

	_, err := p.extendedEvalCtx.QueueJob(p.EvalContext().Ctx(), jobs.Record{
		Description:   jobDesc,
		Username:      p.User(),
		DescriptorIDs: schemas,
		Details: jobspb.SchemaChangeDetails{
			DroppedSchemas:    schemas,
			DroppedTables:     tableDropDetails,
			DroppedTypes:      typeIDs,
			DroppedDatabaseID: descpb.InvalidID,

			FormatVersion: jobspb.DatabaseJobFormatVersion,
		},
		Progress:      jobspb.SchemaChangeProgress{},
		NonCancelable: true,
	})
	return err
}

func (p *planner) removeSchemaComment(ctx context.Context, schemaID descpb.ID) error {
	__antithesis_instrumentation__.Notify(469418)
	_, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.ExecEx(
		ctx,
		"delete-schema-comment",
		p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"DELETE FROM system.comments WHERE type=$1 AND object_id=$2 AND sub_id=0",
		keys.SchemaCommentType,
		schemaID)

	return err
}

func (n *dropSchemaNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(469419)
	return false, nil
}
func (n *dropSchemaNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(469420)
	return tree.Datums{}
}
func (n *dropSchemaNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(469421) }
func (n *dropSchemaNode) ReadingOwnWrites()         { __antithesis_instrumentation__.Notify(469422) }
