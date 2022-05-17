package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
)

type alterTableSetSchemaNode struct {
	newSchema string
	prefix    catalog.ResolvedObjectPrefix
	tableDesc *tabledesc.Mutable
	n         *tree.AlterTableSetSchema
}

func (p *planner) AlterTableSetSchema(
	ctx context.Context, n *tree.AlterTableSetSchema,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(244995)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER TABLE/VIEW/SEQUENCE SET SCHEMA",
	); err != nil {
		__antithesis_instrumentation__.Notify(245004)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245005)
	}
	__antithesis_instrumentation__.Notify(244996)

	tn := n.Name.ToTableName()
	requiredTableKind := tree.ResolveAnyTableKind
	if n.IsView {
		__antithesis_instrumentation__.Notify(245006)
		requiredTableKind = tree.ResolveRequireViewDesc
	} else {
		__antithesis_instrumentation__.Notify(245007)
		if n.IsSequence {
			__antithesis_instrumentation__.Notify(245008)
			requiredTableKind = tree.ResolveRequireSequenceDesc
		} else {
			__antithesis_instrumentation__.Notify(245009)
		}
	}
	__antithesis_instrumentation__.Notify(244997)
	prefix, tableDesc, err := p.ResolveMutableTableDescriptor(
		ctx, &tn, !n.IfExists, requiredTableKind)
	if err != nil {
		__antithesis_instrumentation__.Notify(245010)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245011)
	}
	__antithesis_instrumentation__.Notify(244998)
	if tableDesc == nil {
		__antithesis_instrumentation__.Notify(245012)

		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(245013)
	}
	__antithesis_instrumentation__.Notify(244999)

	if err := checkViewMatchesMaterialized(tableDesc, n.IsView, n.IsMaterialized); err != nil {
		__antithesis_instrumentation__.Notify(245014)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245015)
	}
	__antithesis_instrumentation__.Notify(245000)

	if tableDesc.Temporary {
		__antithesis_instrumentation__.Notify(245016)
		return nil, pgerror.Newf(pgcode.FeatureNotSupported,
			"cannot move objects into or out of temporary schemas")
	} else {
		__antithesis_instrumentation__.Notify(245017)
	}
	__antithesis_instrumentation__.Notify(245001)

	err = p.CheckPrivilege(ctx, tableDesc, privilege.DROP)
	if err != nil {
		__antithesis_instrumentation__.Notify(245018)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245019)
	}
	__antithesis_instrumentation__.Notify(245002)

	for _, dependent := range tableDesc.DependedOnBy {
		__antithesis_instrumentation__.Notify(245020)
		if !dependent.ByID {
			__antithesis_instrumentation__.Notify(245021)
			return nil, p.dependentViewError(
				ctx, string(tableDesc.DescriptorType()), tableDesc.Name,
				tableDesc.ParentID, dependent.ID, "set schema on",
			)
		} else {
			__antithesis_instrumentation__.Notify(245022)
		}
	}
	__antithesis_instrumentation__.Notify(245003)

	return &alterTableSetSchemaNode{
		newSchema: string(n.Schema),
		prefix:    prefix,
		tableDesc: tableDesc,
		n:         n,
	}, nil
}

func (n *alterTableSetSchemaNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(245023)
	telemetry.Inc(n.n.TelemetryCounter())
	ctx := params.ctx
	p := params.p
	tableDesc := n.tableDesc
	oldNameKey := descpb.NameInfo{
		ParentID:       tableDesc.GetParentID(),
		ParentSchemaID: tableDesc.GetParentSchemaID(),
		Name:           tableDesc.GetName(),
	}

	kind := tree.GetTableType(tableDesc.IsSequence(), tableDesc.IsView(), tableDesc.GetIsMaterializedView())
	oldName := tree.MakeTableNameFromPrefix(n.prefix.NamePrefix(), tree.Name(n.tableDesc.GetName()))

	desiredSchemaID, err := p.prepareSetSchema(ctx, n.prefix.Database, tableDesc, n.newSchema)
	if err != nil {
		__antithesis_instrumentation__.Notify(245030)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245031)
	}
	__antithesis_instrumentation__.Notify(245024)

	if desiredSchemaID == oldNameKey.GetParentSchemaID() {
		__antithesis_instrumentation__.Notify(245032)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245033)
	}
	__antithesis_instrumentation__.Notify(245025)

	objectID, err := p.Descriptors().Direct().LookupObjectID(
		ctx, p.txn, tableDesc.GetParentID(), desiredSchemaID, tableDesc.GetName(),
	)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(245034)
		return objectID != descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(245035)
		return pgerror.Newf(pgcode.DuplicateRelation,
			"relation %s already exists in schema %s", tableDesc.GetName(), n.newSchema)
	} else {
		__antithesis_instrumentation__.Notify(245036)
		if err != nil {
			__antithesis_instrumentation__.Notify(245037)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245038)
		}
	}
	__antithesis_instrumentation__.Notify(245026)

	tableDesc.SetParentSchemaID(desiredSchemaID)

	b := p.txn.NewBatch()
	p.renameNamespaceEntry(ctx, b, oldNameKey, tableDesc)

	if err := p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(245039)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245040)
	}
	__antithesis_instrumentation__.Notify(245027)

	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(245041)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245042)
	}
	__antithesis_instrumentation__.Notify(245028)

	newName, err := p.getQualifiedTableName(ctx, tableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(245043)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245044)
	}
	__antithesis_instrumentation__.Notify(245029)

	return p.logEvent(ctx,
		desiredSchemaID,
		&eventpb.SetSchema{
			CommonEventDetails:    eventpb.CommonEventDetails{},
			CommonSQLEventDetails: eventpb.CommonSQLEventDetails{},
			DescriptorName:        oldName.FQString(),
			NewDescriptorName:     newName.FQString(),
			DescriptorType:        kind,
		},
	)
}

func (n *alterTableSetSchemaNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(245045) }

func (n *alterTableSetSchemaNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(245046)
	return false, nil
}
func (n *alterTableSetSchemaNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(245047)
	return tree.Datums{}
}
func (n *alterTableSetSchemaNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(245048)
}
