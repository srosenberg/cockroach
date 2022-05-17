package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/inverted"
	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/constraint"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec/explain"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/span"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

type distSQLSpecExecFactory struct {
	planner *planner
	dsp     *DistSQLPlanner

	planCtx              *PlanningCtx
	singleTenant         bool
	planningMode         distSQLPlanningMode
	gatewaySQLInstanceID base.SQLInstanceID
}

var _ exec.Factory = &distSQLSpecExecFactory{}

type distSQLPlanningMode int

const (
	distSQLDefaultPlanning distSQLPlanningMode = iota

	distSQLLocalOnlyPlanning
)

func newDistSQLSpecExecFactory(p *planner, planningMode distSQLPlanningMode) exec.Factory {
	__antithesis_instrumentation__.Notify(468324)
	e := &distSQLSpecExecFactory{
		planner:              p,
		dsp:                  p.extendedEvalCtx.DistSQLPlanner,
		singleTenant:         p.execCfg.Codec.ForSystemTenant(),
		planningMode:         planningMode,
		gatewaySQLInstanceID: p.extendedEvalCtx.DistSQLPlanner.gatewaySQLInstanceID,
	}
	distribute := DistributionType(DistributionTypeNone)
	if e.planningMode != distSQLLocalOnlyPlanning {
		__antithesis_instrumentation__.Notify(468326)
		distribute = DistributionTypeSystemTenantOnly
	} else {
		__antithesis_instrumentation__.Notify(468327)
	}
	__antithesis_instrumentation__.Notify(468325)
	evalCtx := p.ExtendedEvalContext()
	e.planCtx = e.dsp.NewPlanningCtx(evalCtx.Context, evalCtx, e.planner,
		e.planner.txn, distribute)
	return e
}

func (e *distSQLSpecExecFactory) getPlanCtx(recommendation distRecommendation) *PlanningCtx {
	__antithesis_instrumentation__.Notify(468328)
	distribute := false
	if e.singleTenant && func() bool {
		__antithesis_instrumentation__.Notify(468330)
		return e.planningMode != distSQLLocalOnlyPlanning == true
	}() == true {
		__antithesis_instrumentation__.Notify(468331)
		distribute = shouldDistributeGivenRecAndMode(
			recommendation, e.planner.extendedEvalCtx.SessionData().DistSQLMode,
		)
	} else {
		__antithesis_instrumentation__.Notify(468332)
	}
	__antithesis_instrumentation__.Notify(468329)
	e.planCtx.isLocal = !distribute
	return e.planCtx
}

func (e *distSQLSpecExecFactory) ConstructValues(
	rows [][]tree.TypedExpr, cols colinfo.ResultColumns,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468333)
	if (len(cols) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(468338)
		return len(rows) == 1 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(468339)
		return len(rows) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(468340)
		planCtx := e.getPlanCtx(canDistribute)
		colTypes := getTypesFromResultColumns(cols)
		spec := e.dsp.createValuesSpec(planCtx, colTypes, len(rows), nil)
		physPlan, err := e.dsp.createValuesPlan(planCtx, spec, colTypes)
		if err != nil {
			__antithesis_instrumentation__.Notify(468342)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(468343)
		}
		__antithesis_instrumentation__.Notify(468341)
		physPlan.ResultColumns = cols
		return makePlanMaybePhysical(physPlan, nil), nil
	} else {
		__antithesis_instrumentation__.Notify(468344)
	}
	__antithesis_instrumentation__.Notify(468334)

	recommendation := shouldDistribute
	for _, exprs := range rows {
		__antithesis_instrumentation__.Notify(468345)
		recommendation = recommendation.compose(
			e.checkExprsAndMaybeMergeLastStage(exprs, nil),
		)
		if recommendation == cannotDistribute {
			__antithesis_instrumentation__.Notify(468346)
			break
		} else {
			__antithesis_instrumentation__.Notify(468347)
		}
	}
	__antithesis_instrumentation__.Notify(468335)

	var (
		physPlan         *PhysicalPlan
		err              error
		planNodesToClose []planNode
	)
	planCtx := e.getPlanCtx(recommendation)
	if mustWrapValuesNode(planCtx, true) {
		__antithesis_instrumentation__.Notify(468348)

		v := &valuesNode{
			columns:          cols,
			tuples:           rows,
			specifiedInQuery: true,
		}
		planNodesToClose = []planNode{v}

		ctx := planCtx.EvalContext().Context
		physPlan, err = e.dsp.wrapPlan(ctx, planCtx, v, e.planningMode != distSQLLocalOnlyPlanning)
	} else {
		__antithesis_instrumentation__.Notify(468349)

		colTypes := getTypesFromResultColumns(cols)
		var spec *execinfrapb.ValuesCoreSpec
		spec, err = e.dsp.createValuesSpecFromTuples(planCtx, rows, colTypes)
		if err != nil {
			__antithesis_instrumentation__.Notify(468351)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(468352)
		}
		__antithesis_instrumentation__.Notify(468350)
		physPlan, err = e.dsp.createValuesPlan(planCtx, spec, colTypes)
	}
	__antithesis_instrumentation__.Notify(468336)
	if err != nil {
		__antithesis_instrumentation__.Notify(468353)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468354)
	}
	__antithesis_instrumentation__.Notify(468337)
	physPlan.ResultColumns = cols
	return makePlanMaybePhysical(physPlan, planNodesToClose), nil
}

