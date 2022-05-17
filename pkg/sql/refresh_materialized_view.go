package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type refreshMaterializedViewNode struct {
	n    *tree.RefreshMaterializedView
	desc *tabledesc.Mutable
}

func (p *planner) RefreshMaterializedView(
	ctx context.Context, n *tree.RefreshMaterializedView,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(564986)
	if !p.extendedEvalCtx.TxnIsSingleStmt {
		__antithesis_instrumentation__.Notify(564994)
		return nil, pgerror.Newf(pgcode.InvalidTransactionState, "cannot refresh view in a multi-statement transaction")
	} else {
		__antithesis_instrumentation__.Notify(564995)
	}
	__antithesis_instrumentation__.Notify(564987)
	_, desc, err := p.ResolveMutableTableDescriptorEx(ctx, n.Name, true, tree.ResolveRequireViewDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(564996)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(564997)
	}
	__antithesis_instrumentation__.Notify(564988)
	if !desc.MaterializedView() {
		__antithesis_instrumentation__.Notify(564998)
		return nil, pgerror.Newf(pgcode.WrongObjectType, "%q is not a materialized view", desc.Name)
	} else {
		__antithesis_instrumentation__.Notify(564999)
	}
	__antithesis_instrumentation__.Notify(564989)

	for i := range desc.Mutations {
		__antithesis_instrumentation__.Notify(565000)
		mut := &desc.Mutations[i]
		if mut.GetMaterializedViewRefresh() != nil {
			__antithesis_instrumentation__.Notify(565001)
			return nil, pgerror.Newf(pgcode.ObjectNotInPrerequisiteState, "view is already being refreshed")
		} else {
			__antithesis_instrumentation__.Notify(565002)
		}
	}
	__antithesis_instrumentation__.Notify(564990)

	hasAdminRole, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(565003)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565004)
	}
	__antithesis_instrumentation__.Notify(564991)

	hasOwnership, err := p.HasOwnership(ctx, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(565005)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565006)
	}
	__antithesis_instrumentation__.Notify(564992)

	if !(hasOwnership || func() bool {
		__antithesis_instrumentation__.Notify(565007)
		return hasAdminRole == true
	}() == true) {
		__antithesis_instrumentation__.Notify(565008)
		return nil, pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"must be owner of materialized view %s",
			desc.Name,
		)
	} else {
		__antithesis_instrumentation__.Notify(565009)
	}
	__antithesis_instrumentation__.Notify(564993)

	return &refreshMaterializedViewNode{n: n, desc: desc}, nil
}

func (n *refreshMaterializedViewNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(565010)

	telemetry.Inc(n.n.TelemetryCounter())

	if n.n.Concurrently {
		__antithesis_instrumentation__.Notify(565015)
		params.p.BufferClientNotice(
			params.ctx,
			pgnotice.Newf("CONCURRENTLY is not required as views are refreshed concurrently"),
		)
	} else {
		__antithesis_instrumentation__.Notify(565016)
	}
	__antithesis_instrumentation__.Notify(565011)

	newPrimaryIndex := n.desc.GetPrimaryIndex().IndexDescDeepCopy()
	newIndexes := make([]descpb.IndexDescriptor, len(n.desc.PublicNonPrimaryIndexes()))
	for i, idx := range n.desc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(565017)
		newIndexes[i] = idx.IndexDescDeepCopy()
	}
	__antithesis_instrumentation__.Notify(565012)

	getID := func() descpb.IndexID {
		__antithesis_instrumentation__.Notify(565018)
		res := n.desc.NextIndexID
		n.desc.NextIndexID++
		return res
	}
	__antithesis_instrumentation__.Notify(565013)
	newPrimaryIndex.ID = getID()
	for i := range newIndexes {
		__antithesis_instrumentation__.Notify(565019)
		newIndexes[i].ID = getID()
	}
	__antithesis_instrumentation__.Notify(565014)

	n.desc.AddMaterializedViewRefreshMutation(&descpb.MaterializedViewRefresh{
		NewPrimaryIndex: newPrimaryIndex,
		NewIndexes:      newIndexes,
		AsOf:            params.p.Txn().ReadTimestamp(),
		ShouldBackfill:  n.n.RefreshDataOption != tree.RefreshDataClear,
	})

	return params.p.writeSchemaChange(
		params.ctx,
		n.desc,
		n.desc.ClusterVersion().NextMutationID,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	)
}

func (n *refreshMaterializedViewNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(565020)
	return false, nil
}
func (n *refreshMaterializedViewNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(565021)
	return tree.Datums{}
}
func (n *refreshMaterializedViewNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(565022)
}
func (n *refreshMaterializedViewNode) ReadingOwnWrites() {
	__antithesis_instrumentation__.Notify(565023)
}
