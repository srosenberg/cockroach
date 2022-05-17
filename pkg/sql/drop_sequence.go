package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type dropSequenceNode struct {
	n  *tree.DropSequence
	td []toDelete
}

func (p *planner) DropSequence(ctx context.Context, n *tree.DropSequence) (planNode, error) {
	__antithesis_instrumentation__.Notify(469423)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"DROP SEQUENCE",
	); err != nil {
		__antithesis_instrumentation__.Notify(469427)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469428)
	}
	__antithesis_instrumentation__.Notify(469424)

	td := make([]toDelete, 0, len(n.Names))
	for i := range n.Names {
		__antithesis_instrumentation__.Notify(469429)
		tn := &n.Names[i]
		droppedDesc, err := p.prepareDrop(ctx, tn, !n.IfExists, tree.ResolveRequireSequenceDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(469433)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469434)
		}
		__antithesis_instrumentation__.Notify(469430)
		if droppedDesc == nil {
			__antithesis_instrumentation__.Notify(469435)

			continue
		} else {
			__antithesis_instrumentation__.Notify(469436)
		}
		__antithesis_instrumentation__.Notify(469431)

		if depErr := p.sequenceDependencyError(ctx, droppedDesc, n.DropBehavior); depErr != nil {
			__antithesis_instrumentation__.Notify(469437)
			return nil, depErr
		} else {
			__antithesis_instrumentation__.Notify(469438)
		}
		__antithesis_instrumentation__.Notify(469432)

		td = append(td, toDelete{tn, droppedDesc})
	}
	__antithesis_instrumentation__.Notify(469425)

	if len(td) == 0 {
		__antithesis_instrumentation__.Notify(469439)
		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(469440)
	}
	__antithesis_instrumentation__.Notify(469426)

	return &dropSequenceNode{
		n:  n,
		td: td,
	}, nil
}

func (n *dropSequenceNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(469441) }

func (n *dropSequenceNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(469442)
	telemetry.Inc(sqltelemetry.SchemaChangeDropCounter("sequence"))

	ctx := params.ctx
	for _, toDel := range n.td {
		__antithesis_instrumentation__.Notify(469444)
		droppedDesc := toDel.desc
		err := params.p.dropSequenceImpl(
			ctx, droppedDesc, true, tree.AsStringWithFQNames(n.n, params.Ann()), n.n.DropBehavior,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469446)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469447)
		}
		__antithesis_instrumentation__.Notify(469445)

		if err := params.p.logEvent(params.ctx,
			droppedDesc.ID,
			&eventpb.DropSequence{
				SequenceName: toDel.tn.FQString(),
			}); err != nil {
			__antithesis_instrumentation__.Notify(469448)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469449)
		}
	}
	__antithesis_instrumentation__.Notify(469443)
	return nil
}

func (*dropSequenceNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(469450)
	return false, nil
}
func (*dropSequenceNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(469451)
	return tree.Datums{}
}
func (*dropSequenceNode) Close(context.Context) { __antithesis_instrumentation__.Notify(469452) }

func (p *planner) dropSequenceImpl(
	ctx context.Context,
	seqDesc *tabledesc.Mutable,
	queueJob bool,
	jobDesc string,
	behavior tree.DropBehavior,
) error {
	__antithesis_instrumentation__.Notify(469453)
	if err := removeSequenceOwnerIfExists(ctx, p, seqDesc.ID, seqDesc.GetSequenceOpts()); err != nil {
		__antithesis_instrumentation__.Notify(469456)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469457)
	}
	__antithesis_instrumentation__.Notify(469454)
	if behavior == tree.DropCascade {
		__antithesis_instrumentation__.Notify(469458)
		if err := dropDependentOnSequence(ctx, p, seqDesc); err != nil {
			__antithesis_instrumentation__.Notify(469459)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469460)
		}
	} else {
		__antithesis_instrumentation__.Notify(469461)
	}
	__antithesis_instrumentation__.Notify(469455)
	return p.initiateDropTable(ctx, seqDesc, queueJob, jobDesc)
}

func (p *planner) sequenceDependencyError(
	ctx context.Context, droppedDesc *tabledesc.Mutable, behavior tree.DropBehavior,
) error {
	__antithesis_instrumentation__.Notify(469462)
	if behavior != tree.DropCascade && func() bool {
		__antithesis_instrumentation__.Notify(469464)
		return len(droppedDesc.DependedOnBy) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(469465)
		return pgerror.Newf(
			pgcode.DependentObjectsStillExist,
			"cannot drop sequence %s because other objects depend on it",
			droppedDesc.Name,
		)
	} else {
		__antithesis_instrumentation__.Notify(469466)
	}
	__antithesis_instrumentation__.Notify(469463)
	return nil
}

func (p *planner) canRemoveAllTableOwnedSequences(
	ctx context.Context, desc *tabledesc.Mutable, behavior tree.DropBehavior,
) error {
	__antithesis_instrumentation__.Notify(469467)
	for _, col := range desc.PublicColumns() {
		__antithesis_instrumentation__.Notify(469469)
		err := p.canRemoveOwnedSequencesImpl(ctx, desc, col, behavior, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(469470)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469471)
		}
	}
	__antithesis_instrumentation__.Notify(469468)
	return nil
}

func (p *planner) canRemoveAllColumnOwnedSequences(
	ctx context.Context, desc *tabledesc.Mutable, col catalog.Column, behavior tree.DropBehavior,
) error {
	__antithesis_instrumentation__.Notify(469472)
	return p.canRemoveOwnedSequencesImpl(ctx, desc, col, behavior, true)
}