func (e *distSQLSpecExecFactory) ConstructScan(
	table cat.Table, index cat.Index, params exec.ScanParams, reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468355)
	if table.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(468368)
		return constructVirtualScan(
			e, e.planner, table, index, params, reqOrdering,
			func(d *delayedNode) (exec.Node, error) {
				__antithesis_instrumentation__.Notify(468369)
				planCtx := e.getPlanCtx(cannotDistribute)

				ctx := planCtx.EvalContext().Context
				physPlan, err := e.dsp.wrapPlan(ctx, planCtx, d, e.planningMode != distSQLLocalOnlyPlanning)
				if err != nil {
					__antithesis_instrumentation__.Notify(468371)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(468372)
				}
				__antithesis_instrumentation__.Notify(468370)
				physPlan.ResultColumns = d.columns
				return makePlanMaybePhysical(physPlan, []planNode{d}), nil
			},
		)
	} else {
		__antithesis_instrumentation__.Notify(468373)
	}
	__antithesis_instrumentation__.Notify(468356)

	recommendation := canDistribute
	if params.LocalityOptimized {
		__antithesis_instrumentation__.Notify(468374)
		recommendation = recommendation.compose(cannotDistribute)
	} else {
		__antithesis_instrumentation__.Notify(468375)
	}
	__antithesis_instrumentation__.Notify(468357)
	planCtx := e.getPlanCtx(recommendation)
	p := planCtx.NewPhysicalPlan()

	tabDesc := table.(*optTable).desc
	idx := index.(*optIndex).idx
	colCfg := makeScanColumnsConfig(table, params.NeededCols)

	var sb span.Builder
	sb.Init(e.planner.EvalContext(), e.planner.ExecCfg().Codec, tabDesc, idx)

	cols := make([]catalog.Column, 0, params.NeededCols.Len())
	allCols := tabDesc.AllColumns()
	for ord, ok := params.NeededCols.Next(0); ok; ord, ok = params.NeededCols.Next(ord + 1) {
		__antithesis_instrumentation__.Notify(468376)
		cols = append(cols, allCols[ord])
	}
	__antithesis_instrumentation__.Notify(468358)
	columnIDs := make([]descpb.ColumnID, len(cols))
	for i := range cols {
		__antithesis_instrumentation__.Notify(468377)
		columnIDs[i] = cols[i].GetID()
	}
	__antithesis_instrumentation__.Notify(468359)

	p.ResultColumns = colinfo.ResultColumnsFromColumns(tabDesc.GetID(), cols)

	if params.IndexConstraint != nil && func() bool {
		__antithesis_instrumentation__.Notify(468378)
		return params.IndexConstraint.IsContradiction() == true
	}() == true {
		__antithesis_instrumentation__.Notify(468379)

		return e.ConstructValues([][]tree.TypedExpr{}, p.ResultColumns)
	} else {
		__antithesis_instrumentation__.Notify(468380)
	}
	__antithesis_instrumentation__.Notify(468360)

	var spans roachpb.Spans
	var err error
	if params.InvertedConstraint != nil {
		__antithesis_instrumentation__.Notify(468381)
		spans, err = sb.SpansFromInvertedSpans(params.InvertedConstraint, params.IndexConstraint, nil)
	} else {
		__antithesis_instrumentation__.Notify(468382)
		splitter := span.MakeSplitter(tabDesc, idx, params.NeededCols)
		spans, err = sb.SpansFromConstraint(params.IndexConstraint, splitter)
	}
	__antithesis_instrumentation__.Notify(468361)
	if err != nil {
		__antithesis_instrumentation__.Notify(468383)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468384)
	}
	__antithesis_instrumentation__.Notify(468362)

	isFullTableOrIndexScan := len(spans) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(468385)
		return spans[0].EqualValue(
			tabDesc.IndexSpan(e.planner.ExecCfg().Codec, idx.GetID()),
		) == true
	}() == true
	if err = colCfg.assertValidReqOrdering(reqOrdering); err != nil {
		__antithesis_instrumentation__.Notify(468386)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468387)
	}
	__antithesis_instrumentation__.Notify(468363)

	if isFullTableOrIndexScan {
		__antithesis_instrumentation__.Notify(468388)
		recommendation = recommendation.compose(shouldDistribute)
	} else {
		__antithesis_instrumentation__.Notify(468389)
	}
	__antithesis_instrumentation__.Notify(468364)

	trSpec := physicalplan.NewTableReaderSpec()
	*trSpec = execinfrapb.TableReaderSpec{
		Reverse:                         params.Reverse,
		TableDescriptorModificationTime: tabDesc.GetModificationTime(),
	}
	if err := rowenc.InitIndexFetchSpec(&trSpec.FetchSpec, e.planner.ExecCfg().Codec, tabDesc, idx, columnIDs); err != nil {
		__antithesis_instrumentation__.Notify(468390)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468391)
	}
	__antithesis_instrumentation__.Notify(468365)
	if params.Locking != nil {
		__antithesis_instrumentation__.Notify(468392)
		trSpec.LockingStrength = descpb.ToScanLockingStrength(params.Locking.Strength)
		trSpec.LockingWaitPolicy = descpb.ToScanLockingWaitPolicy(params.Locking.WaitPolicy)
		if trSpec.LockingStrength != descpb.ScanLockingStrength_FOR_NONE {
			__antithesis_instrumentation__.Notify(468393)

			recommendation = cannotDistribute
		} else {
			__antithesis_instrumentation__.Notify(468394)
		}
	} else {
		__antithesis_instrumentation__.Notify(468395)
	}
	__antithesis_instrumentation__.Notify(468366)

	post := execinfrapb.PostProcessSpec{}
	if params.HardLimit != 0 {
		__antithesis_instrumentation__.Notify(468396)
		post.Limit = uint64(params.HardLimit)
	} else {
		__antithesis_instrumentation__.Notify(468397)
		if params.SoftLimit != 0 {
			__antithesis_instrumentation__.Notify(468398)
			trSpec.LimitHint = params.SoftLimit
		} else {
			__antithesis_instrumentation__.Notify(468399)
		}
	}
	__antithesis_instrumentation__.Notify(468367)

	err = e.dsp.planTableReaders(
		planCtx.EvalContext().Context,
		e.getPlanCtx(recommendation),
		p,
		&tableReaderPlanningInfo{
			spec:              trSpec,
			post:              post,
			desc:              tabDesc,
			spans:             spans,
			reverse:           params.Reverse,
			parallelize:       params.Parallelize,
			estimatedRowCount: uint64(params.EstimatedRowCount),
			reqOrdering:       ReqOrdering(reqOrdering),
		},
	)

	return makePlanMaybePhysical(p, nil), err
}

