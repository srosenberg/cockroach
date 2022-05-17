package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type dropViewNode struct {
	n  *tree.DropView
	td []toDelete
}

func (p *planner) DropView(ctx context.Context, n *tree.DropView) (planNode, error) {
	__antithesis_instrumentation__.Notify(469868)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"DROP VIEW",
	); err != nil {
		__antithesis_instrumentation__.Notify(469873)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469874)
	}
	__antithesis_instrumentation__.Notify(469869)

	td := make([]toDelete, 0, len(n.Names))
	for i := range n.Names {
		__antithesis_instrumentation__.Notify(469875)
		tn := &n.Names[i]
		droppedDesc, err := p.prepareDrop(ctx, tn, !n.IfExists, tree.ResolveRequireViewDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(469879)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469880)
		}
		__antithesis_instrumentation__.Notify(469876)
		if droppedDesc == nil {
			__antithesis_instrumentation__.Notify(469881)

			continue
		} else {
			__antithesis_instrumentation__.Notify(469882)
		}
		__antithesis_instrumentation__.Notify(469877)
		if err := checkViewMatchesMaterialized(droppedDesc, true, n.IsMaterialized); err != nil {
			__antithesis_instrumentation__.Notify(469883)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469884)
		}
		__antithesis_instrumentation__.Notify(469878)

		td = append(td, toDelete{tn, droppedDesc})
	}
	__antithesis_instrumentation__.Notify(469870)

	for _, toDel := range td {
		__antithesis_instrumentation__.Notify(469885)
		droppedDesc := toDel.desc
		for _, ref := range droppedDesc.DependedOnBy {
			__antithesis_instrumentation__.Notify(469886)

			if descInSlice(ref.ID, td) {
				__antithesis_instrumentation__.Notify(469888)
				continue
			} else {
				__antithesis_instrumentation__.Notify(469889)
			}
			__antithesis_instrumentation__.Notify(469887)
			if err := p.canRemoveDependentView(ctx, droppedDesc, ref, n.DropBehavior); err != nil {
				__antithesis_instrumentation__.Notify(469890)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(469891)
			}
		}
	}
	__antithesis_instrumentation__.Notify(469871)

	if len(td) == 0 {
		__antithesis_instrumentation__.Notify(469892)
		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(469893)
	}
	__antithesis_instrumentation__.Notify(469872)
	return &dropViewNode{n: n, td: td}, nil
}

func (n *dropViewNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(469894) }

func (n *dropViewNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(469895)
	telemetry.Inc(n.n.TelemetryCounter())

	ctx := params.ctx
	for _, toDel := range n.td {
		__antithesis_instrumentation__.Notify(469897)
		droppedDesc := toDel.desc
		if droppedDesc == nil {
			__antithesis_instrumentation__.Notify(469900)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469901)
		}
		__antithesis_instrumentation__.Notify(469898)

		cascadeDroppedViews, err := params.p.dropViewImpl(
			ctx, droppedDesc, true, tree.AsStringWithFQNames(n.n, params.Ann()), n.n.DropBehavior,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469902)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469903)
		}
		__antithesis_instrumentation__.Notify(469899)

		if err := params.p.logEvent(ctx,
			droppedDesc.ID,
			&eventpb.DropView{
				ViewName:            toDel.tn.FQString(),
				CascadeDroppedViews: cascadeDroppedViews}); err != nil {
			__antithesis_instrumentation__.Notify(469904)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469905)
		}
	}
	__antithesis_instrumentation__.Notify(469896)
	return nil
}

func (*dropViewNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(469906)
	return false, nil
}
func (*dropViewNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(469907)
	return tree.Datums{}
}
func (*dropViewNode) Close(context.Context) { __antithesis_instrumentation__.Notify(469908) }

func descInSlice(descID descpb.ID, td []toDelete) bool {
	__antithesis_instrumentation__.Notify(469909)
	for _, toDel := range td {
		__antithesis_instrumentation__.Notify(469911)
		if descID == toDel.desc.ID {
			__antithesis_instrumentation__.Notify(469912)
			return true
		} else {
			__antithesis_instrumentation__.Notify(469913)
		}
	}
	__antithesis_instrumentation__.Notify(469910)
	return false
}

