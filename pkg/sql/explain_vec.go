package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/colflow"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type explainVecNode struct {
	optColumnsSlot

	options *tree.ExplainOptions
	plan    planComponents

	run struct {
		lines []string

		values tree.Datums

		cleanup func()
	}
}

func (n *explainVecNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(491325)
	n.run.values = make(tree.Datums, 1)
	distSQLPlanner := params.extendedEvalCtx.DistSQLPlanner
	distribution := getPlanDistribution(
		params.ctx, params.p, params.extendedEvalCtx.ExecCfg.NodeID,
		params.extendedEvalCtx.SessionData().DistSQLMode, n.plan.main,
	)
	outerSubqueries := params.p.curPlan.subqueryPlans
	planCtx := newPlanningCtxForExplainPurposes(distSQLPlanner, params, n.plan.subqueryPlans, distribution)
	defer func() {
		__antithesis_instrumentation__.Notify(491330)
		planCtx.planner.curPlan.subqueryPlans = outerSubqueries
	}()
	__antithesis_instrumentation__.Notify(491326)
	physPlan, err := newPhysPlanForExplainPurposes(params.ctx, planCtx, distSQLPlanner, n.plan.main)
	if err != nil {
		__antithesis_instrumentation__.Notify(491331)
		if len(n.plan.subqueryPlans) > 0 {
			__antithesis_instrumentation__.Notify(491333)
			return errors.New("running EXPLAIN (VEC) on this query is " +
				"unsupported because of the presence of subqueries")
		} else {
			__antithesis_instrumentation__.Notify(491334)
		}
		__antithesis_instrumentation__.Notify(491332)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491335)
	}
	__antithesis_instrumentation__.Notify(491327)

	distSQLPlanner.finalizePlanWithRowCount(planCtx, physPlan, n.plan.mainRowCount)
	flows := physPlan.GenerateFlowSpecs()
	flowCtx := newFlowCtxForExplainPurposes(planCtx, params.p)

	if flowCtx.EvalCtx.SessionData().VectorizeMode == sessiondatapb.VectorizeOff {
		__antithesis_instrumentation__.Notify(491336)
		return errors.New("vectorize is set to 'off'")
	} else {
		__antithesis_instrumentation__.Notify(491337)
	}
	__antithesis_instrumentation__.Notify(491328)
	verbose := n.options.Flags[tree.ExplainFlagVerbose]
	willDistribute := physPlan.Distribution.WillDistribute()
	n.run.lines, n.run.cleanup, err = colflow.ExplainVec(
		params.ctx, flowCtx, flows, physPlan.LocalProcessors, nil,
		distSQLPlanner.gatewaySQLInstanceID, verbose, willDistribute,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(491338)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491339)
	}
	__antithesis_instrumentation__.Notify(491329)
	return nil
}

func newFlowCtxForExplainPurposes(planCtx *PlanningCtx, p *planner) *execinfra.FlowCtx {
	__antithesis_instrumentation__.Notify(491340)
	return &execinfra.FlowCtx{
		NodeID:  planCtx.EvalContext().NodeID,
		EvalCtx: planCtx.EvalContext(),
		Cfg: &execinfra.ServerConfig{
			Settings:         p.execCfg.Settings,
			LogicalClusterID: p.DistSQLPlanner().distSQLSrv.ServerConfig.LogicalClusterID,
			VecFDSemaphore:   p.execCfg.DistSQLSrv.VecFDSemaphore,
			NodeDialer:       p.DistSQLPlanner().nodeDialer,
			PodNodeDialer:    p.DistSQLPlanner().podNodeDialer,
		},
		Descriptors: p.Descriptors(),
		DiskMonitor: &mon.BytesMonitor{},
	}
}

func newPlanningCtxForExplainPurposes(
	distSQLPlanner *DistSQLPlanner,
	params runParams,
	subqueryPlans []subquery,
	distribution physicalplan.PlanDistribution,
) *PlanningCtx {
	__antithesis_instrumentation__.Notify(491341)
	distribute := DistributionType(DistributionTypeNone)
	if distribution.WillDistribute() {
		__antithesis_instrumentation__.Notify(491344)
		distribute = DistributionTypeSystemTenantOnly
	} else {
		__antithesis_instrumentation__.Notify(491345)
	}
	__antithesis_instrumentation__.Notify(491342)
	planCtx := distSQLPlanner.NewPlanningCtx(params.ctx, params.extendedEvalCtx,
		params.p, params.p.txn, distribute)
	planCtx.ignoreClose = true
	planCtx.planner.curPlan.subqueryPlans = subqueryPlans
	for i := range planCtx.planner.curPlan.subqueryPlans {
		__antithesis_instrumentation__.Notify(491346)
		p := &planCtx.planner.curPlan.subqueryPlans[i]

		p.started = true
		p.result = tree.DNull
	}
	__antithesis_instrumentation__.Notify(491343)
	return planCtx
}

func (n *explainVecNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(491347)
	if len(n.run.lines) == 0 {
		__antithesis_instrumentation__.Notify(491349)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(491350)
	}
	__antithesis_instrumentation__.Notify(491348)
	n.run.values[0] = tree.NewDString(n.run.lines[0])
	n.run.lines = n.run.lines[1:]
	return true, nil
}

func (n *explainVecNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(491351)
	return n.run.values
}
func (n *explainVecNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(491352)
	n.plan.close(ctx)
	if n.run.cleanup != nil {
		__antithesis_instrumentation__.Notify(491353)
		n.run.cleanup()
	} else {
		__antithesis_instrumentation__.Notify(491354)
	}
}