func (e *distSQLSpecExecFactory) checkExprsAndMaybeMergeLastStage(
	exprs tree.TypedExprs, physPlan *PhysicalPlan,
) distRecommendation {
	__antithesis_instrumentation__.Notify(468400)

	recommendation := shouldDistribute
	if physPlan != nil && func() bool {
		__antithesis_instrumentation__.Notify(468403)
		return !physPlan.IsLastStageDistributed() == true
	}() == true {
		__antithesis_instrumentation__.Notify(468404)
		recommendation = shouldNotDistribute
	} else {
		__antithesis_instrumentation__.Notify(468405)
	}
	__antithesis_instrumentation__.Notify(468401)
	for _, expr := range exprs {
		__antithesis_instrumentation__.Notify(468406)
		if err := checkExpr(expr); err != nil {
			__antithesis_instrumentation__.Notify(468407)
			recommendation = cannotDistribute
			if physPlan != nil {
				__antithesis_instrumentation__.Notify(468409)

				physPlan.EnsureSingleStreamOnGateway()
			} else {
				__antithesis_instrumentation__.Notify(468410)
			}
			__antithesis_instrumentation__.Notify(468408)
			break
		} else {
			__antithesis_instrumentation__.Notify(468411)
		}
	}
	__antithesis_instrumentation__.Notify(468402)
	return recommendation
}

func (e *distSQLSpecExecFactory) ConstructFilter(
	n exec.Node, filter tree.TypedExpr, reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468412)
	physPlan, plan := getPhysPlan(n)
	recommendation := e.checkExprsAndMaybeMergeLastStage([]tree.TypedExpr{filter}, physPlan)

	if err := physPlan.AddFilter(filter, e.getPlanCtx(recommendation), physPlan.PlanToStreamColMap); err != nil {
		__antithesis_instrumentation__.Notify(468414)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468415)
	}
	__antithesis_instrumentation__.Notify(468413)
	physPlan.SetMergeOrdering(e.dsp.convertOrdering(ReqOrdering(reqOrdering), physPlan.PlanToStreamColMap))
	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructInvertedFilter(
	n exec.Node,
	invFilter *inverted.SpanExpression,
	preFiltererExpr tree.TypedExpr,
	preFiltererType *types.T,
	invColumn exec.NodeColumnOrdinal,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468416)
	return nil, unimplemented.NewWithIssue(
		47473, "experimental opt-driven distsql planning: inverted filter")
}

func (e *distSQLSpecExecFactory) ConstructSimpleProject(
	n exec.Node, cols []exec.NodeColumnOrdinal, reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468417)
	physPlan, plan := getPhysPlan(n)
	projection := make([]uint32, len(cols))
	for i := range cols {
		__antithesis_instrumentation__.Notify(468419)
		projection[i] = uint32(cols[physPlan.PlanToStreamColMap[i]])
	}
	__antithesis_instrumentation__.Notify(468418)
	newColMap := identityMap(physPlan.PlanToStreamColMap, len(cols))
	physPlan.AddProjection(
		projection,
		e.dsp.convertOrdering(ReqOrdering(reqOrdering), newColMap),
	)
	physPlan.ResultColumns = getResultColumnsForSimpleProject(
		cols, nil, physPlan.GetResultTypes(), physPlan.ResultColumns,
	)
	physPlan.PlanToStreamColMap = newColMap
	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructSerializingProject(
	n exec.Node, cols []exec.NodeColumnOrdinal, colNames []string,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468420)
	physPlan, plan := getPhysPlan(n)
	physPlan.EnsureSingleStreamOnGateway()
	projection := make([]uint32, len(cols))
	for i := range cols {
		__antithesis_instrumentation__.Notify(468422)
		projection[i] = uint32(cols[physPlan.PlanToStreamColMap[i]])
	}
	__antithesis_instrumentation__.Notify(468421)
	physPlan.AddProjection(projection, execinfrapb.Ordering{})
	physPlan.ResultColumns = getResultColumnsForSimpleProject(cols, colNames, physPlan.GetResultTypes(), physPlan.ResultColumns)
	physPlan.PlanToStreamColMap = identityMap(physPlan.PlanToStreamColMap, len(cols))
	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructRender(
	n exec.Node,
	columns colinfo.ResultColumns,
	exprs tree.TypedExprs,
	reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468423)
	physPlan, plan := getPhysPlan(n)
	recommendation := e.checkExprsAndMaybeMergeLastStage(exprs, physPlan)

	newColMap := identityMap(physPlan.PlanToStreamColMap, len(exprs))
	if err := physPlan.AddRendering(
		exprs, e.getPlanCtx(recommendation), physPlan.PlanToStreamColMap, getTypesFromResultColumns(columns),
		e.dsp.convertOrdering(ReqOrdering(reqOrdering), newColMap),
	); err != nil {
		__antithesis_instrumentation__.Notify(468425)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468426)
	}
	__antithesis_instrumentation__.Notify(468424)
	physPlan.ResultColumns = columns
	physPlan.PlanToStreamColMap = newColMap
	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructApplyJoin(
	joinType descpb.JoinType,
	left exec.Node,
	rightColumns colinfo.ResultColumns,
	onCond tree.TypedExpr,
	planRightSideFn exec.ApplyJoinPlanRightSideFn,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468427)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: apply join")
}

func (e *distSQLSpecExecFactory) ConstructHashJoin(
	joinType descpb.JoinType,
	left, right exec.Node,
	leftEqCols, rightEqCols []exec.NodeColumnOrdinal,
	leftEqColsAreKey, rightEqColsAreKey bool,
	extraOnCond tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468428)
	return e.constructHashOrMergeJoin(
		joinType, left, right, extraOnCond, leftEqCols, rightEqCols,
		leftEqColsAreKey, rightEqColsAreKey,
		ReqOrdering{}, exec.OutputOrdering{},
	)
}

