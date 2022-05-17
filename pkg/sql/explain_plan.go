package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net/url"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/colflow"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec/explain"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/errors"
)

type explainPlanNode struct {
	optColumnsSlot

	options *tree.ExplainOptions

	flags explain.Flags
	plan  *explain.Plan
	run   explainPlanNodeRun
}

type explainPlanNodeRun struct {
	results *valuesNode
}

func (e *explainPlanNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(491243)
	ob := explain.NewOutputBuilder(e.flags)
	plan := e.plan.WrappedPlan.(*planComponents)

	var rows []string
	if e.options.Mode == tree.ExplainGist {
		__antithesis_instrumentation__.Notify(491247)
		rows = []string{e.plan.Gist.String()}
	} else {
		__antithesis_instrumentation__.Notify(491248)

		distribution := getPlanDistribution(
			params.ctx, params.p, params.extendedEvalCtx.ExecCfg.NodeID,
			params.extendedEvalCtx.SessionData().DistSQLMode, plan.main,
		)

		outerSubqueries := params.p.curPlan.subqueryPlans
		distSQLPlanner := params.extendedEvalCtx.DistSQLPlanner
		planCtx := newPlanningCtxForExplainPurposes(distSQLPlanner, params, plan.subqueryPlans, distribution)
		defer func() {
			__antithesis_instrumentation__.Notify(491251)
			planCtx.planner.curPlan.subqueryPlans = outerSubqueries
		}()
		__antithesis_instrumentation__.Notify(491249)
		physicalPlan, err := newPhysPlanForExplainPurposes(params.ctx, planCtx, distSQLPlanner, plan.main)
		var diagramURL url.URL
		var diagramJSON string
		if err != nil {
			__antithesis_instrumentation__.Notify(491252)
			if e.options.Mode == tree.ExplainDistSQL {
				__antithesis_instrumentation__.Notify(491254)
				if len(plan.subqueryPlans) > 0 {
					__antithesis_instrumentation__.Notify(491256)
					return errors.New("running EXPLAIN (DISTSQL) on this query is " +
						"unsupported because of the presence of subqueries")
				} else {
					__antithesis_instrumentation__.Notify(491257)
				}
				__antithesis_instrumentation__.Notify(491255)
				return err
			} else {
				__antithesis_instrumentation__.Notify(491258)
			}
			__antithesis_instrumentation__.Notify(491253)
			ob.AddDistribution(distribution.String())

		} else {
			__antithesis_instrumentation__.Notify(491259)

			distSQLPlanner.finalizePlanWithRowCount(planCtx, physicalPlan, plan.mainRowCount)
			ob.AddDistribution(physicalPlan.Distribution.String())
			flows := physicalPlan.GenerateFlowSpecs()

			ctxSessionData := planCtx.EvalContext().SessionData()
			var willVectorize bool
			if ctxSessionData.VectorizeMode == sessiondatapb.VectorizeOff {
				__antithesis_instrumentation__.Notify(491261)
				willVectorize = false
			} else {
				__antithesis_instrumentation__.Notify(491262)
				willVectorize = true
				for _, flow := range flows {
					__antithesis_instrumentation__.Notify(491263)
					if err := colflow.IsSupported(ctxSessionData.VectorizeMode, flow); err != nil {
						__antithesis_instrumentation__.Notify(491264)
						willVectorize = false
						break
					} else {
						__antithesis_instrumentation__.Notify(491265)
					}
				}
			}
			__antithesis_instrumentation__.Notify(491260)
			ob.AddVectorized(willVectorize)

			if e.options.Mode == tree.ExplainDistSQL {
				__antithesis_instrumentation__.Notify(491266)
				flags := execinfrapb.DiagramFlags{
					ShowInputTypes: e.options.Flags[tree.ExplainFlagTypes],
				}
				diagram, err := execinfrapb.GeneratePlanDiagram(params.p.stmt.String(), flows, flags)
				if err != nil {
					__antithesis_instrumentation__.Notify(491268)
					return err
				} else {
					__antithesis_instrumentation__.Notify(491269)
				}
				__antithesis_instrumentation__.Notify(491267)

				diagramJSON, diagramURL, err = diagram.ToURL()
				if err != nil {
					__antithesis_instrumentation__.Notify(491270)
					return err
				} else {
					__antithesis_instrumentation__.Notify(491271)
				}
			} else {
				__antithesis_instrumentation__.Notify(491272)
			}
		}
		__antithesis_instrumentation__.Notify(491250)

		if e.options.Flags[tree.ExplainFlagJSON] {
			__antithesis_instrumentation__.Notify(491273)

			rows = []string{diagramJSON}
		} else {
			__antithesis_instrumentation__.Notify(491274)
			if err := emitExplain(ob, params.EvalContext(), params.p.ExecCfg().Codec, e.plan); err != nil {
				__antithesis_instrumentation__.Notify(491276)
				return err
			} else {
				__antithesis_instrumentation__.Notify(491277)
			}
			__antithesis_instrumentation__.Notify(491275)
			rows = ob.BuildStringRows()
			if e.options.Mode == tree.ExplainDistSQL {
				__antithesis_instrumentation__.Notify(491278)
				rows = append(rows, "", fmt.Sprintf("Diagram: %s", diagramURL.String()))
			} else {
				__antithesis_instrumentation__.Notify(491279)
			}
		}
	}
	__antithesis_instrumentation__.Notify(491244)

	if params.p.instrumentation.indexRecommendations != nil {
		__antithesis_instrumentation__.Notify(491280)

		rows = append(rows, "")
		rows = append(rows, params.p.instrumentation.indexRecommendations...)
	} else {
		__antithesis_instrumentation__.Notify(491281)
	}
	__antithesis_instrumentation__.Notify(491245)
	v := params.p.newContainerValuesNode(colinfo.ExplainPlanColumns, 0)
	datums := make([]tree.DString, len(rows))
	for i, row := range rows {
		__antithesis_instrumentation__.Notify(491282)
		datums[i] = tree.DString(row)
		if _, err := v.rows.AddRow(params.ctx, tree.Datums{&datums[i]}); err != nil {
			__antithesis_instrumentation__.Notify(491283)
			return err
		} else {
			__antithesis_instrumentation__.Notify(491284)
		}
	}
	__antithesis_instrumentation__.Notify(491246)
	e.run.results = v

	return nil
}

