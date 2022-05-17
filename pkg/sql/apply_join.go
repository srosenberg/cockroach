package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type applyJoinNode struct {
	joinType descpb.JoinType

	input planDataSource

	pred *joinPredicate

	columns colinfo.ResultColumns

	rightTypes []*types.T

	planRightSideFn exec.ApplyJoinPlanRightSideFn

	run struct {
		emptyRight tree.Datums

		leftRow tree.Datums

		leftRowFoundAMatch bool

		rightRows rowContainerHelper

		rightRowsIterator *rowContainerIterator

		out tree.Datums

		done bool
	}
}

func newApplyJoinNode(
	joinType descpb.JoinType,
	left planDataSource,
	rightCols colinfo.ResultColumns,
	pred *joinPredicate,
	planRightSideFn exec.ApplyJoinPlanRightSideFn,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(245239)
	switch joinType {
	case descpb.RightOuterJoin, descpb.FullOuterJoin:
		__antithesis_instrumentation__.Notify(245241)
		return nil, errors.AssertionFailedf("unsupported right outer apply join: %d", redact.Safe(joinType))
	case descpb.ExceptAllJoin, descpb.IntersectAllJoin:
		__antithesis_instrumentation__.Notify(245242)
		return nil, errors.AssertionFailedf("unsupported apply set op: %d", redact.Safe(joinType))
	case descpb.RightSemiJoin, descpb.RightAntiJoin:
		__antithesis_instrumentation__.Notify(245243)
		return nil, errors.AssertionFailedf("unsupported right semi/anti apply join: %d", redact.Safe(joinType))
	default:
		__antithesis_instrumentation__.Notify(245244)
	}
	__antithesis_instrumentation__.Notify(245240)

	return &applyJoinNode{
		joinType:        joinType,
		input:           left,
		pred:            pred,
		rightTypes:      getTypesFromResultColumns(rightCols),
		planRightSideFn: planRightSideFn,
		columns:         pred.cols,
	}, nil
}

func (a *applyJoinNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(245245)

	if a.joinType == descpb.LeftOuterJoin {
		__antithesis_instrumentation__.Notify(245247)
		a.run.emptyRight = make(tree.Datums, len(a.rightTypes))
		for i := range a.run.emptyRight {
			__antithesis_instrumentation__.Notify(245248)
			a.run.emptyRight[i] = tree.DNull
		}
	} else {
		__antithesis_instrumentation__.Notify(245249)
	}
	__antithesis_instrumentation__.Notify(245246)
	a.run.out = make(tree.Datums, len(a.columns))
	a.run.rightRows.Init(a.rightTypes, params.extendedEvalCtx, "apply-join")
	return nil
}