func (p *planner) canRemoveDependentView(
	ctx context.Context,
	from *tabledesc.Mutable,
	ref descpb.TableDescriptor_Reference,
	behavior tree.DropBehavior,
) error {
	__antithesis_instrumentation__.Notify(469914)
	return p.canRemoveDependentViewGeneric(ctx, string(from.DescriptorType()), from.Name, from.ParentID, ref, behavior)
}

func (p *planner) canRemoveDependentViewGeneric(
	ctx context.Context,
	typeName string,
	objName string,
	parentID descpb.ID,
	ref descpb.TableDescriptor_Reference,
	behavior tree.DropBehavior,
) error {
	__antithesis_instrumentation__.Notify(469915)
	viewDesc, err := p.getViewDescForCascade(ctx, typeName, objName, parentID, ref.ID, behavior)
	if err != nil {
		__antithesis_instrumentation__.Notify(469919)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469920)
	}
	__antithesis_instrumentation__.Notify(469916)
	if err := p.CheckPrivilege(ctx, viewDesc, privilege.DROP); err != nil {
		__antithesis_instrumentation__.Notify(469921)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469922)
	}
	__antithesis_instrumentation__.Notify(469917)

	for _, ref := range viewDesc.DependedOnBy {
		__antithesis_instrumentation__.Notify(469923)
		if err := p.canRemoveDependentView(ctx, viewDesc, ref, behavior); err != nil {
			__antithesis_instrumentation__.Notify(469924)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469925)
		}
	}
	__antithesis_instrumentation__.Notify(469918)
	return nil
}

func (p *planner) removeDependentView(
	ctx context.Context, tableDesc, viewDesc *tabledesc.Mutable, jobDesc string,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(469926)

	tableDesc.DependedOnBy = removeMatchingReferences(tableDesc.DependedOnBy, viewDesc.ID)

	return p.dropViewImpl(ctx, viewDesc, true, jobDesc, tree.DropCascade)
}

