package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func PlanAndRunCTAS(
	ctx context.Context,
	dsp *DistSQLPlanner,
	planner *planner,
	txn *kv.Txn,
	isLocal bool,
	in planMaybePhysical,
	out execinfrapb.ProcessorCoreUnion,
	recv *DistSQLReceiver,
) {
	__antithesis_instrumentation__.Notify(467664)
	distribute := DistributionType(DistributionTypeNone)
	if !isLocal {
		__antithesis_instrumentation__.Notify(467667)
		distribute = DistributionTypeSystemTenantOnly
	} else {
		__antithesis_instrumentation__.Notify(467668)
	}
	__antithesis_instrumentation__.Notify(467665)
	planCtx := dsp.NewPlanningCtx(ctx, planner.ExtendedEvalContext(), planner,
		txn, distribute)
	planCtx.stmtType = tree.Rows

	physPlan, cleanup, err := dsp.createPhysPlan(ctx, planCtx, in)
	defer cleanup()
	if err != nil {
		__antithesis_instrumentation__.Notify(467669)
		recv.SetError(errors.Wrapf(err, "constructing distSQL plan"))
		return
	} else {
		__antithesis_instrumentation__.Notify(467670)
	}
	__antithesis_instrumentation__.Notify(467666)
	physPlan.AddNoGroupingStage(
		out, execinfrapb.PostProcessSpec{}, rowexec.CTASPlanResultTypes, execinfrapb.Ordering{},
	)

	physPlan.PlanToStreamColMap = []int{0}

	evalCtxCopy := planner.ExtendedEvalContextCopy()
	dsp.FinalizePlan(planCtx, physPlan)
	dsp.Run(ctx, planCtx, txn, physPlan, recv, evalCtxCopy, nil)()
}