func (e *distSQLSpecExecFactory) ConstructMergeJoin(
	joinType descpb.JoinType,
	left, right exec.Node,
	onCond tree.TypedExpr,
	leftOrdering, rightOrdering colinfo.ColumnOrdering,
	reqOrdering exec.OutputOrdering,
	leftEqColsAreKey, rightEqColsAreKey bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468429)
	leftEqCols, rightEqCols, mergeJoinOrdering, err := getEqualityIndicesAndMergeJoinOrdering(leftOrdering, rightOrdering)
	if err != nil {
		__antithesis_instrumentation__.Notify(468431)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468432)
	}
	__antithesis_instrumentation__.Notify(468430)
	return e.constructHashOrMergeJoin(
		joinType, left, right, onCond, leftEqCols, rightEqCols,
		leftEqColsAreKey, rightEqColsAreKey, mergeJoinOrdering, reqOrdering,
	)
}

func populateAggFuncSpec(
	spec *execinfrapb.AggregatorSpec_Aggregation,
	funcName string,
	distinct bool,
	argCols []exec.NodeColumnOrdinal,
	constArgs []tree.Datum,
	filter exec.NodeColumnOrdinal,
	planCtx *PlanningCtx,
	physPlan *PhysicalPlan,
) (argumentsColumnTypes []*types.T, err error) {
	__antithesis_instrumentation__.Notify(468433)
	funcIdx, err := execinfrapb.GetAggregateFuncIdx(funcName)
	if err != nil {
		__antithesis_instrumentation__.Notify(468438)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468439)
	}
	__antithesis_instrumentation__.Notify(468434)
	spec.Func = execinfrapb.AggregatorSpec_Func(funcIdx)
	spec.Distinct = distinct
	spec.ColIdx = make([]uint32, len(argCols))
	for i, col := range argCols {
		__antithesis_instrumentation__.Notify(468440)
		spec.ColIdx[i] = uint32(col)
	}
	__antithesis_instrumentation__.Notify(468435)
	if filter != tree.NoColumnIdx {
		__antithesis_instrumentation__.Notify(468441)
		filterColIdx := uint32(physPlan.PlanToStreamColMap[filter])
		spec.FilterColIdx = &filterColIdx
	} else {
		__antithesis_instrumentation__.Notify(468442)
	}
	__antithesis_instrumentation__.Notify(468436)
	if len(constArgs) > 0 {
		__antithesis_instrumentation__.Notify(468443)
		spec.Arguments = make([]execinfrapb.Expression, len(constArgs))
		argumentsColumnTypes = make([]*types.T, len(constArgs))
		for k, argument := range constArgs {
			__antithesis_instrumentation__.Notify(468444)
			var err error
			spec.Arguments[k], err = physicalplan.MakeExpression(argument, planCtx, nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(468446)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(468447)
			}
			__antithesis_instrumentation__.Notify(468445)
			argumentsColumnTypes[k] = argument.ResolvedType()
		}
	} else {
		__antithesis_instrumentation__.Notify(468448)
	}
	__antithesis_instrumentation__.Notify(468437)
	return argumentsColumnTypes, nil
}

func (e *distSQLSpecExecFactory) constructAggregators(
	input exec.Node,
	groupCols []exec.NodeColumnOrdinal,
	groupColOrdering colinfo.ColumnOrdering,
	aggregations []exec.AggInfo,
	reqOrdering exec.OutputOrdering,
	isScalar bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468449)
	physPlan, plan := getPhysPlan(input)

	planCtx := e.getPlanCtx(shouldDistribute)
	aggregationSpecs := make([]execinfrapb.AggregatorSpec_Aggregation, len(groupCols)+len(aggregations))
	argumentsColumnTypes := make([][]*types.T, len(groupCols)+len(aggregations))
	var err error
	if len(groupCols) > 0 {
		__antithesis_instrumentation__.Notify(468453)
		argColsScratch := []exec.NodeColumnOrdinal{0}
		noFilter := exec.NodeColumnOrdinal(tree.NoColumnIdx)
		for i, col := range groupCols {
			__antithesis_instrumentation__.Notify(468454)
			spec := &aggregationSpecs[i]
			argColsScratch[0] = col
			_, err = populateAggFuncSpec(
				spec, builtins.AnyNotNull, false, argColsScratch,
				nil, noFilter, planCtx, physPlan,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(468455)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(468456)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(468457)
	}
	__antithesis_instrumentation__.Notify(468450)
	for j := range aggregations {
		__antithesis_instrumentation__.Notify(468458)
		i := len(groupCols) + j
		spec := &aggregationSpecs[i]
		agg := &aggregations[j]
		argumentsColumnTypes[i], err = populateAggFuncSpec(
			spec, agg.FuncName, agg.Distinct, agg.ArgCols,
			agg.ConstArgs, agg.Filter, planCtx, physPlan,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(468459)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(468460)
		}
	}
	__antithesis_instrumentation__.Notify(468451)
	if err := e.dsp.planAggregators(
		planCtx,
		physPlan,
		&aggregatorPlanningInfo{
			aggregations:         aggregationSpecs,
			argumentsColumnTypes: argumentsColumnTypes,
			isScalar:             isScalar,
			groupCols:            convertNodeOrdinalsToInts(groupCols),
			groupColOrdering:     groupColOrdering,
			inputMergeOrdering:   physPlan.MergeOrdering,
			reqOrdering:          ReqOrdering(reqOrdering),
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(468461)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468462)
	}
	__antithesis_instrumentation__.Notify(468452)
	physPlan.ResultColumns = getResultColumnsForGroupBy(physPlan.ResultColumns, groupCols, aggregations)
	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructGroupBy(
	input exec.Node,
	groupCols []exec.NodeColumnOrdinal,
	groupColOrdering colinfo.ColumnOrdering,
	aggregations []exec.AggInfo,
	reqOrdering exec.OutputOrdering,
	groupingOrderType exec.GroupingOrderType,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468463)
	return e.constructAggregators(
		input,
		groupCols,
		groupColOrdering,
		aggregations,
		reqOrdering,
		false,
	)
}

func (e *distSQLSpecExecFactory) ConstructScalarGroupBy(
	input exec.Node, aggregations []exec.AggInfo,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468464)
	return e.constructAggregators(
		input,
		nil,
		nil,
		aggregations,
		exec.OutputOrdering{},
		true,
	)
}

