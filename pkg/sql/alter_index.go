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
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type alterIndexNode struct {
	n         *tree.AlterIndex
	tableDesc *tabledesc.Mutable
	index     catalog.Index
}

func (p *planner) AlterIndex(ctx context.Context, n *tree.AlterIndex) (planNode, error) {
	__antithesis_instrumentation__.Notify(243174)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER INDEX",
	); err != nil {
		__antithesis_instrumentation__.Notify(243177)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243178)
	}
	__antithesis_instrumentation__.Notify(243175)

	tableDesc, index, err := p.getTableAndIndex(ctx, &n.Index, privilege.CREATE)
	if err != nil {
		__antithesis_instrumentation__.Notify(243179)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243180)
	}
	__antithesis_instrumentation__.Notify(243176)
	return &alterIndexNode{n: n, tableDesc: tableDesc, index: index}, nil
}

func (n *alterIndexNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(243181) }

func (n *alterIndexNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(243182)

	descriptorChanged := false
	origNumMutations := len(n.tableDesc.Mutations)

	for _, cmd := range n.n.Cmds {
		__antithesis_instrumentation__.Notify(243188)
		switch t := cmd.(type) {
		case *tree.AlterIndexPartitionBy:
			__antithesis_instrumentation__.Notify(243189)
			telemetry.Inc(sqltelemetry.SchemaChangeAlterCounterWithExtra("index", "partition_by"))
			if n.tableDesc.GetLocalityConfig() != nil {
				__antithesis_instrumentation__.Notify(243197)
				return pgerror.Newf(
					pgcode.FeatureNotSupported,
					"cannot change the partitioning of an index if the table is part of a multi-region database",
				)
			} else {
				__antithesis_instrumentation__.Notify(243198)
			}
			__antithesis_instrumentation__.Notify(243190)
			if n.tableDesc.PartitionAllBy {
				__antithesis_instrumentation__.Notify(243199)
				return pgerror.Newf(
					pgcode.FeatureNotSupported,
					"cannot change the partitioning of an index if the table has PARTITION ALL BY defined",
				)
			} else {
				__antithesis_instrumentation__.Notify(243200)
			}
			__antithesis_instrumentation__.Notify(243191)
			if n.index.GetPartitioning().NumImplicitColumns() > 0 {
				__antithesis_instrumentation__.Notify(243201)
				return unimplemented.New(
					"ALTER INDEX PARTITION BY",
					"cannot ALTER INDEX PARTITION BY on an index which already has implicit column partitioning",
				)
			} else {
				__antithesis_instrumentation__.Notify(243202)
			}
			__antithesis_instrumentation__.Notify(243192)
			if n.index.IsSharded() {
				__antithesis_instrumentation__.Notify(243203)
				return pgerror.Newf(
					pgcode.FeatureNotSupported,
					"cannot set explicit partitioning with ALTER INDEX PARTITION BY on a hash sharded index",
				)
			} else {
				__antithesis_instrumentation__.Notify(243204)
			}
			__antithesis_instrumentation__.Notify(243193)
			allowImplicitPartitioning := params.p.EvalContext().SessionData().ImplicitColumnPartitioningEnabled || func() bool {
				__antithesis_instrumentation__.Notify(243205)
				return n.tableDesc.IsLocalityRegionalByRow() == true
			}() == true
			alteredIndexDesc := n.index.IndexDescDeepCopy()
			newImplicitCols, newPartitioning, err := CreatePartitioning(
				params.ctx,
				params.extendedEvalCtx.Settings,
				params.EvalContext(),
				n.tableDesc,
				alteredIndexDesc,
				t.PartitionBy,
				nil,
				allowImplicitPartitioning,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(243206)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243207)
			}
			__antithesis_instrumentation__.Notify(243194)
			if newPartitioning.NumImplicitColumns > 0 {
				__antithesis_instrumentation__.Notify(243208)
				return unimplemented.New(
					"ALTER INDEX PARTITION BY",
					"cannot ALTER INDEX and change the partitioning to contain implicit columns",
				)
			} else {
				__antithesis_instrumentation__.Notify(243209)
			}
			__antithesis_instrumentation__.Notify(243195)
			isIndexAltered := tabledesc.UpdateIndexPartitioning(&alteredIndexDesc, n.index.Primary(), newImplicitCols, newPartitioning)
			if isIndexAltered {
				__antithesis_instrumentation__.Notify(243210)
				oldPartitioning := n.index.GetPartitioning().DeepCopy()
				if n.index.Primary() {
					__antithesis_instrumentation__.Notify(243212)
					n.tableDesc.SetPrimaryIndex(alteredIndexDesc)
				} else {
					__antithesis_instrumentation__.Notify(243213)
					n.tableDesc.SetPublicNonPrimaryIndex(n.index.Ordinal(), alteredIndexDesc)
				}
				__antithesis_instrumentation__.Notify(243211)
				n.index = n.tableDesc.ActiveIndexes()[n.index.Ordinal()]
				descriptorChanged = true
				if err := deleteRemovedPartitionZoneConfigs(
					params.ctx,
					params.p.txn,
					n.tableDesc,
					n.index.GetID(),
					oldPartitioning,
					n.index.GetPartitioning(),
					params.extendedEvalCtx.ExecCfg,
				); err != nil {
					__antithesis_instrumentation__.Notify(243214)
					return err
				} else {
					__antithesis_instrumentation__.Notify(243215)
				}
			} else {
				__antithesis_instrumentation__.Notify(243216)
			}
		default:
			__antithesis_instrumentation__.Notify(243196)
			return errors.AssertionFailedf(
				"unsupported alter command: %T", cmd)
		}

	}
	__antithesis_instrumentation__.Notify(243183)

	version := params.ExecCfg().Settings.Version.ActiveVersion(params.ctx)
	if err := n.tableDesc.AllocateIDs(params.ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(243217)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243218)
	}
	__antithesis_instrumentation__.Notify(243184)

	addedMutations := len(n.tableDesc.Mutations) > origNumMutations
	if !addedMutations && func() bool {
		__antithesis_instrumentation__.Notify(243219)
		return !descriptorChanged == true
	}() == true {
		__antithesis_instrumentation__.Notify(243220)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(243221)
	}
	__antithesis_instrumentation__.Notify(243185)
	mutationID := descpb.InvalidMutationID
	if addedMutations {
		__antithesis_instrumentation__.Notify(243222)
		mutationID = n.tableDesc.ClusterVersion().NextMutationID
	} else {
		__antithesis_instrumentation__.Notify(243223)
	}
	__antithesis_instrumentation__.Notify(243186)
	if err := params.p.writeSchemaChange(
		params.ctx, n.tableDesc, mutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(243224)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243225)
	}
	__antithesis_instrumentation__.Notify(243187)

	return params.p.logEvent(params.ctx,
		n.tableDesc.ID,
		&eventpb.AlterIndex{
			TableName:  n.n.Index.Table.FQString(),
			IndexName:  n.index.GetName(),
			MutationID: uint32(mutationID),
		})
}

func (n *alterIndexNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(243226)
	return false, nil
}
func (n *alterIndexNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(243227)
	return tree.Datums{}
}
func (n *alterIndexNode) Close(context.Context) { __antithesis_instrumentation__.Notify(243228) }