func (p *planner) dropViewImpl(
	ctx context.Context,
	viewDesc *tabledesc.Mutable,
	queueJob bool,
	jobDesc string,
	behavior tree.DropBehavior,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(469927)
	var cascadeDroppedViews []string

	dependedOn := append([]descpb.ID(nil), viewDesc.DependsOn...)
	for _, depID := range dependedOn {
		__antithesis_instrumentation__.Notify(469933)
		dependencyDesc, err := p.Descriptors().GetMutableTableVersionByID(ctx, depID, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(469936)
			return cascadeDroppedViews,
				errors.Wrapf(err, "error resolving dependency relation ID %d", depID)
		} else {
			__antithesis_instrumentation__.Notify(469937)
		}
		__antithesis_instrumentation__.Notify(469934)

		if dependencyDesc.Dropped() {
			__antithesis_instrumentation__.Notify(469938)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469939)
		}
		__antithesis_instrumentation__.Notify(469935)
		dependencyDesc.DependedOnBy = removeMatchingReferences(dependencyDesc.DependedOnBy, viewDesc.ID)
		if err := p.writeSchemaChange(
			ctx, dependencyDesc, descpb.InvalidMutationID,
			fmt.Sprintf("removing references for view %s from table %s(%d)",
				viewDesc.Name, dependencyDesc.Name, dependencyDesc.ID),
		); err != nil {
			__antithesis_instrumentation__.Notify(469940)
			return cascadeDroppedViews, err
		} else {
			__antithesis_instrumentation__.Notify(469941)
		}

	}
	__antithesis_instrumentation__.Notify(469928)
	viewDesc.DependsOn = nil

	typesDependedOn := append([]descpb.ID(nil), viewDesc.DependsOnTypes...)
	backRefJobDesc := fmt.Sprintf("updating type back references %v for table %d", typesDependedOn, viewDesc.ID)
	if err := p.removeTypeBackReferences(ctx, typesDependedOn, viewDesc.ID, backRefJobDesc); err != nil {
		__antithesis_instrumentation__.Notify(469942)
		return cascadeDroppedViews, err
	} else {
		__antithesis_instrumentation__.Notify(469943)
	}
	__antithesis_instrumentation__.Notify(469929)

	if behavior == tree.DropCascade {
		__antithesis_instrumentation__.Notify(469944)
		dependedOnBy := append([]descpb.TableDescriptor_Reference(nil), viewDesc.DependedOnBy...)
		for _, ref := range dependedOnBy {
			__antithesis_instrumentation__.Notify(469945)
			dependentDesc, err := p.getViewDescForCascade(
				ctx, string(viewDesc.DescriptorType()), viewDesc.Name, viewDesc.ParentID, ref.ID, behavior,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(469948)
				return cascadeDroppedViews, err
			} else {
				__antithesis_instrumentation__.Notify(469949)
			}
			__antithesis_instrumentation__.Notify(469946)

			qualifiedView, err := p.getQualifiedTableName(ctx, dependentDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(469950)
				return cascadeDroppedViews, err
			} else {
				__antithesis_instrumentation__.Notify(469951)
			}
			__antithesis_instrumentation__.Notify(469947)

			if !dependentDesc.Dropped() {
				__antithesis_instrumentation__.Notify(469952)
				cascadedViews, err := p.dropViewImpl(ctx, dependentDesc, queueJob, "dropping dependent view", behavior)
				if err != nil {
					__antithesis_instrumentation__.Notify(469954)
					return cascadeDroppedViews, err
				} else {
					__antithesis_instrumentation__.Notify(469955)
				}
				__antithesis_instrumentation__.Notify(469953)
				cascadeDroppedViews = append(cascadeDroppedViews, cascadedViews...)
				cascadeDroppedViews = append(cascadeDroppedViews, qualifiedView.FQString())
			} else {
				__antithesis_instrumentation__.Notify(469956)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(469957)
	}
	__antithesis_instrumentation__.Notify(469930)

	if err := p.removeBackRefsFromAllTypesInTable(ctx, viewDesc); err != nil {
		__antithesis_instrumentation__.Notify(469958)
		return cascadeDroppedViews, err
	} else {
		__antithesis_instrumentation__.Notify(469959)
	}
	__antithesis_instrumentation__.Notify(469931)

	if err := p.initiateDropTable(ctx, viewDesc, queueJob, jobDesc); err != nil {
		__antithesis_instrumentation__.Notify(469960)
		return cascadeDroppedViews, err
	} else {
		__antithesis_instrumentation__.Notify(469961)
	}
	__antithesis_instrumentation__.Notify(469932)

	return cascadeDroppedViews, nil
}

func (p *planner) getViewDescForCascade(
	ctx context.Context,
	typeName string,
	objName string,
	parentID, viewID descpb.ID,
	behavior tree.DropBehavior,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(469962)
	viewDesc, err := p.Descriptors().GetMutableTableVersionByID(ctx, viewID, p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(469965)
		log.Warningf(ctx, "unable to retrieve descriptor for view %d: %v", viewID, err)
		return nil, errors.Wrapf(err, "error resolving dependent view ID %d", viewID)
	} else {
		__antithesis_instrumentation__.Notify(469966)
	}
	__antithesis_instrumentation__.Notify(469963)
	if behavior != tree.DropCascade {
		__antithesis_instrumentation__.Notify(469967)
		viewName := viewDesc.Name
		if viewDesc.ParentID != parentID {
			__antithesis_instrumentation__.Notify(469969)
			var err error
			viewFQName, err := p.getQualifiedTableName(ctx, viewDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(469971)
				log.Warningf(ctx, "unable to retrieve qualified name of view %d: %v", viewID, err)
				return nil, sqlerrors.NewDependentObjectErrorf(
					"cannot drop %s %q because a view depends on it", typeName, objName)
			} else {
				__antithesis_instrumentation__.Notify(469972)
			}
			__antithesis_instrumentation__.Notify(469970)
			viewName = viewFQName.FQString()
		} else {
			__antithesis_instrumentation__.Notify(469973)
		}
		__antithesis_instrumentation__.Notify(469968)
		return nil, errors.WithHintf(
			sqlerrors.NewDependentObjectErrorf("cannot drop %s %q because view %q depends on it",
				typeName, objName, viewName),
			"you can drop %s instead.", viewName)
	} else {
		__antithesis_instrumentation__.Notify(469974)
	}
	__antithesis_instrumentation__.Notify(469964)
	return viewDesc, nil
}