func (e *distSQLSpecExecFactory) ConstructDistinct(
	input exec.Node,
	distinctCols, orderedCols exec.NodeColumnOrdinalSet,
	reqOrdering exec.OutputOrdering,
	nullsAreDistinct bool,
	errorOnDup string,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468465)
	physPlan, plan := getPhysPlan(input)
	spec := e.dsp.createDistinctSpec(
		convertFastIntSetToUint32Slice(distinctCols),
		convertFastIntSetToUint32Slice(orderedCols),
		nullsAreDistinct,
		errorOnDup,
		e.dsp.convertOrdering(ReqOrdering(reqOrdering), physPlan.PlanToStreamColMap),
	)
	e.dsp.addDistinctProcessors(physPlan, spec)

	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructHashSetOp(
	typ tree.UnionType, all bool, left, right exec.Node,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468466)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: hash set op")
}

func (e *distSQLSpecExecFactory) ConstructStreamingSetOp(
	typ tree.UnionType,
	all bool,
	left, right exec.Node,
	streamingOrdering colinfo.ColumnOrdering,
	reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468467)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: streaming set op")
}

func (e *distSQLSpecExecFactory) ConstructUnionAll(
	left, right exec.Node, reqOrdering exec.OutputOrdering, hardLimit uint64,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468468)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: union all")
}

func (e *distSQLSpecExecFactory) ConstructSort(
	input exec.Node, ordering exec.OutputOrdering, alreadyOrderedPrefix int,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468469)
	physPlan, plan := getPhysPlan(input)
	e.dsp.addSorters(physPlan, colinfo.ColumnOrdering(ordering), alreadyOrderedPrefix, 0)

	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructOrdinality(
	input exec.Node, colName string,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468470)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: ordinality")
}

func (e *distSQLSpecExecFactory) ConstructIndexJoin(
	input exec.Node,
	table cat.Table,
	keyCols []exec.NodeColumnOrdinal,
	tableCols exec.TableColumnOrdinalSet,
	reqOrdering exec.OutputOrdering,
	limitHint int64,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468471)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: index join")
}

func (e *distSQLSpecExecFactory) ConstructLookupJoin(
	joinType descpb.JoinType,
	input exec.Node,
	table cat.Table,
	index cat.Index,
	eqCols []exec.NodeColumnOrdinal,
	eqColsAreKey bool,
	lookupExpr tree.TypedExpr,
	remoteLookupExpr tree.TypedExpr,
	lookupCols exec.TableColumnOrdinalSet,
	onCond tree.TypedExpr,
	isFirstJoinInPairedJoiner bool,
	isSecondJoinInPairedJoiner bool,
	reqOrdering exec.OutputOrdering,
	locking *tree.LockingItem,
	limitHint int64,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468472)

	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: lookup join")
}

func (e *distSQLSpecExecFactory) ConstructInvertedJoin(
	joinType descpb.JoinType,
	invertedExpr tree.TypedExpr,
	input exec.Node,
	table cat.Table,
	index cat.Index,
	prefixEqCols []exec.NodeColumnOrdinal,
	lookupCols exec.TableColumnOrdinalSet,
	onCond tree.TypedExpr,
	isFirstJoinInPairedJoiner bool,
	reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468473)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: inverted join")
}

func (e *distSQLSpecExecFactory) constructZigzagJoinSide(
	planCtx *PlanningCtx,
	table cat.Table,
	index cat.Index,
	wantedCols exec.TableColumnOrdinalSet,
	fixedVals []tree.TypedExpr,
	eqCols []exec.TableColumnOrdinal,
) (zigzagPlanningSide, error) {
	__antithesis_instrumentation__.Notify(468474)
	desc := table.(*optTable).desc
	colCfg := scanColumnsConfig{wantedColumns: make([]tree.ColumnID, 0, wantedCols.Len())}
	for c, ok := wantedCols.Next(0); ok; c, ok = wantedCols.Next(c + 1) {
		__antithesis_instrumentation__.Notify(468479)
		colCfg.wantedColumns = append(colCfg.wantedColumns, desc.PublicColumns()[c].GetID())
	}
	__antithesis_instrumentation__.Notify(468475)
	cols, err := initColsForScan(desc, colCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(468480)
		return zigzagPlanningSide{}, err
	} else {
		__antithesis_instrumentation__.Notify(468481)
	}
	__antithesis_instrumentation__.Notify(468476)
	typs := make([]*types.T, len(fixedVals))
	for i := range typs {
		__antithesis_instrumentation__.Notify(468482)
		typs[i] = fixedVals[i].ResolvedType()
	}
	__antithesis_instrumentation__.Notify(468477)
	valuesSpec, err := e.dsp.createValuesSpecFromTuples(planCtx, [][]tree.TypedExpr{fixedVals}, typs)
	if err != nil {
		__antithesis_instrumentation__.Notify(468483)
		return zigzagPlanningSide{}, err
	} else {
		__antithesis_instrumentation__.Notify(468484)
	}
	__antithesis_instrumentation__.Notify(468478)

	return zigzagPlanningSide{
		desc:        desc,
		index:       index.(*optIndex).idx,
		cols:        cols,
		eqCols:      convertTableOrdinalsToInts(eqCols),
		fixedValues: valuesSpec,
	}, nil
}