func (a *applyJoinNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(245250)
	if a.run.done {
		__antithesis_instrumentation__.Notify(245252)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(245253)
	}
	__antithesis_instrumentation__.Notify(245251)

	for {
		__antithesis_instrumentation__.Notify(245254)
		if a.run.rightRowsIterator != nil {
			__antithesis_instrumentation__.Notify(245260)

			for {
				__antithesis_instrumentation__.Notify(245262)

				rrow, err := a.run.rightRowsIterator.Next()
				if err != nil {
					__antithesis_instrumentation__.Notify(245268)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(245269)
				}
				__antithesis_instrumentation__.Notify(245263)
				if rrow == nil {
					__antithesis_instrumentation__.Notify(245270)

					break
				} else {
					__antithesis_instrumentation__.Notify(245271)
				}
				__antithesis_instrumentation__.Notify(245264)

				predMatched, err := a.pred.eval(params.EvalContext(), a.run.leftRow, rrow)
				if err != nil {
					__antithesis_instrumentation__.Notify(245272)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(245273)
				}
				__antithesis_instrumentation__.Notify(245265)
				if !predMatched {
					__antithesis_instrumentation__.Notify(245274)

					continue
				} else {
					__antithesis_instrumentation__.Notify(245275)
				}
				__antithesis_instrumentation__.Notify(245266)

				a.run.leftRowFoundAMatch = true
				if a.joinType == descpb.LeftAntiJoin || func() bool {
					__antithesis_instrumentation__.Notify(245276)
					return a.joinType == descpb.LeftSemiJoin == true
				}() == true {
					__antithesis_instrumentation__.Notify(245277)

					break
				} else {
					__antithesis_instrumentation__.Notify(245278)
				}
				__antithesis_instrumentation__.Notify(245267)

				a.pred.prepareRow(a.run.out, a.run.leftRow, rrow)
				return true, nil
			}
			__antithesis_instrumentation__.Notify(245261)

			if err := a.clearRightRows(params); err != nil {
				__antithesis_instrumentation__.Notify(245279)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(245280)
			}
		} else {
			__antithesis_instrumentation__.Notify(245281)
		}
		__antithesis_instrumentation__.Notify(245255)

		foundAMatch := a.run.leftRowFoundAMatch
		a.run.leftRowFoundAMatch = false

		if a.run.leftRow != nil {
			__antithesis_instrumentation__.Notify(245282)

			if foundAMatch {
				__antithesis_instrumentation__.Notify(245283)
				if a.joinType == descpb.LeftSemiJoin {
					__antithesis_instrumentation__.Notify(245284)

					a.pred.prepareRow(a.run.out, a.run.leftRow, nil)
					a.run.leftRow = nil
					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(245285)
				}
			} else {
				__antithesis_instrumentation__.Notify(245286)

				switch a.joinType {
				case descpb.LeftOuterJoin:
					__antithesis_instrumentation__.Notify(245287)
					a.pred.prepareRow(a.run.out, a.run.leftRow, a.run.emptyRight)
					a.run.leftRow = nil
					return true, nil
				case descpb.LeftAntiJoin:
					__antithesis_instrumentation__.Notify(245288)
					a.pred.prepareRow(a.run.out, a.run.leftRow, nil)
					a.run.leftRow = nil
					return true, nil
				default:
					__antithesis_instrumentation__.Notify(245289)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(245290)
		}
		__antithesis_instrumentation__.Notify(245256)

		ok, err := a.input.plan.Next(params)
		if err != nil {
			__antithesis_instrumentation__.Notify(245291)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(245292)
		}
		__antithesis_instrumentation__.Notify(245257)
		if !ok {
			__antithesis_instrumentation__.Notify(245293)

			a.run.done = true
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(245294)
		}
		__antithesis_instrumentation__.Notify(245258)

		leftRow := a.input.plan.Values()
		a.run.leftRow = leftRow

		p, err := a.planRightSideFn(newExecFactory(params.p), leftRow)
		if err != nil {
			__antithesis_instrumentation__.Notify(245295)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(245296)
		}
		__antithesis_instrumentation__.Notify(245259)
		plan := p.(*planComponents)

		if err := a.runRightSidePlan(params, plan); err != nil {
			__antithesis_instrumentation__.Notify(245297)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(245298)
		}

	}
}

func (a *applyJoinNode) clearRightRows(params runParams) error {
	__antithesis_instrumentation__.Notify(245299)
	if err := a.run.rightRows.Clear(params.ctx); err != nil {
		__antithesis_instrumentation__.Notify(245301)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245302)
	}
	__antithesis_instrumentation__.Notify(245300)
	a.run.rightRowsIterator.Close()
	a.run.rightRowsIterator = nil
	return nil
}

func (a *applyJoinNode) runRightSidePlan(params runParams, plan *planComponents) error {
	__antithesis_instrumentation__.Notify(245303)
	rowResultWriter := NewRowResultWriter(&a.run.rightRows)
	if err := runPlanInsidePlan(params, plan, rowResultWriter); err != nil {
		__antithesis_instrumentation__.Notify(245305)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245306)
	}
	__antithesis_instrumentation__.Notify(245304)
	a.run.rightRowsIterator = newRowContainerIterator(params.ctx, a.run.rightRows, a.rightTypes)
	return nil
}