func emitExplain(
	ob *explain.OutputBuilder,
	evalCtx *tree.EvalContext,
	codec keys.SQLCodec,
	explainPlan *explain.Plan,
) (err error) {
	__antithesis_instrumentation__.Notify(491285)

	defer func() {
		__antithesis_instrumentation__.Notify(491289)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(491290)

			if ok, e := errorutil.ShouldCatch(r); ok && func() bool {
				__antithesis_instrumentation__.Notify(491291)
				return !buildutil.CrdbTestBuild == true
			}() == true {
				__antithesis_instrumentation__.Notify(491292)
				err = e
			} else {
				__antithesis_instrumentation__.Notify(491293)

				panic(r)
			}
		} else {
			__antithesis_instrumentation__.Notify(491294)
		}
	}()
	__antithesis_instrumentation__.Notify(491286)

	if explainPlan == nil {
		__antithesis_instrumentation__.Notify(491295)
		return errors.AssertionFailedf("no plan")
	} else {
		__antithesis_instrumentation__.Notify(491296)
	}
	__antithesis_instrumentation__.Notify(491287)

	spanFormatFn := func(table cat.Table, index cat.Index, scanParams exec.ScanParams) string {
		__antithesis_instrumentation__.Notify(491297)
		if table.IsVirtualTable() {
			__antithesis_instrumentation__.Notify(491301)
			return "<virtual table spans>"
		} else {
			__antithesis_instrumentation__.Notify(491302)
		}
		__antithesis_instrumentation__.Notify(491298)
		tabDesc := table.(*optTable).desc
		idx := index.(*optIndex).idx
		spans, err := generateScanSpans(evalCtx, codec, tabDesc, idx, scanParams)
		if err != nil {
			__antithesis_instrumentation__.Notify(491303)
			return err.Error()
		} else {
			__antithesis_instrumentation__.Notify(491304)
		}
		__antithesis_instrumentation__.Notify(491299)

		skip := 2
		if !codec.ForSystemTenant() {
			__antithesis_instrumentation__.Notify(491305)
			skip = 4
		} else {
			__antithesis_instrumentation__.Notify(491306)
		}
		__antithesis_instrumentation__.Notify(491300)
		return catalogkeys.PrettySpans(idx, spans, skip)
	}
	__antithesis_instrumentation__.Notify(491288)

	return explain.Emit(explainPlan, ob, spanFormatFn)
}

func (e *explainPlanNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(491307)
	return e.run.results.Next(params)
}
func (e *explainPlanNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(491308)
	return e.run.results.Values()
}

func (e *explainPlanNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(491309)
	closeNode := func(n exec.Node) {
		__antithesis_instrumentation__.Notify(491313)
		switch n := n.(type) {
		case planNode:
			__antithesis_instrumentation__.Notify(491314)
			n.Close(ctx)
		case planMaybePhysical:
			__antithesis_instrumentation__.Notify(491315)
			n.Close(ctx)
		default:
			__antithesis_instrumentation__.Notify(491316)
			panic(errors.AssertionFailedf("unknown plan node type %T", n))
		}
	}
	__antithesis_instrumentation__.Notify(491310)

	closeNode(e.plan.Root.WrappedNode())
	for i := range e.plan.Subqueries {
		__antithesis_instrumentation__.Notify(491317)
		closeNode(e.plan.Subqueries[i].Root.(*explain.Node).WrappedNode())
	}
	__antithesis_instrumentation__.Notify(491311)
	for i := range e.plan.Checks {
		__antithesis_instrumentation__.Notify(491318)
		closeNode(e.plan.Checks[i].WrappedNode())
	}
	__antithesis_instrumentation__.Notify(491312)
	if e.run.results != nil {
		__antithesis_instrumentation__.Notify(491319)
		e.run.results.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(491320)
	}
}

func newPhysPlanForExplainPurposes(
	ctx context.Context, planCtx *PlanningCtx, distSQLPlanner *DistSQLPlanner, plan planMaybePhysical,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(491321)
	if plan.isPhysicalPlan() {
		__antithesis_instrumentation__.Notify(491323)
		return plan.physPlan.PhysicalPlan, nil
	} else {
		__antithesis_instrumentation__.Notify(491324)
	}
	__antithesis_instrumentation__.Notify(491322)
	physPlan, err := distSQLPlanner.createPhysPlanForPlanNode(ctx, planCtx, plan.planNode)

	planCtx.getCleanupFunc()()
	return physPlan, err
}