func (e *distSQLSpecExecFactory) ConstructZigzagJoin(
	leftTable cat.Table,
	leftIndex cat.Index,
	leftCols exec.TableColumnOrdinalSet,
	leftFixedVals []tree.TypedExpr,
	leftEqCols []exec.TableColumnOrdinal,
	rightTable cat.Table,
	rightIndex cat.Index,
	rightCols exec.TableColumnOrdinalSet,
	rightFixedVals []tree.TypedExpr,
	rightEqCols []exec.TableColumnOrdinal,
	onCond tree.TypedExpr,
	reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468485)

	planCtx := e.getPlanCtx(cannotDistribute)

	sides := make([]zigzagPlanningSide, 2)
	var err error
	sides[0], err = e.constructZigzagJoinSide(planCtx, leftTable, leftIndex, leftCols, leftFixedVals, leftEqCols)
	if err != nil {
		__antithesis_instrumentation__.Notify(468489)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468490)
	}
	__antithesis_instrumentation__.Notify(468486)
	sides[1], err = e.constructZigzagJoinSide(planCtx, rightTable, rightIndex, rightCols, rightFixedVals, rightEqCols)
	if err != nil {
		__antithesis_instrumentation__.Notify(468491)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468492)
	}
	__antithesis_instrumentation__.Notify(468487)

	leftResultColumns := colinfo.ResultColumnsFromColumns(sides[0].desc.GetID(), sides[0].cols)
	rightResultColumns := colinfo.ResultColumnsFromColumns(sides[1].desc.GetID(), sides[1].cols)
	resultColumns := make(colinfo.ResultColumns, 0, len(leftResultColumns)+len(rightResultColumns))
	resultColumns = append(resultColumns, leftResultColumns...)
	resultColumns = append(resultColumns, rightResultColumns...)
	p, err := e.dsp.planZigzagJoin(planCtx, zigzagPlanningInfo{
		sides:       sides,
		columns:     resultColumns,
		onCond:      onCond,
		reqOrdering: ReqOrdering(reqOrdering),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(468493)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468494)
	}
	__antithesis_instrumentation__.Notify(468488)
	p.ResultColumns = resultColumns
	return makePlanMaybePhysical(p, nil), nil
}

func (e *distSQLSpecExecFactory) ConstructLimit(
	input exec.Node, limitExpr, offsetExpr tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468495)
	physPlan, plan := getPhysPlan(input)

	recommendation := e.checkExprsAndMaybeMergeLastStage(nil, physPlan)
	count, offset, err := evalLimit(e.planner.EvalContext(), limitExpr, offsetExpr)
	if err != nil {
		__antithesis_instrumentation__.Notify(468498)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468499)
	}
	__antithesis_instrumentation__.Notify(468496)
	if err = physPlan.AddLimit(count, offset, e.getPlanCtx(recommendation)); err != nil {
		__antithesis_instrumentation__.Notify(468500)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468501)
	}
	__antithesis_instrumentation__.Notify(468497)

	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructTopK(
	input exec.Node, k int64, ordering exec.OutputOrdering, alreadyOrderedPrefix int,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468502)
	physPlan, plan := getPhysPlan(input)
	if k <= 0 {
		__antithesis_instrumentation__.Notify(468504)
		return nil, errors.New("negative or zero value for LIMIT")
	} else {
		__antithesis_instrumentation__.Notify(468505)
	}
	__antithesis_instrumentation__.Notify(468503)

	e.dsp.addSorters(physPlan, colinfo.ColumnOrdering(ordering), alreadyOrderedPrefix, k)

	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructMax1Row(
	input exec.Node, errorText string,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468506)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: max1row")
}

func (e *distSQLSpecExecFactory) ConstructProjectSet(
	n exec.Node, exprs tree.TypedExprs, zipCols colinfo.ResultColumns, numColsPerGen []int,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468507)
	physPlan, plan := getPhysPlan(n)
	cols := append(plan.physPlan.ResultColumns, zipCols...)
	err := e.dsp.addProjectSet(
		physPlan,

		e.getPlanCtx(cannotDistribute),
		&projectSetPlanningInfo{
			columns:         cols,
			numColsInSource: len(plan.physPlan.ResultColumns),
			exprs:           exprs,
			numColsPerGen:   numColsPerGen,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(468509)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468510)
	}
	__antithesis_instrumentation__.Notify(468508)
	physPlan.ResultColumns = cols
	return plan, nil
}

func (e *distSQLSpecExecFactory) ConstructWindow(
	input exec.Node, window exec.WindowInfo,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468511)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: window")
}

func (e *distSQLSpecExecFactory) ConstructPlan(
	root exec.Node,
	subqueries []exec.Subquery,
	cascades []exec.Cascade,
	checks []exec.Node,
	rootRowCount int64,
) (exec.Plan, error) {
	__antithesis_instrumentation__.Notify(468512)
	if len(subqueries) != 0 {
		__antithesis_instrumentation__.Notify(468516)
		return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: subqueries")
	} else {
		__antithesis_instrumentation__.Notify(468517)
	}
	__antithesis_instrumentation__.Notify(468513)
	if len(cascades) != 0 {
		__antithesis_instrumentation__.Notify(468518)
		return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: cascades")
	} else {
		__antithesis_instrumentation__.Notify(468519)
	}
	__antithesis_instrumentation__.Notify(468514)
	if len(checks) != 0 {
		__antithesis_instrumentation__.Notify(468520)
		return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: checks")
	} else {
		__antithesis_instrumentation__.Notify(468521)
	}
	__antithesis_instrumentation__.Notify(468515)
	return constructPlan(e.planner, root, subqueries, cascades, checks, rootRowCount)
}

func (e *distSQLSpecExecFactory) ConstructExplainOpt(
	plan string, envOpts exec.ExplainEnvData,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468522)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: explain opt")
}

