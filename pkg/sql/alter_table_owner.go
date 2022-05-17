package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
)

type alterTableOwnerNode struct {
	owner  security.SQLUsername
	desc   *tabledesc.Mutable
	n      *tree.AlterTableOwner
	prefix catalog.ResolvedObjectPrefix
}

func (p *planner) AlterTableOwner(ctx context.Context, n *tree.AlterTableOwner) (planNode, error) {
	__antithesis_instrumentation__.Notify(244953)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER TABLE/VIEW/SEQUENCE OWNER",
	); err != nil {
		__antithesis_instrumentation__.Notify(244960)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244961)
	}
	__antithesis_instrumentation__.Notify(244954)

	tn := n.Name.ToTableName()

	requiredTableKind := tree.ResolveAnyTableKind
	if n.IsView {
		__antithesis_instrumentation__.Notify(244962)
		requiredTableKind = tree.ResolveRequireViewDesc
	} else {
		__antithesis_instrumentation__.Notify(244963)
		if n.IsSequence {
			__antithesis_instrumentation__.Notify(244964)
			requiredTableKind = tree.ResolveRequireSequenceDesc
		} else {
			__antithesis_instrumentation__.Notify(244965)
		}
	}
	__antithesis_instrumentation__.Notify(244955)
	prefix, tableDesc, err := p.ResolveMutableTableDescriptor(
		ctx, &tn, !n.IfExists, requiredTableKind)
	if err != nil {
		__antithesis_instrumentation__.Notify(244966)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244967)
	}
	__antithesis_instrumentation__.Notify(244956)
	if tableDesc == nil {
		__antithesis_instrumentation__.Notify(244968)

		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(244969)
	}
	__antithesis_instrumentation__.Notify(244957)

	if err := checkViewMatchesMaterialized(tableDesc, n.IsView, n.IsMaterialized); err != nil {
		__antithesis_instrumentation__.Notify(244970)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244971)
	}
	__antithesis_instrumentation__.Notify(244958)

	owner, err := n.Owner.ToSQLUsername(p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(244972)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244973)
	}
	__antithesis_instrumentation__.Notify(244959)
	return &alterTableOwnerNode{
		owner:  owner,
		desc:   tableDesc,
		n:      n,
		prefix: prefix,
	}, nil
}

func (n *alterTableOwnerNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(244974)
	telemetry.Inc(n.n.TelemetryCounter())
	ctx := params.ctx
	p := params.p
	tableDesc := n.desc
	newOwner := n.owner
	oldOwner := n.desc.GetPrivileges().Owner()

	if err := p.checkCanAlterToNewOwner(ctx, tableDesc, newOwner); err != nil {
		__antithesis_instrumentation__.Notify(244980)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244981)
	}
	__antithesis_instrumentation__.Notify(244975)

	if err := p.canCreateOnSchema(
		ctx, tableDesc.GetParentSchemaID(), tableDesc.ParentID, newOwner, checkPublicSchema); err != nil {
		__antithesis_instrumentation__.Notify(244982)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244983)
	}
	__antithesis_instrumentation__.Notify(244976)

	tbNameWithSchema := tree.MakeTableNameWithSchema(
		tree.Name(n.prefix.Database.GetName()),
		tree.Name(n.prefix.Schema.GetName()),
		tree.Name(tableDesc.GetName()),
	)

	if err := p.setNewTableOwner(ctx, tableDesc, tbNameWithSchema, newOwner); err != nil {
		__antithesis_instrumentation__.Notify(244984)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244985)
	}
	__antithesis_instrumentation__.Notify(244977)

	if newOwner == oldOwner {
		__antithesis_instrumentation__.Notify(244986)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(244987)
	}
	__antithesis_instrumentation__.Notify(244978)

	if err := p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(244988)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244989)
	}
	__antithesis_instrumentation__.Notify(244979)

	return nil
}

func (p *planner) setNewTableOwner(
	ctx context.Context,
	desc *tabledesc.Mutable,
	tbNameWithSchema tree.TableName,
	newOwner security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(244990)
	privs := desc.GetPrivileges()
	privs.SetOwner(newOwner)

	return p.logEvent(ctx,
		desc.ID,
		&eventpb.AlterTableOwner{
			TableName: tbNameWithSchema.FQString(),
			Owner:     newOwner.Normalized(),
		})
}

func (n *alterTableOwnerNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(244991) }

func (n *alterTableOwnerNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(244992)
	return false, nil
}
func (n *alterTableOwnerNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(244993)
	return tree.Datums{}
}
func (n *alterTableOwnerNode) Close(context.Context) { __antithesis_instrumentation__.Notify(244994) }