func (p *planner) canRemoveOwnedSequencesImpl(
	ctx context.Context,
	desc *tabledesc.Mutable,
	col catalog.Column,
	behavior tree.DropBehavior,
	isColumnDrop bool,
) error {
	__antithesis_instrumentation__.Notify(469473)
	for i := 0; i < col.NumOwnsSequences(); i++ {
		__antithesis_instrumentation__.Notify(469475)
		sequenceID := col.GetOwnsSequenceID(i)
		seqDesc, err := p.LookupTableByID(ctx, sequenceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(469481)

			if errors.Is(err, catalog.ErrDescriptorDropped) || func() bool {
				__antithesis_instrumentation__.Notify(469483)
				return pgerror.GetPGCode(err) == pgcode.UndefinedTable == true
			}() == true {
				__antithesis_instrumentation__.Notify(469484)
				log.Eventf(ctx, "swallowing error ensuring owned sequences can be removed: %s", err.Error())
				continue
			} else {
				__antithesis_instrumentation__.Notify(469485)
			}
			__antithesis_instrumentation__.Notify(469482)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469486)
		}
		__antithesis_instrumentation__.Notify(469476)

		var firstDep *descpb.TableDescriptor_Reference
		multipleIterationErr := seqDesc.ForeachDependedOnBy(func(dep *descpb.TableDescriptor_Reference) error {
			__antithesis_instrumentation__.Notify(469487)
			if firstDep != nil {
				__antithesis_instrumentation__.Notify(469489)
				return iterutil.StopIteration()
			} else {
				__antithesis_instrumentation__.Notify(469490)
			}
			__antithesis_instrumentation__.Notify(469488)
			firstDep = dep
			return nil
		})
		__antithesis_instrumentation__.Notify(469477)

		if firstDep == nil {
			__antithesis_instrumentation__.Notify(469491)

			continue
		} else {
			__antithesis_instrumentation__.Notify(469492)
		}
		__antithesis_instrumentation__.Notify(469478)

		if multipleIterationErr == nil && func() bool {
			__antithesis_instrumentation__.Notify(469493)
			return firstDep.ID == desc.ID == true
		}() == true {
			__antithesis_instrumentation__.Notify(469494)

			if !isColumnDrop {
				__antithesis_instrumentation__.Notify(469496)

				continue
			} else {
				__antithesis_instrumentation__.Notify(469497)
			}
			__antithesis_instrumentation__.Notify(469495)

			if len(firstDep.ColumnIDs) == 1 && func() bool {
				__antithesis_instrumentation__.Notify(469498)
				return firstDep.ColumnIDs[0] == col.GetID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(469499)

				continue
			} else {
				__antithesis_instrumentation__.Notify(469500)
			}
		} else {
			__antithesis_instrumentation__.Notify(469501)
		}
		__antithesis_instrumentation__.Notify(469479)

		if behavior == tree.DropCascade {
			__antithesis_instrumentation__.Notify(469502)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469503)
		}
		__antithesis_instrumentation__.Notify(469480)

		return pgerror.Newf(
			pgcode.DependentObjectsStillExist,
			"cannot drop table %s because other objects depend on it",
			desc.Name,
		)
	}
	__antithesis_instrumentation__.Notify(469474)
	return nil
}

func dropDependentOnSequence(ctx context.Context, p *planner, seqDesc *tabledesc.Mutable) error {
	__antithesis_instrumentation__.Notify(469504)
	for _, dependent := range seqDesc.DependedOnBy {
		__antithesis_instrumentation__.Notify(469506)
		tblDesc, err := p.Descriptors().GetMutableTableByID(ctx, p.txn, dependent.ID,
			tree.ObjectLookupFlags{
				CommonLookupFlags: tree.CommonLookupFlags{
					IncludeOffline: true,
					IncludeDropped: true,
				},
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(469512)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469513)
		}
		__antithesis_instrumentation__.Notify(469507)

		if tblDesc.Dropped() {
			__antithesis_instrumentation__.Notify(469514)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469515)
		}
		__antithesis_instrumentation__.Notify(469508)

		if tblDesc.IsView() {
			__antithesis_instrumentation__.Notify(469516)
			_, err = p.dropViewImpl(ctx, tblDesc, false, "", tree.DropCascade)
			if err != nil {
				__antithesis_instrumentation__.Notify(469518)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469519)
			}
			__antithesis_instrumentation__.Notify(469517)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469520)
		}
		__antithesis_instrumentation__.Notify(469509)

		colsToDropDefault := make(map[descpb.ColumnID]struct{})
		for _, colID := range dependent.ColumnIDs {
			__antithesis_instrumentation__.Notify(469521)
			colsToDropDefault[colID] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(469510)

		for _, column := range tblDesc.PublicColumns() {
			__antithesis_instrumentation__.Notify(469522)
			if _, ok := colsToDropDefault[column.GetID()]; ok {
				__antithesis_instrumentation__.Notify(469523)
				column.ColumnDesc().DefaultExpr = nil
				if err := p.removeSequenceDependencies(ctx, tblDesc, column); err != nil {
					__antithesis_instrumentation__.Notify(469524)
					return err
				} else {
					__antithesis_instrumentation__.Notify(469525)
				}
			} else {
				__antithesis_instrumentation__.Notify(469526)
			}
		}
		__antithesis_instrumentation__.Notify(469511)

		jobDesc := fmt.Sprintf(
			"removing default expressions using sequence %q since it is being dropped",
			seqDesc.Name,
		)
		if err := p.writeSchemaChange(
			ctx, tblDesc, descpb.InvalidMutationID, jobDesc,
		); err != nil {
			__antithesis_instrumentation__.Notify(469527)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469528)
		}
	}
	__antithesis_instrumentation__.Notify(469505)
	return nil
}