func (e *distSQLSpecExecFactory) ConstructExplain(
	options *tree.ExplainOptions,
	stmtType tree.StatementReturnType,
	buildFn exec.BuildPlanForExplainFn,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468523)
	if options.Flags[tree.ExplainFlagEnv] {
		__antithesis_instrumentation__.Notify(468528)
		return nil, errors.New("ENV only supported with (OPT) option")
	} else {
		__antithesis_instrumentation__.Notify(468529)
	}
	__antithesis_instrumentation__.Notify(468524)

	newFactory := newDistSQLSpecExecFactory(e.planner, e.planningMode)
	plan, err := buildFn(newFactory)

	newFactory.(*distSQLSpecExecFactory).planCtx.getCleanupFunc()()
	if err != nil {
		__antithesis_instrumentation__.Notify(468530)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468531)
	}
	__antithesis_instrumentation__.Notify(468525)

	p := plan.(*explain.Plan).WrappedPlan.(*planComponents)
	var explainNode planNode
	if options.Mode == tree.ExplainVec {
		__antithesis_instrumentation__.Notify(468532)
		explainNode = &explainVecNode{
			options: options,
			plan:    *p,
		}
	} else {
		__antithesis_instrumentation__.Notify(468533)
		if options.Mode == tree.ExplainDDL {
			__antithesis_instrumentation__.Notify(468534)
			explainNode = &explainDDLNode{
				options: options,
				plan:    *p,
			}
		} else {
			__antithesis_instrumentation__.Notify(468535)
			explainNode = &explainPlanNode{
				options: options,
				flags:   explain.MakeFlags(options),
				plan:    plan.(*explain.Plan),
			}
		}
	}
	__antithesis_instrumentation__.Notify(468526)

	planCtx := e.getPlanCtx(cannotDistribute)
	physPlan, err := e.dsp.wrapPlan(planCtx.EvalContext().Context, planCtx, explainNode, e.planningMode != distSQLLocalOnlyPlanning)
	if err != nil {
		__antithesis_instrumentation__.Notify(468536)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468537)
	}
	__antithesis_instrumentation__.Notify(468527)
	physPlan.ResultColumns = planColumns(explainNode)

	physPlan.Distribution = p.main.physPlan.Distribution
	return makePlanMaybePhysical(physPlan, []planNode{explainNode}), nil
}

func (e *distSQLSpecExecFactory) ConstructShowTrace(
	typ tree.ShowTraceType, compact bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468538)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: show trace")
}

func (e *distSQLSpecExecFactory) ConstructInsert(
	input exec.Node,
	table cat.Table,
	arbiterIndexes cat.IndexOrdinals,
	arbiterConstraints cat.UniqueOrdinals,
	insertCols exec.TableColumnOrdinalSet,
	returnCols exec.TableColumnOrdinalSet,
	checkCols exec.CheckOrdinalSet,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468539)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: insert")
}

func (e *distSQLSpecExecFactory) ConstructInsertFastPath(
	rows [][]tree.TypedExpr,
	table cat.Table,
	insertCols exec.TableColumnOrdinalSet,
	returnCols exec.TableColumnOrdinalSet,
	checkCols exec.CheckOrdinalSet,
	fkChecks []exec.InsertFastPathFKCheck,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468540)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: insert fast path")
}

func (e *distSQLSpecExecFactory) ConstructUpdate(
	input exec.Node,
	table cat.Table,
	fetchCols exec.TableColumnOrdinalSet,
	updateCols exec.TableColumnOrdinalSet,
	returnCols exec.TableColumnOrdinalSet,
	checks exec.CheckOrdinalSet,
	passthrough colinfo.ResultColumns,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468541)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: update")
}

func (e *distSQLSpecExecFactory) ConstructUpsert(
	input exec.Node,
	table cat.Table,
	arbiterIndexes cat.IndexOrdinals,
	arbiterConstraints cat.UniqueOrdinals,
	canaryCol exec.NodeColumnOrdinal,
	insertCols exec.TableColumnOrdinalSet,
	fetchCols exec.TableColumnOrdinalSet,
	updateCols exec.TableColumnOrdinalSet,
	returnCols exec.TableColumnOrdinalSet,
	checks exec.CheckOrdinalSet,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468542)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: upsert")
}

func (e *distSQLSpecExecFactory) ConstructDelete(
	input exec.Node,
	table cat.Table,
	fetchCols exec.TableColumnOrdinalSet,
	returnCols exec.TableColumnOrdinalSet,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468543)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: delete")
}

func (e *distSQLSpecExecFactory) ConstructDeleteRange(
	table cat.Table,
	needed exec.TableColumnOrdinalSet,
	indexConstraint *constraint.Constraint,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468544)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: delete range")
}

func (e *distSQLSpecExecFactory) ConstructCreateTable(
	schema cat.Schema, ct *tree.CreateTable,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468545)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: create table")
}

func (e *distSQLSpecExecFactory) ConstructCreateTableAs(
	input exec.Node, schema cat.Schema, ct *tree.CreateTable,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468546)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: create table")
}

func (e *distSQLSpecExecFactory) ConstructCreateView(
	schema cat.Schema,
	viewName *cat.DataSourceName,
	ifNotExists bool,
	replace bool,
	persistence tree.Persistence,
	materialized bool,
	viewQuery string,
	columns colinfo.ResultColumns,
	deps opt.ViewDeps,
	typeDeps opt.ViewTypeDeps,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468547)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: create view")
}

func (e *distSQLSpecExecFactory) ConstructSequenceSelect(sequence cat.Sequence) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468548)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: sequence select")
}

func (e *distSQLSpecExecFactory) ConstructSaveTable(
	input exec.Node, table *cat.DataSourceName, colNames []string,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468549)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: save table")
}

func (e *distSQLSpecExecFactory) ConstructErrorIfRows(
	input exec.Node, mkErr exec.MkErrFn,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468550)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: error if rows")
}

func (e *distSQLSpecExecFactory) ConstructOpaque(metadata opt.OpaqueMetadata) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468551)
	plan, err := constructOpaque(metadata)
	if err != nil {
		__antithesis_instrumentation__.Notify(468554)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468555)
	}
	__antithesis_instrumentation__.Notify(468552)
	planCtx := e.getPlanCtx(cannotDistribute)
	physPlan, err := e.dsp.wrapPlan(planCtx.EvalContext().Context, planCtx, plan, e.planningMode != distSQLLocalOnlyPlanning)
	if err != nil {
		__antithesis_instrumentation__.Notify(468556)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468557)
	}
	__antithesis_instrumentation__.Notify(468553)
	physPlan.ResultColumns = planColumns(plan)
	return makePlanMaybePhysical(physPlan, []planNode{plan}), nil
}