func runPlanInsidePlan(params runParams, plan *planComponents, resultWriter rowResultWriter) error {
	__antithesis_instrumentation__.Notify(245307)
	recv := MakeDistSQLReceiver(
		params.ctx, resultWriter, tree.Rows,
		params.ExecCfg().RangeDescriptorCache,
		params.p.Txn(),
		params.ExecCfg().Clock,
		params.p.extendedEvalCtx.Tracing,
		params.p.ExecCfg().ContentionRegistry,
		nil,
	)
	defer recv.Release()

	if len(plan.subqueryPlans) != 0 {
		__antithesis_instrumentation__.Notify(245310)

		if len(params.p.curPlan.subqueryPlans) != 0 {
			__antithesis_instrumentation__.Notify(245313)
			return unimplemented.NewWithIssue(66447, `apply joins with subqueries in the "inner" and "outer" contexts are not supported`)
		} else {
			__antithesis_instrumentation__.Notify(245314)
		}
		__antithesis_instrumentation__.Notify(245311)

		oldSubqueries := params.p.curPlan.subqueryPlans
		params.p.curPlan.subqueryPlans = plan.subqueryPlans
		defer func() {
			__antithesis_instrumentation__.Notify(245315)
			params.p.curPlan.subqueryPlans = oldSubqueries
		}()
		__antithesis_instrumentation__.Notify(245312)

		subqueryResultMemAcc := params.p.EvalContext().Mon.MakeBoundAccount()
		defer subqueryResultMemAcc.Close(params.ctx)
		if !params.p.extendedEvalCtx.ExecCfg.DistSQLPlanner.PlanAndRunSubqueries(
			params.ctx,
			params.p,
			params.extendedEvalCtx.copy,
			plan.subqueryPlans,
			recv,
			&subqueryResultMemAcc,
		) {
			__antithesis_instrumentation__.Notify(245316)
			return resultWriter.Err()
		} else {
			__antithesis_instrumentation__.Notify(245317)
		}
	} else {
		__antithesis_instrumentation__.Notify(245318)
	}
	__antithesis_instrumentation__.Notify(245308)

	evalCtx := params.p.ExtendedEvalContextCopy()
	plannerCopy := *params.p
	distributePlan := getPlanDistribution(
		params.ctx, &plannerCopy, plannerCopy.execCfg.NodeID, plannerCopy.SessionData().DistSQLMode, plan.main,
	)
	distributeType := DistributionType(DistributionTypeNone)
	if distributePlan.WillDistribute() {
		__antithesis_instrumentation__.Notify(245319)
		distributeType = DistributionTypeSystemTenantOnly
	} else {
		__antithesis_instrumentation__.Notify(245320)
	}
	__antithesis_instrumentation__.Notify(245309)
	planCtx := params.p.extendedEvalCtx.ExecCfg.DistSQLPlanner.NewPlanningCtx(
		params.ctx, evalCtx, &plannerCopy, params.p.txn, distributeType)
	planCtx.planner.curPlan.planComponents = *plan
	planCtx.ExtendedEvalCtx.Planner = &plannerCopy
	planCtx.stmtType = recv.stmtType

	params.p.extendedEvalCtx.ExecCfg.DistSQLPlanner.PlanAndRun(
		params.ctx, evalCtx, planCtx, params.p.Txn(), plan.main, recv,
	)()
	return resultWriter.Err()
}

func (a *applyJoinNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(245321)
	return a.run.out
}

func (a *applyJoinNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(245322)
	a.input.plan.Close(ctx)
	a.run.rightRows.Close(ctx)
	if a.run.rightRowsIterator != nil {
		__antithesis_instrumentation__.Notify(245323)
		a.run.rightRowsIterator.Close()
		a.run.rightRowsIterator = nil
	} else {
		__antithesis_instrumentation__.Notify(245324)
	}
}