func (e *distSQLSpecExecFactory) ConstructAlterTableSplit(
	index cat.Index, input exec.Node, expiration tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468558)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: alter table split")
}

func (e *distSQLSpecExecFactory) ConstructAlterTableUnsplit(
	index cat.Index, input exec.Node,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468559)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: alter table unsplit")
}

func (e *distSQLSpecExecFactory) ConstructAlterTableUnsplitAll(index cat.Index) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468560)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: alter table unsplit all")
}

func (e *distSQLSpecExecFactory) ConstructAlterTableRelocate(
	index cat.Index, input exec.Node, relocateSubject tree.RelocateSubject,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468561)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: alter table relocate")
}

func (e *distSQLSpecExecFactory) ConstructAlterRangeRelocate(
	input exec.Node,
	relocateSubject tree.RelocateSubject,
	toStoreID tree.TypedExpr,
	fromStoreID tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468562)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: alter range relocate")
}

func (e *distSQLSpecExecFactory) ConstructBuffer(input exec.Node, label string) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468563)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: buffer")
}

func (e *distSQLSpecExecFactory) ConstructScanBuffer(
	ref exec.Node, label string,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468564)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: scan buffer")
}

func (e *distSQLSpecExecFactory) ConstructRecursiveCTE(
	initial exec.Node, fn exec.RecursiveCTEIterationFn, label string, deduplicate bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468565)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: recursive CTE")
}

func (e *distSQLSpecExecFactory) ConstructControlJobs(
	command tree.JobCommand, input exec.Node, reason tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468566)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: control jobs")
}

func (e *distSQLSpecExecFactory) ConstructControlSchedules(
	command tree.ScheduleCommand, input exec.Node,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468567)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: control jobs")
}

func (e *distSQLSpecExecFactory) ConstructCancelQueries(
	input exec.Node, ifExists bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468568)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: cancel queries")
}

func (e *distSQLSpecExecFactory) ConstructCancelSessions(
	input exec.Node, ifExists bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468569)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: cancel sessions")
}

func (e *distSQLSpecExecFactory) ConstructCreateStatistics(
	cs *tree.CreateStats,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468570)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: create statistics")
}

func (e *distSQLSpecExecFactory) ConstructExport(
	input exec.Node,
	fileName tree.TypedExpr,
	fileFormat string,
	options []exec.KVOption,
	notNullColsSet exec.NodeColumnOrdinalSet,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468571)
	return nil, unimplemented.NewWithIssue(47473, "experimental opt-driven distsql planning: export")
}

func getPhysPlan(n exec.Node) (*PhysicalPlan, planMaybePhysical) {
	__antithesis_instrumentation__.Notify(468572)
	plan := n.(planMaybePhysical)
	return plan.physPlan.PhysicalPlan, plan
}

func (e *distSQLSpecExecFactory) constructHashOrMergeJoin(
	joinType descpb.JoinType,
	left, right exec.Node,
	onCond tree.TypedExpr,
	leftEqCols, rightEqCols []exec.NodeColumnOrdinal,
	leftEqColsAreKey, rightEqColsAreKey bool,
	mergeJoinOrdering colinfo.ColumnOrdering,
	reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(468573)
	leftPhysPlan, leftPlan := getPhysPlan(left)
	rightPhysPlan, rightPlan := getPhysPlan(right)
	resultColumns := getJoinResultColumns(joinType, leftPhysPlan.ResultColumns, rightPhysPlan.ResultColumns)
	leftMap, rightMap := leftPhysPlan.PlanToStreamColMap, rightPhysPlan.PlanToStreamColMap
	helper := &joinPlanningHelper{
		numLeftOutCols:          len(leftPhysPlan.GetResultTypes()),
		numRightOutCols:         len(rightPhysPlan.GetResultTypes()),
		numAllLeftCols:          len(leftPhysPlan.GetResultTypes()),
		leftPlanToStreamColMap:  leftMap,
		rightPlanToStreamColMap: rightMap,
	}
	post, joinToStreamColMap := helper.joinOutColumns(joinType, resultColumns)

	planCtx := e.getPlanCtx(shouldDistribute)
	onExpr, err := helper.remapOnExpr(planCtx, onCond)
	if err != nil {
		__antithesis_instrumentation__.Notify(468575)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468576)
	}
	__antithesis_instrumentation__.Notify(468574)

	leftEqColsRemapped := eqCols(leftEqCols, leftMap)
	rightEqColsRemapped := eqCols(rightEqCols, rightMap)
	info := joinPlanningInfo{
		leftPlan:                 leftPhysPlan,
		rightPlan:                rightPhysPlan,
		joinType:                 joinType,
		joinResultTypes:          getTypesFromResultColumns(resultColumns),
		onExpr:                   onExpr,
		post:                     post,
		joinToStreamColMap:       joinToStreamColMap,
		leftEqCols:               leftEqColsRemapped,
		rightEqCols:              rightEqColsRemapped,
		leftEqColsAreKey:         leftEqColsAreKey,
		rightEqColsAreKey:        rightEqColsAreKey,
		leftMergeOrd:             distsqlOrdering(mergeJoinOrdering, leftEqColsRemapped),
		rightMergeOrd:            distsqlOrdering(mergeJoinOrdering, rightEqColsRemapped),
		leftPlanDistribution:     leftPhysPlan.Distribution,
		rightPlanDistribution:    rightPhysPlan.Distribution,
		allowPartialDistribution: e.planningMode != distSQLLocalOnlyPlanning,
	}
	p := e.dsp.planJoiners(planCtx, &info, ReqOrdering(reqOrdering))
	p.ResultColumns = resultColumns
	return makePlanMaybePhysical(p, append(leftPlan.physPlan.planNodesToClose, rightPlan.physPlan.planNodesToClose...)), nil
}
